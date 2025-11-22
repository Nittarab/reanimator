package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

var incidentCounter int

// generateIncidentID generates a unique incident ID for testing
func generateIncidentID() string {
	incidentCounter++
	return fmt.Sprintf("test_inc_%d_%d", time.Now().UnixNano(), incidentCounter)
}

// **Feature: ai-sre-platform, Property 13: Audit trail completeness**
// **Validates: Requirements 14.1, 20.3**
// For any incident processed through the system, the audit trail should contain
// records for all state transitions (received, workflow_triggered, in_progress, pr_created, etc.)
func TestProperty_AuditTrailCompleteness(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("test database not configured")
	}
	defer db.Close()

	repo := NewIncidentRepository(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("audit trail contains all state transitions", prop.ForAll(
		func(numTransitions int) bool {
			// Clean up before each test iteration
			cleanupTestData(db.DB)

			// Generate a unique incident ID
			incidentID := generateIncidentID()

			// Create initial incident
			incident := &models.Incident{
				ID:           incidentID,
				ServiceName:  "test-service",
				Repository:   "test/repo",
				ErrorMessage: "test error message",
				Severity:     "high",
				Status:       models.StatusPending,
				Provider:     "test",
				ProviderData: map[string]interface{}{},
			}

			// Create the incident (this logs the initial event)
			if err := repo.Create(incident); err != nil {
				t.Logf("Failed to create incident: %v", err)
				return false
			}

			// Apply status transitions
			statuses := []models.IncidentStatus{
				models.StatusWorkflowTriggered,
				models.StatusInProgress,
				models.StatusPRCreated,
				models.StatusResolved,
			}
			
			for i := 0; i < numTransitions && i < len(statuses); i++ {
				if err := repo.UpdateStatus(incidentID, statuses[i]); err != nil {
					t.Logf("Failed to update status: %v", err)
					return false
				}
				// Small delay to ensure distinct timestamps
				time.Sleep(1 * time.Millisecond)
			}

			// Retrieve all events for this incident
			events, err := repo.GetEventsByIncidentID(incidentID)
			if err != nil {
				t.Logf("Failed to get events: %v", err)
				return false
			}

			// Check that we have at least the creation event
			if len(events) < 1 {
				t.Logf("Expected at least 1 event (creation), got %d", len(events))
				return false
			}

			// Check that the first event is incident_received
			if events[0].EventType != models.EventIncidentReceived {
				t.Logf("Expected first event to be incident_received, got %s", events[0].EventType)
				return false
			}

			// Check that we have events for all status transitions
			// We expect: 1 creation event + numTransitions status change events
			expectedEventCount := 1 + numTransitions
			if len(events) != expectedEventCount {
				t.Logf("Expected %d events, got %d", expectedEventCount, len(events))
				return false
			}

			// Verify events are in chronological order
			for i := 1; i < len(events); i++ {
				if events[i].CreatedAt.Before(events[i-1].CreatedAt) {
					t.Logf("Events are not in chronological order")
					return false
				}
			}

			// Verify all events have the correct incident ID
			for _, event := range events {
				if event.IncidentID != incidentID {
					t.Logf("Event has incorrect incident ID: expected %s, got %s", incidentID, event.IncidentID)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 4),
	))

	properties.TestingRun(t)
}

// genIncidentStatus generates random incident statuses
func genIncidentStatus() gopter.Gen {
	statuses := []models.IncidentStatus{
		models.StatusWorkflowTriggered,
		models.StatusInProgress,
		models.StatusPRCreated,
		models.StatusResolved,
		models.StatusFailed,
	}
	return gen.OneConstOf(
		statuses[0], statuses[1], statuses[2], statuses[3], statuses[4],
	)
}

// **Feature: ai-sre-platform, Property 14: Incident filtering correctness**
// **Validates: Requirements 14.3, 19.5**
// For any filter criteria (status, service, repository, time range), the returned
// incidents should all match the filter conditions
func TestProperty_IncidentFilteringCorrectness(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("test database not configured")
	}
	defer db.Close()

	repo := NewIncidentRepository(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("filtered incidents match filter criteria", prop.ForAll(
		func(incidents []testIncident, filterStatus models.IncidentStatus) bool {
			// Clean up before each test iteration
			cleanupTestData(db.DB)

			// Create all incidents
			for _, inc := range incidents {
				incident := &models.Incident{
					ID:           inc.ID,
					ServiceName:  inc.ServiceName,
					Repository:   inc.Repository,
					ErrorMessage: "test error",
					Severity:     "high",
					Status:       inc.Status,
					Provider:     "test",
					ProviderData: map[string]interface{}{},
				}
				if err := repo.Create(incident); err != nil {
					t.Logf("Failed to create incident: %v", err)
					return false
				}
			}

			// Apply filter by status
			filter := &IncidentFilter{
				Status: &filterStatus,
			}

			filtered, err := repo.ListWithFilter(filter)
			if err != nil {
				t.Logf("Failed to list with filter: %v", err)
				return false
			}

			// Verify all returned incidents match the filter
			for _, incident := range filtered {
				if incident.Status != filterStatus {
					t.Logf("Incident %s has status %s, expected %s", incident.ID, incident.Status, filterStatus)
					return false
				}
			}

			// Count how many incidents should match
			expectedCount := 0
			for _, inc := range incidents {
				if inc.Status == filterStatus {
					expectedCount++
				}
			}

			if len(filtered) != expectedCount {
				t.Logf("Expected %d incidents, got %d", expectedCount, len(filtered))
				return false
			}

			return true
		},
		genTestIncidents(),
		genIncidentStatus(),
	))

	properties.Property("filtered incidents by service name match criteria", prop.ForAll(
		func(incidents []testIncident, filterService string) bool {
			// Clean up before each test iteration
			cleanupTestData(db.DB)

			// Create all incidents
			for _, inc := range incidents {
				incident := &models.Incident{
					ID:           inc.ID,
					ServiceName:  inc.ServiceName,
					Repository:   inc.Repository,
					ErrorMessage: "test error",
					Severity:     "high",
					Status:       inc.Status,
					Provider:     "test",
					ProviderData: map[string]interface{}{},
				}
				if err := repo.Create(incident); err != nil {
					t.Logf("Failed to create incident: %v", err)
					return false
				}
			}

			// Apply filter by service name
			filter := &IncidentFilter{
				ServiceName: &filterService,
			}

			filtered, err := repo.ListWithFilter(filter)
			if err != nil {
				t.Logf("Failed to list with filter: %v", err)
				return false
			}

			// Verify all returned incidents match the filter
			for _, incident := range filtered {
				if incident.ServiceName != filterService {
					t.Logf("Incident %s has service %s, expected %s", incident.ID, incident.ServiceName, filterService)
					return false
				}
			}

			// Count how many incidents should match
			expectedCount := 0
			for _, inc := range incidents {
				if inc.ServiceName == filterService {
					expectedCount++
				}
			}

			if len(filtered) != expectedCount {
				t.Logf("Expected %d incidents, got %d", expectedCount, len(filtered))
				return false
			}

			return true
		},
		genTestIncidents(),
		gen.OneConstOf("service-a", "service-b", "service-c"),
	))

	properties.Property("filtered incidents by time range match criteria", prop.ForAll(
		func(numIncidents int) bool {
			// Clean up before each test iteration
			cleanupTestData(db.DB)

			now := time.Now()
			
			// Create incidents with different timestamps
			for i := 0; i < numIncidents; i++ {
				incident := &models.Incident{
					ID:           generateIncidentID(),
					ServiceName:  "test-service",
					Repository:   "test/repo",
					ErrorMessage: "test error",
					Severity:     "high",
					Status:       models.StatusPending,
					Provider:     "test",
					ProviderData: map[string]interface{}{},
				}
				if err := repo.Create(incident); err != nil {
					t.Logf("Failed to create incident: %v", err)
					return false
				}
				
				// Update the created_at timestamp to be in the past
				pastTime := now.Add(-time.Duration(i+1) * time.Hour)
				_, err := db.DB.Exec("UPDATE incidents SET created_at = $1 WHERE id = $2", pastTime, incident.ID)
				if err != nil {
					t.Logf("Failed to update timestamp: %v", err)
					return false
				}
			}

			// Filter for incidents in the last 3 hours
			startTime := now.Add(-3 * time.Hour)
			filter := &IncidentFilter{
				StartTime: &startTime,
			}

			filtered, err := repo.ListWithFilter(filter)
			if err != nil {
				t.Logf("Failed to list with filter: %v", err)
				return false
			}

			// Verify all returned incidents are within the time range
			for _, incident := range filtered {
				if incident.CreatedAt.Before(startTime) {
					t.Logf("Incident %s created at %v is before start time %v", incident.ID, incident.CreatedAt, startTime)
					return false
				}
			}

			// Count how many incidents should match (first 3 in our case, or less if we have fewer)
			expectedCount := numIncidents
			if expectedCount > 3 {
				expectedCount = 3
			}

			if len(filtered) != expectedCount {
				t.Logf("Expected %d incidents, got %d", expectedCount, len(filtered))
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// testIncident is a helper struct for generating test incidents
type testIncident struct {
	ID          string
	ServiceName string
	Repository  string
	Status      models.IncidentStatus
}

// genTestIncidents generates a slice of test incidents with unique IDs
func genTestIncidents() gopter.Gen {
	return gen.IntRange(2, 10).Map(func(count int) []testIncident {
		incidents := make([]testIncident, count)
		services := []string{"service-a", "service-b", "service-c"}
		repos := []string{"repo-1", "repo-2", "repo-3"}
		statuses := []models.IncidentStatus{
			models.StatusPending,
			models.StatusWorkflowTriggered,
			models.StatusInProgress,
			models.StatusPRCreated,
			models.StatusResolved,
			models.StatusFailed,
		}
		
		for i := 0; i < count; i++ {
			incidents[i] = testIncident{
				ID:          generateIncidentID(),
				ServiceName: services[i%len(services)],
				Repository:  repos[i%len(repos)],
				Status:      statuses[i%len(statuses)],
			}
		}
		return incidents
	})
}

// **Feature: ai-sre-platform, Property 15: Statistics computation accuracy**
// **Validates: Requirements 14.4**
// For any set of incidents, computing success rate should equal (successful incidents / total incidents)
// and mean time to resolution should equal the average of (completed_at - created_at) for resolved incidents
func TestProperty_StatisticsComputationAccuracy(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("test database not configured")
	}
	defer db.Close()

	repo := NewIncidentRepository(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("success rate is correctly computed", prop.ForAll(
		func(resolvedCount, failedCount, pendingCount int) bool {
			// Clean up before each test iteration
			cleanupTestData(db.DB)

			// Create incidents with different statuses
			totalCount := resolvedCount + failedCount + pendingCount
			
			if totalCount == 0 {
				return true // Skip empty case
			}

			// Create resolved incidents
			for i := 0; i < resolvedCount; i++ {
				incident := &models.Incident{
					ID:           generateIncidentID(),
					ServiceName:  "test-service",
					Repository:   "test/repo",
					ErrorMessage: "test error",
					Severity:     "high",
					Status:       models.StatusResolved,
					Provider:     "test",
					ProviderData: map[string]interface{}{},
				}
				if err := repo.Create(incident); err != nil {
					t.Logf("Failed to create incident: %v", err)
					return false
				}
			}

			// Create failed incidents
			for i := 0; i < failedCount; i++ {
				incident := &models.Incident{
					ID:           generateIncidentID(),
					ServiceName:  "test-service",
					Repository:   "test/repo",
					ErrorMessage: "test error",
					Severity:     "high",
					Status:       models.StatusFailed,
					Provider:     "test",
					ProviderData: map[string]interface{}{},
				}
				if err := repo.Create(incident); err != nil {
					t.Logf("Failed to create incident: %v", err)
					return false
				}
			}

			// Create pending incidents
			for i := 0; i < pendingCount; i++ {
				incident := &models.Incident{
					ID:           generateIncidentID(),
					ServiceName:  "test-service",
					Repository:   "test/repo",
					ErrorMessage: "test error",
					Severity:     "high",
					Status:       models.StatusPending,
					Provider:     "test",
					ProviderData: map[string]interface{}{},
				}
				if err := repo.Create(incident); err != nil {
					t.Logf("Failed to create incident: %v", err)
					return false
				}
			}

			// Get statistics
			stats, err := repo.GetStatistics(nil)
			if err != nil {
				t.Logf("Failed to get statistics: %v", err)
				return false
			}

			// Verify total count
			if stats.TotalIncidents != totalCount {
				t.Logf("Expected total %d, got %d", totalCount, stats.TotalIncidents)
				return false
			}

			// Verify resolved count
			if stats.ResolvedIncidents != resolvedCount {
				t.Logf("Expected resolved %d, got %d", resolvedCount, stats.ResolvedIncidents)
				return false
			}

			// Verify failed count
			if stats.FailedIncidents != failedCount {
				t.Logf("Expected failed %d, got %d", failedCount, stats.FailedIncidents)
				return false
			}

			// Verify success rate
			expectedSuccessRate := float64(resolvedCount) / float64(totalCount)
			if !floatEquals(stats.SuccessRate, expectedSuccessRate, 0.001) {
				t.Logf("Expected success rate %.3f, got %.3f", expectedSuccessRate, stats.SuccessRate)
				return false
			}

			return true
		},
		gen.IntRange(0, 10),
		gen.IntRange(0, 10),
		gen.IntRange(0, 10),
	))

	properties.Property("mean time to resolve is correctly computed", prop.ForAll(
		func(numIncidents int) bool {
			// Clean up before each test iteration
			cleanupTestData(db.DB)

			now := time.Now()
			var totalSeconds float64

			// Create incidents with specific resolution times
			for i := 0; i < numIncidents; i++ {
				incidentID := generateIncidentID()
				seconds := (i + 1) * 60 // 60, 120, 180, etc. seconds
				createdAt := now.Add(-time.Duration(seconds) * time.Second)
				completedAt := now

				incident := &models.Incident{
					ID:           incidentID,
					ServiceName:  "test-service",
					Repository:   "test/repo",
					ErrorMessage: "test error",
					Severity:     "high",
					Status:       models.StatusResolved,
					Provider:     "test",
					ProviderData: map[string]interface{}{},
				}
				if err := repo.Create(incident); err != nil {
					t.Logf("Failed to create incident: %v", err)
					return false
				}

				// Update timestamps
				_, err := db.DB.Exec(
					"UPDATE incidents SET created_at = $1, completed_at = $2 WHERE id = $3",
					createdAt, completedAt, incidentID,
				)
				if err != nil {
					t.Logf("Failed to update timestamps: %v", err)
					return false
				}

				totalSeconds += float64(seconds)
			}

			// Get statistics
			stats, err := repo.GetStatistics(nil)
			if err != nil {
				t.Logf("Failed to get statistics: %v", err)
				return false
			}

			// Verify mean time to resolve
			expectedMTTR := totalSeconds / float64(numIncidents)
			if !floatEquals(stats.MeanTimeToResolve, expectedMTTR, 1.0) {
				t.Logf("Expected MTTR %.2f, got %.2f", expectedMTTR, stats.MeanTimeToResolve)
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// floatEquals checks if two floats are approximately equal
func floatEquals(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}
