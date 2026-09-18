package entity

import (
	"encoding/json"
	"time"
)

type TrainingMode string

const (
	TrainingModeDeterministic  TrainingMode = "deterministic"
	TrainingModeVirtualPartner TrainingMode = "virtual_partner"
)

func (m TrainingMode) Valid() bool {
	return m == TrainingModeDeterministic || m == TrainingModeVirtualPartner
}

type TrainingStatus string

const (
	TrainingStatusInProgress      TrainingStatus = "in_progress"
	TrainingStatusReadyToComplete TrainingStatus = "ready_to_complete"
	TrainingStatusCompleted       TrainingStatus = "completed"
	TrainingStatusAbandoned       TrainingStatus = "abandoned"
)

type TrainingMessage struct {
	ID        string
	SessionID string
	Author    string
	Text      string
	CreatedAt time.Time
}

type TrainingEvent struct {
	ID             string
	SessionID      string
	ClientEventID  string
	SequenceNumber int32
	NodeID         string
	ActionID       string
	SemanticAction string
	NextNodeID     *string
	CreatedAt      time.Time
}

type TrainingSession struct {
	ID                string
	UserID            string
	ScenarioID        string
	ScenarioVersionID *string
	ScenarioVersion   string
	ScenarioSnapshot  json.RawMessage
	Mode              TrainingMode
	Channel           string
	IsPreview         bool
	Status            TrainingStatus
	CurrentNodeID     *string
	Outcome           json.RawMessage
	Analysis          json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CompletedAt       *time.Time
}

type StartTrainingSession struct {
	ScenarioID string
	Mode       TrainingMode
	Channel    string
	Preview    bool
}

type SelectTrainingAction struct {
	SessionID     string
	ActionID      string
	ClientEventID string
}

type SendTrainingMessage struct {
	SessionID string
	Text      string
	MessageID string
}

type AvailableAction struct {
	ID             string `json:"action_id"`
	Label          string `json:"label"`
	SemanticAction string `json:"semantic_action"`
	NextNodeID     string `json:"next_node_id"`
}

type TrainingNode struct {
	ID               string            `json:"node_id"`
	Speaker          string            `json:"speaker"`
	Text             string            `json:"text"`
	AvailableActions []AvailableAction `json:"available_actions"`
	Terminal         bool              `json:"terminal"`
}

type TrainingActionResult struct {
	Event    TrainingEvent
	NextNode *TrainingNode
	Status   TrainingStatus
}

type TrainingMessageResult struct {
	LearnerMessage  TrainingMessage
	PartnerMessage  TrainingMessage
	SemanticActions []string
}

type TrainingCompletion struct {
	Outcome  json.RawMessage
	Analysis json.RawMessage
}
