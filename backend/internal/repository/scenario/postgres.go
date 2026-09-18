package scenario

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/scenario/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type scenarioRepository struct {
	queries *sqlc.Queries
}

func NewScenarioRepository(qdb sqlc.DBTX) *scenarioRepository {
	return &scenarioRepository{queries: sqlc.New(qdb)}
}

func (r *scenarioRepository) CreateScenario(ctx context.Context, scenario entity.Scenario) (*entity.Scenario, error) {
	id, err := pgxconv.StringToUUID(scenario.ID)
	if err != nil {
		return nil, err
	}
	organizationID, err := pgxconv.StringToUUID(valueOfStringPtr(scenario.OrganizationID))
	if err != nil {
		return nil, err
	}
	createdBy, err := pgxconv.StringToUUID(valueOfStringPtr(scenario.CreatedBy))
	if err != nil {
		return nil, err
	}
	row, err := r.getQueries(ctx).CreateScenario(ctx, sqlc.CreateScenarioParams{
		ID:                   id,
		OrganizationID:       organizationID,
		CreatedBy:            createdBy,
		Version:              scenario.Version,
		Title:                scenario.Title,
		ModeSupport:          scenario.ModeSupport,
		Context:              []byte(scenario.Context),
		Roles:                []byte(scenario.Roles),
		Goals:                []byte(scenario.Goals),
		Interests:            []byte(scenario.Interests),
		Constraints:          []byte(scenario.Constraints),
		Parameters:           []byte(scenario.Parameters),
		PrivateInformation:   []byte(scenario.PrivateInformation),
		Difficulty:           sqlc.ScenariosScenarioDifficulty(scenario.Difficulty),
		Competencies:         scenario.Competencies,
		ObservationRules:     []byte(scenario.ObservationRules),
		CompletionConditions: []byte(scenario.CompletionConditions),
		Channels:             scenario.Channels,
		Nodes:                []byte(scenario.Nodes),
	})
	if err != nil {
		return nil, err
	}
	return scenarioFromSQL(row), nil
}

func (r *scenarioRepository) GetScenarioByID(ctx context.Context, id string) (*entity.Scenario, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetScenarioByID(ctx, pgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrScenarioNotFound
	}
	if err != nil {
		return nil, err
	}
	return scenarioFromSQL(row), nil
}

func (r *scenarioRepository) ListScenariosByOrganization(ctx context.Context, organizationID string, status *entity.ScenarioStatus, limit, offset int32) ([]entity.Scenario, error) {
	pgOrganizationID, err := pgxconv.StringToUUID(organizationID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	statusValue := ""
	if status != nil {
		statusValue = string(*status)
	}
	rows, err := r.getQueries(ctx).ListScenariosByOrganization(ctx, sqlc.ListScenariosByOrganizationParams{
		OrganizationID: pgOrganizationID,
		Status:         statusValue,
		LimitCount:     limit,
		OffsetCount:    offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Scenario, 0, len(rows))
	for _, row := range rows {
		items = append(items, *scenarioFromSQL(row))
	}
	return items, nil
}

func (r *scenarioRepository) UpdateScenario(ctx context.Context, scenario entity.Scenario) (*entity.Scenario, error) {
	id, err := pgxconv.StringToUUID(scenario.ID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).UpdateScenario(ctx, sqlc.UpdateScenarioParams{
		ID:                   id,
		Version:              scenario.Version,
		Title:                scenario.Title,
		ModeSupport:          scenario.ModeSupport,
		Context:              []byte(scenario.Context),
		Roles:                []byte(scenario.Roles),
		Goals:                []byte(scenario.Goals),
		Interests:            []byte(scenario.Interests),
		Constraints:          []byte(scenario.Constraints),
		Parameters:           []byte(scenario.Parameters),
		PrivateInformation:   []byte(scenario.PrivateInformation),
		Difficulty:           sqlc.ScenariosScenarioDifficulty(scenario.Difficulty),
		Competencies:         scenario.Competencies,
		ObservationRules:     []byte(scenario.ObservationRules),
		CompletionConditions: []byte(scenario.CompletionConditions),
		Channels:             scenario.Channels,
		Nodes:                []byte(scenario.Nodes),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrScenarioNotFound
	}
	if err != nil {
		return nil, err
	}
	return scenarioFromSQL(row), nil
}

func (r *scenarioRepository) SaveScenarioValidation(ctx context.Context, id string, result []byte) (*entity.Scenario, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).SaveScenarioValidation(ctx, sqlc.SaveScenarioValidationParams{ID: pgID, ValidationResult: result})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrScenarioNotFound
	}
	if err != nil {
		return nil, err
	}
	return scenarioFromSQL(row), nil
}

func (r *scenarioRepository) SetScenarioStatus(ctx context.Context, id string, status entity.ScenarioStatus) (*entity.Scenario, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).SetScenarioStatus(ctx, sqlc.SetScenarioStatusParams{ID: pgID, Status: sqlc.ScenariosScenarioStatus(status)})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrScenarioNotFound
	}
	if err != nil {
		return nil, err
	}
	return scenarioFromSQL(row), nil
}

