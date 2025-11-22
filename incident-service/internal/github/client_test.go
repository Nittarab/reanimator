package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// TestProperty6_WorkflowDispatchIncludesRequiredContext verifies that workflow dispatch
// includes all required context fields
// **Feature: ai-sre-platform, Property 6: Workflow dispatch includes required context**
// **Validates: Requirements 3.2**
func TestProperty6_WorkflowDispatchIncludesRequiredContext(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("workflow dispatch includes required fields", prop.ForAll(
		func(incidentID, serviceName, errorMessage string, createdAt time.Time, hasStackTrace bool) bool {
			// Create a test incident
			var stackTrace *string
			if hasStackTrace {
				st := "at function() line 42"
				stackTrace = &st
			}

			incident := &models.Incident{
				ID:           incidentID,
				ServiceName:  serviceName,
				Repository:   "test-org/test-repo",
				ErrorMessage: errorMessage,
				StackTrace:   stackTrace,
				Status:       models.StatusPending,
				CreatedAt:    createdAt,
			}

			// Create a test server that captures the request
			var capturedRequest WorkflowDispatchRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &capturedRequest)
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			// Create client and dispatch workflow
			client := NewClient(server.URL, "test-token", "test-workflow.yml", 2)
			ctx := context.Background()
			_, err := client.DispatchWorkflow(ctx, incident, "main")

			if err != nil {
				t.Logf("Dispatch failed: %v", err)
				return false
			}

			// Verify all required fields are present and non-empty
			inputs := capturedRequest.Inputs
			
			if inputs.IncidentID == "" {
				t.Logf("Missing incident_id")
				return false
			}
			if inputs.IncidentID != incidentID {
				t.Logf("incident_id mismatch: expected %s, got %s", incidentID, inputs.IncidentID)
				return false
			}

			if inputs.ErrorMessage == "" {
				t.Logf("Missing error_message")
				return false
			}
			if inputs.ErrorMessage != errorMessage {
				t.Logf("error_message mismatch: expected %s, got %s", errorMessage, inputs.ErrorMessage)
				return false
			}

			if inputs.ServiceName == "" {
				t.Logf("Missing service_name")
				return false
			}
			if inputs.ServiceName != serviceName {
				t.Logf("service_name mismatch: expected %s, got %s", serviceName, inputs.ServiceName)
				return false
			}

			if inputs.Timestamp == "" {
				t.Logf("Missing timestamp")
				return false
			}
			// Verify timestamp is valid RFC3339
			_, err = time.Parse(time.RFC3339, inputs.Timestamp)
			if err != nil {
				t.Logf("Invalid timestamp format: %s", inputs.Timestamp)
				return false
			}

			// Verify stack trace is included if present
			if hasStackTrace && inputs.StackTrace == "" {
				t.Logf("Stack trace was present but not included in inputs")
				return false
			}

			return true
		},
		gen.Identifier().SuchThat(func(s string) bool { return s != "" }),
		gen.Identifier().SuchThat(func(s string) bool { return s != "" }),
		gen.AnyString().SuchThat(func(s string) bool { return s != "" }),
		gen.Time(),
		gen.Bool(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_RetryWithExponentialBackoff verifies that failed workflow dispatch
// attempts are retried with exponential backoff
// **Feature: ai-sre-platform, Property 7: Retry with exponential backoff**
// **Validates: Requirements 3.3**
func TestProperty7_RetryWithExponentialBackoff(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing-sensitive test in short mode")
	}
	
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 5 // Reduce iterations since each test takes 3+ seconds
	properties := gopter.NewProperties(parameters)

	properties.Property("retries up to 3 times with exponential backoff", prop.ForAll(
		func(incidentID, serviceName, errorMessage string) bool {
			incident := &models.Incident{
				ID:           incidentID,
				ServiceName:  serviceName,
				Repository:   "test-org/test-repo",
				ErrorMessage: errorMessage,
				Status:       models.StatusPending,
				CreatedAt:    time.Now(),
			}

			// Track attempts and timing
			attemptCount := 0
			attemptTimes := []time.Time{}
			
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attemptCount++
				attemptTimes = append(attemptTimes, time.Now())
				// Always fail to trigger retries
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", "test-workflow.yml", 2)
			ctx := context.Background()
			
			startTime := time.Now()
			_, err := client.DispatchWorkflow(ctx, incident, "main")
			totalDuration := time.Since(startTime)

			// Should fail after 3 attempts
			if err == nil {
				t.Logf("Expected error after retries, got nil")
				return false
			}

			// Verify exactly 3 attempts were made
			if attemptCount != 3 {
				t.Logf("Expected 3 attempts, got %d", attemptCount)
				return false
			}

			// Verify exponential backoff timing
			// Expected delays: 0s (first attempt), 1s, 2s
			// Total should be at least 3 seconds
			if totalDuration < 3*time.Second {
				t.Logf("Total duration %v is less than expected 3s", totalDuration)
				return false
			}

			// Verify delays between attempts are increasing
			if len(attemptTimes) >= 2 {
				delay1 := attemptTimes[1].Sub(attemptTimes[0])
				if delay1 < 900*time.Millisecond { // Allow some tolerance
					t.Logf("First retry delay %v is less than expected ~1s", delay1)
					return false
				}
			}

			if len(attemptTimes) >= 3 {
				delay2 := attemptTimes[2].Sub(attemptTimes[1])
				if delay2 < 1900*time.Millisecond { // Allow some tolerance
					t.Logf("Second retry delay %v is less than expected ~2s", delay2)
					return false
				}
			}

			return true
		},
		gen.Identifier().SuchThat(func(s string) bool { return s != "" }),
		gen.Identifier().SuchThat(func(s string) bool { return s != "" }),
		gen.AnyString().SuchThat(func(s string) bool { return s != "" }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty8_ConcurrencyLimitEnforcement verifies that the concurrency limit
// is enforced and excess incidents are queued
// **Feature: ai-sre-platform, Property 8: Concurrency limit enforcement**
// **Validates: Requirements 3.4, 12.2, 12.3**
func TestProperty8_ConcurrencyLimitEnforcement(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("excess incidents are queued when limit reached", prop.ForAll(
		func(maxConcurrency uint8, extraIncidents uint8) bool {
			// Ensure reasonable values
			if maxConcurrency == 0 {
				maxConcurrency = 1
			}
			if maxConcurrency > 10 {
				maxConcurrency = 10
			}
			if extraIncidents > 20 {
				extraIncidents = 20
			}

			totalIncidents := int(maxConcurrency) + int(extraIncidents)
			repository := "test-org/test-repo"

			// Create a test server that always succeeds
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", "test-workflow.yml", int(maxConcurrency))
			ctx := context.Background()

			// Dispatch totalIncidents incidents
			successCount := 0
			queuedCount := 0

			for i := 0; i < totalIncidents; i++ {
				incident := &models.Incident{
					ID:           fmt.Sprintf("inc_%d", i),
					ServiceName:  "test-service",
					Repository:   repository,
					ErrorMessage: "test error",
					Status:       models.StatusPending,
					CreatedAt:    time.Now(),
				}

				_, err := client.DispatchWorkflow(ctx, incident, "main")
				if err != nil {
					if err.Error() == "concurrency limit reached, incident queued" {
						queuedCount++
					} else {
						t.Logf("Unexpected error: %v", err)
						return false
					}
				} else {
					successCount++
				}
			}

			// Verify that exactly maxConcurrency incidents were dispatched
			if successCount != int(maxConcurrency) {
				t.Logf("Expected %d successful dispatches, got %d", maxConcurrency, successCount)
				return false
			}

			// Verify that exactly extraIncidents were queued
			if queuedCount != int(extraIncidents) {
				t.Logf("Expected %d queued incidents, got %d", extraIncidents, queuedCount)
				return false
			}

			// Verify active count matches maxConcurrency
			activeCount := client.GetActiveCount(repository)
			if activeCount != int(maxConcurrency) {
				t.Logf("Expected active count %d, got %d", maxConcurrency, activeCount)
				return false
			}

			// Verify queued count matches extraIncidents
			queuedCountFromClient := client.GetQueuedCount(repository)
			if queuedCountFromClient != int(extraIncidents) {
				t.Logf("Expected queued count %d, got %d", extraIncidents, queuedCountFromClient)
				return false
			}

			// Test that decrementing active count processes queued incidents
			if extraIncidents > 0 {
				nextIncident := client.DecrementActive(repository)
				if nextIncident == nil {
					t.Logf("Expected next incident from queue, got nil")
					return false
				}

				// Verify active count decreased
				newActiveCount := client.GetActiveCount(repository)
				if newActiveCount != int(maxConcurrency)-1 {
					t.Logf("Expected active count %d after decrement, got %d", int(maxConcurrency)-1, newActiveCount)
					return false
				}

				// Verify queued count decreased
				newQueuedCount := client.GetQueuedCount(repository)
				if newQueuedCount != int(extraIncidents)-1 {
					t.Logf("Expected queued count %d after processing, got %d", int(extraIncidents)-1, newQueuedCount)
					return false
				}
			}

			return true
		},
		gen.UInt8Range(1, 10),
		gen.UInt8Range(0, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
