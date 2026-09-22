package curator

import (
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type curatorTeamResponse struct {
	Team []entity.CuratorTeamMember `json:"team"`
}

type createAssignmentRequest struct {
	LearnerID  string `json:"learner_id"`
	ScenarioID string `json:"scenario_id"`
}

type assignmentResponse struct {
	AssignmentID  string                  `json:"assignment_id"`
	CuratorID     string                  `json:"curator_id"`
	LearnerID     string                  `json:"learner_id"`
	ScenarioID    string                  `json:"scenario_id"`
	ScenarioTitle string                  `json:"scenario_title"`
	Status        entity.AssignmentStatus `json:"status"`
	CreatedAt     time.Time               `json:"created_at"`
	CompletedAt   *time.Time              `json:"completed_at"`
}

type assignmentListResponse struct {
	Assignments []assignmentResponse `json:"assignments"`
}

type teamSessionListResponse struct {
	Sessions []entity.TeamSessionListItem `json:"sessions"`
}
