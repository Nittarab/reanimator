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

// DatadogAdapter handles Datadog webhook payloads
type DatadogAdapter struct {
	secret string
}

// NewDatadogAdapter creates a new Datadog adapter
func NewDatadogAdapter() *DatadogAdapter {
	return &DatadogAdapter{
		secret: os.Getenv("DATADOG_WEBHOOK_SECRET"),
	}
}

// ProviderName returns the provider name
func (a *DatadogAdapter) ProviderName() string {
	return "datadog"
}

// Validate validates the webhook signature
func (a *DatadogAdapter) Validate(r *http.Request) error {
	if a.secret == "" {
		// If no secret is configured, skip validation
		return nil
	}

	signature := r.Header.Get("X-Datadog-Signature")
	if signature == "" {
		return fmt.Errorf("missing X-Datadog-Signature header")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(a.secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// Parse transforms Datadog payload to internal Incident
func (a *DatadogAdapter) Parse(body []byte) (*models.Incident, error) {
	var payload DatadogPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse datadog payload: %w", err)
	}

	// Validate required fields
	if payload.ID == "" {
		return nil, fmt.Errorf("missing required field: id")
	}
	if payload.Title == "" {
		return nil, fmt.Errorf("missing required field: title")
	}

	// Extract service name from tags
	serviceName := extractServiceFromTags(payload.Tags)
	if serviceName == "" {
		serviceName = "unknown"
	}

	// Map priority to severity
	severity := mapDatadogSeverity(payload.Priority)

	// Construct error message
	errorMessage := payload.Title
	if payload.Body != "" {
		errorMessage = fmt.Sprintf("%s: %s", payload.Title, payload.Body)
	}

	// Extract stack trace if present in body
	var stackTrace *string
	if strings.Contains(payload.Body, "Traceback") || strings.Contains(payload.Body, "at ") {
		stackTrace = &payload.Body
	}

	// Create incident ID
	incidentID := fmt.Sprintf("inc_dd_%s", payload.ID)

	// Store provider data
	providerData := map[string]interface{}{
		"alert_id":     payload.ID,
		"alert_type":   payload.AlertType,
		"tags":         payload.Tags,
		"date_happened": payload.DateHappened,
	}
	if payload.Snapshot != "" {
		providerData["snapshot_url"] = payload.Snapshot
	}

	incident := &models.Incident{
		ID:           incidentID,
		ServiceName:  serviceName,
		Repository:   "", // Will be mapped later
		ErrorMessage: errorMessage,
		StackTrace:   stackTrace,
		Severity:     severity,
		Status:       models.StatusPending,
		Provider:     "datadog",
		ProviderData: providerData,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return incident, nil
}

// DatadogPayload represents a Datadog webhook payload
type DatadogPayload struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Body           string   `json:"body"`
	AlertType      string   `json:"alert_type"`
	Priority       string   `json:"priority"`
	Tags           []string `json:"tags"`
	DateHappened   int64    `json:"date_happened"`
	AggregationKey string   `json:"aggregation_key"`
	SourceTypeName string   `json:"source_type_name"`
	Snapshot       string   `json:"snapshot"`
}

// extractServiceFromTags extracts service name from Datadog tags
func extractServiceFromTags(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "service:") {
			return strings.TrimPrefix(tag, "service:")
		}
	}
	return ""
}

// mapDatadogSeverity maps Datadog priority to internal severity
func mapDatadogSeverity(priority string) string {
	switch strings.ToLower(priority) {
	case "p1":
		return "critical"
	case "p2":
		return "high"
	case "p3":
		return "medium"
	case "p4", "low":
		return "low"
	default:
		return "medium"
	}
}
