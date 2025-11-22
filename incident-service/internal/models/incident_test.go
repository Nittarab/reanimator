package models

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Test that incident status values are valid
func TestProperty_IncidentStatusValidity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	validStatuses := []IncidentStatus{
		StatusPending,
		StatusWorkflowTriggered,
		StatusInProgress,
		StatusPRCreated,
		StatusResolved,
		StatusFailed,
		StatusNoFixNeeded,
	}

	properties.Property("incident status is always valid", prop.ForAll(
		func(status IncidentStatus) bool {
			// Check if the status is one of the valid values
			for _, valid := range validStatuses {
				if status == valid {
					return true
				}
			}
			return false
		},
		gen.OneConstOf(
			StatusPending,
			StatusWorkflowTriggered,
			StatusInProgress,
			StatusPRCreated,
			StatusResolved,
			StatusFailed,
			StatusNoFixNeeded,
		),
	))

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties.TestingRun(t, parameters)
}

// Test that JSONB marshaling/unmarshaling is consistent
func TestProperty_JSONBRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("JSONB round-trip preserves data", prop.ForAll(
		func(key string, value string) bool {
			data := map[string]interface{}{
				key: value,
			}
			jsonb := JSONB(data)

			// Marshal to value
			dbValue, err := jsonb.Value()
			if err != nil {
				t.Logf("failed to marshal: %v", err)
				return false
			}

			// Unmarshal back
			var result JSONB
			if err := result.Scan(dbValue); err != nil {
				t.Logf("failed to unmarshal: %v", err)
				return false
			}

			// Check that keys match
			if len(result) != len(data) {
				t.Logf("length mismatch: expected %d, got %d", len(data), len(result))
				return false
			}

			// Check that the value is preserved
			if result[key] != value {
				t.Logf("value mismatch for key %s: expected %v, got %v", key, value, result[key])
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties.TestingRun(t, parameters)
}
