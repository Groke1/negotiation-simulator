package entity

import (
	"encoding/json"
	"time"
)

type AIGenerationOptions struct {
	Instructions string
}

type AIGenerationStatus string

const (
	AIGenerationStatusPending   AIGenerationStatus = "pending"
	AIGenerationStatusRunning   AIGenerationStatus = "running"
	AIGenerationStatusSucceeded AIGenerationStatus = "succeeded"
	AIGenerationStatusFailed    AIGenerationStatus = "failed"
)

type AIGenerationKind string

const (
	AIGenerationKindExpandContext    AIGenerationKind = "expand_context"
	AIGenerationKindGenerateDraft    AIGenerationKind = "generate_draft"
	AIGenerationKindGenerateBlock    AIGenerationKind = "generate_block"
	AIGenerationKindGenerateDialogue AIGenerationKind = "generate_dialogue"
)

type AIGenerationRequest struct {
	ID            string             `json:"request_id"`
	ScenarioID    string             `json:"scenario_id"`
	UserID        string             `json:"-"`
	BlockID       string             `json:"block_id,omitempty"`
	LearnerRoleID string             `json:"learner_role_id,omitempty"`
	Kind          AIGenerationKind   `json:"kind"`
	Status        AIGenerationStatus `json:"status"`
	Result        json.RawMessage    `json:"result,omitempty"`
	ErrorCode     string             `json:"error_code,omitempty"`
	ErrorMessage  string             `json:"error_message,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type ScenarioDraftProposal struct {
	ExpandedContext string
	Blocks          []ScenarioBlock
}

type BlockContentProposal struct {
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	ExpandedContext string               `json:"expanded_context"`
	LeadRoleID      string               `json:"lead_role_id"`
	LeadProbability int                  `json:"lead_probability"`
	RoleGoals       []BlockRoleGoal      `json:"role_goals"`
	Behavior        string               `json:"behavior"`
	RoleBehaviors   []BlockRoleBehavior  `json:"role_behaviors"`
	Result          *ScenarioEndResult   `json:"result,omitempty"`
	ActionRules     []SemanticActionRule `json:"action_rules"`
	Constraints     []string             `json:"constraints"`
}

type SemanticClassification struct {
	SemanticActionCode string  `json:"semantic_action_code"`
	Confidence         float64 `json:"confidence"`
	Reason             string  `json:"reason"`
}

type VirtualPartnerTurn struct {
	PartnerSemanticActionCode string                     `json:"partner_semantic_action_code"`
	Content                   map[string]any             `json:"content"`
	StateUpdates              VirtualPartnerStateUpdates `json:"state_updates"`
	Transition                VirtualPartnerTransition   `json:"transition"`
	Completion                VirtualPartnerCompletion   `json:"completion"`
}

type VirtualPartnerStateUpdates struct {
	RevealedFacts    []string       `json:"revealed_facts"`
	ParameterUpdates map[string]any `json:"parameter_updates"`
	ShortSummary     string         `json:"short_summary"`
}

type VirtualPartnerTransition struct {
	TransitionID *string `json:"transition_id"`
	Reason       string  `json:"reason"`
}

type VirtualPartnerCompletion struct {
	Completed bool    `json:"completed"`
	Reason    *string `json:"reason"`
}

type VirtualPartnerRuntimeTurn struct {
	LearnerSemanticActionCode string         `json:"learner_semantic_action_code"`
	LearnerConfidence         float64        `json:"learner_confidence"`
	PartnerSemanticActionCode string         `json:"partner_semantic_action_code"`
	Content                   map[string]any `json:"content"`
	Decision                  string         `json:"decision"`
	TransitionID              *string        `json:"transition_id"`
	PatienceDelta             int            `json:"patience_delta"`
}

type TrainingStateEvaluation struct {
	Decision          string                     `json:"decision"`
	TransitionID      *string                    `json:"transition_id"`
	TerminationReason *string                    `json:"termination_reason"`
	Reason            string                     `json:"reason"`
	StateUpdates      VirtualPartnerStateUpdates `json:"state_updates"`
}

type TrainingAssessment struct {
	Process            []TrainingProcessAssessment `json:"process"`
	Timeline           []TrainingTimelineItem      `json:"timeline"`
	MessageFeedback    []TrainingMessageFeedback   `json:"message_feedback"`
	Highlights         TrainingHighlights          `json:"highlights"`
	Strengths          []TrainingAnalysisNote      `json:"strengths"`
	Improvements       []TrainingAnalysisNote      `json:"improvements"`
	ClusterAssessments []TrainingClusterAssessment `json:"cluster_assessments"`
}

type TrainingProcessAssessment struct {
	Order               int      `json:"order"`
	Summary             string   `json:"summary"`
	SemanticActionCodes []string `json:"semantic_action_codes"`
}

type TrainingAnalysisNote struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TrainingTimelineItem struct {
	Order               int      `json:"order"`
	BlockID             string   `json:"block_id"`
	BlockName           string   `json:"block_name"`
	TransitionID        *string  `json:"transition_id"`
	FromBlockID         string   `json:"from_block_id"`
	ToBlockID           string   `json:"to_block_id"`
	MessageIDs          []string `json:"message_ids"`
	Summary             string   `json:"summary"`
	Outcome             string   `json:"outcome"`
	SemanticActionCodes []string `json:"semantic_action_codes"`
}

type TrainingMessageFeedback struct {
	MessageID          string   `json:"message_id"`
	BlockID            string   `json:"block_id"`
	SemanticActionCode string   `json:"semantic_action_code"`
	Score              int      `json:"score"`
	Level              string   `json:"level"`
	Summary            string   `json:"summary"`
	Strengths          []string `json:"strengths"`
	Improvements       []string `json:"improvements"`
	RelatedClusterIDs  []string `json:"related_cluster_ids"`
}

type TrainingHighlights struct {
	Strengths    []TrainingHighlight `json:"strengths"`
	Improvements []TrainingHighlight `json:"improvements"`
}

type TrainingHighlight struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	MessageIDs  []string `json:"message_ids"`
	ClusterIDs  []string `json:"cluster_ids"`
}

type TrainingClusterAssessment struct {
	ClusterID        string                    `json:"cluster_id"`
	Status           string                    `json:"status"`
	Level            int                       `json:"level"`
	EvidenceEpisodes []TrainingEvidenceEpisode `json:"evidence_episodes"`
	Summary          string                    `json:"summary"`
	Comment          string                    `json:"comment"`
	Recommendation   string                    `json:"recommendation"`
}

type TrainingEvidenceEpisode struct {
	MessageIDs          []string `json:"message_ids"`
	Summary             string   `json:"summary"`
	SemanticActionCodes []string `json:"semantic_action_codes"`
}
