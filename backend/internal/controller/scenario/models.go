package scenario

import (
	"encoding/json"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type scenarioRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateScenarioRequest struct {
	Title             string                    `json:"title"`
	Description       string                    `json:"description"`
	ExpandedContext   string                    `json:"expanded_context"`
	Difficulty        entity.ScenarioDifficulty `json:"difficulty"`
	Tags              []string                  `json:"tags"`
	SupportedRuntimes []string                  `json:"supported_runtimes"`
	SupportedChannels []string                  `json:"supported_channels"`
	TargetClusterIDs  []string                  `json:"target_cluster_ids"`
	Roles             json.RawMessage           `json:"roles"`
	Blocks            json.RawMessage           `json:"blocks"`
}

type scenarioSummaryResponse struct {
	ScenarioID        string                `json:"scenario_id"`
	Title             string                `json:"title"`
	Description       string                `json:"description"`
	Status            entity.ScenarioStatus `json:"status"`
	Difficulty        string                `json:"difficulty"`
	Tags              []string              `json:"tags"`
	SupportedRuntimes []string              `json:"supported_runtimes"`
	SupportedChannels []string              `json:"supported_channels"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

type scenarioListResponse struct {
	Scenarios []scenarioSummaryResponse `json:"scenarios"`
}

type scenarioResponse struct {
	ScenarioID        string                    `json:"scenario_id"`
	Title             string                    `json:"title"`
	Description       string                    `json:"description"`
	ExpandedContext   string                    `json:"expanded_context"`
	Status            entity.ScenarioStatus     `json:"status"`
	Difficulty        entity.ScenarioDifficulty `json:"difficulty"`
	Tags              []string                  `json:"tags"`
	SupportedRuntimes []string                  `json:"supported_runtimes"`
	SupportedChannels []string                  `json:"supported_channels"`
	TargetClusterIDs  []string                  `json:"target_cluster_ids"`
	Roles             []entity.ScenarioRole     `json:"roles"`
	Blocks            []entity.ScenarioBlock    `json:"blocks"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	PublishedAt       *time.Time                `json:"published_at,omitempty"`
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
	ScenarioID  string    `json:"scenario_id"`
	Status      string    `json:"status"`
	PublishedAt time.Time `json:"published_at"`
	Warnings    []any     `json:"warnings"`
}

type competencyClusterListResponse struct {
	Clusters []competencyClusterResponse `json:"clusters"`
}

type competencyClusterResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type semanticActionListResponse struct {
	Actions []semanticActionResponse `json:"actions"`
}

type semanticActionResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
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
