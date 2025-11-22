package adapters

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// PagerDutyAdapter handles PagerDuty webhook payloads
type PagerDutyAdapter struct {
	secret string
}

// NewPagerDutyAdapter creates a new PagerDuty adapter
func NewPagerDutyAdapter() *PagerDutyAdapter {
	return &PagerDutyAdapter{
		secret: os.Getenv("PAGERDUTY_WEBHOOK_SECRET"),
	}
}

// ProviderName returns the provider name
func (a *PagerDutyAdapter) ProviderName() string {
	return "pagerduty"
}

// Validate validates the webhook signature
func (a *PagerDutyAdapter) Validate(r *http.Request) error {
	if a.secret == "" {
		// If no secret is configured, skip validation
		return nil
	}

	signature := r.Header.Get("X-PagerDuty-Signature")
	if signature == "" {
		return fmt.Errorf("missing X-PagerDuty-Signature header")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(a.secret))
	mac.Write(body)
	expectedSignature := "v1=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// Parse transforms PagerDuty payload to internal Incident
func (a *PagerDutyAdapter) Parse(body []byte) (*models.Incident, error) {
	var payload PagerDutyPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse pagerduty payload: %w", err)
	}

	// Only process incident.triggered events
	if payload.Event.EventType != "incident.triggered" {
		return nil, fmt.Errorf("unsupported event type: %s", payload.Event.EventType)
	}

	data := payload.Event.Data

	// Extract service name
	serviceName := data.Service.Summary
	if serviceName == "" {
		serviceName = "unknown"
	}

	// Map urgency to severity
	severity := mapPagerDutySeverity(data.Urgency)

	// Extract error message
	errorMessage := data.Title

	// Extract stack trace from body details
	var stackTrace *string
	if data.Body.Details != "" {
		if strings.Contains(data.Body.Details, "Stack trace:") || 
		   strings.Contains(data.Body.Details, "at ") ||
		   strings.Contains(data.Body.Details, "Traceback") {
			stackTrace = &data.Body.Details
		}
	}

	// Create incident ID
	incidentID := fmt.Sprintf("inc_pd_%s", data.ID)

	// Store provider data
	providerData := map[string]interface{}{
		"incident_id":  data.ID,
		"incident_url": data.HTMLURL,
		"service_id":   data.Service.ID,
		"urgency":      data.Urgency,
	}

	incident := &models.Incident{
		ID:           incidentID,
		ServiceName:  serviceName,
		Repository:   "", // Will be mapped later
		ErrorMessage: errorMessage,
		StackTrace:   stackTrace,
		Severity:     severity,
		Status:       models.StatusPending,
		Provider:     "pagerduty",
		ProviderData: providerData,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return incident, nil
}

// PagerDutyPayload represents a PagerDuty webhook payload
type PagerDutyPayload struct {
	Event PagerDutyEvent `json:"event"`
}

// PagerDutyEvent represents a PagerDuty event
type PagerDutyEvent struct {
	ID           string                `json:"id"`
	EventType    string                `json:"event_type"`
	ResourceType string                `json:"resource_type"`
	OccurredAt   string                `json:"occurred_at"`
	Data         PagerDutyIncidentData `json:"data"`
}

// PagerDutyIncidentData represents PagerDuty incident data
type PagerDutyIncidentData struct {
	ID      string                  `json:"id"`
	Type    string                  `json:"type"`
	Title   string                  `json:"title"`
	Service PagerDutyService        `json:"service"`
	Urgency string                  `json:"urgency"`
	Body    PagerDutyIncidentBody   `json:"body"`
	HTMLURL string                  `json:"html_url"`
}

// PagerDutyService represents a PagerDuty service
type PagerDutyService struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
}

// PagerDutyIncidentBody represents incident body details
type PagerDutyIncidentBody struct {
	Details string `json:"details"`
}

// mapPagerDutySeverity maps PagerDuty urgency to internal severity
func mapPagerDutySeverity(urgency string) string {
	switch strings.ToLower(urgency) {
	case "high":
		return "critical"
	case "low":
		return "medium"
	default:
		return "medium"
	}
}
