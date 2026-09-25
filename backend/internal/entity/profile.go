package entity

type CompetencyCluster struct {
	ID          string
	Code        string
	Name        string
	Description string
}

type TrainingClusterResult struct {
	ClusterID      string  `json:"cluster_id"`
	Score          float64 `json:"score"`
	EvidenceWeight float64 `json:"-"`
	Comment        string  `json:"comment,omitempty"`
}

type ProfileChangeSource string

const (
	ProfileChangeSourceTraining      ProfileChangeSource = "training"
	ProfileChangeSourceArenaFeedback ProfileChangeSource = "arena_feedback"
)
