package models

import (
	"fmt"
	"time"
)

// IncidentService handles incident business logic
type IncidentService struct {
	repo              IncidentRepository
	serviceMappings   map[string]ServiceMapping
	deduplicationTime time.Duration
}

// IncidentRepository defines the interface for incident persistence
type IncidentRepository interface {
	Create(incident *Incident) error
	GetByID(id string) (*Incident, error)
	Update(incident *Incident) error
	UpdateStatus(id string, status IncidentStatus) error
	List() ([]*Incident, error)
	FindDuplicateIncident(serviceName, errorMessage string, timeWindow time.Duration) (*Incident, error)
}

// ServiceMapping maps a service name to a repository
type ServiceMapping struct {
	ServiceName string
	Repository  string
	Branch      string
}

// NewIncidentService creates a new incident service
func NewIncidentService(repo IncidentRepository, mappings []ServiceMapping, deduplicationTime time.Duration) *IncidentService {
	// Build a map for fast lookups
	mappingMap := make(map[string]ServiceMapping)
	for _, mapping := range mappings {
		mappingMap[mapping.ServiceName] = mapping
	}

	return &IncidentService{
		repo:              repo,
		serviceMappings:   mappingMap,
		deduplicationTime: deduplicationTime,
	}
}

// CreateIncident creates a new incident with deduplication and service mapping
func (s *IncidentService) CreateIncident(incident *Incident) (*Incident, error) {
	// Check for duplicates within the time window
	duplicate, err := s.repo.FindDuplicateIncident(incident.ServiceName, incident.ErrorMessage, s.deduplicationTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	// If duplicate found, update and return it
	if duplicate != nil {
		duplicate.UpdatedAt = time.Now()
		if err := s.repo.Update(duplicate); err != nil {
			return nil, fmt.Errorf("failed to update duplicate incident: %w", err)
		}
		return duplicate, nil
	}

	// Map service to repository
	mapping, found := s.serviceMappings[incident.ServiceName]
	if found {
		incident.Repository = mapping.Repository
		incident.Status = StatusPending
	} else {
		// Service not mapped - mark as requiring manual setup
		incident.Repository = ""
		incident.Status = StatusFailed
	}

	// Create the incident
	if err := s.repo.Create(incident); err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	return incident, nil
}

// GetIncident retrieves an incident by ID
func (s *IncidentService) GetIncident(id string) (*Incident, error) {
	return s.repo.GetByID(id)
}

// ListIncidents retrieves all incidents
func (s *IncidentService) ListIncidents() ([]*Incident, error) {
	return s.repo.List()
}

// UpdateIncidentStatus updates the status of an incident
func (s *IncidentService) UpdateIncidentStatus(id string, status IncidentStatus) error {
	return s.repo.UpdateStatus(id, status)
}

// LookupRepository looks up the repository for a service name
func (s *IncidentService) LookupRepository(serviceName string) (string, bool) {
	mapping, found := s.serviceMappings[serviceName]
	if !found {
		return "", false
	}
	return mapping.Repository, true
}

// TransitionStatus validates and performs a status transition
func (s *IncidentService) TransitionStatus(incident *Incident, newStatus IncidentStatus) error {
	// Validate state transitions
	validTransitions := map[IncidentStatus][]IncidentStatus{
		StatusPending: {StatusWorkflowTriggered, StatusFailed},
		StatusWorkflowTriggered: {StatusInProgress, StatusFailed},
		StatusInProgress: {StatusPRCreated, StatusFailed, StatusNoFixNeeded},
		StatusPRCreated: {StatusResolved, StatusFailed},
		StatusFailed: {StatusPending}, // Allow retry
		StatusNoFixNeeded: {},
		StatusResolved: {},
	}

	allowed := false
	for _, validStatus := range validTransitions[incident.Status] {
		if validStatus == newStatus {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("invalid status transition from %s to %s", incident.Status, newStatus)
	}

	incident.Status = newStatus
	incident.UpdatedAt = time.Now()

	// Update timestamps based on status
	switch newStatus {
	case StatusWorkflowTriggered:
		now := time.Now()
		incident.TriggeredAt = &now
	case StatusResolved, StatusFailed, StatusNoFixNeeded:
		now := time.Now()
		incident.CompletedAt = &now
	}

	return s.repo.Update(incident)
}
