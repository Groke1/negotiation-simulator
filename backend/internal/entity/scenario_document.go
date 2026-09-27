package entity

import (
	"encoding/json"
	"strings"
	"time"
)

type ScenarioDocument struct {
	ScenarioID        string             `json:"scenario_id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	ExpandedContext   string             `json:"expanded_context,omitempty"`
	Status            ScenarioStatus     `json:"status"`
	Difficulty        ScenarioDifficulty `json:"difficulty"`
	Tags              []string           `json:"tags"`
	SupportedRuntimes []string           `json:"supported_runtimes"`
	SupportedChannels []string           `json:"supported_channels"`
	TargetClusterIDs  []string           `json:"target_cluster_ids"`
	Roles             []ScenarioRole     `json:"roles"`
	Blocks            []ScenarioBlock    `json:"blocks"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	PublishedAt       *time.Time         `json:"published_at,omitempty"`
}

type ScenarioContextDocument struct {
	Description     string   `json:"description"`
	ExpandedContext string   `json:"expanded_context"`
	Tags            []string `json:"tags"`
}

type ScenarioRole struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Goal               string `json:"goal"`
	Interests          string `json:"interests"`
	Constraints        string `json:"constraints"`
	PrivateInformation string `json:"private_information"`
}

type PublicScenarioInfo struct {
	ScenarioID      string                `json:"scenario_id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	ExpandedContext string                `json:"expanded_context"`
	Roles           []PublicScenarioRole  `json:"roles"`
	Blocks          []PublicScenarioBlock `json:"blocks"`
}

type PublicScenarioRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Goal        string `json:"goal"`
	Interests   string `json:"interests"`
	Constraints string `json:"constraints"`
}

type PublicScenarioBlock struct {
	ID          string                     `json:"id"`
	Type        string                     `json:"type"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Transitions []PublicScenarioTransition `json:"transitions"`
}

type PublicScenarioTransition struct {
	ID            string `json:"id"`
	TargetBlockID string `json:"target_block_id"`
}

