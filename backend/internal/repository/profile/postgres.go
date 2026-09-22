package profile

import (
	"context"
	"encoding/json"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/profile/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(qdb sqlc.DBTX) *Repository {
	return &Repository{queries: sqlc.New(qdb)}
}

func (r *Repository) ListCompetencyClusters(ctx context.Context) ([]entity.CompetencyCluster, error) {
	rows, err := r.getQueries(ctx).ListCompetencyClusters(ctx)
	if err != nil {
		return nil, err
	}
	clusters := make([]entity.CompetencyCluster, 0, len(rows))
	for _, row := range rows {
		clusters = append(clusters, entity.CompetencyCluster{
			ID:          pgxconv.UUIDToString(row.ID),
			Code:        row.Code,
			Name:        row.Name,
			Description: row.Description,
		})
	}
	return clusters, nil
}

func (r *Repository) RecalculateAfterTraining(ctx context.Context, userID, sourceID string, score int) ([]entity.ProfileClusterChange, error) {
	return r.recalculate(ctx, userID, sourceID, entity.ProfileChangeSourceTraining, func(current entity.ProfileCluster) int {
		target := score
		if target < 0 {
			target = 0
		}
		if target > 100 {
			target = 100
		}
		return clampScore(current.Score + (target-current.Score)/4)
	})
}

func (r *Repository) RecalculateAfterArenaFeedback(ctx context.Context, userID, sourceID string, ratings map[string]int) ([]entity.ProfileClusterChange, error) {
	return r.recalculate(ctx, userID, sourceID, entity.ProfileChangeSourceArenaFeedback, func(current entity.ProfileCluster) int {
		target, ok := ratings[current.ClusterID]
		if !ok {
			target, ok = ratings[current.Code]
		}
		if !ok {
			return current.Score
		}
		if target < 0 {
			target = 0
		}
		if target > 100 {
			target = 100
		}
		return clampScore((current.Score*3 + target) / 4)
	})
}

func (r *Repository) recalculate(ctx context.Context, userID, sourceID string, source entity.ProfileChangeSource, next func(entity.ProfileCluster) int) ([]entity.ProfileClusterChange, error) {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	pgSourceID, err := pgxconv.StringToUUID(sourceID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListProfileForUpdate(ctx, pgUserID)
	if err != nil {
		return nil, err
	}
	changes := make([]entity.ProfileClusterChange, 0, len(rows))
	for _, row := range rows {
		clusterID := pgxconv.UUIDToString(row.ClusterID)
		before := int(row.Score)
		after := next(entity.ProfileCluster{ClusterID: clusterID, Code: row.Code, Name: row.Name, Score: before})
		if after == before {
			continue
		}
		if err := r.getQueries(ctx).UpdateProfileScore(ctx, sqlc.UpdateProfileScoreParams{UserID: pgUserID, ClusterID: row.ClusterID, Score: int16(after)}); err != nil {
			return nil, err
		}
		changes = append(changes, entity.ProfileClusterChange{ClusterID: clusterID, Before: before, After: after, Delta: after - before})
	}
	if len(changes) == 0 {
		return changes, nil
	}
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(changes)
	if err != nil {
		return nil, err
	}
	if err := r.getQueries(ctx).CreateProfileHistory(ctx, sqlc.CreateProfileHistoryParams{ID: pgID, UserID: pgUserID, Source: string(source), SourceID: pgSourceID, Changes: data}); err != nil {
		return nil, err
	}
	return changes, nil
}

func (r *Repository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
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
