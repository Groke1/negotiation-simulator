package entity

import (
	"encoding/json"
	"time"
)

type ArenaAssignmentStatus string

const (
	ArenaAssignmentScheduled  ArenaAssignmentStatus = "scheduled"
	ArenaAssignmentReady      ArenaAssignmentStatus = "ready"
	ArenaAssignmentInProgress ArenaAssignmentStatus = "in_progress"
	ArenaAssignmentCompleted  ArenaAssignmentStatus = "completed"
	ArenaAssignmentCancelled  ArenaAssignmentStatus = "cancelled"
)

type ArenaSessionStatus string

const (
	ArenaSessionWaiting  ArenaSessionStatus = "waiting_for_participants"
	ArenaSessionProgress ArenaSessionStatus = "in_progress"
	ArenaSessionComplete ArenaSessionStatus = "completed"
	ArenaSessionReviewed ArenaSessionStatus = "reviewed"
)

type ArenaParticipantRole string

const (
	ArenaParticipantLearner ArenaParticipantRole = "learner"
	ArenaParticipantPartner ArenaParticipantRole = "partner"
)

func (r ArenaParticipantRole) Valid() bool {
	return r == ArenaParticipantLearner || r == ArenaParticipantPartner
}

type ArenaParticipant struct {
	UserID   string
	Role     ArenaParticipantRole
	JoinedAt *time.Time
}

type ArenaAssignment struct {
	ID                string
	OrganizationID    string
	ScenarioID        string
	ScenarioVersionID string
	ScenarioTitle     string
	ScenarioVersion   string
	Participants      []ArenaParticipant
	ScheduledAt       time.Time
	Status            ArenaAssignmentStatus
	SessionID         *string
	CreatedBy         string
	CreatedAt         time.Time
}

type ArenaSession struct {
	ID              string
	AssignmentID    string
	ScenarioID      string
	ScenarioTitle   string
	ScenarioVersion string
	Channel         string
	Status          ArenaSessionStatus
	ScheduledAt     time.Time
	Participants    []ArenaParticipant
	StartedAt       *time.Time
	CompletedAt     *time.Time
}

type ArenaMessage struct {
	ID        string
	SessionID string
	SenderID  string
	Text      string
	CreatedAt time.Time
}

type ArenaFeedback struct {
	ID            string
	SessionID     string
	ParticipantID string
	ExpertID      string
	Comment       string
	Ratings       json.RawMessage
	CreatedAt     time.Time
}

type CreateArenaAssignment struct {
	ScenarioID   string
	Participants []ArenaParticipant
	ScheduledAt  time.Time
}

type CreateArenaFeedback struct {
	SessionID     string
	ParticipantID string
	Comment       string
	Ratings       json.RawMessage
}
