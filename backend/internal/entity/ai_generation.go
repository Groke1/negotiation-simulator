package entity

type AIGenerationOptions struct {
	Instructions string
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

type TrainingClusterAssessment struct {
	ClusterID        string                    `json:"cluster_id"`
	Status           string                    `json:"status"`
	Level            int                       `json:"level"`
	EvidenceEpisodes []TrainingEvidenceEpisode `json:"evidence_episodes"`
	Comment          string                    `json:"comment"`
}

type TrainingEvidenceEpisode struct {
	Summary             string   `json:"summary"`
	SemanticActionCodes []string `json:"semantic_action_codes"`
}
