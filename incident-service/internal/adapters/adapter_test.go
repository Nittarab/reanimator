package adapters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: ai-sre-platform, Property 2: Provider format transformation**
// **Validates: Requirements 1.2**
//
// Property: For any valid provider-specific incident payload, transforming it to the internal
// Incident structure should produce a valid Incident with all required fields populated.
func TestProperty_ProviderFormatTransformation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Test Datadog adapter
	properties.Property("Datadog: valid payload produces valid incident", prop.ForAll(
		func(id, title, body, priority string, tags []string, dateHappened int64) bool {
			// Create a valid Datadog payload
			payload := DatadogPayload{
				ID:           id,
				Title:        title,
				Body:         body,
				AlertType:    "error",
				Priority:     priority,
				Tags:         tags,
				DateHappened: dateHappened,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Logf("Failed to marshal payload: %v", err)
				return false
			}

			adapter := NewDatadogAdapter()
			incident, err := adapter.Parse(payloadBytes)
			if err != nil {
				t.Logf("Failed to parse payload: %v", err)
				return false
			}

			// Verify all required fields are populated
			if incident.ID == "" {
				t.Logf("Missing incident ID")
				return false
			}
			if incident.ServiceName == "" {
				t.Logf("Missing service name")
				return false
			}
			if incident.ErrorMessage == "" {
				t.Logf("Missing error message")
				return false
			}
			if incident.Severity == "" {
				t.Logf("Missing severity")
				return false
			}
			if incident.Status == "" {
				t.Logf("Missing status")
				return false
			}
			if incident.Provider != "datadog" {
				t.Logf("Wrong provider: %s", incident.Provider)
				return false
			}
			if incident.ProviderData == nil {
				t.Logf("Missing provider data")
				return false
			}
			if incident.CreatedAt.IsZero() {
				t.Logf("Missing created_at timestamp")
				return false
			}
			if incident.UpdatedAt.IsZero() {
				t.Logf("Missing updated_at timestamp")
				return false
			}

			return true
		},
		gen.Identifier(),                                    // id
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }), // title
		gen.AlphaString(),                                   // body
		gen.OneConstOf("P1", "P2", "P3", "P4", "low"),      // priority
		gen.SliceOf(gen.Const("service:test-service")),     // tags with service
		gen.Int64Range(1000000000, 9999999999),             // dateHappened
	))

	// Test PagerDuty adapter
	properties.Property("PagerDuty: valid payload produces valid incident", prop.ForAll(
		func(incidentID, title, serviceName, urgency string) bool {
			// Create a valid PagerDuty payload
			payload := PagerDutyPayload{
				Event: PagerDutyEvent{
					ID:           "evt_" + incidentID,
					EventType:    "incident.triggered",
					ResourceType: "incident",
					OccurredAt:   "2024-01-15T10:00:00Z",
					Data: PagerDutyIncidentData{
						ID:    incidentID,
						Type:  "incident",
						Title: title,
						Service: PagerDutyService{
							ID:      "svc_123",
							Summary: serviceName,
						},
						Urgency: urgency,
						Body: PagerDutyIncidentBody{
							Details: "Error details",
						},
						HTMLURL: "https://example.pagerduty.com/incidents/" + incidentID,
					},
				},
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Logf("Failed to marshal payload: %v", err)
				return false
			}

			adapter := NewPagerDutyAdapter()
			incident, err := adapter.Parse(payloadBytes)
			if err != nil {
				t.Logf("Failed to parse payload: %v", err)
				return false
			}

			// Verify all required fields are populated
			if incident.ID == "" {
				t.Logf("Missing incident ID")
				return false
			}
			if incident.ServiceName == "" {
				t.Logf("Missing service name")
				return false
			}
			if incident.ErrorMessage == "" {
				t.Logf("Missing error message")
				return false
			}
			if incident.Severity == "" {
				t.Logf("Missing severity")
				return false
			}
			if incident.Status == "" {
				t.Logf("Missing status")
				return false
			}
			if incident.Provider != "pagerduty" {
				t.Logf("Wrong provider: %s", incident.Provider)
				return false
			}
			if incident.ProviderData == nil {
				t.Logf("Missing provider data")
				return false
			}

			return true
		},
		gen.Identifier(),                                    // incidentID
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }), // title
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }), // serviceName
		gen.OneConstOf("high", "low"),                       // urgency
	))

	// Test Grafana adapter
	properties.Property("Grafana: valid payload produces valid incident", prop.ForAll(
		func(title, message, ruleID, ruleName, state string, labels map[string]string) bool {
			// Ensure service label exists
			if labels == nil {
				labels = make(map[string]string)
			}
			labels["service"] = "test-service"

			// Create a valid Grafana payload
			payload := GrafanaPayload{
				Title:    title,
				State:    state,
				Message:  message,
				RuleID:   ruleID,
				RuleName: ruleName,
				RuleURL:  "https://grafana.example.com/rules/" + ruleID,
				Labels:   labels,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Logf("Failed to marshal payload: %v", err)
				return false
			}

			adapter := NewGrafanaAdapter()
			incident, err := adapter.Parse(payloadBytes)
			if err != nil {
				t.Logf("Failed to parse payload: %v", err)
				return false
			}

			// Verify all required fields are populated
			if incident.ID == "" {
				t.Logf("Missing incident ID")
				return false
			}
			if incident.ServiceName == "" {
				t.Logf("Missing service name")
				return false
			}
			if incident.ErrorMessage == "" {
				t.Logf("Missing error message")
				return false
			}
			if incident.Severity == "" {
				t.Logf("Missing severity")
				return false
			}
			if incident.Status == "" {
				t.Logf("Missing status")
				return false
			}
			if incident.Provider != "grafana" {
				t.Logf("Wrong provider: %s", incident.Provider)
				return false
			}
			if incident.ProviderData == nil {
				t.Logf("Missing provider data")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }), // title
		gen.AlphaString(),                                   // message
		gen.Identifier(),                                    // ruleID
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }), // ruleName
		gen.OneConstOf("alerting", "firing"),               // state
		gen.MapOf(gen.Identifier(), gen.AlphaString()),     // labels
	))

	// Test Sentry adapter
	properties.Property("Sentry: valid payload produces valid incident", prop.ForAll(
		func(issueID, title, project, level string) bool {
			// Create a valid Sentry payload
			payload := SentryPayload{
				Action: "created",
				Data: SentryData{
					Issue: SentryIssue{
						ID:       issueID,
						Title:    title,
						Culprit:  "app/services/user-service.js in getUser",
						Level:    level,
						Platform: "javascript",
						Project:  project,
					},
					Event: SentryEvent{
						EventID:   "evt_" + issueID,
						Timestamp: "2024-01-15T10:00:00Z",
						Tags:      [][]string{{"service", project}},
					},
				},
				URL: fmt.Sprintf("https://sentry.io/issues/%s/", issueID),
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Logf("Failed to marshal payload: %v", err)
				return false
			}

			adapter := NewSentryAdapter()
			incident, err := adapter.Parse(payloadBytes)
			if err != nil {
				t.Logf("Failed to parse payload: %v", err)
				return false
			}

			// Verify all required fields are populated
			if incident.ID == "" {
				t.Logf("Missing incident ID")
				return false
			}
			if incident.ServiceName == "" {
				t.Logf("Missing service name")
				return false
			}
			if incident.ErrorMessage == "" {
				t.Logf("Missing error message")
				return false
			}
			if incident.Severity == "" {
				t.Logf("Missing severity")
				return false
			}
			if incident.Status == "" {
				t.Logf("Missing status")
				return false
			}
			if incident.Provider != "sentry" {
				t.Logf("Wrong provider: %s", incident.Provider)
				return false
			}
			if incident.ProviderData == nil {
				t.Logf("Missing provider data")
				return false
			}

			return true
		},
		gen.Identifier(),                                    // issueID
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }), // title
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }), // project
		gen.OneConstOf("fatal", "error", "warning", "info"), // level
	))

	properties.TestingRun(t)
}

