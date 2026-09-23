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

type ScenarioBlock struct {
	ID                    string                `json:"id"`
	Type                  string                `json:"type"`
	Name                  string                `json:"name"`
	Description           string                `json:"description"`
	ExpandedContext       string                `json:"expanded_context"`
	LeadRoleID            string                `json:"lead_role_id"`
	LeadProbability       int                   `json:"lead_probability"`
	RoleGoals             []BlockRoleGoal       `json:"role_goals"`
	Behavior              string                `json:"behavior"`
	ActionRules           []SemanticActionRule  `json:"action_rules"`
	Constraints           []string              `json:"constraints"`
	DeterministicDialogue DeterministicDialogue `json:"deterministic_dialogue"`
	Transitions           []ScenarioTransition  `json:"transitions"`
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
	ID            string `json:"id"`
	TargetBlockID string `json:"target_block_id"`
	Condition     string `json:"condition"`
	Note          string `json:"note"`
	Weight        int    `json:"weight"`
	Forbidden     bool   `json:"forbidden"`
}

type DeterministicDialogue struct {
	StartNodeID string                      `json:"start_node_id"`
	Nodes       []DeterministicDialogueNode `json:"nodes"`
}

type DeterministicDialogueNode struct {
	ID              string                     `json:"id"`
	IncomingMessage *DeterministicMessage      `json:"incoming_message"`
	ReplyOptions    []DeterministicReplyOption `json:"reply_options"`
}

type DeterministicMessage struct {
	RoleID             string `json:"role_id"`
	Text               string `json:"text"`
	SemanticActionCode string `json:"semantic_action_code"`
}

type DeterministicReplyOption struct {
	ID                 string `json:"id"`
	RoleID             string `json:"role_id"`
	Text               string `json:"text"`
	SemanticActionCode string `json:"semantic_action_code"`
	NextNodeID         string `json:"next_node_id"`
	TransitionID       string `json:"transition_id"`
}

func ScenarioDocumentFromScenario(s Scenario) ScenarioDocument {
	context := ScenarioContextDocument{}
	_ = json.Unmarshal(s.Context, &context)
	roles := []ScenarioRole{}
	_ = json.Unmarshal(s.Roles, &roles)
	blocks := []ScenarioBlock{}
	_ = json.Unmarshal(s.Nodes, &blocks)
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

func ScenarioContextJSON(description, expandedContext string, tags []string) json.RawMessage {
	raw, _ := json.Marshal(ScenarioContextDocument{Description: strings.TrimSpace(description), ExpandedContext: strings.TrimSpace(expandedContext), Tags: tags})
	return raw
}
