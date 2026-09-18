package arena

import (
	"encoding/json"
	"time"
)

type participantRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}
type createAssignmentRequest struct {
	ScenarioID   string               `json:"scenario_id"`
	Participants []participantRequest `json:"participants"`
	ScheduledAt  time.Time            `json:"scheduled_at"`
}
type feedbackRequest struct {
	ParticipantID string          `json:"participant_id"`
	Comment       string          `json:"comment"`
	Ratings       json.RawMessage `json:"ratings"`
}
type wsMessageSend struct {
	Type      string `json:"type"`
	MessageID string `json:"message_id"`
	Text      string `json:"text"`
}
type participantResponse struct {
	UserID   string     `json:"user_id"`
	Role     string     `json:"role"`
	JoinedAt *time.Time `json:"joined_at,omitempty"`
}
type assignmentResponse struct {
	ArenaAssignmentID string                `json:"arena_assignment_id"`
	ScenarioID        string                `json:"scenario_id"`
	ScenarioTitle     string                `json:"scenario_title"`
	ScenarioVersion   string                `json:"scenario_version"`
	Participants      []participantResponse `json:"participants"`
	ScheduledAt       time.Time             `json:"scheduled_at"`
	Status            string                `json:"status"`
	ArenaSessionID    *string               `json:"arena_session_id"`
	CreatedBy         string                `json:"created_by"`
	CreatedAt         time.Time             `json:"created_at"`
}
type sessionResponse struct {
	ArenaSessionID    string                `json:"arena_session_id"`
	ArenaAssignmentID string                `json:"arena_assignment_id"`
	ScenarioID        string                `json:"scenario_id"`
	ScenarioTitle     string                `json:"scenario_title"`
	ScenarioVersion   string                `json:"scenario_version"`
	Channel           string                `json:"channel"`
	Status            string                `json:"status"`
	ScheduledAt       time.Time             `json:"scheduled_at"`
	Participants      []participantResponse `json:"participants"`
	StartedAt         *time.Time            `json:"started_at"`
	CompletedAt       *time.Time            `json:"completed_at"`
}
type messageResponse struct {
	MessageID string    `json:"message_id"`
	SenderID  string    `json:"sender_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
type feedbackResponse struct {
	FeedbackID     string          `json:"feedback_id"`
	ArenaSessionID string          `json:"arena_session_id"`
	ParticipantID  string          `json:"participant_id"`
	ExpertID       string          `json:"expert_id"`
	Comment        string          `json:"comment"`
	Ratings        json.RawMessage `json:"ratings"`
	CreatedAt      time.Time       `json:"created_at"`
}
