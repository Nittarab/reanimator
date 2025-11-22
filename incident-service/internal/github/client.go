package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// Client handles GitHub API interactions
type Client struct {
	apiURL     string
	token      string
	httpClient *http.Client
	workflow   string

	// Concurrency tracking
	mu                  sync.RWMutex
	activeWorkflows     map[string]int // repository -> active count
	queuedIncidents     map[string][]*models.Incident // repository -> queued incidents
	maxWorkflowsPerRepo int
}

// WorkflowDispatchInput represents the inputs for a workflow dispatch
type WorkflowDispatchInput struct {
	IncidentID   string `json:"incident_id"`
	ErrorMessage string `json:"error_message"`
	StackTrace   string `json:"stack_trace"`
	ServiceName  string `json:"service_name"`
	Timestamp    string `json:"timestamp"`
	MCPConfig    string `json:"mcp_config,omitempty"`
}

// WorkflowDispatchRequest represents the GitHub workflow dispatch API request
type WorkflowDispatchRequest struct {
	Ref    string                 `json:"ref"`
	Inputs WorkflowDispatchInput  `json:"inputs"`
}

// NewClient creates a new GitHub API client
func NewClient(apiURL, token, workflow string, maxWorkflowsPerRepo int) *Client {
	return &Client{
		apiURL:              apiURL,
		token:               token,
		workflow:            workflow,
		httpClient:          &http.Client{Timeout: 30 * time.Second},
		activeWorkflows:     make(map[string]int),
		queuedIncidents:     make(map[string][]*models.Incident),
		maxWorkflowsPerRepo: maxWorkflowsPerRepo,
	}
}

// DispatchWorkflow triggers a GitHub Actions workflow for an incident
// Returns workflow run ID if successful, error otherwise
func (c *Client) DispatchWorkflow(ctx context.Context, incident *models.Incident, branch string) (int64, error) {
	// Check concurrency limit
	if !c.canDispatch(incident.Repository) {
		c.queueIncident(incident)
		return 0, fmt.Errorf("concurrency limit reached, incident queued")
	}

	// Prepare workflow inputs
	inputs := WorkflowDispatchInput{
		IncidentID:   incident.ID,
		ErrorMessage: incident.ErrorMessage,
		ServiceName:  incident.ServiceName,
		Timestamp:    incident.CreatedAt.Format(time.RFC3339),
	}

	if incident.StackTrace != nil {
		inputs.StackTrace = *incident.StackTrace
	}

	request := WorkflowDispatchRequest{
		Ref:    branch,
		Inputs: inputs,
	}

	// Retry logic with exponential backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := c.dispatchWorkflowAttempt(ctx, incident.Repository, request)
		if err == nil {
			// Success - increment active workflow count
			c.incrementActive(incident.Repository)
			// We don't have the run ID from the dispatch API, return 0
			return 0, nil
		}

		lastErr = err
	}

	return 0, fmt.Errorf("workflow dispatch failed after 3 attempts: %w", lastErr)
}

// dispatchWorkflowAttempt makes a single attempt to dispatch a workflow
func (c *Client) dispatchWorkflowAttempt(ctx context.Context, repository string, request WorkflowDispatchRequest) error {
	// Build API URL: /repos/{owner}/{repo}/actions/workflows/{workflow_id}/dispatches
	url := fmt.Sprintf("%s/repos/%s/actions/workflows/%s/dispatches", c.apiURL, repository, c.workflow)

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// canDispatch checks if a workflow can be dispatched for the given repository
func (c *Client) canDispatch(repository string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	active := c.activeWorkflows[repository]
	return active < c.maxWorkflowsPerRepo
}

// queueIncident adds an incident to the queue for a repository
func (c *Client) queueIncident(incident *models.Incident) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queuedIncidents[incident.Repository] = append(c.queuedIncidents[incident.Repository], incident)
}

// incrementActive increments the active workflow count for a repository
func (c *Client) incrementActive(repository string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.activeWorkflows[repository]++
}

// DecrementActive decrements the active workflow count and returns the next queued incident if any
func (c *Client) DecrementActive(repository string) *models.Incident {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.activeWorkflows[repository] > 0 {
		c.activeWorkflows[repository]--
	}

	// Check if there are queued incidents
	queue := c.queuedIncidents[repository]
	if len(queue) == 0 {
		return nil
	}

	// Pop the first incident from the queue
	incident := queue[0]
	c.queuedIncidents[repository] = queue[1:]

	return incident
}

// GetActiveCount returns the number of active workflows for a repository
func (c *Client) GetActiveCount(repository string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.activeWorkflows[repository]
}

// GetQueuedCount returns the number of queued incidents for a repository
func (c *Client) GetQueuedCount(repository string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.queuedIncidents[repository])
}
