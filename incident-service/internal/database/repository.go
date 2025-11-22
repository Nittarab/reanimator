package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// IncidentRepository handles incident database operations
type IncidentRepository struct {
	db *DB
}

// NewIncidentRepository creates a new incident repository
func NewIncidentRepository(db *DB) *IncidentRepository {
	return &IncidentRepository{db: db}
}

// Create inserts a new incident into the database
func (r *IncidentRepository) Create(incident *models.Incident) error {
	providerDataJSON, err := json.Marshal(incident.ProviderData)
	if err != nil {
		return fmt.Errorf("failed to marshal provider data: %w", err)
	}

	query := `
		INSERT INTO incidents (
			id, service_name, repository, error_message, stack_trace,
			severity, status, provider, provider_data, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	incident.CreatedAt = now
	incident.UpdatedAt = now

	_, err = r.db.Exec(
		query,
		incident.ID,
		incident.ServiceName,
		incident.Repository,
		incident.ErrorMessage,
		incident.StackTrace,
		incident.Severity,
		incident.Status,
		incident.Provider,
		providerDataJSON,
		incident.CreatedAt,
		incident.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

// GetByID retrieves an incident by its ID
func (r *IncidentRepository) GetByID(id string) (*models.Incident, error) {
	query := `
		SELECT 
			id, service_name, repository, error_message, stack_trace,
			severity, status, provider, provider_data, workflow_run_id,
			pull_request_url, diagnosis, created_at, updated_at,
			triggered_at, completed_at
		FROM incidents
		WHERE id = $1
	`

	var incident models.Incident
	var providerDataJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&incident.ID,
		&incident.ServiceName,
		&incident.Repository,
		&incident.ErrorMessage,
		&incident.StackTrace,
		&incident.Severity,
		&incident.Status,
		&incident.Provider,
		&providerDataJSON,
		&incident.WorkflowRunID,
		&incident.PullRequestURL,
		&incident.Diagnosis,
		&incident.CreatedAt,
		&incident.UpdatedAt,
		&incident.TriggeredAt,
		&incident.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("incident not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	if err := json.Unmarshal(providerDataJSON, &incident.ProviderData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider data: %w", err)
	}

	return &incident, nil
}

// Update updates an existing incident
func (r *IncidentRepository) Update(incident *models.Incident) error {
	providerDataJSON, err := json.Marshal(incident.ProviderData)
	if err != nil {
		return fmt.Errorf("failed to marshal provider data: %w", err)
	}

	query := `
		UPDATE incidents
		SET service_name = $2, repository = $3, error_message = $4,
		    stack_trace = $5, severity = $6, status = $7, provider = $8,
		    provider_data = $9, workflow_run_id = $10, pull_request_url = $11,
		    diagnosis = $12, updated_at = $13, triggered_at = $14, completed_at = $15
		WHERE id = $1
	`

	incident.UpdatedAt = time.Now()

	_, err = r.db.Exec(
		query,
		incident.ID,
		incident.ServiceName,
		incident.Repository,
		incident.ErrorMessage,
		incident.StackTrace,
		incident.Severity,
		incident.Status,
		incident.Provider,
		providerDataJSON,
		incident.WorkflowRunID,
		incident.PullRequestURL,
		incident.Diagnosis,
		incident.UpdatedAt,
		incident.TriggeredAt,
		incident.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	return nil
}

// List retrieves all incidents with optional filtering
func (r *IncidentRepository) List() ([]*models.Incident, error) {
	query := `
		SELECT 
			id, service_name, repository, error_message, stack_trace,
			severity, status, provider, provider_data, workflow_run_id,
			pull_request_url, diagnosis, created_at, updated_at,
			triggered_at, completed_at
		FROM incidents
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*models.Incident
	for rows.Next() {
		var incident models.Incident
		var providerDataJSON []byte

		err := rows.Scan(
			&incident.ID,
			&incident.ServiceName,
			&incident.Repository,
			&incident.ErrorMessage,
			&incident.StackTrace,
			&incident.Severity,
			&incident.Status,
			&incident.Provider,
			&providerDataJSON,
			&incident.WorkflowRunID,
			&incident.PullRequestURL,
			&incident.Diagnosis,
			&incident.CreatedAt,
			&incident.UpdatedAt,
			&incident.TriggeredAt,
			&incident.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		if err := json.Unmarshal(providerDataJSON, &incident.ProviderData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider data: %w", err)
		}

		incidents = append(incidents, &incident)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incidents: %w", err)
	}

	return incidents, nil
}

// FindDuplicateIncident finds an existing incident with the same service and error within the time window
func (r *IncidentRepository) FindDuplicateIncident(serviceName, errorMessage string, timeWindow time.Duration) (*models.Incident, error) {
	query := `
		SELECT 
			id, service_name, repository, error_message, stack_trace,
			severity, status, provider, provider_data, workflow_run_id,
			pull_request_url, diagnosis, created_at, updated_at,
			triggered_at, completed_at
		FROM incidents
		WHERE service_name = $1 
		  AND error_message = $2
		  AND created_at > $3
		ORDER BY created_at DESC
		LIMIT 1
	`

	cutoffTime := time.Now().Add(-timeWindow)
	var incident models.Incident
	var providerDataJSON []byte

	err := r.db.QueryRow(query, serviceName, errorMessage, cutoffTime).Scan(
		&incident.ID,
		&incident.ServiceName,
		&incident.Repository,
		&incident.ErrorMessage,
		&incident.StackTrace,
		&incident.Severity,
		&incident.Status,
		&incident.Provider,
		&providerDataJSON,
		&incident.WorkflowRunID,
		&incident.PullRequestURL,
		&incident.Diagnosis,
		&incident.CreatedAt,
		&incident.UpdatedAt,
		&incident.TriggeredAt,
		&incident.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No duplicate found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicate incident: %w", err)
	}

	if err := json.Unmarshal(providerDataJSON, &incident.ProviderData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider data: %w", err)
	}

	return &incident, nil
}

// UpdateStatus updates the status of an incident
func (r *IncidentRepository) UpdateStatus(id string, status models.IncidentStatus) error {
	query := `
		UPDATE incidents
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update incident status: %w", err)
	}

	return nil
}
