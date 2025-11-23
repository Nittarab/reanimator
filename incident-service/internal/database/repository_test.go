package database

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// **Feature: ai-sre-platform, Property 1: Incident persistence round-trip**
// **Validates: Requirements 1.5**
//
// For any valid incident received via webhook, storing it to the database
// and then querying it by ID should return an equivalent incident with all
// required fields (unique identifier, timestamp, status) populated.
func TestProperty_IncidentPersistenceRoundTrip(t *testing.T) {
	// Skip if no test database is configured
	db := setupTestDB(t)
	if db == nil {
		t.Skip("test database not configured")
	}
	defer db.Close()

	repo := NewIncidentRepository(db)

	properties := gopter.NewProperties(nil)

	properties.Property("incident persistence round-trip", prop.ForAll(
		func(incident *models.Incident) bool {
			// Store the incident
			err := repo.Create(incident)
			if err != nil {
				t.Logf("failed to create incident: %v", err)
				return false
			}

			// Retrieve the incident
			retrieved, err := repo.GetByID(incident.ID)
			if err != nil {
				t.Logf("failed to retrieve incident: %v", err)
				return false
			}

			// Verify all required fields are populated and match
			if retrieved.ID != incident.ID {
				t.Logf("ID mismatch: expected %s, got %s", incident.ID, retrieved.ID)
				return false
			}

			if retrieved.ServiceName != incident.ServiceName {
				t.Logf("ServiceName mismatch: expected %s, got %s", incident.ServiceName, retrieved.ServiceName)
				return false
			}

			if retrieved.ErrorMessage != incident.ErrorMessage {
				t.Logf("ErrorMessage mismatch: expected %s, got %s", incident.ErrorMessage, retrieved.ErrorMessage)
				return false
			}

			if retrieved.Status != incident.Status {
				t.Logf("Status mismatch: expected %s, got %s", incident.Status, retrieved.Status)
				return false
			}

			if retrieved.Provider != incident.Provider {
				t.Logf("Provider mismatch: expected %s, got %s", incident.Provider, retrieved.Provider)
				return false
			}

			if retrieved.Severity != incident.Severity {
				t.Logf("Severity mismatch: expected %s, got %s", incident.Severity, retrieved.Severity)
				return false
			}

			// Verify timestamps are populated
			if retrieved.CreatedAt.IsZero() {
				t.Logf("CreatedAt is zero")
				return false
			}

			if retrieved.UpdatedAt.IsZero() {
				t.Logf("UpdatedAt is zero")
				return false
			}

			// Verify unique identifier is present
			if retrieved.ID == "" {
				t.Logf("ID is empty")
				return false
			}

			return true
		},
		genIncident(),
	))

	// Run at least 100 iterations as specified in the design
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties.TestingRun(t, parameters)
}

// genIncident generates random valid incidents for property-based testing
func genIncident() gopter.Gen {
	return gopter.CombineGens(
		gen.Identifier(),                                                                    // ID
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }), // ServiceName
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 200 }), // ErrorMessage
		gen.OneConstOf("low", "medium", "high", "critical"),                                 // Severity
		gen.OneConstOf(models.StatusPending, models.StatusWorkflowTriggered, models.StatusInProgress), // Status
		gen.OneConstOf("datadog", "pagerduty", "grafana", "sentry"),                        // Provider
	).Map(func(values []interface{}) *models.Incident {
		id := values[0].(string)
		serviceName := values[1].(string)
		errorMessage := values[2].(string)
		severity := values[3].(string)
		status := values[4].(models.IncidentStatus)
		provider := values[5].(string)

		return &models.Incident{
			ID:           fmt.Sprintf("inc_%s_%s", provider, id),
			ServiceName:  serviceName,
			Repository:   "",
			ErrorMessage: errorMessage,
			Severity:     severity,
			Status:       status,
			Provider:     provider,
			ProviderData: map[string]interface{}{
				"test": "data",
			},
		}
	})
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *DB {
	// Use environment variable for test database
	dsn := getTestDatabaseDSN()
	if dsn == "" {
		return nil
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Logf("failed to connect to test database: %v (skipping test)", err)
		return nil
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Logf("failed to ping test database: %v (skipping test)", err)
		db.Close()
		return nil
	}

	// Create test schema
	if err := setupTestSchema(db); err != nil {
		t.Logf("failed to setup test schema: %v (skipping test)", err)
		db.Close()
		return nil
	}

	// Clean up before test
	if err := cleanupTestData(db); err != nil {
		t.Logf("failed to cleanup test data: %v (skipping test)", err)
		db.Close()
		return nil
	}

	// Register cleanup
	t.Cleanup(func() {
		cleanupTestData(db)
		db.Close()
	})

	return &DB{db}
}

