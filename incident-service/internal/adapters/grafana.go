package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// GrafanaAdapter handles Grafana webhook payloads
type GrafanaAdapter struct {
	secret string
}

// NewGrafanaAdapter creates a new Grafana adapter
func NewGrafanaAdapter() *GrafanaAdapter {
	return &GrafanaAdapter{
		secret: os.Getenv("GRAFANA_WEBHOOK_SECRET"),
	}
}

// ProviderName returns the provider name
func (a *GrafanaAdapter) ProviderName() string {
	return "grafana"
}

// Validate validates the webhook (optional secret)
func (a *GrafanaAdapter) Validate(r *http.Request) error {
	if a.secret == "" {
		// If no secret is configured, skip validation
		return nil
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("missing Authorization header")
	}

	// Grafana sends "Bearer <secret>"
	expectedAuth := "Bearer " + a.secret
	if authHeader != expectedAuth {
		return fmt.Errorf("invalid authorization")
	}

	return nil
}

// Parse transforms Grafana payload to internal Incident
func (a *GrafanaAdapter) Parse(body []byte) (*models.Incident, error) {
	var payload GrafanaPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse grafana payload: %w", err)
	}

	// Only process firing alerts
	if payload.State != "alerting" && payload.State != "firing" {
		return nil, fmt.Errorf("unsupported alert state: %s", payload.State)
	}

	// Extract service name from labels
	serviceName := extractServiceFromLabels(payload.Labels)
	if serviceName == "" {
		serviceName = payload.RuleName
	}

	// Map state to severity
	severity := mapGrafanaSeverity(payload.State, payload.Labels)

	// Construct error message from title and message
	errorMessage := payload.Title
	if payload.Message != "" {
		errorMessage = fmt.Sprintf("%s: %s", payload.Title, payload.Message)
	}

	// Extract stack trace from annotations or query results
	var stackTrace *string
	if stackTraceStr := extractStackTraceFromGrafana(payload); stackTraceStr != "" {
		stackTrace = &stackTraceStr
	}

	// Create incident ID
	incidentID := fmt.Sprintf("inc_grafana_%s_%d", payload.RuleID, time.Now().Unix())

	// Store provider data
	providerData := map[string]interface{}{
		"rule_id":   payload.RuleID,
		"rule_name": payload.RuleName,
		"state":     payload.State,
		"labels":    payload.Labels,
	}
	if payload.RuleURL != "" {
		providerData["rule_url"] = payload.RuleURL
	}

	incident := &models.Incident{
		ID:           incidentID,
		ServiceName:  serviceName,
		Repository:   "", // Will be mapped later
		ErrorMessage: errorMessage,
		StackTrace:   stackTrace,
		Severity:     severity,
		Status:       models.StatusPending,
		Provider:     "grafana",
		ProviderData: providerData,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return incident, nil
}

// GrafanaPayload represents a Grafana webhook payload
type GrafanaPayload struct {
	Title       string            `json:"title"`
	State       string            `json:"state"`
	Message     string            `json:"message"`
	RuleID      string            `json:"ruleId"`
	RuleName    string            `json:"ruleName"`
	RuleURL     string            `json:"ruleUrl"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// extractServiceFromLabels extracts service name from Grafana labels
func extractServiceFromLabels(labels map[string]string) string {
	// Try common label names
	if service, ok := labels["service"]; ok {
		return service
	}
	if service, ok := labels["app"]; ok {
		return service
	}
	if service, ok := labels["application"]; ok {
		return service
	}
	return ""
}

// mapGrafanaSeverity maps Grafana alert state to internal severity
func mapGrafanaSeverity(state string, labels map[string]string) string {
	// Check if severity is explicitly set in labels
	if severity, ok := labels["severity"]; ok {
		switch strings.ToLower(severity) {
		case "critical", "high", "medium", "low":
			return strings.ToLower(severity)
		}
	}

	// Default mapping based on state
	switch strings.ToLower(state) {
	case "alerting", "firing":
		return "high"
	default:
		return "medium"
	}
}

// extractStackTraceFromGrafana attempts to extract stack trace from annotations
func extractStackTraceFromGrafana(payload GrafanaPayload) string {
	// Check annotations for stack trace
	if stackTrace, ok := payload.Annotations["stack_trace"]; ok {
		return stackTrace
	}
	if stackTrace, ok := payload.Annotations["error"]; ok {
		if strings.Contains(stackTrace, "at ") || strings.Contains(stackTrace, "Traceback") {
			return stackTrace
		}
	}
	
	// Check message for stack trace patterns
	if strings.Contains(payload.Message, "at ") || strings.Contains(payload.Message, "Traceback") {
		return payload.Message
	}

	return ""
}
