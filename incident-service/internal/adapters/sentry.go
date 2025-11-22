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

// SentryAdapter handles Sentry webhook payloads
type SentryAdapter struct {
	secret string
}

// NewSentryAdapter creates a new Sentry adapter
func NewSentryAdapter() *SentryAdapter {
	return &SentryAdapter{
		secret: os.Getenv("SENTRY_WEBHOOK_SECRET"),
	}
}

// ProviderName returns the provider name
func (a *SentryAdapter) ProviderName() string {
	return "sentry"
}

// Validate validates the webhook signature
func (a *SentryAdapter) Validate(r *http.Request) error {
	if a.secret == "" {
		// If no secret is configured, skip validation
		return nil
	}

	signature := r.Header.Get("Sentry-Hook-Signature")
	if signature == "" {
		return fmt.Errorf("missing Sentry-Hook-Signature header")
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

// Parse transforms Sentry payload to internal Incident
func (a *SentryAdapter) Parse(body []byte) (*models.Incident, error) {
	var payload SentryPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse sentry payload: %w", err)
	}

	// Only process created events
	if payload.Action != "created" {
		return nil, fmt.Errorf("unsupported action: %s", payload.Action)
	}

	// Extract service name from tags or project
	serviceName := extractServiceFromSentryTags(payload.Data.Event.Tags)
	if serviceName == "" {
		serviceName = payload.Data.Issue.Project
	}
	if serviceName == "" {
		serviceName = "unknown"
	}

	// Map level to severity
	severity := mapSentrySeverity(payload.Data.Issue.Level)

	// Extract error message
	errorMessage := payload.Data.Issue.Title

	// Extract stack trace
	var stackTrace *string
	if stackTraceStr := extractStackTraceFromSentry(payload.Data.Event); stackTraceStr != "" {
		stackTrace = &stackTraceStr
	}

	// Create incident ID
	incidentID := fmt.Sprintf("inc_sentry_%s", payload.Data.Issue.ID)

	// Store provider data
	providerData := map[string]interface{}{
		"issue_id":   payload.Data.Issue.ID,
		"event_id":   payload.Data.Event.EventID,
		"issue_url":  payload.URL,
		"platform":   payload.Data.Issue.Platform,
		"culprit":    payload.Data.Issue.Culprit,
	}

	incident := &models.Incident{
		ID:           incidentID,
		ServiceName:  serviceName,
		Repository:   "", // Will be mapped later
		ErrorMessage: errorMessage,
		StackTrace:   stackTrace,
		Severity:     severity,
		Status:       models.StatusPending,
		Provider:     "sentry",
		ProviderData: providerData,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return incident, nil
}

// SentryPayload represents a Sentry webhook payload
type SentryPayload struct {
	Action string     `json:"action"`
	Data   SentryData `json:"data"`
	URL    string     `json:"url"`
}

// SentryData represents Sentry data
type SentryData struct {
	Issue SentryIssue `json:"issue"`
	Event SentryEvent `json:"event"`
}

// SentryIssue represents a Sentry issue
type SentryIssue struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Culprit  string `json:"culprit"`
	Level    string `json:"level"`
	Platform string `json:"platform"`
	Project  string `json:"project"`
}

// SentryEvent represents a Sentry event
type SentryEvent struct {
	EventID   string                   `json:"event_id"`
	Timestamp string                   `json:"timestamp"`
	Exception *SentryException         `json:"exception"`
	Tags      [][]string               `json:"tags"`
}

// SentryException represents exception data
type SentryException struct {
	Values []SentryExceptionValue `json:"values"`
}

// SentryExceptionValue represents an exception value
type SentryExceptionValue struct {
	Type       string              `json:"type"`
	Value      string              `json:"value"`
	Stacktrace *SentryStacktrace   `json:"stacktrace"`
}

// SentryStacktrace represents a stack trace
type SentryStacktrace struct {
	Frames []SentryFrame `json:"frames"`
}

// SentryFrame represents a stack frame
type SentryFrame struct {
	Filename string `json:"filename"`
	Function string `json:"function"`
	Lineno   int    `json:"lineno"`
}

// extractServiceFromSentryTags extracts service name from Sentry tags
func extractServiceFromSentryTags(tags [][]string) string {
	for _, tag := range tags {
		if len(tag) == 2 && tag[0] == "service" {
			return tag[1]
		}
		if len(tag) == 2 && tag[0] == "app" {
			return tag[1]
		}
	}
	return ""
}

// mapSentrySeverity maps Sentry level to internal severity
func mapSentrySeverity(level string) string {
	switch strings.ToLower(level) {
	case "fatal":
		return "critical"
	case "error":
		return "high"
	case "warning":
		return "medium"
	case "info", "debug":
		return "low"
	default:
		return "medium"
	}
}

// extractStackTraceFromSentry formats Sentry exception into a stack trace string
func extractStackTraceFromSentry(event SentryEvent) string {
	if event.Exception == nil || len(event.Exception.Values) == 0 {
		return ""
	}

	var sb strings.Builder
	
	for _, exc := range event.Exception.Values {
		if exc.Type != "" && exc.Value != "" {
			sb.WriteString(fmt.Sprintf("%s: %s\n", exc.Type, exc.Value))
		}
		
		if exc.Stacktrace != nil && len(exc.Stacktrace.Frames) > 0 {
			for _, frame := range exc.Stacktrace.Frames {
				sb.WriteString(fmt.Sprintf("  at %s (%s:%d)\n", 
					frame.Function, frame.Filename, frame.Lineno))
			}
		}
	}

	return strings.TrimSpace(sb.String())
}
