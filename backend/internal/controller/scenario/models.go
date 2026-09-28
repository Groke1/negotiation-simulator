package scenario

import (
	"encoding/json"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type scenarioRequest struct {
	Title string `json:"title"`
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
	ScenarioID          string                `json:"scenario_id"`
	Title               string                `json:"title"`
	Description         string                `json:"description"`
	Status              entity.ScenarioStatus `json:"status"`
	Difficulty          string                `json:"difficulty"`
	Tags                []string              `json:"tags"`
	SupportedRuntimes   []string              `json:"supported_runtimes"`
	SupportedChannels   []string              `json:"supported_channels"`
	CreatedBy           string                `json:"created_by_methodist_id"`
	IsOwnedByMe         bool                  `json:"is_owned_by_me"`
	UpdatedAt           time.Time             `json:"updated_at"`
	IsFavorite          bool                  `json:"is_favorite"`
	IsAssignedByCurator bool                  `json:"is_assigned_by_curator"`
	IsCompleted         bool                  `json:"is_completed"`
	AttemptsCount       int                   `json:"attempts_count"`
	BestAttempt         *bestTrainingAttempt  `json:"best_attempt"`
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
	Blocks            []scenarioBlockResponse   `json:"blocks"`
	CreatedBy         string                    `json:"created_by_methodist_id"`
	IsOwnedByMe       bool                      `json:"is_owned_by_me"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	PublishedAt       *time.Time                `json:"published_at,omitempty"`
}

type publishScenarioResponse struct {
	ScenarioID  string                           `json:"scenario_id"`
	Status      string                           `json:"status"`
	PublishedAt time.Time                        `json:"published_at"`
	Warnings    []entity.ScenarioValidationIssue `json:"warnings"`
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

type bestTrainingAttempt struct {
	SessionID   string    `json:"session_id"`
	Score       int       `json:"score"`
	CompletedAt time.Time `json:"completed_at"`
}

type aiGenerationOptionsRequest struct {
	Instructions string `json:"instructions"`
}

type generateDialogueRequest struct {
	LearnerRoleID string `json:"learner_role_id"`
	Instructions  string `json:"instructions"`
}

type expandContextResponse struct {
	ExpandedContext string `json:"expanded_context"`
}

type scenarioDraftProposalResponse struct {
	Blocks []scenarioBlockResponse `json:"blocks"`
}

type scenarioBlockResponse struct {
	ID                    string                        `json:"id"`
	Type                  string                        `json:"type"`
	Name                  string                        `json:"name"`
	Description           string                        `json:"description"`
	ExpandedContext       string                        `json:"expanded_context,omitempty"`
	LeadRoleID            string                        `json:"lead_role_id,omitempty"`
	RoleGoals             []entity.BlockRoleGoal        `json:"role_goals"`
	RoleBehaviors         []entity.BlockRoleBehavior    `json:"role_behaviors"`
	Result                *entity.ScenarioEndResult     `json:"result,omitempty"`
	DeterministicDialogue *entity.DeterministicDialogue `json:"deterministic_dialogue,omitempty"`
	Transitions           []entity.ScenarioTransition   `json:"transitions"`
}

type generateScenarioDraftResponse struct {
	Proposal   scenarioDraftProposalResponse `json:"proposal"`
	Validation *entity.ScenarioValidation    `json:"validation"`
}

type generateBlockResponse struct {
	BlockID    string                      `json:"block_id"`
	Proposal   entity.BlockContentProposal `json:"proposal"`
	Validation *entity.ScenarioValidation  `json:"validation"`
}

type generateDialogueResponse struct {
	BlockID       string                       `json:"block_id"`
	LearnerRoleID string                       `json:"learner_role_id"`
	Proposal      entity.DeterministicDialogue `json:"proposal"`
	Validation    *entity.ScenarioValidation   `json:"validation"`
}
