package config

import (
	"testing"
)

func TestRuleEngine_Evaluate(t *testing.T) {
	tests := []struct {
		name          string
		rules         []CustomRule
		incident      IncidentData
		expectMatches int
	}{
		{
			name: "service name match",
			rules: []CustomRule{
				{
					Name:    "test-rule",
					Enabled: true,
					Conditions: RuleConditions{
						ServiceName: stringPtr("payment-service"),
					},
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			incident: IncidentData{
				ServiceName:  "payment-service",
				ErrorMessage: "payment failed",
				Severity:     "high",
			},
			expectMatches: 1,
		},
		{
			name: "error pattern match",
			rules: []CustomRule{
				{
					Name:    "pattern-rule",
					Enabled: true,
					Conditions: RuleConditions{
						ErrorPattern: stringPtr(".*payment.*failed.*"),
					},
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			incident: IncidentData{
				ServiceName:  "payment-service",
				ErrorMessage: "payment processing failed",
				Severity:     "high",
			},
			expectMatches: 1,
		},
		{
			name: "no match - wrong service",
			rules: []CustomRule{
				{
					Name:    "test-rule",
					Enabled: true,
					Conditions: RuleConditions{
						ServiceName: stringPtr("payment-service"),
					},
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			incident: IncidentData{
				ServiceName:  "user-service",
				ErrorMessage: "error",
				Severity:     "high",
			},
			expectMatches: 0,
		},
		{
			name: "disabled rule should not match",
			rules: []CustomRule{
				{
					Name:    "disabled-rule",
					Enabled: false,
					Conditions: RuleConditions{
						ServiceName: stringPtr("payment-service"),
					},
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			incident: IncidentData{
				ServiceName:  "payment-service",
				ErrorMessage: "error",
				Severity:     "high",
			},
			expectMatches: 0,
		},
		{
			name: "metadata match",
			rules: []CustomRule{
				{
					Name:    "metadata-rule",
					Enabled: true,
					Conditions: RuleConditions{
						Metadata: map[string]string{
							"environment": "test",
						},
					},
					Actions: RuleActions{
						SkipRemediation: true,
					},
				},
			},
			incident: IncidentData{
				ServiceName:  "test-service",
				ErrorMessage: "error",
				Severity:     "high",
				Metadata: map[string]string{
					"environment": "test",
				},
			},
			expectMatches: 1,
		},
		{
			name: "multiple conditions - all must match",
			rules: []CustomRule{
				{
					Name:    "multi-condition-rule",
					Enabled: true,
					Conditions: RuleConditions{
						ServiceName:  stringPtr("payment-service"),
						ErrorPattern: stringPtr(".*timeout.*"),
						Severity:     stringPtr("high"),
					},
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			incident: IncidentData{
				ServiceName:  "payment-service",
				ErrorMessage: "connection timeout",
				Severity:     "high",
			},
			expectMatches: 1,
		},
		{
			name: "multiple conditions - one fails",
			rules: []CustomRule{
				{
					Name:    "multi-condition-rule",
					Enabled: true,
					Conditions: RuleConditions{
						ServiceName:  stringPtr("payment-service"),
						ErrorPattern: stringPtr(".*timeout.*"),
						Severity:     stringPtr("critical"),
					},
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			incident: IncidentData{
				ServiceName:  "payment-service",
				ErrorMessage: "connection timeout",
				Severity:     "high", // Doesn't match critical
			},
			expectMatches: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewRuleEngine(tt.rules)
			matches := engine.Evaluate(&tt.incident)

			if len(matches) != tt.expectMatches {
				t.Errorf("Evaluate() returned %d matches, want %d", len(matches), tt.expectMatches)
			}
		})
	}
}

func TestApplyActions(t *testing.T) {
	tests := []struct {
		name             string
		incident         IncidentData
		matches          []RuleMatch
		expectedSeverity string
		expectedMetadata map[string]string
	}{
		{
			name: "apply severity change",
			incident: IncidentData{
				ServiceName: "test-service",
				Severity:    "high",
			},
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			expectedSeverity: "critical",
		},
		{
			name: "add metadata",
			incident: IncidentData{
				ServiceName: "test-service",
				Severity:    "high",
				Metadata:    map[string]string{},
			},
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						AddMetadata: map[string]string{
							"team":     "payments",
							"priority": "high",
						},
					},
				},
			},
			expectedSeverity: "high",
			expectedMetadata: map[string]string{
				"team":     "payments",
				"priority": "high",
			},
		},
		{
			name: "multiple actions",
			incident: IncidentData{
				ServiceName: "test-service",
				Severity:    "medium",
			},
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
						AddMetadata: map[string]string{
							"escalated": "true",
						},
					},
				},
			},
			expectedSeverity: "critical",
			expectedMetadata: map[string]string{
				"escalated": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ApplyActions(&tt.incident, tt.matches)

			if tt.incident.Severity != tt.expectedSeverity {
				t.Errorf("Severity = %s, want %s", tt.incident.Severity, tt.expectedSeverity)
			}

			if tt.expectedMetadata != nil {
				for key, expectedValue := range tt.expectedMetadata {
					actualValue, exists := tt.incident.Metadata[key]
					if !exists {
						t.Errorf("Metadata key %s not found", key)
					} else if actualValue != expectedValue {
						t.Errorf("Metadata[%s] = %s, want %s", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestShouldSkipRemediation(t *testing.T) {
	tests := []struct {
		name     string
		matches  []RuleMatch
		expected bool
	}{
		{
			name: "skip remediation",
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						SkipRemediation: true,
					},
				},
			},
			expected: true,
		},
		{
			name: "don't skip remediation",
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			expected: false,
		},
		{
			name: "one rule says skip",
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
				{
					Actions: RuleActions{
						SkipRemediation: true,
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldSkipRemediation(tt.matches)
			if result != tt.expected {
				t.Errorf("ShouldSkipRemediation() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetRepositoryOverride(t *testing.T) {
	tests := []struct {
		name     string
		matches  []RuleMatch
		expected *string
	}{
		{
			name: "has repository override",
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						SetRepository: stringPtr("org/custom-repo"),
					},
				},
			},
			expected: stringPtr("org/custom-repo"),
		},
		{
			name: "no repository override",
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						SetSeverity: stringPtr("critical"),
					},
				},
			},
			expected: nil,
		},
		{
			name: "first override wins",
			matches: []RuleMatch{
				{
					Actions: RuleActions{
						SetRepository: stringPtr("org/first-repo"),
					},
				},
				{
					Actions: RuleActions{
						SetRepository: stringPtr("org/second-repo"),
					},
				},
			},
			expected: stringPtr("org/first-repo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRepositoryOverride(tt.matches)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetRepositoryOverride() = %v, want nil", *result)
				}
			} else {
				if result == nil {
					t.Errorf("GetRepositoryOverride() = nil, want %s", *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("GetRepositoryOverride() = %s, want %s", *result, *tt.expected)
				}
			}
		})
	}
}