func (r *scenarioRepository) CreateScenarioVersion(ctx context.Context, scenarioID, version string, snapshot []byte) (*entity.ScenarioVersion, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, err
	}
	pgScenarioID, err := pgxconv.StringToUUID(scenarioID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateScenarioVersion(ctx, sqlc.CreateScenarioVersionParams{ID: pgID, ScenarioID: pgScenarioID, Version: version, Snapshot: snapshot})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "scenario_versions_scenario_id_version_key" {
			return nil, entity.ErrScenarioVersionAlreadyExists
		}
		return nil, err
	}
	return &entity.ScenarioVersion{ID: pgxconv.UUIDToString(row.ID), Version: row.Version, Snapshot: row.Snapshot, PublishedAt: row.PublishedAt.Time}, nil
}

func (r *scenarioRepository) ListScenarioVersions(ctx context.Context, scenarioID string, limit, offset int32) ([]entity.ScenarioVersion, error) {
	pgScenarioID, err := pgxconv.StringToUUID(scenarioID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListScenarioVersions(ctx, sqlc.ListScenarioVersionsParams{ScenarioID: pgScenarioID, LimitCount: limit, OffsetCount: offset})
	if err != nil {
		return nil, err
	}
	versions := make([]entity.ScenarioVersion, 0, len(rows))
	for _, row := range rows {
		versions = append(versions, entity.ScenarioVersion{ID: pgxconv.UUIDToString(row.ID), Version: row.Version, Snapshot: row.Snapshot, PublishedAt: row.PublishedAt.Time})
	}
	return versions, nil
}

func (r *scenarioRepository) GetScenarioVersion(ctx context.Context, scenarioID, version string) (*entity.ScenarioVersion, error) {
	pgScenarioID, err := pgxconv.StringToUUID(scenarioID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetScenarioVersion(ctx, sqlc.GetScenarioVersionParams{ScenarioID: pgScenarioID, Version: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrScenarioVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &entity.ScenarioVersion{ID: pgxconv.UUIDToString(row.ID), Version: row.Version, Snapshot: row.Snapshot, PublishedAt: row.PublishedAt.Time}, nil
}

func (r *scenarioRepository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func scenarioFromSQL(row sqlc.ScenariosScenario) *entity.Scenario {
	return &entity.Scenario{
		ID:                   pgxconv.UUIDToString(row.ID),
		OrganizationID:       pgxconv.PgUUIDtoString(row.OrganizationID),
		CreatedBy:            pgxconv.PgUUIDtoString(row.CreatedBy),
		Version:              row.Version,
		Title:                row.Title,
		Status:               entity.ScenarioStatus(row.Status),
		ModeSupport:          row.ModeSupport,
		Context:              row.Context,
		Roles:                row.Roles,
		Goals:                row.Goals,
		Interests:            row.Interests,
		Constraints:          row.Constraints,
		Parameters:           row.Parameters,
		PrivateInformation:   row.PrivateInformation,
		Difficulty:           entity.ScenarioDifficulty(row.Difficulty),
		Competencies:         row.Competencies,
		ObservationRules:     row.ObservationRules,
		CompletionConditions: row.CompletionConditions,
		Channels:             row.Channels,
		Nodes:                row.Nodes,
		Revision:             row.Revision,
		ValidatedRevision:    int64FromPg(row.ValidatedRevision),
		ValidationResult:     row.ValidationResult,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
	}
}

func int64FromPg(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func valueOfStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
