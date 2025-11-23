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

// Create inserts a new incident into the database and logs the creation event
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

	// Log the incident creation event
	event := &models.IncidentEvent{
		IncidentID: incident.ID,
		EventType:  models.EventIncidentReceived,
		EventData: map[string]interface{}{
			"provider":     incident.Provider,
			"service_name": incident.ServiceName,
			"severity":     incident.Severity,
		},
	}

	if err := r.LogEvent(event); err != nil {
		// Log error but don't fail the incident creation
		return fmt.Errorf("failed to log incident creation event: %w", err)
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

// IncidentFilter represents filtering options for incident queries
type IncidentFilter struct {
	Status      *models.IncidentStatus
	ServiceName *string
	Repository  *string
	StartTime   *time.Time
	EndTime     *time.Time
}

// List retrieves all incidents with optional filtering
func (r *IncidentRepository) List() ([]*models.Incident, error) {
	return r.ListWithFilter(nil)
}

// ListWithFilter retrieves incidents with optional filtering
func (r *IncidentRepository) ListWithFilter(filter *IncidentFilter) ([]*models.Incident, error) {
	query := `
		SELECT 
			id, service_name, repository, error_message, stack_trace,
			severity, status, provider, provider_data, workflow_run_id,
			pull_request_url, diagnosis, created_at, updated_at,
			triggered_at, completed_at
		FROM incidents
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if filter != nil {
		if filter.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argCount)
			args = append(args, *filter.Status)
			argCount++
		}
		if filter.ServiceName != nil {
			query += fmt.Sprintf(" AND service_name = $%d", argCount)
			args = append(args, *filter.ServiceName)
			argCount++
		}
		if filter.Repository != nil {
			query += fmt.Sprintf(" AND repository = $%d", argCount)
			args = append(args, *filter.Repository)
			argCount++
		}
		if filter.StartTime != nil {
			query += fmt.Sprintf(" AND created_at >= $%d", argCount)
			args = append(args, *filter.StartTime)
			argCount++
		}
		if filter.EndTime != nil {
			query += fmt.Sprintf(" AND created_at <= $%d", argCount)
			args = append(args, *filter.EndTime)
			// argCount++ not needed after last parameter
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
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

// UpdateStatus updates the status of an incident and logs the status change event
func (r *IncidentRepository) UpdateStatus(id string, status models.IncidentStatus) error {
	// Get the current incident to check the old status
	incident, err := r.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get incident for status update: %w", err)
	}

	oldStatus := incident.Status

	query := `
		UPDATE incidents
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err = r.db.Exec(query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update incident status: %w", err)
	}

	// Log the status change event
	var eventType models.IncidentEventType
	switch status {
	case models.StatusWorkflowTriggered:
		eventType = models.EventWorkflowTriggered
	case models.StatusInProgress:
		eventType = models.EventWorkflowInProgress
	case models.StatusPRCreated:
		eventType = models.EventPRCreated
	case models.StatusResolved:
		eventType = models.EventIncidentResolved
	case models.StatusFailed:
		eventType = models.EventIncidentFailed
	default:
		eventType = models.EventStatusChanged
	}

	event := &models.IncidentEvent{
		IncidentID: id,
		EventType:  eventType,
		EventData: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": status,
		},
	}

	if err := r.LogEvent(event); err != nil {
		// Log error but don't fail the status update
		return fmt.Errorf("failed to log status change event: %w", err)
	}

	return nil
}

// LogEvent logs an event in the incident lifecycle for audit trail
func (r *IncidentRepository) LogEvent(event *models.IncidentEvent) error {
	eventDataJSON, err := json.Marshal(event.EventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	query := `
		INSERT INTO incident_events (incident_id, event_type, event_data, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	event.CreatedAt = now

	err = r.db.QueryRow(query, event.IncidentID, event.EventType, eventDataJSON, event.CreatedAt).Scan(&event.ID)
	if err != nil {
		return fmt.Errorf("failed to log event: %w", err)
	}

	return nil
}

// GetEventsByIncidentID retrieves all events for a specific incident
func (r *IncidentRepository) GetEventsByIncidentID(incidentID string) ([]*models.IncidentEvent, error) {
	query := `
		SELECT id, incident_id, event_type, event_data, created_at
		FROM incident_events
		WHERE incident_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	var events []*models.IncidentEvent
	for rows.Next() {
		var event models.IncidentEvent
		var eventDataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.IncidentID,
			&event.EventType,
			&eventDataJSON,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if err := json.Unmarshal(eventDataJSON, &event.EventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// IncidentStatistics represents aggregated statistics about incidents
type IncidentStatistics struct {
	TotalIncidents    int     `json:"total_incidents"`
	ResolvedIncidents int     `json:"resolved_incidents"`
	FailedIncidents   int     `json:"failed_incidents"`
	SuccessRate       float64 `json:"success_rate"`
	MeanTimeToResolve float64 `json:"mean_time_to_resolve_seconds"`
}

// GetStatistics computes aggregated statistics for incidents
func (r *IncidentRepository) GetStatistics(filter *IncidentFilter) (*IncidentStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'resolved' OR status = 'pr_created' THEN 1 END) as resolved,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
			AVG(EXTRACT(EPOCH FROM (completed_at - created_at))) as avg_resolution_time
		FROM incidents
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if filter != nil {
		if filter.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argCount)
			args = append(args, *filter.Status)
			argCount++
		}
		if filter.ServiceName != nil {
			query += fmt.Sprintf(" AND service_name = $%d", argCount)
			args = append(args, *filter.ServiceName)
			argCount++
		}
		if filter.Repository != nil {
			query += fmt.Sprintf(" AND repository = $%d", argCount)
			args = append(args, *filter.Repository)
			argCount++
		}
		if filter.StartTime != nil {
			query += fmt.Sprintf(" AND created_at >= $%d", argCount)
			args = append(args, *filter.StartTime)
			argCount++
		}
		if filter.EndTime != nil {
			query += fmt.Sprintf(" AND created_at <= $%d", argCount)
			args = append(args, *filter.EndTime)
			// argCount++ not needed after last parameter
		}
	}

	var stats IncidentStatistics
	var avgResolutionTime sql.NullFloat64

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalIncidents,
		&stats.ResolvedIncidents,
		&stats.FailedIncidents,
		&avgResolutionTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Calculate success rate
	if stats.TotalIncidents > 0 {
		stats.SuccessRate = float64(stats.ResolvedIncidents) / float64(stats.TotalIncidents)
	}

	// Set mean time to resolve
	if avgResolutionTime.Valid {
		stats.MeanTimeToResolve = avgResolutionTime.Float64
	}

	return &stats, nil
}

// DeleteOldIncidents deletes incidents older than the retention period
func (r *IncidentRepository) DeleteOldIncidents(retentionPeriod time.Duration) (int64, error) {
	query := `
		DELETE FROM incidents
		WHERE created_at < $1
	`

	cutoffTime := time.Now().Add(-retentionPeriod)
	result, err := r.db.Exec(query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old incidents: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
