package user

import (
	"context"
	"errors"

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
