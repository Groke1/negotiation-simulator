package entity

type CompetencyCluster struct {
	ID          string
	Code        string
	Name        string
	Description string
}

type ProfileChangeSource string

const (
	ProfileChangeSourceTraining      ProfileChangeSource = "training"
	ProfileChangeSourceArenaFeedback ProfileChangeSource = "arena_feedback"
)
