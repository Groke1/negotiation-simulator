package entity

import (
	"encoding/json"
	"time"
)

type ArenaAssignmentStatus string

const (
	ArenaAssignmentScheduled  ArenaAssignmentStatus = "scheduled"
	ArenaAssignmentInProgress ArenaAssignmentStatus = "in_progress"
	ArenaAssignmentCompleted  ArenaAssignmentStatus = "completed"
	ArenaAssignmentCancelled  ArenaAssignmentStatus = "cancelled"
)

type ArenaSessionStatus string

const (
	ArenaSessionWaiting    ArenaSessionStatus = "waiting_for_participants"
	ArenaSessionInProgress ArenaSessionStatus = "in_progress"
	ArenaSessionCompleted  ArenaSessionStatus = "completed"
	ArenaSessionCancelled  ArenaSessionStatus = "cancelled"
)

type ArenaParticipant struct {
	UserID         string `json:"user_id"`
	ScenarioRoleID string `json:"scenario_role_id"`
}

type ArenaAssignment struct {
	ID            string                `json:"arena_assignment_id"`
	ScenarioID    string                `json:"scenario_id"`
	ScenarioTitle string                `json:"scenario_title"`
	Participants  []ArenaParticipant    `json:"participants"`
	Channel       string                `json:"channel"`
	ScheduledAt   time.Time             `json:"scheduled_at"`
	Status        ArenaAssignmentStatus `json:"status"`
	SessionID     *string               `json:"arena_session_id"`
	CreatedBy     string                `json:"created_by"`
	CreatedAt     time.Time             `json:"created_at"`
}

type CreateArenaAssignment struct {
	ScenarioID   string
	CreatedBy    string
	Participants []ArenaParticipant
	Channel      string
	ScheduledAt  time.Time
}

type ArenaSession struct {
	ID            string                    `json:"arena_session_id"`
	AssignmentID  string                    `json:"arena_assignment_id"`
	ScenarioID    string                    `json:"scenario_id"`
	ScenarioTitle string                    `json:"scenario_title"`
	Scenario      *PublicScenarioInfo       `json:"scenario,omitempty"`
	MyRole        *ScenarioRole             `json:"-"`
	OpponentRole  *PublicScenarioRole       `json:"-"`
	Channel       string                    `json:"channel"`
	Status        ArenaSessionStatus        `json:"status"`
	ScheduledAt   time.Time                 `json:"scheduled_at"`
	Participants  []ArenaSessionParticipant `json:"participants"`
	Joined        []ArenaSessionParticipant `json:"-"`
	CreatedBy     string                    `json:"created_by"`
	StartedAt     *time.Time                `json:"started_at"`
	CompletedAt   *time.Time                `json:"completed_at"`
}

type ArenaSessionParticipant struct {
	UserID         string     `json:"user_id"`
	ScenarioRoleID string     `json:"scenario_role_id"`
	JoinedAt       *time.Time `json:"joined_at"`
}

type ArenaMessage struct {
	ID                  string          `json:"message_id"`
	SessionID           string          `json:"arena_session_id,omitempty"`
	SenderID            string          `json:"sender_id"`
	Content             json.RawMessage `json:"content"`
	SemanticActionCode  *string         `json:"semantic_action_code,omitempty"`
	SemanticAnnotations json.RawMessage `json:"semantic_annotations,omitempty"`
	Classification      json.RawMessage `json:"classification"`
	CreatedAt           time.Time       `json:"created_at"`
}

type CreateArenaMessage struct {
	ID                  string
	SessionID           string
	SenderID            string
	Content             json.RawMessage
	SemanticActionCode  *string
	SemanticAnnotations json.RawMessage
}

type ClassifyArenaMessage struct {
	SessionID           string
	MessageID           string
	ClassifierID        string
	SemanticActionCode  *string
	SemanticAnnotations json.RawMessage
}

type ArenaMessageClassification struct {
	MessageID                   string          `json:"message_id"`
	SelectedSemanticActionCode  *string         `json:"selected_semantic_action_code,omitempty"`
	ExpectedSemanticActionCode  *string         `json:"expected_semantic_action_code,omitempty"`
	SelectedSemanticAnnotations json.RawMessage `json:"selected_semantic_annotations,omitempty"`
	ExpectedSemanticAnnotations json.RawMessage `json:"expected_semantic_annotations,omitempty"`
	Correct                     bool            `json:"correct"`
}

type ArenaFeedbackRating struct {
	ClusterID string `json:"cluster_id"`
	Score     int    `json:"score"`
	Comment   string `json:"comment,omitempty"`
}

type ArenaFeedback struct {
	ID            string                 `json:"feedback_id"`
	SessionID     string                 `json:"arena_session_id"`
	ParticipantID string                 `json:"participant_id"`
	CuratorID     string                 `json:"curator_id"`
	Comment       string                 `json:"comment"`
	Ratings       []ArenaFeedbackRating  `json:"ratings"`
	ProfileChange []ProfileClusterChange `json:"profile_change"`
	CreatedAt     time.Time              `json:"created_at"`
}

type UserArenaHistoryItem struct {
	ArenaSessionID    string             `json:"arena_session_id"`
	ArenaAssignmentID string             `json:"arena_assignment_id"`
	ScenarioID        string             `json:"scenario_id"`
	ScenarioTitle     string             `json:"scenario_title"`
	ScheduledAt       time.Time          `json:"scheduled_at"`
	StartedAt         *time.Time         `json:"started_at"`
	CompletedAt       *time.Time         `json:"completed_at"`
	Status            ArenaSessionStatus `json:"status"`
	HasFeedback       bool               `json:"has_feedback"`
}

func ArenaParticipantsJSON(participants []ArenaParticipant) json.RawMessage {
	raw, _ := json.Marshal(participants)
	return raw
}
