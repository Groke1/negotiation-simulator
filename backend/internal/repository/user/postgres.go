package user

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/user/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type userRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(qdb sqlc.DBTX) *userRepository {
	return &userRepository{queries: sqlc.New(qdb)}
}

func (r *userRepository) CreateUser(ctx context.Context, user entity.User) (*entity.User, error) {
	id, err := pgxconv.StringToUUID(user.ID)
	if err != nil {
		return nil, err
	}
	row, err := r.getQueries(ctx).CreateUser(ctx, sqlc.CreateUserParams{
		ID:           id,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "ux_users_username":
				return nil, entity.ErrUsernameAlreadyExists
			case "ux_users_email":
				return nil, entity.ErrEmailAlreadyExists
			}
		}
		return nil, err
	}
	return userFromSQL(row), nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.getQueries(ctx).GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return userFromSQL(row), nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetUserByID(ctx, pgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return userFromSQL(row), nil
}

func (r *userRepository) AddUserRole(ctx context.Context, userID string, role entity.ProductRole) error {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return entity.ErrInvalidRequest
	}
	err = r.getQueries(ctx).AddUserRole(ctx, sqlc.AddUserRoleParams{
		UserID: pgUserID,
		Role:   sqlc.UsersProductRole(role),
	})
	if err != nil {
		if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "user_roles_pkey":
				return entity.ErrRoleAlreadyAssigned
			case "user_roles_user_id_fkey":
				return entity.ErrUserNotFound
			}
		}
		return err
	}
	return nil
}

func (r *userRepository) GetUserRoles(ctx context.Context, userID string) ([]entity.ProductRole, error) {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).GetUserRoles(ctx, pgUserID)
	if err != nil {
		return nil, err
	}
	roles := make([]entity.ProductRole, 0, len(rows))
	for _, role := range rows {
		roles = append(roles, entity.ProductRole(role))
	}
	return roles, nil
}

func (r *userRepository) ListProfile(ctx context.Context, userID string) ([]entity.ProfileCluster, error) {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListProfile(ctx, pgUserID)
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

func (r *userRepository) ListProfileHistory(ctx context.Context, userID string, from, to *time.Time, limit, offset int32) ([]entity.ProfileHistoryEvent, error) {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListProfileHistory(ctx, sqlc.ListProfileHistoryParams{UserID: pgUserID, FromAt: pgxconv.TimeToTimestamptz(from), ToAt: pgxconv.TimeToTimestamptz(to), LimitCount: limit, OffsetCount: offset})
	if err != nil {
		return nil, err
	}
	result := make([]entity.ProfileHistoryEvent, 0, len(rows))
	for _, row := range rows {
		var changes []entity.ProfileClusterChange
		if err := json.Unmarshal(row.Changes, &changes); err != nil {
			return nil, err
		}
		result = append(result, entity.ProfileHistoryEvent{ID: pgxconv.UUIDToString(row.ID), Source: row.Source, SourceID: pgxconv.UUIDToString(row.SourceID), Changes: changes, CreatedAt: row.CreatedAt.Time})
	}
	return result, nil
}

func (r *userRepository) ListOrganizationUsers(ctx context.Context, requesterID string, limit, offset int32) ([]entity.OrganizationUser, error) {
	pgRequesterID, err := pgxconv.StringToUUID(requesterID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListOrganizationUsers(ctx, sqlc.ListOrganizationUsersParams{RequesterID: pgRequesterID, LimitCount: limit, OffsetCount: offset})
	if err != nil {
		return nil, err
	}
	users := make([]entity.OrganizationUser, 0, len(rows))
	for _, row := range rows {
		roles := make([]entity.ProductRole, 0, len(row.Roles))
		for _, role := range row.Roles {
			roles = append(roles, entity.ProductRole(role))
		}
		users = append(users, entity.OrganizationUser{ID: pgxconv.UUIDToString(row.ID), Username: row.Username, Email: row.Email, Roles: roles, CreatedAt: row.CreatedAt.Time})
	}
	return users, nil
}

func (r *userRepository) SetUserOrganization(ctx context.Context, userID, organizationID string) (*entity.User, error) {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	pgOrganizationID, err := pgxconv.StringToUUID(organizationID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).SetUserOrganization(ctx, sqlc.SetUserOrganizationParams{
		UserID:         pgUserID,
		OrganizationID: pgOrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrOrganizationAlreadySet
	}
	if err != nil {
		return nil, err
	}
	return userFromSQL(row), nil
}

func (r *userRepository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func userFromSQL(row sqlc.UsersUser) *entity.User {
	return &entity.User{
		ID:             pgxconv.UUIDToString(row.ID),
		OrganizationID: pgxconv.PgUUIDtoString(row.OrganizationID),
		Username:       row.Username,
		Email:          row.Email,
		PasswordHash:   row.PasswordHash,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}
