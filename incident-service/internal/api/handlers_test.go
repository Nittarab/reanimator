package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/ai-sre-platform/incident-service/internal/config"
	"github.com/your-org/ai-sre-platform/incident-service/internal/database"
	"github.com/your-org/ai-sre-platform/incident-service/internal/github"
	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// TestHandleWorkflowStatus_Success tests the workflow status webhook handler
func TestHandleWorkflowStatus_Success(t *testing.T) {
	// Skip if no test database is configured
	db, err := database.Connect("postgres://localhost/ai_sre_test?sslmode=disable")
	if err != nil {
		t.Skipf("test database not configured: %v", err)
	}
	defer db.Close()

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		GitHub: config.GitHubConfig{
			APIURL:       "https://api.github.com",
			Token:        "test-token",
			WorkflowName: "test-workflow.yml",
		},
		Concurrency: config.ConcurrencyConfig{
			MaxWorkflowsPerRepo: 2,
		},
	}

	// Create GitHub client
	githubClient := github.NewClient(
		cfg.GitHub.APIURL,
		cfg.GitHub.Token,
		cfg.GitHub.WorkflowName,
		cfg.Concurrency.MaxWorkflowsPerRepo,
	)

	// Create server (without Redis for this test)
	redis, _ := database.ConnectRedis("localhost:6379", "", 0)
	server := NewServer(cfg, db, redis, githubClient)
	repository := database.NewIncidentRepository(db)

	// Create a test incident
	incident := &models.Incident{
		ID:           "test-incident-123",
		ServiceName:  "test-service",
		Repository:   "test-org/test-repo",
		ErrorMessage: "test error",
		Status:       models.StatusInProgress,
		Provider:     "test",
		ProviderData: map[string]interface{}{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repository.Create(incident); err != nil {
		t.Fatalf("failed to create test incident: %v", err)
	}

	// Clean up after test
	defer func() {
		db.Exec("DELETE FROM incident_events WHERE incident_id = $1", incident.ID)
		db.Exec("DELETE FROM incidents WHERE id = $1", incident.ID)
	}()

	// Create workflow status payload
	payload := WorkflowStatusPayload{
		IncidentID:     incident.ID,
		Status:         "success",
		PullRequestURL: "https://github.com/test-org/test-repo/pull/123",
		Diagnosis:      "Fixed the bug by adding null check",
		Repository:     incident.Repository,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/webhooks/workflow-status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call the handler
	server.handleWorkflowStatus(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify incident was updated
	updatedIncident, err := repository.GetByID(incident.ID)
	if err != nil {
		t.Fatalf("failed to get updated incident: %v", err)
	}

	if updatedIncident.Status != models.StatusPRCreated {
		t.Errorf("expected status pr_created, got %s", updatedIncident.Status)
	}

	if updatedIncident.PullRequestURL == nil || *updatedIncident.PullRequestURL != payload.PullRequestURL {
		t.Errorf("expected PR URL %s, got %v", payload.PullRequestURL, updatedIncident.PullRequestURL)
	}

	if updatedIncident.Diagnosis == nil || *updatedIncident.Diagnosis != payload.Diagnosis {
		t.Errorf("expected diagnosis %s, got %v", payload.Diagnosis, updatedIncident.Diagnosis)
	}

	if updatedIncident.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

// TestHandleWorkflowStatus_Failed tests the workflow status webhook handler with failed status
func TestHandleWorkflowStatus_Failed(t *testing.T) {
	// Skip if no test database is configured
	db, err := database.Connect("postgres://localhost/ai_sre_test?sslmode=disable")
	if err != nil {
		t.Skipf("test database not configured: %v", err)
	}
	defer db.Close()

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		GitHub: config.GitHubConfig{
			APIURL:       "https://api.github.com",
			Token:        "test-token",
			WorkflowName: "test-workflow.yml",
		},
		Concurrency: config.ConcurrencyConfig{
			MaxWorkflowsPerRepo: 2,
		},
	}

	// Create GitHub client
	githubClient := github.NewClient(
		cfg.GitHub.APIURL,
		cfg.GitHub.Token,
		cfg.GitHub.WorkflowName,
		cfg.Concurrency.MaxWorkflowsPerRepo,
	)

	// Create server (without Redis for this test)
	redis, _ := database.ConnectRedis("localhost:6379", "", 0)
	server := NewServer(cfg, db, redis, githubClient)
	repository := database.NewIncidentRepository(db)

	// Create a test incident
	incident := &models.Incident{
		ID:           "test-incident-456",
		ServiceName:  "test-service",
		Repository:   "test-org/test-repo",
		ErrorMessage: "test error",
		Status:       models.StatusInProgress,
		Provider:     "test",
		ProviderData: map[string]interface{}{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repository.Create(incident); err != nil {
		t.Fatalf("failed to create test incident: %v", err)
	}

	// Clean up after test
	defer func() {
		db.Exec("DELETE FROM incident_events WHERE incident_id = $1", incident.ID)
		db.Exec("DELETE FROM incidents WHERE id = $1", incident.ID)
	}()

	// Create workflow status payload with failed status
	payload := WorkflowStatusPayload{
		IncidentID: incident.ID,
		Status:     "failed",
		Diagnosis:  "Could not determine root cause",
		Repository: incident.Repository,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/webhooks/workflow-status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call the handler
	server.handleWorkflowStatus(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify incident was updated
	updatedIncident, err := repository.GetByID(incident.ID)
	if err != nil {
		t.Fatalf("failed to get updated incident: %v", err)
	}

	if updatedIncident.Status != models.StatusFailed {
		t.Errorf("expected status failed, got %s", updatedIncident.Status)
	}

	if updatedIncident.PullRequestURL != nil {
		t.Errorf("expected no PR URL for failed status, got %v", updatedIncident.PullRequestURL)
	}

	if updatedIncident.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

// TestHandleWorkflowStatus_InvalidPayload tests the workflow status webhook handler with invalid payload
func TestHandleWorkflowStatus_InvalidPayload(t *testing.T) {
	// Create minimal server for this test
	cfg := &config.Config{}
	githubClient := github.NewClient("https://api.github.com", "test-token", "test-workflow.yml", 2)
	
	// Use nil for db and redis since we won't reach that code
	server := &Server{
		config:       cfg,
		githubClient: githubClient,
		logger:       NewLogger(),
		metrics:      NewMetrics(),
	}

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
	}{
		{
			name:           "invalid JSON",
			payload:        "{invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing incident_id",
			payload:        `{"status": "success", "repository": "test-org/test-repo"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing status",
			payload:        `{"incident_id": "test-123", "repository": "test-org/test-repo"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing repository",
			payload:        `{"incident_id": "test-123", "status": "success"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/webhooks/workflow-status", bytes.NewReader([]byte(tt.payload)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleWorkflowStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