// getTestDatabaseDSN returns the test database connection string
func getTestDatabaseDSN() string {
	// Check for test database environment variable
	if dsn := getEnvOrDefault("TEST_DATABASE_URL", ""); dsn != "" {
		return dsn
	}

	// Default test database configuration
	host := getEnvOrDefault("TEST_DATABASE_HOST", "localhost")
	port := getEnvOrDefault("TEST_DATABASE_PORT", "5432")
	user := getEnvOrDefault("TEST_DATABASE_USER", "postgres")
	password := getEnvOrDefault("TEST_DATABASE_PASSWORD", "postgres")
	dbname := getEnvOrDefault("TEST_DATABASE_NAME", "ai_sre_test")
	sslmode := getEnvOrDefault("TEST_DATABASE_SSL_MODE", "disable")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupTestSchema creates the test database schema
func setupTestSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS incidents (
			id VARCHAR(255) PRIMARY KEY,
			service_name VARCHAR(255) NOT NULL,
			repository VARCHAR(255) NOT NULL DEFAULT '',
			error_message TEXT NOT NULL,
			stack_trace TEXT,
			severity VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL,
			provider VARCHAR(50) NOT NULL,
			provider_data JSONB NOT NULL DEFAULT '{}',
			workflow_run_id BIGINT,
			pull_request_url TEXT,
			diagnosis TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			triggered_at TIMESTAMP,
			completed_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS incident_events (
			id SERIAL PRIMARY KEY,
			incident_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(100) NOT NULL,
			event_data JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
		);
	`

	_, err := db.Exec(schema)
	return err
}

// cleanupTestData removes all test data
func cleanupTestData(db *sql.DB) error {
	// Delete in order due to foreign key constraints
	// Use TRUNCATE for faster cleanup and to reset sequences
	_, err := db.Exec("TRUNCATE TABLE incident_events CASCADE")
	if err != nil {
		// Fallback to DELETE if TRUNCATE fails
		_, err = db.Exec("DELETE FROM incident_events")
		if err != nil {
			return err
		}
	}
	_, err = db.Exec("TRUNCATE TABLE incidents CASCADE")
	if err != nil {
		// Fallback to DELETE if TRUNCATE fails
		_, err = db.Exec("DELETE FROM incidents")
	}
	return err
}

// Unit test for basic incident creation
func TestIncidentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("test database not configured")
	}
	defer db.Close()

	repo := NewIncidentRepository(db)

	incident := &models.Incident{
		ID:           "inc_test_001",
		ServiceName:  "test-service",
		Repository:   "org/test-repo",
		ErrorMessage: "test error",
		Severity:     "high",
		Status:       models.StatusPending,
		Provider:     "datadog",
		ProviderData: map[string]interface{}{
			"alert_id": "12345",
		},
	}

	err := repo.Create(incident)
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	// Verify incident was created
	retrieved, err := repo.GetByID(incident.ID)
	if err != nil {
		t.Fatalf("failed to retrieve incident: %v", err)
	}

	if retrieved.ID != incident.ID {
		t.Errorf("expected ID %s, got %s", incident.ID, retrieved.ID)
	}

	if retrieved.ServiceName != incident.ServiceName {
		t.Errorf("expected ServiceName %s, got %s", incident.ServiceName, retrieved.ServiceName)
	}
}

// Unit test for incident retrieval
func TestIncidentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("test database not configured")
	}
	defer db.Close()

	repo := NewIncidentRepository(db)

	// Create test incident
	incident := &models.Incident{
		ID:           "inc_test_002",
		ServiceName:  "test-service",
		Repository:   "org/test-repo",
		ErrorMessage: "test error",
		Severity:     "high",
		Status:       models.StatusPending,
		Provider:     "datadog",
		ProviderData: map[string]interface{}{},
	}

	err := repo.Create(incident)
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	// Test retrieval
	retrieved, err := repo.GetByID(incident.ID)
	if err != nil {
		t.Fatalf("failed to retrieve incident: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved incident is nil")
	}

	// Test non-existent incident
	_, err = repo.GetByID("non_existent_id")
	if err == nil {
		t.Error("expected error for non-existent incident")
	}
}

// Unit test for incident update
func TestIncidentRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("test database not configured")
	}
	defer db.Close()

	repo := NewIncidentRepository(db)

	// Create test incident
	incident := &models.Incident{
		ID:           "inc_test_003",
		ServiceName:  "test-service",
		Repository:   "org/test-repo",
		ErrorMessage: "test error",
		Severity:     "high",
		Status:       models.StatusPending,
		Provider:     "datadog",
		ProviderData: map[string]interface{}{},
	}

	err := repo.Create(incident)
	if err != nil {
		t.Fatalf("failed to create incident: %v", err)
	}

	// Update incident
	incident.Status = models.StatusWorkflowTriggered
	now := time.Now()
	incident.TriggeredAt = &now

	err = repo.Update(incident)
	if err != nil {
		t.Fatalf("failed to update incident: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(incident.ID)
	if err != nil {
		t.Fatalf("failed to retrieve incident: %v", err)
	}

	if retrieved.Status != models.StatusWorkflowTriggered {
		t.Errorf("expected status %s, got %s", models.StatusWorkflowTriggered, retrieved.Status)
	}

	if retrieved.TriggeredAt == nil {
		t.Error("expected TriggeredAt to be set")
	}
}
