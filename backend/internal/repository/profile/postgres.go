package profile

import (
	"context"

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

func (r *Repository) ListProfileForUpdate(ctx context.Context, userID string) ([]entity.ProfileCluster, error) {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListProfileForUpdate(ctx, pgUserID)
	if err != nil {
		return nil, err
	}
	profile := make([]entity.ProfileCluster, 0, len(rows))
	for _, row := range rows {
		profile = append(profile, entity.ProfileCluster{
			ClusterID:     pgxconv.UUIDToString(row.ClusterID),
			Code:          row.Code,
			Name:          row.Name,
			Score:         int(row.Score),
			ScoreRaw:      row.ScoreRaw,
			Confidence:    int(row.Confidence),
			ConfidenceRaw: row.ConfidenceRaw,
			EvidenceMass:  row.EvidenceMass,
			UpdatedAt:     pgxconv.TimestamptzToTime(row.UpdatedAt),
		})
	}
	return profile, nil
}

func (r *Repository) UpdateProfileState(ctx context.Context, userID, clusterID string, score, evidenceMass, confidence float64) error {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return entity.ErrInvalidRequest
	}
	pgClusterID, err := pgxconv.StringToUUID(clusterID)
	if err != nil {
		return entity.ErrInvalidRequest
	}
	return r.getQueries(ctx).UpdateProfileState(ctx, sqlc.UpdateProfileStateParams{UserID: pgUserID, ClusterID: pgClusterID, Score: score, EvidenceMass: evidenceMass, Confidence: confidence})
}

func (r *Repository) CreateProfileHistory(ctx context.Context, userID, sourceID string, source entity.ProfileChangeSource, changes []byte) error {
	id, err := entity.NewUUID()
	if err != nil {
		return err
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return err
	}
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return entity.ErrInvalidRequest
	}
	pgSourceID, err := pgxconv.StringToUUID(sourceID)
	if err != nil {
		return entity.ErrInvalidRequest
	}
	return r.getQueries(ctx).CreateProfileHistory(ctx, sqlc.CreateProfileHistoryParams{ID: pgID, UserID: pgUserID, Source: string(source), SourceID: pgSourceID, Changes: changes})
}

func (r *Repository) ListSemanticActions(ctx context.Context) ([]entity.SemanticActionDefinition, error) {
	rows, err := r.getQueries(ctx).ListSemanticActions(ctx)
	if err != nil {
		return nil, err
	}
	actions := make([]entity.SemanticActionDefinition, 0, len(rows))
	for _, row := range rows {
		actions = append(actions, entity.SemanticActionDefinition{Code: row.Code, Name: row.Name, Description: row.Description})
	}
	return actions, nil
}

func (r *Repository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}
