package profile

import (
	"context"
	"encoding/json"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type repository interface {
	ListCompetencyClusters(ctx context.Context) ([]entity.CompetencyCluster, error)
	ListProfileForUpdate(ctx context.Context, userID string) ([]entity.ProfileCluster, error)
	UpdateProfileScore(ctx context.Context, userID, clusterID string, score int) error
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

func (s *Service) RecalculateAfterTraining(ctx context.Context, userID, sourceID string, score int) ([]entity.ProfileClusterChange, error) {
	target := clampScore(score)
	return s.recalculate(ctx, userID, sourceID, entity.ProfileChangeSourceTraining, func(current entity.ProfileCluster) int {
		return clampScore(current.Score + (target-current.Score)/4)
	})
}

func (s *Service) RecalculateAfterArenaFeedback(ctx context.Context, userID, sourceID string, ratings map[string]int) ([]entity.ProfileClusterChange, error) {
	return s.recalculate(ctx, userID, sourceID, entity.ProfileChangeSourceArenaFeedback, func(current entity.ProfileCluster) int {
		target, ok := ratings[current.ClusterID]
		if !ok {
			target, ok = ratings[current.Code]
		}
		if !ok {
			return current.Score
		}
		return clampScore((current.Score*3 + clampScore(target)) / 4)
	})
}

func (s *Service) recalculate(ctx context.Context, userID, sourceID string, source entity.ProfileChangeSource, next func(entity.ProfileCluster) int) ([]entity.ProfileClusterChange, error) {
	profile, err := s.repository.ListProfileForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}
	changes := make([]entity.ProfileClusterChange, 0, len(profile))
	for _, current := range profile {
		before := current.Score
		after := next(current)
		if after == before {
			continue
		}
		if err := s.repository.UpdateProfileScore(ctx, userID, current.ClusterID, after); err != nil {
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

func clampScore(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
