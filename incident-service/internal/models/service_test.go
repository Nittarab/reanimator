package models

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// MockIncidentRepository is a mock implementation for testing
type MockIncidentRepository struct {
	incidents map[string]*Incident
	created   []*Incident
}

func NewMockIncidentRepository() *MockIncidentRepository {
	return &MockIncidentRepository{
		incidents: make(map[string]*Incident),
		created:   make([]*Incident, 0),
	}
}

func (m *MockIncidentRepository) Create(incident *Incident) error {
	// Set timestamps like the real repository does
	now := time.Now()
	incident.CreatedAt = now
	incident.UpdatedAt = now
	
	m.incidents[incident.ID] = incident
	m.created = append(m.created, incident)
	return nil
}

func (m *MockIncidentRepository) GetByID(id string) (*Incident, error) {
	incident, ok := m.incidents[id]
	if !ok {
		return nil, nil
	}
	return incident, nil
}

func (m *MockIncidentRepository) Update(incident *Incident) error {
	m.incidents[incident.ID] = incident
	return nil
}

func (m *MockIncidentRepository) UpdateStatus(id string, status IncidentStatus) error {
	if incident, ok := m.incidents[id]; ok {
		incident.Status = status
		incident.UpdatedAt = time.Now()
	}
	return nil
}

func (m *MockIncidentRepository) List() ([]*Incident, error) {
	incidents := make([]*Incident, 0, len(m.incidents))
	for _, incident := range m.incidents {
		incidents = append(incidents, incident)
	}
	return incidents, nil
}

func (m *MockIncidentRepository) FindDuplicateIncident(serviceName, errorMessage string, timeWindow time.Duration) (*Incident, error) {
	cutoffTime := time.Now().Add(-timeWindow)
	for _, incident := range m.incidents {
		if incident.ServiceName == serviceName &&
			incident.ErrorMessage == errorMessage &&
			incident.CreatedAt.After(cutoffTime) {
			return incident, nil
		}
	}
	return nil, nil
}

