package config

import (
	"regexp"
)

// RuleEngine evaluates custom rules against incidents
type RuleEngine struct {
	rules []*CustomRule
}

// NewRuleEngine creates a new rule engine with the given rules
func NewRuleEngine(rules []CustomRule) *RuleEngine {
	// Filter to only enabled rules
	enabledRules := make([]*CustomRule, 0, len(rules))
	for i := range rules {
		if rules[i].Enabled {
			enabledRules = append(enabledRules, &rules[i])
		}
	}

	return &RuleEngine{
		rules: enabledRules,
	}
}

// IncidentData represents the data needed to evaluate rules
type IncidentData struct {
	ServiceName  string
	ErrorMessage string
	Severity     string
	Provider     string
	Metadata     map[string]string
}

// RuleMatch represents a rule that matched an incident
type RuleMatch struct {
	Rule    *CustomRule
	Actions RuleActions
}

// Evaluate evaluates all rules against the incident and returns matches
func (e *RuleEngine) Evaluate(incident *IncidentData) []RuleMatch {
	matches := make([]RuleMatch, 0)

	for _, rule := range e.rules {
		if e.matchesRule(incident, rule) {
			matches = append(matches, RuleMatch{
				Rule:    rule,
				Actions: rule.Actions,
			})
		}
	}

	return matches
}

// matchesRule checks if an incident matches a rule's conditions
func (e *RuleEngine) matchesRule(incident *IncidentData, rule *CustomRule) bool {
	conditions := &rule.Conditions

	// Check service name
	if conditions.ServiceName != nil {
		if incident.ServiceName != *conditions.ServiceName {
			return false
		}
	}

	// Check error pattern
	if conditions.ErrorPattern != nil && *conditions.ErrorPattern != "" {
		matched, err := regexp.MatchString(*conditions.ErrorPattern, incident.ErrorMessage)
		if err != nil || !matched {
			return false
		}
	}

	// Check severity
	if conditions.Severity != nil {
		if incident.Severity != *conditions.Severity {
			return false
		}
	}

	// Check provider
	if conditions.Provider != nil {
		if incident.Provider != *conditions.Provider {
			return false
		}
	}

	// Check metadata
	for key, value := range conditions.Metadata {
		incidentValue, exists := incident.Metadata[key]
		if !exists || incidentValue != value {
			return false
		}
	}

	return true
}

// ApplyActions applies rule actions to incident data
func ApplyActions(incident *IncidentData, matches []RuleMatch) {
	for _, match := range matches {
		actions := match.Actions

		// Apply severity change
		if actions.SetSeverity != nil {
			incident.Severity = *actions.SetSeverity
		}

		// Add metadata
		if incident.Metadata == nil {
			incident.Metadata = make(map[string]string)
		}
		for key, value := range actions.AddMetadata {
			incident.Metadata[key] = value
		}
	}
}

// ShouldSkipRemediation checks if any matching rule indicates remediation should be skipped
func ShouldSkipRemediation(matches []RuleMatch) bool {
	for _, match := range matches {
		if match.Actions.SkipRemediation {
			return true
		}
	}
	return false
}

// GetRepositoryOverride returns the repository override from the first matching rule that specifies one
func GetRepositoryOverride(matches []RuleMatch) *string {
	for _, match := range matches {
		if match.Actions.SetRepository != nil {
			return match.Actions.SetRepository
		}
	}
	return nil
}
