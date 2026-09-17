package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/auth/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type authRepository struct {
	queries *sqlc.Queries
}

func NewAuthRepository(qdb sqlc.DBTX) *authRepository {
	return &authRepository{queries: sqlc.New(qdb)}
}

func (r *authRepository) CreateRefreshToken(ctx context.Context, token entity.RefreshToken) (*entity.RefreshToken, error) {
	id, err := pgxconv.StringToUUID(token.ID)
	if err != nil {
		return nil, err
	}
	userID, err := pgxconv.StringToUUID(token.UserID)
	if err != nil {
		return nil, err
	}

	row, err := r.getQueries(ctx).CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		ID:         id,
		UserID:     userID,
		TokenHash:  token.TokenHash,
		ActiveRole: sqlc.UsersProductRole(token.ActiveRole),
		ExpiresAt:  pgxconv.TimeToRequiredTimestamptz(token.ExpiresAt),
	})
	if err != nil {
		return nil, err
	}
	return refreshTokenFromSQL(row), nil
}

func (r *authRepository) GetRefreshTokenByHashForUpdate(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	row, err := r.getQueries(ctx).GetRefreshTokenByHashForUpdate(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrRefreshTokenInvalid
	}
	if err != nil {
		return nil, err
	}
	return refreshTokenFromSQL(row), nil
}

func (r *authRepository) MarkRefreshTokenUsed(ctx context.Context, id string) error {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return err
	}
	return r.getQueries(ctx).MarkRefreshTokenUsed(ctx, pgID)
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, id string) error {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return err
	}
	return r.getQueries(ctx).RevokeRefreshToken(ctx, pgID)
}

func (r *authRepository) UpdateRefreshTokenActiveRole(ctx context.Context, sessionID, userID string, role entity.ProductRole) (*entity.RefreshToken, error) {
	pgSessionID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, err
	}
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, err
	}
	row, err := r.getQueries(ctx).UpdateRefreshTokenActiveRole(ctx, sqlc.UpdateRefreshTokenActiveRoleParams{
		ID:         pgSessionID,
		UserID:     pgUserID,
		ActiveRole: sqlc.UsersProductRole(role),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrAccessTokenInvalid
	}
	if err != nil {
		return nil, err
	}
	return refreshTokenFromSQL(row), nil
}

func (r *authRepository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func refreshTokenFromSQL(row sqlc.UsersRefreshToken) *entity.RefreshToken {
	return &entity.RefreshToken{
		ID:         pgxconv.UUIDToString(row.ID),
		UserID:     pgxconv.UUIDToString(row.UserID),
		TokenHash:  row.TokenHash,
		ActiveRole: entity.ProductRole(row.ActiveRole),
		ExpiresAt:  row.ExpiresAt.Time,
		RevokedAt:  pgxconv.TimestamptzToTime(row.RevokedAt),
		UsedAt:     pgxconv.TimestamptzToTime(row.UsedAt),
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}
