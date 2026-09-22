package entity

import "time"

type AssignmentStatus string

const (
	AssignmentStatusPending   AssignmentStatus = "pending"
	AssignmentStatusCompleted AssignmentStatus = "completed"
)

type Assignment struct {
	ID            string           `json:"assignment_id"`
	CuratorID     string           `json:"curator_id"`
	LearnerID     string           `json:"learner_id"`
	ScenarioID    string           `json:"scenario_id"`
	ScenarioTitle string           `json:"scenario_title"`
	Status        AssignmentStatus `json:"status"`
	CreatedAt     time.Time        `json:"created_at"`
	CompletedAt   *time.Time       `json:"completed_at"`
}

type CreateAssignment struct {
	CuratorID  string
	LearnerID  string
	ScenarioID string
}

type AssignmentFilter struct {
	CuratorID string
	Status    *AssignmentStatus
	Limit     int32
	Offset    int32
}

type CuratorTeamMember struct {
	UserID        string           `json:"user_id"`
	Name          string           `json:"name"`
	SessionsCount int              `json:"sessions_count"`
	LastSessionAt *time.Time       `json:"last_session_at"`
	AverageScore  *int             `json:"average_score"`
	Clusters      []ProfileCluster `json:"clusters"`
}

type TeamSessionListItem struct {
	Type          string     `json:"type"`
	SessionID     string     `json:"session_id"`
	UserIDs       []string   `json:"user_ids"`
	ScenarioID    string     `json:"scenario_id"`
	ScenarioTitle string     `json:"scenario_title"`
	Status        string     `json:"status"`
	Score         *int       `json:"score"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type TeamAnalytics struct {
	UsersCount              int                    `json:"users_count"`
	TrainingSessionsCount   int                    `json:"training_sessions_count"`
	ArenaSessionsCount      int                    `json:"arena_sessions_count"`
	PendingAssignmentsCount int                    `json:"pending_assignments_count"`
	Clusters                []TeamAnalyticsCluster `json:"clusters"`
}

type TeamAnalyticsCluster struct {
	ClusterID    string `json:"cluster_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	AverageScore int    `json:"average_score"`
}