// **Feature: ai-sre-platform, Property 4: Service-to-repository lookup consistency**
// **Validates: Requirements 2.2**
func TestProperty_ServiceLookupConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("For any configured service name, looking up the repository should always return the same result", prop.ForAll(
		func(serviceName string, repository string, branch string) bool {
			// Create service mappings
			mappings := []ServiceMapping{
				{
					ServiceName: serviceName,
					Repository:  repository,
					Branch:      branch,
				},
			}

			// Create incident service
			repo := NewMockIncidentRepository()
			service := NewIncidentService(repo, mappings, 5*time.Minute)

			// Perform multiple lookups
			result1, found1 := service.LookupRepository(serviceName)
			result2, found2 := service.LookupRepository(serviceName)
			result3, found3 := service.LookupRepository(serviceName)

			// All lookups should return the same result
			if found1 != found2 || found2 != found3 {
				return false
			}

			if !found1 {
				return true // Service not found is consistent
			}

			// All results should be identical
			return result1 == result2 && result2 == result3 && result1 == repository
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: ai-sre-platform, Property 5: Deduplication within time window**
// **Validates: Requirements 2.3**
func TestProperty_DeduplicationWithinTimeWindow(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("For any two incidents with the same service and error within time window, only one should exist", prop.ForAll(
		func(serviceName string, errorMessage string, repository string) bool {
			// Create service mappings
			mappings := []ServiceMapping{
				{
					ServiceName: serviceName,
					Repository:  repository,
					Branch:      "main",
				},
			}

			// Create incident service with 5 minute deduplication window
			repo := NewMockIncidentRepository()
			service := NewIncidentService(repo, mappings, 5*time.Minute)

			// Create first incident
			incident1 := &Incident{
				ID:           "inc_1",
				ServiceName:  serviceName,
				ErrorMessage: errorMessage,
				Severity:     "high",
				Provider:     "test",
				ProviderData: make(map[string]interface{}),
			}

			result1, err := service.CreateIncident(incident1)
			if err != nil {
				return false
			}

			// Create second incident with same service and error (should be deduplicated)
			incident2 := &Incident{
				ID:           "inc_2",
				ServiceName:  serviceName,
				ErrorMessage: errorMessage,
				Severity:     "high",
				Provider:     "test",
				ProviderData: make(map[string]interface{}),
			}

			result2, err := service.CreateIncident(incident2)
			if err != nil {
				return false
			}

			// Both should return the same incident (the first one)
			if result1.ID != result2.ID {
				return false
			}

			// Only one incident should be created in the repository
			if len(repo.created) != 1 {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test for status state machine
func TestIncidentStatusTransitions(t *testing.T) {
	repo := NewMockIncidentRepository()
	service := NewIncidentService(repo, []ServiceMapping{}, 5*time.Minute)

	tests := []struct {
		name        string
		fromStatus  IncidentStatus
		toStatus    IncidentStatus
		shouldError bool
	}{
		{"pending to workflow_triggered", StatusPending, StatusWorkflowTriggered, false},
		{"pending to failed", StatusPending, StatusFailed, false},
		{"workflow_triggered to in_progress", StatusWorkflowTriggered, StatusInProgress, false},
		{"workflow_triggered to failed", StatusWorkflowTriggered, StatusFailed, false},
		{"in_progress to pr_created", StatusInProgress, StatusPRCreated, false},
		{"in_progress to failed", StatusInProgress, StatusFailed, false},
		{"in_progress to no_fix_needed", StatusInProgress, StatusNoFixNeeded, false},
		{"pr_created to resolved", StatusPRCreated, StatusResolved, false},
		{"pr_created to failed", StatusPRCreated, StatusFailed, false},
		{"failed to pending", StatusFailed, StatusPending, false},
		{"pending to resolved", StatusPending, StatusResolved, true}, // Invalid
		{"resolved to pending", StatusResolved, StatusPending, true}, // Invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incident := &Incident{
				ID:           "test-" + tt.name,
				ServiceName:  "test-service",
				ErrorMessage: "test error",
				Status:       tt.fromStatus,
				Severity:     "high",
				Provider:     "test",
				ProviderData: make(map[string]interface{}),
			}

			// Create the incident in the repo
			repo.Create(incident)

			err := service.TransitionStatus(incident, tt.toStatus)

			if tt.shouldError && err == nil {
				t.Errorf("expected error for transition from %s to %s, but got none", tt.fromStatus, tt.toStatus)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error for transition from %s to %s: %v", tt.fromStatus, tt.toStatus, err)
			}

			if !tt.shouldError && incident.Status != tt.toStatus {
				t.Errorf("expected status %s, got %s", tt.toStatus, incident.Status)
			}
		})
	}
}

// Unit test for service mapping
func TestServiceMapping(t *testing.T) {
	mappings := []ServiceMapping{
		{ServiceName: "api-gateway", Repository: "org/api-gateway", Branch: "main"},
		{ServiceName: "user-service", Repository: "org/user-service", Branch: "main"},
	}

	repo := NewMockIncidentRepository()
	service := NewIncidentService(repo, mappings, 5*time.Minute)

	tests := []struct {
		name            string
		serviceName     string
		expectedRepo    string
		expectedFound   bool
		expectedStatus  IncidentStatus
	}{
		{"mapped service", "api-gateway", "org/api-gateway", true, StatusPending},
		{"another mapped service", "user-service", "org/user-service", true, StatusPending},
		{"unmapped service", "unknown-service", "", false, StatusFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incident := &Incident{
				ID:           "test-" + tt.name,
				ServiceName:  tt.serviceName,
				ErrorMessage: "test error",
				Severity:     "high",
				Provider:     "test",
				ProviderData: make(map[string]interface{}),
			}

			result, err := service.CreateIncident(incident)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Repository != tt.expectedRepo {
				t.Errorf("expected repository %s, got %s", tt.expectedRepo, result.Repository)
			}

			if result.Status != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, result.Status)
			}

			// Verify lookup
			repo, found := service.LookupRepository(tt.serviceName)
			if found != tt.expectedFound {
				t.Errorf("expected found=%v, got %v", tt.expectedFound, found)
			}
			if found && repo != tt.expectedRepo {
				t.Errorf("expected repository %s, got %s", tt.expectedRepo, repo)
			}
		})
	}
}
