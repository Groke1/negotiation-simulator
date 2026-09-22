package arena

import (
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type createAssignmentRequest struct {
	ScenarioID   string                    `json:"scenario_id"`
	Participants []entity.ArenaParticipant `json:"participants"`
	Channel      string                    `json:"channel"`
	ScheduledAt  time.Time                 `json:"scheduled_at"`
}

type assignmentListResponse struct {
	Arenas []entity.ArenaAssignment `json:"arenas"`
}

type createAssignmentResponse struct {
	ArenaAssignmentID string                       `json:"arena_assignment_id"`
	Status            entity.ArenaAssignmentStatus `json:"status"`
}

type joinArenaResponse struct {
	ArenaSessionID string                           `json:"arena_session_id"`
	Status         entity.ArenaSessionStatus        `json:"status"`
	Participants   []entity.ArenaSessionParticipant `json:"participants"`
	WebsocketPath  string                           `json:"websocket_path"`
	Preparation    arenaPreparation                 `json:"preparation"`
}

type arenaPreparation struct {
	ScenarioID      string              `json:"scenario_id"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	ExpandedContext string              `json:"expanded_context"`
	MyRole          scenarioRolePrivate `json:"my_role"`
	OpponentRole    scenarioRoleVisible `json:"opponent_role"`
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
}

type messageListResponse struct {
	Messages []entity.ArenaMessage `json:"messages"`
}

type feedbackRequest struct {
	ParticipantID string                       `json:"participant_id"`
	Comment       string                       `json:"comment"`
	Ratings       []entity.ArenaFeedbackRating `json:"ratings"`
}

type feedbackListResponse struct {
	Feedback []entity.ArenaFeedback `json:"feedback"`
}

type userAssignmentsResponse struct {
	Arenas []entity.ArenaAssignment `json:"arenas"`
}

type userHistoryResponse struct {
	Arenas []entity.UserArenaHistoryItem `json:"arenas"`
}

type sessionDetailsResponse struct {
	Session  entity.ArenaSession    `json:"session"`
	Messages []entity.ArenaMessage  `json:"messages"`
	Feedback []entity.ArenaFeedback `json:"feedback"`
	Media    []any                  `json:"media"`
}

type wsMessageSend struct {
	Type               string `json:"type"`
	MessageID          string `json:"message_id"`
	Text               string `json:"text"`
	SemanticActionCode string `json:"semantic_action_code"`
}
