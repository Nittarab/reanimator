package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// IncidentStatus represents the current state of an incident
type IncidentStatus string

const (
	StatusPending           IncidentStatus = "pending"
	StatusWorkflowTriggered IncidentStatus = "workflow_triggered"
	StatusInProgress        IncidentStatus = "in_progress"
	StatusPRCreated         IncidentStatus = "pr_created"
	StatusResolved          IncidentStatus = "resolved"
	StatusFailed            IncidentStatus = "failed"
	StatusNoFixNeeded       IncidentStatus = "no_fix_needed"
)

// Incident represents an incident notification from an observability platform
type Incident struct {
	ID             string                 `json:"id" db:"id"`
	ServiceName    string                 `json:"service_name" db:"service_name"`
	Repository     string                 `json:"repository" db:"repository"`
	ErrorMessage   string                 `json:"error_message" db:"error_message"`
	StackTrace     *string                `json:"stack_trace,omitempty" db:"stack_trace"`
	Severity       string                 `json:"severity" db:"severity"`
	Status         IncidentStatus         `json:"status" db:"status"`
	Provider       string                 `json:"provider" db:"provider"`
	ProviderData   map[string]interface{} `json:"provider_data" db:"provider_data"`
	WorkflowRunID  *int64                 `json:"workflow_run_id,omitempty" db:"workflow_run_id"`
	PullRequestURL *string                `json:"pull_request_url,omitempty" db:"pull_request_url"`
	Diagnosis      *string                `json:"diagnosis,omitempty" db:"diagnosis"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	TriggeredAt    *time.Time             `json:"triggered_at,omitempty" db:"triggered_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// JSONB is a custom type for PostgreSQL JSONB columns
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
