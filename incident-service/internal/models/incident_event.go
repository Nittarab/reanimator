package models

import (
	"time"
)

// IncidentEventType represents the type of event in the incident lifecycle
type IncidentEventType string

const (
	EventIncidentReceived       IncidentEventType = "incident_received"
	EventWorkflowTriggered      IncidentEventType = "workflow_triggered"
	EventWorkflowInProgress     IncidentEventType = "workflow_in_progress"
	EventPRCreated              IncidentEventType = "pr_created"
	EventIncidentResolved       IncidentEventType = "incident_resolved"
	EventIncidentFailed         IncidentEventType = "incident_failed"
	EventManualTrigger          IncidentEventType = "manual_trigger"
	EventStatusChanged          IncidentEventType = "status_changed"
	EventDuplicateDetected      IncidentEventType = "duplicate_detected"
	EventQueuedForRemediation   IncidentEventType = "queued_for_remediation"
	EventDequeuedForRemediation IncidentEventType = "dequeued_for_remediation"
)

// IncidentEvent represents an event in the incident lifecycle for audit trail
type IncidentEvent struct {
	ID         int64                  `json:"id" db:"id"`
	IncidentID string                 `json:"incident_id" db:"incident_id"`
	EventType  IncidentEventType      `json:"event_type" db:"event_type"`
	EventData  map[string]interface{} `json:"event_data" db:"event_data"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}
