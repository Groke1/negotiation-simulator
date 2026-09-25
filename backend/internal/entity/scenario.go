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
	ScenarioDifficultyEasy   ScenarioDifficulty = "easy"
	ScenarioDifficultyMedium ScenarioDifficulty = "medium"
	ScenarioDifficultyHard   ScenarioDifficulty = "hard"
)

func (d ScenarioDifficulty) Valid() bool {
	switch d {
	case ScenarioDifficultyEasy, ScenarioDifficultyMedium, ScenarioDifficultyHard:
		return true
	default:
		return false
	}
}

type Scenario struct {
	ID                   string
	OrganizationID       *string
	CreatedBy            *string
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
	PublishedAt          *time.Time
}

type CreateScenario struct {
	UserID               string
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
	UserID            string
	Status            *ScenarioStatus
	Difficulty        *ScenarioDifficulty
	Tag               *string
	Runtime           *string
	Channel           *string
	Favorite          *bool
	AssignedByCurator *bool
	Completed         *bool
	OwnedByMe         bool
	Limit             int32
	Offset            int32
}

type ScenarioValidation struct {
	Valid         bool                      `json:"valid"`
	ErrorsCount   int                       `json:"errors_count"`
	WarningsCount int                       `json:"warnings_count"`
	Issues        []ScenarioValidationIssue `json:"issues"`
}

type ScenarioValidationIssue struct {
	Severity   string  `json:"severity"`
	Code       string  `json:"code"`
	Message    string  `json:"message"`
	EntityType string  `json:"entity_type"`
	EntityID   *string `json:"entity_id,omitempty"`
	Field      *string `json:"field,omitempty"`
}

type SemanticActionDefinition struct {
	Code        string
	Name        string
	Description string
}

type ScenarioPersonalStats struct {
	IsFavorite          bool
	IsAssignedByCurator bool
	IsCompleted         bool
	AttemptsCount       int
	BestAttempt         *BestTrainingAttempt
}

type BestTrainingAttempt struct {
	SessionID   string
	Score       int
	CompletedAt time.Time
}
