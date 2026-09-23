package scenario

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

func (r *scenarioRepository) ListScenariosByOrganization(ctx context.Context, organizationID string, params entity.ListScenarios) ([]entity.Scenario, error) {
	pgOrganizationID, err := pgxconv.StringToUUID(organizationID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	pgUserID, err := pgxconv.StringToUUID(params.UserID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	statusValue := ""
	if params.Status != nil {
		statusValue = string(*params.Status)
	}
	difficultyValue := ""
	if params.Difficulty != nil {
		difficultyValue = string(*params.Difficulty)
	}
	rows, err := r.getQueries(ctx).ListScenariosByOrganization(ctx, sqlc.ListScenariosByOrganizationParams{
		OrganizationID:    pgOrganizationID,
		UserID:            pgUserID,
		Status:            statusValue,
		Difficulty:        difficultyValue,
		Tag:               stringValue(params.Tag),
		Runtime:           stringValue(params.Runtime),
		Channel:           stringValue(params.Channel),
		OwnedByMe:         params.OwnedByMe,
		Favorite:          boolValue(params.Favorite),
		AssignedByCurator: boolValue(params.AssignedByCurator),
		Completed:         boolValue(params.Completed),
		LimitCount:        params.Limit,
		OffsetCount:       params.Offset,
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

func (r *scenarioRepository) AddFavorite(ctx context.Context, userID, scenarioID string) error {
	pgUserID, pgScenarioID, err := userScenarioUUIDs(userID, scenarioID)
	if err != nil {
		return err
	}
	return r.getQueries(ctx).AddScenarioFavorite(ctx, sqlc.AddScenarioFavoriteParams{UserID: pgUserID, ScenarioID: pgScenarioID})
}

func (r *scenarioRepository) RemoveFavorite(ctx context.Context, userID, scenarioID string) error {
	pgUserID, pgScenarioID, err := userScenarioUUIDs(userID, scenarioID)
	if err != nil {
		return err
	}
	return r.getQueries(ctx).RemoveScenarioFavorite(ctx, sqlc.RemoveScenarioFavoriteParams{UserID: pgUserID, ScenarioID: pgScenarioID})
}

func (r *scenarioRepository) GetPersonalStats(ctx context.Context, userID, scenarioID string) (*entity.ScenarioPersonalStats, error) {
	pgUserID, pgScenarioID, err := userScenarioUUIDs(userID, scenarioID)
	if err != nil {
		return nil, err
	}
	row, err := r.getQueries(ctx).GetScenarioPersonalStats(ctx, sqlc.GetScenarioPersonalStatsParams{UserID: pgUserID, ScenarioID: pgScenarioID})
	if err != nil {
		return nil, err
	}
	stats := &entity.ScenarioPersonalStats{
		IsFavorite:          row.IsFavorite,
		IsAssignedByCurator: row.IsAssignedByCurator,
		IsCompleted:         row.IsCompleted,
		AttemptsCount:       int(row.AttemptsCount),
	}
	if row.BestSessionID.Valid && row.BestCompletedAt.Valid {
		stats.BestAttempt = &entity.BestTrainingAttempt{
			SessionID:   pgxconv.UUIDToString(row.BestSessionID),
			Score:       int(row.BestScore),
			CompletedAt: row.BestCompletedAt.Time,
		}
	}
	return stats, nil
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
		PublishedAt:          pgxconv.TimestamptzToTime(row.PublishedAt),
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

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolValue(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *value, Valid: true}
}

func userScenarioUUIDs(userID, scenarioID string) (pgtype.UUID, pgtype.UUID, error) {
	pgUserID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, entity.ErrInvalidRequest
	}
	pgScenarioID, err := pgxconv.StringToUUID(scenarioID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, entity.ErrInvalidRequest
	}
	return pgUserID, pgScenarioID, nil
}
