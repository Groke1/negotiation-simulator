package organizations

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/organizations/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type organizationRepository struct {
	queries *sqlc.Queries
}

func NewOrganizationRepository(qdb sqlc.DBTX) *organizationRepository {
	return &organizationRepository{queries: sqlc.New(qdb)}
}

func (r *organizationRepository) CreateOrganization(ctx context.Context, org entity.Organization) (*entity.Organization, error) {
	id, err := pgxconv.StringToUUID(org.ID)
	if err != nil {
		return nil, err
	}
	row, err := r.getQueries(ctx).CreateOrganization(ctx, sqlc.CreateOrganizationParams{ID: id, Name: org.Name})
	if err != nil {
		if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) && pgErr.ConstraintName == "ux_organizations_name" {
			return nil, entity.ErrOrganizationAlreadyExists
		}
		return nil, err
	}
	return organizationFromSQL(row), nil
}

func (r *organizationRepository) GetOrganizationByID(ctx context.Context, id string) (*entity.Organization, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetOrganizationByID(ctx, pgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrOrganizationNotFound
	}
	if err != nil {
		return nil, err
	}
	return organizationFromSQL(row), nil
}

func (r *organizationRepository) ListOrganizations(ctx context.Context, params entity.ListOrganizations) ([]entity.Organization, error) {
	rows, err := r.getQueries(ctx).ListOrganizations(ctx, sqlc.ListOrganizationsParams{
		Query:       params.Query,
		LimitCount:  params.Limit,
		OffsetCount: params.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Organization, 0, len(rows))
	for _, row := range rows {
		items = append(items, *organizationFromSQL(row))
	}
	return items, nil
}

func (r *organizationRepository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func organizationFromSQL(row sqlc.OrganizationsOrganization) *entity.Organization {
	return &entity.Organization{
		ID:        pgxconv.UUIDToString(row.ID),
		Name:      row.Name,
		CreatedAt: row.CreatedAt.Time,
	}
}