type ScenarioBlock struct {
	ID                    string                 `json:"id"`
	Type                  string                 `json:"type"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	ExpandedContext       string                 `json:"expanded_context"`
	LeadRoleID            string                 `json:"lead_role_id"`
	LeadProbability       int                    `json:"lead_probability"`
	RoleGoals             []BlockRoleGoal        `json:"role_goals"`
	Behavior              string                 `json:"behavior"`
	ActionRules           []SemanticActionRule   `json:"action_rules"`
	Constraints           []string               `json:"constraints"`
	Result                *ScenarioEndResult     `json:"result,omitempty"`
	DeterministicDialogue *DeterministicDialogue `json:"deterministic_dialogue,omitempty"`
	Transitions           []ScenarioTransition   `json:"transitions"`
}

type ScenarioEndResult struct {
	Summary     string   `json:"summary"`
	GoalsMet    []string `json:"goals_met"`
	GoalsMissed []string `json:"goals_missed"`
}

type BlockRoleGoal struct {
	RoleID string `json:"role_id"`
	Goal   string `json:"goal"`
}

type SemanticActionRule struct {
	SemanticActionCode string `json:"semantic_action_code"`
	Reaction           string `json:"reaction"`
}

type ScenarioTransition struct {
	ID              string `json:"id"`
	TargetBlockID   string `json:"target_block_id"`
	Condition       string `json:"condition"`
	Note            string `json:"note"`
	FeedbackComment string `json:"feedback_comment"`
	Weight          int    `json:"weight"`
	Forbidden       bool   `json:"forbidden"`
}

type DeterministicDialogue struct {
	LearnerRoleID string                      `json:"learner_role_id"`
	StartNodeID   string                      `json:"start_node_id"`
	Nodes         []DeterministicDialogueNode `json:"nodes"`
}

type DeterministicDialogueNode struct {
	ID                string                          `json:"id"`
	IncomingMessage   *DeterministicMessage           `json:"incoming_message"`
	AssessmentTargets []DeterministicAssessmentTarget `json:"assessment_targets"`
	ReplyOptions      []DeterministicReplyOption      `json:"reply_options"`
}

type DeterministicAssessmentTarget struct {
	ClusterID string  `json:"cluster_id"`
	Weight    float64 `json:"weight"`
}

type DeterministicMessage struct {
	RoleID             string `json:"role_id"`
	Text               string `json:"text"`
	SemanticActionCode string `json:"semantic_action_code"`
}

type DeterministicReplyOption struct {
	ID                 string                         `json:"id"`
	RoleID             string                         `json:"role_id"`
	Text               string                         `json:"text"`
	SemanticActionCode string                         `json:"semantic_action_code"`
	Assessments        []DeterministicReplyAssessment `json:"assessments"`
	NextNodeID         string                         `json:"next_node_id"`
	TransitionID       string                         `json:"transition_id"`
}

type DeterministicReplyAssessment struct {
	ClusterID string `json:"cluster_id"`
	Level     int    `json:"level"`
	Comment   string `json:"comment"`
}

func ScenarioDocumentFromScenario(s Scenario) ScenarioDocument {
	context := ScenarioContextDocument{}
	_ = json.Unmarshal(s.Context, &context)
	roles := []ScenarioRole{}
	_ = json.Unmarshal(s.Roles, &roles)
	blocks := []ScenarioBlock{}
	_ = json.Unmarshal(s.Nodes, &blocks)
	NormalizeScenarioBlocks(blocks)
	if context.Description == "" {
		var legacy struct {
			Description string `json:"description"`
		}
		_ = json.Unmarshal(s.Context, &legacy)
		context.Description = legacy.Description
	}
	return ScenarioDocument{
		ScenarioID:        s.ID,
		Title:             s.Title,
		Description:       context.Description,
		ExpandedContext:   context.ExpandedContext,
		Status:            s.Status,
		Difficulty:        s.Difficulty,
		Tags:              context.Tags,
		SupportedRuntimes: s.ModeSupport,
		SupportedChannels: s.Channels,
		TargetClusterIDs:  s.Competencies,
		Roles:             roles,
		Blocks:            blocks,
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
		PublishedAt:       s.PublishedAt,
	}
}

func PublicScenarioInfoFromDocument(doc ScenarioDocument) PublicScenarioInfo {
	roles := make([]PublicScenarioRole, 0, len(doc.Roles))
	for _, role := range doc.Roles {
		roles = append(roles, PublicScenarioRole{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Goal:        role.Goal,
			Interests:   role.Interests,
			Constraints: role.Constraints,
		})
	}
	blocks := make([]PublicScenarioBlock, 0, len(doc.Blocks))
	for _, block := range doc.Blocks {
		blocks = append(blocks, PublicScenarioBlock{
			ID:          block.ID,
			Type:        block.Type,
			Name:        block.Name,
			Description: block.Description,
			Transitions: publicScenarioTransitions(block.Transitions),
		})
	}
	return PublicScenarioInfo{
		ScenarioID:      doc.ScenarioID,
		Title:           doc.Title,
		Description:     doc.Description,
		ExpandedContext: doc.ExpandedContext,
		Roles:           roles,
		Blocks:          blocks,
	}
}

func PublicScenarioBlockByID(doc ScenarioDocument, blockID string) *PublicScenarioBlock {
	for _, block := range doc.Blocks {
		if block.ID != blockID {
			continue
		}
		return &PublicScenarioBlock{
			ID:          block.ID,
			Type:        block.Type,
			Name:        block.Name,
			Description: block.Description,
			Transitions: publicScenarioTransitions(block.Transitions),
		}
	}
	return nil
}

func publicScenarioTransitions(transitions []ScenarioTransition) []PublicScenarioTransition {
	items := make([]PublicScenarioTransition, 0, len(transitions))
	for _, transition := range transitions {
		if transition.Forbidden {
			continue
		}
		items = append(items, PublicScenarioTransition{ID: transition.ID, TargetBlockID: transition.TargetBlockID})
	}
	return items
}

func NormalizeScenarioBlocks(blocks []ScenarioBlock) {
	for i := range blocks {
		if blocks[i].RoleGoals == nil {
			blocks[i].RoleGoals = []BlockRoleGoal{}
		}
		if blocks[i].ActionRules == nil {
			blocks[i].ActionRules = []SemanticActionRule{}
		}
		if blocks[i].Constraints == nil {
			blocks[i].Constraints = []string{}
		}
		if blocks[i].Transitions == nil {
			blocks[i].Transitions = []ScenarioTransition{}
		}
		dialogue := blocks[i].DeterministicDialogue
		if dialogue == nil {
			continue
		}
		if len(dialogue.Nodes) == 0 && dialogue.StartNodeID == "" && dialogue.LearnerRoleID == "" {
			blocks[i].DeterministicDialogue = nil
			continue
		}
		if dialogue.Nodes == nil {
			dialogue.Nodes = []DeterministicDialogueNode{}
		}
		for j := range dialogue.Nodes {
			if dialogue.Nodes[j].AssessmentTargets == nil {
				dialogue.Nodes[j].AssessmentTargets = []DeterministicAssessmentTarget{}
			}
			if dialogue.Nodes[j].ReplyOptions == nil {
				dialogue.Nodes[j].ReplyOptions = []DeterministicReplyOption{}
			}
			for k := range dialogue.Nodes[j].ReplyOptions {
				if dialogue.Nodes[j].ReplyOptions[k].Assessments == nil {
					dialogue.Nodes[j].ReplyOptions[k].Assessments = []DeterministicReplyAssessment{}
				}
			}
		}
	}
}

func ScenarioContextJSON(description, expandedContext string, tags []string) json.RawMessage {
	raw, _ := json.Marshal(ScenarioContextDocument{Description: strings.TrimSpace(description), ExpandedContext: strings.TrimSpace(expandedContext), Tags: tags})
	return raw
}
