package scenario

import (
	"encoding/json"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type scenarioRequest struct {
	Version              string                    `json:"version"`
	Title                string                    `json:"title"`
	ModeSupport          []string                  `json:"mode_support"`
	Context              json.RawMessage           `json:"context"`
	Roles                json.RawMessage           `json:"roles"`
	Goals                json.RawMessage           `json:"goals"`
	Interests            json.RawMessage           `json:"interests"`
	Constraints          json.RawMessage           `json:"constraints"`
	Parameters           json.RawMessage           `json:"parameters"`
	PrivateInformation   json.RawMessage           `json:"private_information"`
	Difficulty           entity.ScenarioDifficulty `json:"difficulty"`
	Competencies         []string                  `json:"competencies"`
	ObservationRules     json.RawMessage           `json:"observation_rules"`
	CompletionConditions json.RawMessage           `json:"completion_conditions"`
	Channels             []string                  `json:"channels"`
	Nodes                json.RawMessage           `json:"nodes"`
}

type updateScenarioRequest struct {
	Version              *string                    `json:"version"`
	Title                *string                    `json:"title"`
	ModeSupport          *[]string                  `json:"mode_support"`
	Context              *json.RawMessage           `json:"context"`
	Roles                *json.RawMessage           `json:"roles"`
	Goals                *json.RawMessage           `json:"goals"`
	Interests            *json.RawMessage           `json:"interests"`
	Constraints          *json.RawMessage           `json:"constraints"`
	Parameters           *json.RawMessage           `json:"parameters"`
	PrivateInformation   *json.RawMessage           `json:"private_information"`
	Difficulty           *entity.ScenarioDifficulty `json:"difficulty"`
	Competencies         *[]string                  `json:"competencies"`
	ObservationRules     *json.RawMessage           `json:"observation_rules"`
	CompletionConditions *json.RawMessage           `json:"completion_conditions"`
	Channels             *[]string                  `json:"channels"`
	Nodes                *json.RawMessage           `json:"nodes"`
}

type scenarioCreatedResponse struct {
	ID string `json:"id"`
}

type scenarioSummaryResponse struct {
	ScenarioID string                `json:"scenario_id"`
	Title      string                `json:"title"`
	Version    string                `json:"version"`
	Status     entity.ScenarioStatus `json:"status"`
	Difficulty string                `json:"difficulty"`
	Channels   []string              `json:"channels"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

type scenarioListResponse struct {
	Scenarios []scenarioSummaryResponse `json:"scenarios"`
}

type scenarioResponse struct {
	ScenarioID           string                    `json:"scenario_id"`
	Version              string                    `json:"version"`
	Title                string                    `json:"title"`
	ModeSupport          []string                  `json:"mode_support"`
	Context              json.RawMessage           `json:"context"`
	Roles                json.RawMessage           `json:"roles"`
	Goals                json.RawMessage           `json:"goals"`
	Interests            json.RawMessage           `json:"interests"`
	Constraints          json.RawMessage           `json:"constraints"`
	Parameters           json.RawMessage           `json:"parameters"`
	PrivateInformation   json.RawMessage           `json:"private_information"`
	Difficulty           entity.ScenarioDifficulty `json:"difficulty"`
	Competencies         []string                  `json:"competencies"`
	ObservationRules     json.RawMessage           `json:"observation_rules"`
	CompletionConditions json.RawMessage           `json:"completion_conditions"`
	Channels             []string                  `json:"channels"`
	Nodes                json.RawMessage           `json:"nodes"`
}

type generateScenarioRequest struct {
	Prompt string `json:"prompt"`
}

type generateScenarioResponse struct {
	ScenarioDraft  scenarioResponse `json:"scenario_draft"`
	AIGenerationID string           `json:"ai_generation_id"`
	Provider       string           `json:"provider"`
}

type publishScenarioResponse struct {
	Version     string    `json:"version"`
	PublishedAt time.Time `json:"published_at"`
}

type scenarioVersionSummaryResponse struct {
	Version     string    `json:"version"`
	PublishedAt time.Time `json:"published_at"`
}

type scenarioVersionListResponse struct {
	Versions []scenarioVersionSummaryResponse `json:"versions"`
}

type scenarioVersionResponse struct {
	Version     string          `json:"version"`
	Snapshot    json.RawMessage `json:"snapshot"`
	PublishedAt time.Time       `json:"published_at"`
}
