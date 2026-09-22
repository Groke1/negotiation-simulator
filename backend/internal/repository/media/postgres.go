package media

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/media/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(qdb sqlc.DBTX) *Repository {
	return &Repository{queries: sqlc.New(qdb)}
}

func (r *Repository) CreateMedia(ctx context.Context, params entity.CreateMediaAsset) (*entity.MediaAsset, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, err
	}
	ownerID, err := pgxconv.StringToUUID(params.OwnerID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateMedia(ctx, sqlc.CreateMediaParams{
		ID:              pgID,
		OwnerID:         ownerID,
		MediaType:       string(params.MediaType),
		ContentType:     params.ContentType,
		Payload:         params.Payload,
		DurationSeconds: durationArg(params.DurationSeconds),
	})
	if err != nil {
		return nil, err
	}
	return mediaFromSQL(row.ID, row.OwnerID, row.MediaType, row.ContentType, row.DurationSeconds, row.CreatedAt), nil
}

func (r *Repository) GetMedia(ctx context.Context, id string) (*entity.MediaAsset, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetMedia(ctx, pgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrMediaNotFound
	}
	if err != nil {
		return nil, err
	}
	return mediaFromSQL(row.ID, row.OwnerID, row.MediaType, row.ContentType, row.DurationSeconds, row.CreatedAt), nil
}

func (r *Repository) GetMediaContent(ctx context.Context, id string) (*entity.MediaContent, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetMediaContent(ctx, pgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrMediaNotFound
	}
	if err != nil {
		return nil, err
	}
	return &entity.MediaContent{OwnerID: pgxconv.UUIDToString(row.OwnerID), ContentType: row.ContentType, Payload: row.Payload}, nil
}

func (r *Repository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func durationArg(value *int) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*value), Valid: true}
}

func mediaFromSQL(id, ownerID pgtype.UUID, mediaType, contentType string, duration pgtype.Int4, createdAt pgtype.Timestamptz) *entity.MediaAsset {
	return &entity.MediaAsset{
		ID:              pgxconv.UUIDToString(id),
		OwnerID:         pgxconv.UUIDToString(ownerID),
		MediaType:       entity.MediaType(mediaType),
		ContentType:     contentType,
		DurationSeconds: int4ToInt(duration),
		CreatedAt:       createdAt.Time,
	}
}

func int4ToInt(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int32)
	return &v
}
