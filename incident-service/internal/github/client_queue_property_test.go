package github

import (
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// TestProperty12_WorkflowCompletionUpdatesQueue verifies that when a workflow completes,
// the active workflow count is decremented and the next queued incident is processed
// **Feature: ai-sre-platform, Property 12: Workflow completion updates queue**
// **Validates: Requirements 12.5**
func TestProperty12_WorkflowCompletionUpdatesQueue(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("workflow completion decrements active count and processes queue", prop.ForAll(
		func(maxConcurrency uint8, queuedIncidents uint8) bool {
			// Ensure reasonable values
			if maxConcurrency == 0 {
				maxConcurrency = 1
			}
			if maxConcurrency > 10 {
				maxConcurrency = 10
			}
			if queuedIncidents > 20 {
				queuedIncidents = 20
			}

			repository := "test-org/test-repo"
			client := NewClient("https://api.github.com", "test-token", "test-workflow.yml", int(maxConcurrency))

			// Manually set up the state: maxConcurrency active workflows and queuedIncidents in queue
			client.mu.Lock()
			client.activeWorkflows[repository] = int(maxConcurrency)
			
			// Create queued incidents
			for i := 0; i < int(queuedIncidents); i++ {
				incident := &models.Incident{
					ID:           fmt.Sprintf("queued_inc_%d", i),
					ServiceName:  "test-service",
					Repository:   repository,
					ErrorMessage: fmt.Sprintf("queued error %d", i),
					Status:       models.StatusPending,
					CreatedAt:    time.Now(),
				}
				client.queuedIncidents[repository] = append(client.queuedIncidents[repository], incident)
			}
			client.mu.Unlock()

			// Verify initial state
			initialActive := client.GetActiveCount(repository)
			if initialActive != int(maxConcurrency) {
				t.Logf("Initial active count mismatch: expected %d, got %d", maxConcurrency, initialActive)
				return false
			}

			initialQueued := client.GetQueuedCount(repository)
			if initialQueued != int(queuedIncidents) {
				t.Logf("Initial queued count mismatch: expected %d, got %d", queuedIncidents, initialQueued)
				return false
			}

			// Simulate workflow completion by calling DecrementActive
			nextIncident := client.DecrementActive(repository)

			// Verify active count was decremented
			newActive := client.GetActiveCount(repository)
			expectedActive := int(maxConcurrency) - 1
			if newActive != expectedActive {
				t.Logf("Active count after decrement: expected %d, got %d", expectedActive, newActive)
				return false
			}

			// Verify queue behavior based on whether there were queued incidents
			if queuedIncidents > 0 {
				// Should return the next incident from the queue
				if nextIncident == nil {
					t.Logf("Expected next incident from queue, got nil")
					return false
				}

				// Verify it's the first incident that was queued
				if nextIncident.ID != "queued_inc_0" {
					t.Logf("Expected first queued incident (queued_inc_0), got %s", nextIncident.ID)
					return false
				}

				// Verify queue count was decremented
				newQueued := client.GetQueuedCount(repository)
				expectedQueued := int(queuedIncidents) - 1
				if newQueued != expectedQueued {
					t.Logf("Queued count after processing: expected %d, got %d", expectedQueued, newQueued)
					return false
				}
			} else {
				// No queued incidents, should return nil
				if nextIncident != nil {
					t.Logf("Expected nil when queue is empty, got incident %s", nextIncident.ID)
					return false
				}

				// Queue count should remain 0
				newQueued := client.GetQueuedCount(repository)
				if newQueued != 0 {
					t.Logf("Expected queued count 0, got %d", newQueued)
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

// TestProperty12_MultipleCompletions verifies that multiple workflow completions
// correctly process the queue in FIFO order
func TestProperty12_MultipleCompletions(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("multiple completions process queue in FIFO order", prop.ForAll(
		func(maxConcurrency uint8, queuedIncidents uint8) bool {
			// Ensure reasonable values
			if maxConcurrency == 0 {
				maxConcurrency = 1
			}
			if maxConcurrency > 5 {
				maxConcurrency = 5
			}
			if queuedIncidents < 2 || queuedIncidents > 10 {
				queuedIncidents = 5
			}

			repository := "test-org/test-repo"
			client := NewClient("https://api.github.com", "test-token", "test-workflow.yml", int(maxConcurrency))

			// Set up initial state
			client.mu.Lock()
			client.activeWorkflows[repository] = int(maxConcurrency)
			
			// Create queued incidents with predictable IDs
			expectedOrder := []string{}
			for i := 0; i < int(queuedIncidents); i++ {
				incidentID := fmt.Sprintf("inc_%03d", i)
				expectedOrder = append(expectedOrder, incidentID)
				incident := &models.Incident{
					ID:           incidentID,
					ServiceName:  "test-service",
					Repository:   repository,
					ErrorMessage: fmt.Sprintf("error %d", i),
					Status:       models.StatusPending,
					CreatedAt:    time.Now(),
				}
				client.queuedIncidents[repository] = append(client.queuedIncidents[repository], incident)
			}
			client.mu.Unlock()

			// Process all queued incidents by calling DecrementActive multiple times
			processedOrder := []string{}
			for i := 0; i < int(queuedIncidents); i++ {
				nextIncident := client.DecrementActive(repository)
				if nextIncident == nil {
					t.Logf("Expected incident at position %d, got nil", i)
					return false
				}
				processedOrder = append(processedOrder, nextIncident.ID)
			}

			// Verify FIFO order
			if len(processedOrder) != len(expectedOrder) {
				t.Logf("Processed count mismatch: expected %d, got %d", len(expectedOrder), len(processedOrder))
				return false
			}

			for i := range expectedOrder {
				if processedOrder[i] != expectedOrder[i] {
					t.Logf("Order mismatch at position %d: expected %s, got %s", i, expectedOrder[i], processedOrder[i])
					return false
				}
			}

			// Verify queue is now empty
			finalQueued := client.GetQueuedCount(repository)
			if finalQueued != 0 {
				t.Logf("Expected empty queue, got %d incidents", finalQueued)
				return false
			}

			// Verify active count was decremented correctly
			// Each DecrementActive call reduces the count by 1, so after queuedIncidents calls,
			// the active count should be maxConcurrency - queuedIncidents (but not below 0)
			finalActive := client.GetActiveCount(repository)
			expectedFinalActive := int(maxConcurrency) - int(queuedIncidents)
			if expectedFinalActive < 0 {
				expectedFinalActive = 0
			}
			if finalActive != expectedFinalActive {
				t.Logf("Final active count: expected %d, got %d", expectedFinalActive, finalActive)
				return false
			}

			return true
		},
		gen.UInt8Range(1, 5),
		gen.UInt8Range(2, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty12_NoUnderflow verifies that decrementing active count
// never goes below zero
func TestProperty12_NoUnderflow(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("active count never goes below zero", prop.ForAll(
		func(initialActive uint8, decrements uint8) bool {
			if initialActive > 10 {
				initialActive = 10
			}
			if decrements > 20 {
				decrements = 20
			}

			repository := "test-org/test-repo"
			client := NewClient("https://api.github.com", "test-token", "test-workflow.yml", 10)

			// Set initial active count
			client.mu.Lock()
			client.activeWorkflows[repository] = int(initialActive)
			client.mu.Unlock()

			// Perform decrements
			for i := 0; i < int(decrements); i++ {
				client.DecrementActive(repository)
				
				// Check that active count is never negative
				activeCount := client.GetActiveCount(repository)
				if activeCount < 0 {
					t.Logf("Active count went negative: %d", activeCount)
					return false
				}
			}

			// Final active count should be max(0, initialActive - decrements)
			finalActive := client.GetActiveCount(repository)
			expectedFinal := int(initialActive) - int(decrements)
			if expectedFinal < 0 {
				expectedFinal = 0
			}

			if finalActive != expectedFinal {
				t.Logf("Final active count: expected %d, got %d", expectedFinal, finalActive)
				return false
			}

			return true
		},
		gen.UInt8Range(0, 10),
		gen.UInt8Range(0, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
