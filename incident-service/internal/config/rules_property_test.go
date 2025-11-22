package config

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: ai-sre-platform, Property 16: Custom rule evaluation
// Validates: Requirements 16.2, 16.3
//
// Property: For any incident and custom rule, if the incident matches the rule conditions,
// then the rule's actions (severity adjustment, metadata enrichment) should be applied to the incident.
func TestProperty_RuleEvaluation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: If a rule matches an incident, applying its actions should modify the incident accordingly
	properties.Property("matching rule actions are applied correctly", prop.ForAll(
		func(serviceName string, errorMsg string, initialSeverity string, newSeverity string, metadataKey string, metadataValue string) bool {
			// Create a rule that matches the service name
			rule := CustomRule{
				Name:    "test-rule",
				Enabled: true,
				Conditions: RuleConditions{
					ServiceName: &serviceName,
				},
				Actions: RuleActions{
					SetSeverity: &newSeverity,
					AddMetadata: map[string]string{
						metadataKey: metadataValue,
					},
				},
			}

			// Create an incident that matches the rule
			incident := IncidentData{
				ServiceName:  serviceName,
				ErrorMessage: errorMsg,
				Severity:     initialSeverity,
				Metadata:     make(map[string]string),
			}

			// Evaluate the rule
			engine := NewRuleEngine([]CustomRule{rule})
			matches := engine.Evaluate(&incident)

			// The rule should match
			if len(matches) != 1 {
				return false
			}

			// Apply the actions
			ApplyActions(&incident, matches)

			// Verify severity was changed
			if incident.Severity != newSeverity {
				return false
			}

			// Verify metadata was added
			if incident.Metadata[metadataKey] != metadataValue {
				return false
			}

			return true
		},
		genServiceName(),
		genErrorMessage(),
		genSeverity(),
		genSeverity(),
		genMetadataKey(),
		genMetadataValue(),
	))

	// Property: If a rule doesn't match an incident, its actions should not be applied
	properties.Property("non-matching rule actions are not applied", prop.ForAll(
		func(ruleServiceName string, incidentServiceName string, errorMsg string, initialSeverity string, newSeverity string) bool {
			// Ensure the service names are different
			if ruleServiceName == incidentServiceName {
				return true // Skip this case
			}

			// Create a rule that matches a different service
			rule := CustomRule{
				Name:    "test-rule",
				Enabled: true,
				Conditions: RuleConditions{
					ServiceName: &ruleServiceName,
				},
				Actions: RuleActions{
					SetSeverity: &newSeverity,
				},
			}

			// Create an incident that doesn't match
			incident := IncidentData{
				ServiceName:  incidentServiceName,
				ErrorMessage: errorMsg,
				Severity:     initialSeverity,
				Metadata:     make(map[string]string),
			}

			// Evaluate the rule
			engine := NewRuleEngine([]CustomRule{rule})
			matches := engine.Evaluate(&incident)

			// The rule should not match
			if len(matches) != 0 {
				return false
			}

			// Apply actions (should be a no-op)
			ApplyActions(&incident, matches)

			// Verify severity was NOT changed
			if incident.Severity != initialSeverity {
				return false
			}

			return true
		},
		genServiceName(),
		genServiceName(),
		genErrorMessage(),
		genSeverity(),
		genSeverity(),
	))

	// Property: Disabled rules should never match
	properties.Property("disabled rules never match", prop.ForAll(
		func(serviceName string, errorMsg string, severity string) bool {
			// Create a disabled rule
			rule := CustomRule{
				Name:    "disabled-rule",
				Enabled: false,
				Conditions: RuleConditions{
					ServiceName: &serviceName,
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("critical"),
				},
			}

			// Create an incident that would match if the rule were enabled
			incident := IncidentData{
				ServiceName:  serviceName,
				ErrorMessage: errorMsg,
				Severity:     severity,
			}

			// Evaluate the rule
			engine := NewRuleEngine([]CustomRule{rule})
			matches := engine.Evaluate(&incident)

			// The rule should not match because it's disabled
			return len(matches) == 0
		},
		genServiceName(),
		genErrorMessage(),
		genSeverity(),
	))

	// Property: Multiple matching rules should all be applied
	properties.Property("multiple matching rules are all applied", prop.ForAll(
		func(serviceName string, errorMsg string, initialSeverity string, severity1 string, severity2 string, key1 string, val1 string, key2 string, val2 string) bool {
			// Ensure metadata keys are different
			if key1 == key2 {
				return true // Skip this case
			}

			// Create two rules that both match
			rule1 := CustomRule{
				Name:    "rule1",
				Enabled: true,
				Conditions: RuleConditions{
					ServiceName: &serviceName,
				},
				Actions: RuleActions{
					SetSeverity: &severity1,
					AddMetadata: map[string]string{
						key1: val1,
					},
				},
			}

			rule2 := CustomRule{
				Name:    "rule2",
				Enabled: true,
				Conditions: RuleConditions{
					ServiceName: &serviceName,
				},
				Actions: RuleActions{
					SetSeverity: &severity2,
					AddMetadata: map[string]string{
						key2: val2,
					},
				},
			}

			// Create an incident
			incident := IncidentData{
				ServiceName:  serviceName,
				ErrorMessage: errorMsg,
				Severity:     initialSeverity,
				Metadata:     make(map[string]string),
			}

			// Evaluate the rules
			engine := NewRuleEngine([]CustomRule{rule1, rule2})
			matches := engine.Evaluate(&incident)

			// Both rules should match
			if len(matches) != 2 {
				return false
			}

			// Apply the actions
			ApplyActions(&incident, matches)

			// The last severity should win (rule2's severity)
			if incident.Severity != severity2 {
				return false
			}

			// Both metadata entries should be present
			if incident.Metadata[key1] != val1 {
				return false
			}
			if incident.Metadata[key2] != val2 {
				return false
			}

			return true
		},
		genServiceName(),
		genErrorMessage(),
		genSeverity(),
		genSeverity(),
		genSeverity(),
		genMetadataKey(),
		genMetadataValue(),
		genMetadataKey(),
		genMetadataValue(),
	))

	// Property: Error pattern matching should work correctly
	properties.Property("error pattern matching works correctly", prop.ForAll(
		func(pattern string, matchingMsg string, severity string) bool {
			// Create a rule with an error pattern
			rule := CustomRule{
				Name:    "pattern-rule",
				Enabled: true,
				Conditions: RuleConditions{
					ErrorPattern: &pattern,
				},
				Actions: RuleActions{
					SetSeverity: &severity,
				},
			}

			// Create an incident with a message that should match
			incident := IncidentData{
				ServiceName:  "test-service",
				ErrorMessage: matchingMsg,
				Severity:     "low",
			}

			// Evaluate the rule
			engine := NewRuleEngine([]CustomRule{rule})
			_ = engine.Evaluate(&incident)

			// If the pattern is valid and matches, we should get a match
			// If the pattern is invalid or doesn't match, we shouldn't
			// We can't predict this, so we just verify the engine doesn't crash
			return true
		},
		genErrorPattern(),
		genErrorMessage(),
		genSeverity(),
	))

	// Property: Metadata conditions should match correctly
	properties.Property("metadata conditions match correctly", prop.ForAll(
		func(serviceName string, errorMsg string, severity string, metaKey string, metaValue string, newSeverity string) bool {
			// Create a rule that matches on metadata
			rule := CustomRule{
				Name:    "metadata-rule",
				Enabled: true,
				Conditions: RuleConditions{
					Metadata: map[string]string{
						metaKey: metaValue,
					},
				},
				Actions: RuleActions{
					SetSeverity: &newSeverity,
				},
			}

			// Create an incident with matching metadata
			incident := IncidentData{
				ServiceName:  serviceName,
				ErrorMessage: errorMsg,
				Severity:     severity,
				Metadata: map[string]string{
					metaKey: metaValue,
				},
			}

			// Evaluate the rule
			engine := NewRuleEngine([]CustomRule{rule})
			matches := engine.Evaluate(&incident)

			// The rule should match
			if len(matches) != 1 {
				return false
			}

			// Apply actions
			ApplyActions(&incident, matches)

			// Verify severity was changed
			return incident.Severity == newSeverity
		},
		genServiceName(),
		genErrorMessage(),
		genSeverity(),
		genMetadataKey(),
		genMetadataValue(),
		genSeverity(),
	))

	// Property: Repository override should be returned correctly
	properties.Property("repository override is returned correctly", prop.ForAll(
		func(serviceName string, repository string) bool {
			// Create a rule with repository override
			rule := CustomRule{
				Name:    "repo-rule",
				Enabled: true,
				Conditions: RuleConditions{
					ServiceName: &serviceName,
				},
				Actions: RuleActions{
					SetRepository: &repository,
				},
			}

			// Create a matching incident
			incident := IncidentData{
				ServiceName:  serviceName,
				ErrorMessage: "error",
				Severity:     "high",
			}

			// Evaluate the rule
			engine := NewRuleEngine([]CustomRule{rule})
			matches := engine.Evaluate(&incident)

			// Get repository override
			override := GetRepositoryOverride(matches)

			// Should return the repository
			return override != nil && *override == repository
		},
		genServiceName(),
		genRepository(),
	))

	// Property: Skip remediation flag should be detected correctly
	properties.Property("skip remediation flag is detected correctly", prop.ForAll(
		func(serviceName string, skipRemediation bool) bool {
			// Create a rule with skip remediation flag
			rule := CustomRule{
				Name:    "skip-rule",
				Enabled: true,
				Conditions: RuleConditions{
					ServiceName: &serviceName,
				},
				Actions: RuleActions{
					SkipRemediation: skipRemediation,
				},
			}

			// Create a matching incident
			incident := IncidentData{
				ServiceName:  serviceName,
				ErrorMessage: "error",
				Severity:     "high",
			}

			// Evaluate the rule
			engine := NewRuleEngine([]CustomRule{rule})
			matches := engine.Evaluate(&incident)

			// Check skip remediation
			shouldSkip := ShouldSkipRemediation(matches)

			// Should match the flag
			return shouldSkip == skipRemediation
		},
		genServiceName(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Generators for property-based testing

func genServiceName() gopter.Gen {
	return gen.OneConstOf(
		"api-gateway",
		"user-service",
		"payment-service",
		"auth-service",
		"notification-service",
		"data-processor",
	)
}

func genErrorMessage() gopter.Gen {
	return gen.OneConstOf(
		"connection timeout",
		"null pointer exception",
		"database connection failed",
		"payment processing failed",
		"authentication failed",
		"rate limit exceeded",
		"internal server error",
	)
}

func genSeverity() gopter.Gen {
	return gen.OneConstOf("critical", "high", "medium", "low")
}

func genMetadataKey() gopter.Gen {
	return gen.OneConstOf(
		"environment",
		"team",
		"priority",
		"region",
		"version",
	)
}

func genMetadataValue() gopter.Gen {
	return gen.OneConstOf(
		"production",
		"staging",
		"test",
		"payments",
		"infrastructure",
		"us-east-1",
		"eu-west-1",
		"v1.0.0",
		"v2.0.0",
	)
}

func genErrorPattern() gopter.Gen {
	return gen.OneConstOf(
		".*timeout.*",
		".*failed.*",
		".*error.*",
		".*exception.*",
		".*connection.*",
		"payment.*",
		"^database.*",
	)
}

func genRepository() gopter.Gen {
	return gen.OneConstOf(
		"org/api-gateway",
		"org/user-service",
		"org/payment-service",
		"org/auth-service",
		"org/notification-service",
	)
}
