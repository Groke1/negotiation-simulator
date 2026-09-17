package entity

import (
	"encoding/json"
	"time"
)

type ScenarioStatus string

const (
	ScenarioStatusDraft     ScenarioStatus = "draft"
	ScenarioStatusPublished ScenarioStatus = "published"
	ScenarioStatusArchived  ScenarioStatus = "archived"
)

func (s ScenarioStatus) Valid() bool {
	switch s {
	case ScenarioStatusDraft, ScenarioStatusPublished, ScenarioStatusArchived:
		return true
	default:
		return false
	}
}

type ScenarioDifficulty string

const (
	ScenarioDifficultyL1 ScenarioDifficulty = "L1"
	ScenarioDifficultyL2 ScenarioDifficulty = "L2"
	ScenarioDifficultyL3 ScenarioDifficulty = "L3"
)

func (d ScenarioDifficulty) Valid() bool {
	switch d {
	case ScenarioDifficultyL1, ScenarioDifficultyL2, ScenarioDifficultyL3:
		return true
	default:
		return false
	}
}

type Scenario struct {
	ID                   string
	OrganizationID       *string
	CreatedBy            *string
	Version              string
	Title                string
	Status               ScenarioStatus
	ModeSupport          []string
	Context              json.RawMessage
	Roles                json.RawMessage
	Goals                json.RawMessage
	Interests            json.RawMessage
	Constraints          json.RawMessage
	Parameters           json.RawMessage
	PrivateInformation   json.RawMessage
	Difficulty           ScenarioDifficulty
	Competencies         []string
	ObservationRules     json.RawMessage
	CompletionConditions json.RawMessage
	Channels             []string
	Nodes                json.RawMessage
	Revision             int64
	ValidatedRevision    *int64
	ValidationResult     json.RawMessage
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type CreateScenario struct {
	UserID               string
	Version              string
	Title                string
	ModeSupport          []string
	Context              json.RawMessage
	Roles                json.RawMessage
	Goals                json.RawMessage
	Interests            json.RawMessage
	Constraints          json.RawMessage
	Parameters           json.RawMessage
	PrivateInformation   json.RawMessage
	Difficulty           ScenarioDifficulty
	Competencies         []string
	ObservationRules     json.RawMessage
	CompletionConditions json.RawMessage
	Channels             []string
	Nodes                json.RawMessage
}

type UpdateScenario struct {
	ScenarioID           string
	UserID               string
	Version              *string
	Title                *string
	ModeSupport          *[]string
	Context              *json.RawMessage
	Roles                *json.RawMessage
	Goals                *json.RawMessage
	Interests            *json.RawMessage
	Constraints          *json.RawMessage
	Parameters           *json.RawMessage
	PrivateInformation   *json.RawMessage
	Difficulty           *ScenarioDifficulty
	Competencies         *[]string
	ObservationRules     *json.RawMessage
	CompletionConditions *json.RawMessage
	Channels             *[]string
	Nodes                *json.RawMessage
}

type ListScenarios struct {
	UserID string
	Status *ScenarioStatus
	Limit  int32
	Offset int32
}

type ValidationStructure struct {
	HasGoal                 bool `json:"has_goal"`
	HasConflict             bool `json:"has_conflict"`
	HasRoles                bool `json:"has_roles"`
	HasCompletionConditions bool `json:"has_completion_conditions"`
}

type ValidationGraph struct {
	DeadEnds         []string   `json:"dead_ends"`
	UnreachableNodes []string   `json:"unreachable_nodes"`
	Cycles           [][]string `json:"cycles"`
}

type UnderCoveredIndicator struct {
	Indicator     string `json:"indicator"`
	ContextsFound int    `json:"contexts_found"`
	Required      int    `json:"required"`
}

type ValidationCoverage struct {
	UnderCoveredIndicators []UnderCoveredIndicator `json:"under_covered_indicators"`
}

type ScenarioValidation struct {
	Valid     bool                `json:"valid"`
	Structure ValidationStructure `json:"structure"`
	Graph     ValidationGraph     `json:"graph"`
	Coverage  ValidationCoverage  `json:"coverage"`
}

type ScenarioVersion struct {
	Version     string
	Snapshot    json.RawMessage
	PublishedAt time.Time
}
