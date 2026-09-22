package auth

import (
	"context"
	"errors"
	"time"

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
		ActiveRole: nullableRole(token.ActiveRole),
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

func (r *authRepository) RotateRefreshToken(ctx context.Context, id, currentTokenHash, newTokenHash string, expiresAt time.Time) (*entity.RefreshToken, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.getQueries(ctx).RotateRefreshToken(ctx, sqlc.RotateRefreshTokenParams{
		ID:               pgID,
		CurrentTokenHash: currentTokenHash,
		NewTokenHash:     newTokenHash,
		ExpiresAt:        pgxconv.TimeToRequiredTimestamptz(expiresAt),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrRefreshTokenInvalid
	}
	if err != nil {
		return nil, err
	}
	return refreshTokenFromSQL(row), nil
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
		ActiveRole: nullableRole(role),
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
		ActiveRole: activeRoleFromSQL(row.ActiveRole),
		ExpiresAt:  row.ExpiresAt.Time,
		RevokedAt:  pgxconv.TimestamptzToTime(row.RevokedAt),
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}

func nullableRole(role entity.ProductRole) sqlc.NullUsersProductRole {
	if role == entity.NoActiveRole {
		return sqlc.NullUsersProductRole{}
	}
	return sqlc.NullUsersProductRole{UsersProductRole: sqlc.UsersProductRole(role), Valid: true}
}

func activeRoleFromSQL(role sqlc.NullUsersProductRole) entity.ProductRole {
	if !role.Valid {
		return entity.NoActiveRole
	}
	return entity.ProductRole(role.UsersProductRole)
}
