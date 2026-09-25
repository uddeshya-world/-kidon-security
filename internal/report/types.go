package report

import "time"

// MissionData holds the state of the current assessment.
// TitanClass is the class of the agent (Titan) under assessment.
type MissionData struct {
	Timestamp     time.Time      `json:"timestamp"`
	TitanClass    string         `json:"titan_class"`
	StaticIssues  []Issue        `json:"static_issues"`
	AttackResults []AttackResult `json:"attack_results"`
	RuntimeAlerts []string       `json:"runtime_alerts"`
}

// Issue represents a static analysis finding
type Issue struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

// AttackResult represents an offensive simulation result
type AttackResult struct {
	Type     string `json:"type"`
	Prompt   string `json:"prompt"`
	Success  bool   `json:"success"`
	Severity string `json:"severity"`
}
