-- Create incident_events table for audit trail
CREATE TABLE IF NOT EXISTS incident_events (
    id SERIAL PRIMARY KEY,
    incident_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);

-- Create indexes for common queries
CREATE INDEX idx_incident_events_incident_id ON incident_events(incident_id);
CREATE INDEX idx_incident_events_created_at ON incident_events(created_at DESC);
CREATE INDEX idx_incident_events_event_type ON incident_events(event_type);
