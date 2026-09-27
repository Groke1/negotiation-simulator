package training

import (
	"encoding/json"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type startTrainingSessionRequest struct {
	ScenarioID     string                 `json:"scenario_id"`
	ScenarioRoleID string                 `json:"scenario_role_id"`
	Mode           entity.TrainingMode    `json:"mode"`
	Channel        entity.TrainingChannel `json:"channel"`
	AssignmentID   *string                `json:"assignment_id"`
}

type trainingSessionListResponse struct {
	Sessions []trainingSessionListItem `json:"sessions"`
}

type trainingSessionListItem struct {
	SessionID      string                 `json:"session_id"`
	ScenarioID     string                 `json:"scenario_id"`
	ScenarioTitle  string                 `json:"scenario_title"`
	ScenarioRoleID string                 `json:"scenario_role_id"`
	Mode           entity.TrainingMode    `json:"mode"`
	Channel        entity.TrainingChannel `json:"channel"`
	Status         entity.TrainingStatus  `json:"status"`
	Score          *int                   `json:"score"`
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at"`
}

type trainingPreparation struct {
	ScenarioID      string                       `json:"scenario_id"`
	Title           string                       `json:"title"`
	Description     string                       `json:"description"`
	ExpandedContext string                       `json:"expanded_context"`
	MyRole          scenarioRolePrivate          `json:"my_role"`
	PartnerRole     scenarioRoleVisible          `json:"partner_role"`
	Roles           []entity.PublicScenarioRole  `json:"roles"`
	Blocks          []entity.PublicScenarioBlock `json:"blocks"`
}

type scenarioRolePrivate struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Goal               string `json:"goal"`
	Interests          string `json:"interests"`
	Constraints        string `json:"constraints"`
	PrivateInformation string `json:"private_information"`
}

type scenarioRoleVisible struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Goal        string `json:"goal"`
	Interests   string `json:"interests"`
	Constraints string `json:"constraints"`
}

type deterministicCurrentNode struct {
	NodeID           string                         `json:"node_id"`
	BlockID          string                         `json:"block_id"`
	IncomingMessage  *partnerDeterministicMessage   `json:"incoming_message"`
	AvailableActions []deterministicAvailableAction `json:"available_actions"`
}

type partnerDeterministicMessage struct {
	MessageID string `json:"message_id"`
	Text      string `json:"text"`
}

type deterministicAvailableAction struct {
	ActionID string `json:"action_id"`
	Text     string `json:"text"`
}

type trainingMessage struct {
	MessageID           string          `json:"message_id"`
	Author              string          `json:"author"`
	Content             json.RawMessage `json:"content"`
	SemanticActionCode  *string         `json:"semantic_action_code"`
	SemanticAnnotations json.RawMessage `json:"semantic_annotations"`
	Classification      json.RawMessage `json:"classification"`
	CreatedAt           time.Time       `json:"created_at"`
}

type trainingSessionResponse struct {
	SessionID      string                      `json:"session_id"`
	ScenarioID     string                      `json:"scenario_id"`
	ScenarioTitle  string                      `json:"scenario_title"`
	ScenarioRoleID string                      `json:"scenario_role_id"`
	AssignmentID   *string                     `json:"assignment_id"`
	Mode           entity.TrainingMode         `json:"mode"`
	Channel        entity.TrainingChannel      `json:"channel"`
	Status         entity.TrainingStatus       `json:"status"`
	IsEnd          bool                        `json:"is_end"`
	Scenario       entity.PublicScenarioInfo   `json:"scenario"`
	Preparation    trainingPreparation         `json:"preparation"`
	CurrentBlock   *entity.PublicScenarioBlock `json:"current_block"`
	CurrentNode    *deterministicCurrentNode   `json:"current_node"`
	Messages       []trainingMessage           `json:"messages"`
	Score          *int                        `json:"score"`
	StartedAt      time.Time                   `json:"started_at"`
	CompletedAt    *time.Time                  `json:"completed_at"`
}

type trainingActionRequest struct {
	ActionID           string `json:"action_id"`
	SemanticActionCode string `json:"semantic_action_code"`
	ClientEventID      string `json:"client_event_id"`
}

type trainingActionResponse struct {
	EventID        string                    `json:"event_id"`
	Classification json.RawMessage           `json:"classification"`
	NextNode       *deterministicCurrentNode `json:"next_node"`
	Status         entity.TrainingStatus     `json:"status"`
	IsEnd          bool                      `json:"is_end"`
}

type trainingMessageRequest struct {
	MessageID           string          `json:"message_id"`
	Content             json.RawMessage `json:"content"`
	SemanticActionCode  *string         `json:"semantic_action_code"`
	SemanticAnnotations json.RawMessage `json:"semantic_annotations"`
}

type trainingMessageResponse struct {
	LearnerMessage trainingMessage       `json:"learner_message"`
	PartnerMessage trainingMessage       `json:"partner_message"`
	Status         entity.TrainingStatus `json:"status"`
	IsEnd          bool                  `json:"is_end"`
}

type classifyMessageRequest struct {
	SemanticActionCode string `json:"semantic_action_code"`
}

type classifyMessageResponse struct {
	MessageID                  string `json:"message_id"`
	SelectedSemanticActionCode string `json:"selected_semantic_action_code"`
	ExpectedSemanticActionCode string `json:"expected_semantic_action_code"`
	Correct                    bool   `json:"correct"`
}