// **Feature: ai-sre-platform, Property 3: Malformed data error handling**
// **Validates: Requirements 1.4**
//
// Property: For any malformed or invalid webhook payload, the Incident Service should
// return an error without crashing.
func TestProperty_MalformedDataErrorHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Test that random bytes don't crash the parsers
	properties.Property("Random bytes don't crash Datadog parser", prop.ForAll(
		func(randomBytes []byte) bool {
			adapter := NewDatadogAdapter()
			_, err := adapter.Parse(randomBytes)
			// We expect an error, but the parser should not panic
			return err != nil
		},
		gen.SliceOf(gen.UInt8()),
	))

	properties.Property("Random bytes don't crash PagerDuty parser", prop.ForAll(
		func(randomBytes []byte) bool {
			adapter := NewPagerDutyAdapter()
			_, err := adapter.Parse(randomBytes)
			// We expect an error, but the parser should not panic
			return err != nil
		},
		gen.SliceOf(gen.UInt8()),
	))

	properties.Property("Random bytes don't crash Grafana parser", prop.ForAll(
		func(randomBytes []byte) bool {
			adapter := NewGrafanaAdapter()
			_, err := adapter.Parse(randomBytes)
			// We expect an error, but the parser should not panic
			return err != nil
		},
		gen.SliceOf(gen.UInt8()),
	))

	properties.Property("Random bytes don't crash Sentry parser", prop.ForAll(
		func(randomBytes []byte) bool {
			adapter := NewSentryAdapter()
			_, err := adapter.Parse(randomBytes)
			// We expect an error, but the parser should not panic
			return err != nil
		},
		gen.SliceOf(gen.UInt8()),
	))

	// Test that invalid JSON structures return errors
	properties.Property("Invalid JSON returns error for Datadog", prop.ForAll(
		func(invalidJSON string) bool {
			adapter := NewDatadogAdapter()
			_, err := adapter.Parse([]byte(invalidJSON))
			return err != nil
		},
		gen.OneConstOf(
			"not json at all",
			"{incomplete json",
			"[1,2,3]",
			"null",
			"123",
			`{"wrong": "structure"}`,
		),
	))

	properties.Property("Invalid JSON returns error for PagerDuty", prop.ForAll(
		func(invalidJSON string) bool {
			adapter := NewPagerDutyAdapter()
			_, err := adapter.Parse([]byte(invalidJSON))
			return err != nil
		},
		gen.OneConstOf(
			"not json at all",
			"{incomplete json",
			"[1,2,3]",
			"null",
			"123",
			`{"wrong": "structure"}`,
		),
	))

	properties.Property("Invalid JSON returns error for Grafana", prop.ForAll(
		func(invalidJSON string) bool {
			adapter := NewGrafanaAdapter()
			_, err := adapter.Parse([]byte(invalidJSON))
			return err != nil
		},
		gen.OneConstOf(
			"not json at all",
			"{incomplete json",
			"[1,2,3]",
			"null",
			"123",
			`{"wrong": "structure"}`,
		),
	))

	properties.Property("Invalid JSON returns error for Sentry", prop.ForAll(
		func(invalidJSON string) bool {
			adapter := NewSentryAdapter()
			_, err := adapter.Parse([]byte(invalidJSON))
			return err != nil
		},
		gen.OneConstOf(
			"not json at all",
			"{incomplete json",
			"[1,2,3]",
			"null",
			"123",
			`{"wrong": "structure"}`,
		),
	))

	// Test that unsupported event types return errors
	properties.Property("PagerDuty rejects non-triggered events", prop.ForAll(
		func(eventType string) bool {
			payload := PagerDutyPayload{
				Event: PagerDutyEvent{
					EventType: eventType,
					Data: PagerDutyIncidentData{
						ID:    "test",
						Title: "test",
						Service: PagerDutyService{
							Summary: "test",
						},
					},
				},
			}

			payloadBytes, _ := json.Marshal(payload)
			adapter := NewPagerDutyAdapter()
			_, err := adapter.Parse(payloadBytes)
			
			// Should return error for non-triggered events
			return err != nil
		},
		gen.OneConstOf(
			"incident.resolved",
			"incident.acknowledged",
			"incident.escalated",
			"incident.reassigned",
		),
	))

	properties.Property("Grafana rejects non-alerting states", prop.ForAll(
		func(state string) bool {
			payload := GrafanaPayload{
				Title:    "test",
				State:    state,
				RuleID:   "test",
				RuleName: "test",
				Labels:   map[string]string{"service": "test"},
			}

			payloadBytes, _ := json.Marshal(payload)
			adapter := NewGrafanaAdapter()
			_, err := adapter.Parse(payloadBytes)
			
			// Should return error for non-alerting states
			return err != nil
		},
		gen.OneConstOf(
			"ok",
			"paused",
			"pending",
			"no_data",
		),
	))

	properties.Property("Sentry rejects non-created actions", prop.ForAll(
		func(action string) bool {
			payload := SentryPayload{
				Action: action,
				Data: SentryData{
					Issue: SentryIssue{
						ID:    "test",
						Title: "test",
					},
					Event: SentryEvent{
						EventID: "test",
					},
				},
			}

			payloadBytes, _ := json.Marshal(payload)
			adapter := NewSentryAdapter()
			_, err := adapter.Parse(payloadBytes)
			
			// Should return error for non-created actions
			return err != nil
		},
		gen.OneConstOf(
			"resolved",
			"assigned",
			"ignored",
			"deleted",
		),
	))

	properties.TestingRun(t)
}
