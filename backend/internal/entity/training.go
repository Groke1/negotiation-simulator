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

type TrainingChannel string

const (
	TrainingChannelChat         TrainingChannel = "chat"
	TrainingChannelEmail        TrainingChannel = "email"
	TrainingChannelVoiceMessage TrainingChannel = "voice_message"
)

type TrainingStatus string

const (
	TrainingStatusInProgress      TrainingStatus = "in_progress"
	TrainingStatusReadyToComplete TrainingStatus = "ready_to_complete"
	TrainingStatusCompleted       TrainingStatus = "completed"
	TrainingStatusAbandoned       TrainingStatus = "abandoned"
)

type StartTrainingSession struct {
	UserID         string
	ScenarioID     string
	ScenarioRoleID string
	Mode           TrainingMode
	Channel        TrainingChannel
	AssignmentID   *string
}

type TrainingSession struct {
	ID             string
	UserID         string
	ScenarioID     string
	ScenarioTitle  string
	ScenarioRoleID string
	Mode           TrainingMode
	Channel        TrainingChannel
	Status         TrainingStatus
	AssignmentID   *string
	Snapshot       json.RawMessage
	State          json.RawMessage
	Analysis       json.RawMessage
	Score          *int
	CreatedAt      time.Time
	CompletedAt    *time.Time
}

type TrainingMessage struct {
	ID                  string
	SessionID           string
	Author              string
	Content             json.RawMessage
	SemanticActionCode  *string
	SemanticAnnotations json.RawMessage
	ExpectedActionCode  *string
	Classification      json.RawMessage
	CreatedAt           time.Time
}

type TrainingSessionFilter struct {
	UserID     string
	ScenarioID *string
	Status     *TrainingStatus
	Mode       *TrainingMode
	Channel    *TrainingChannel
	Limit      int32
	Offset     int32
}

type CreateTrainingMessage struct {
	ID                  string
	SessionID           string
	Author              string
	Content             json.RawMessage
	SemanticActionCode  *string
	SemanticAnnotations json.RawMessage
	ExpectedActionCode  *string
	Classification      json.RawMessage
}

type TrainingSessionDetails struct {
	Session  TrainingSession
	Messages []TrainingMessage
}

type TrainingActionResult struct {
	EventID        string
	Classification json.RawMessage
	NextState      json.RawMessage
	Status         TrainingStatus
}

type TrainingMessageExchange struct {
	LearnerMessage TrainingMessage
	PartnerMessage TrainingMessage
	Status         TrainingStatus
}

type TrainingCompletion struct {
	Session       TrainingSession
	Score         int
	Outcome       json.RawMessage
	ProfileChange []ProfileClusterChange
}
