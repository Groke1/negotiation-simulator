package profile

import (
	"context"
	"encoding/json"
	"math"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

const (
	deterministicSourceWeight = 1.0
	arenaSourceWeight         = 1.5
	profileEvidenceCap        = 6.0
	confidenceFullMass        = 6.0
)

type repository interface {
	ListCompetencyClusters(ctx context.Context) ([]entity.CompetencyCluster, error)
	ListSemanticActions(ctx context.Context) ([]entity.SemanticActionDefinition, error)
	ListProfileForUpdate(ctx context.Context, userID string) ([]entity.ProfileCluster, error)
	UpdateProfileState(ctx context.Context, userID, clusterID string, score, evidenceMass, confidence float64) error
	CreateProfileHistory(ctx context.Context, userID, sourceID string, source entity.ProfileChangeSource, changes []byte) error
}

type Service struct {
	repository repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListCompetencyClusters(ctx context.Context) ([]entity.CompetencyCluster, error) {
	return s.repository.ListCompetencyClusters(ctx)
}

func (s *Service) ListSemanticActions(ctx context.Context) ([]entity.SemanticActionDefinition, error) {
	return s.repository.ListSemanticActions(ctx)
}

func (s *Service) RecalculateAfterTraining(ctx context.Context, userID, sourceID string, results []entity.TrainingClusterResult) ([]entity.ProfileClusterChange, error) {
	observations := make(map[string]profileObservation, len(results))
	for _, result := range results {
		if result.ClusterID == "" || result.EvidenceWeight <= 0 {
			continue
		}
		observations[result.ClusterID] = profileObservation{
			Score:  clampFloat(result.Score),
			Weight: deterministicSourceWeight * result.EvidenceWeight,
		}
	}
	return s.recalculate(ctx, userID, sourceID, entity.ProfileChangeSourceTraining, observations)
}

func (s *Service) RecalculateAfterArenaFeedback(ctx context.Context, userID, sourceID string, ratings map[string]int) ([]entity.ProfileClusterChange, error) {
	observations := make(map[string]profileObservation, len(ratings))
	for clusterID, score := range ratings {
		observations[clusterID] = profileObservation{Score: clampFloat(float64(score)), Weight: arenaSourceWeight}
	}
	return s.recalculate(ctx, userID, sourceID, entity.ProfileChangeSourceArenaFeedback, observations)
}

type profileObservation struct {
	Score  float64
	Weight float64
}

func (s *Service) recalculate(ctx context.Context, userID, sourceID string, source entity.ProfileChangeSource, observations map[string]profileObservation) ([]entity.ProfileClusterChange, error) {
	profile, err := s.repository.ListProfileForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}
	changes := make([]entity.ProfileClusterChange, 0, len(profile))
	for _, current := range profile {
		observation, ok := observations[current.ClusterID]
		if !ok {
			observation, ok = observations[current.Code]
		}
		if !ok || observation.Weight <= 0 {
			continue
		}
		before := rawScore(current)
		beforeMass := clampEvidenceMass(current.EvidenceMass)
		after := weightedScore(before, beforeMass, observation.Score, observation.Weight)
		afterMass := clampEvidenceMass(beforeMass + observation.Weight)
		confidence := confidenceFromMass(afterMass)
		if err := s.repository.UpdateProfileState(ctx, userID, current.ClusterID, after, afterMass, confidence); err != nil {
			return nil, err
		}
		changes = append(changes, entity.ProfileClusterChange{ClusterID: current.ClusterID, Before: before, After: after, Delta: after - before})
	}
	if len(changes) == 0 {
		return changes, nil
	}
	data, err := json.Marshal(changes)
	if err != nil {
		return nil, err
	}
	if err := s.repository.CreateProfileHistory(ctx, userID, sourceID, source, data); err != nil {
		return nil, err
	}
	return changes, nil
}

func rawScore(current entity.ProfileCluster) float64 {
	if current.ScoreRaw > 0 || current.Score != 0 {
		return clampFloat(current.ScoreRaw)
	}
	return clampFloat(float64(current.Score))
}

func weightedScore(currentScore, currentMass, observationScore, observationWeight float64) float64 {
	if currentMass <= 0 {
		return clampFloat(observationScore)
	}
	return clampFloat((currentScore*currentMass + observationScore*observationWeight) / (currentMass + observationWeight))
}

func confidenceFromMass(mass float64) float64 {
	if confidenceFullMass <= 0 {
		return 0
	}
	return clampFloat((mass / confidenceFullMass) * 100)
}

func clampEvidenceMass(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > profileEvidenceCap {
		return profileEvidenceCap
	}
	return value
}

func clampFloat(value float64) float64 {
	if math.IsNaN(value) || value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
