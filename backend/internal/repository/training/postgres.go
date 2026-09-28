package training

import (
	"context"
	"errors"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/training/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(qdb sqlc.DBTX) *Repository {
	return &Repository{queries: sqlc.New(qdb)}
}

func (r *Repository) CreateSession(ctx context.Context, params entity.StartTrainingSession, snapshot, state []byte) (*entity.TrainingSession, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, err
	}
	userID, err := pgxconv.StringToUUID(params.UserID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	scenarioID, err := pgxconv.StringToUUID(params.ScenarioID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	assignmentID, err := optionalUUID(params.AssignmentID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateTrainingSession(ctx, sqlc.CreateTrainingSessionParams{
		ID:             pgID,
		UserID:         userID,
		ScenarioID:     scenarioID,
		ScenarioRoleID: params.ScenarioRoleID,
		Mode:           string(params.Mode),
		Channel:        string(params.Channel),
		Status:         string(entity.TrainingStatusInProgress),
		AssignmentID:   assignmentID,
		Snapshot:       snapshot,
		State:          state,
	})
	if err != nil {
		return nil, err
	}
	return sessionFromSQL(row, ""), nil
}

func (r *Repository) GetSession(ctx context.Context, id string) (*entity.TrainingSession, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetTrainingSession(ctx, pgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTrainingSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return sessionFromGetRow(row), nil
}

func (r *Repository) ListSessions(ctx context.Context, filter entity.TrainingSessionFilter) ([]entity.TrainingSession, error) {
	userID, err := pgxconv.StringToUUID(filter.UserID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	scenarioID, err := optionalUUID(filter.ScenarioID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListTrainingSessions(ctx, sqlc.ListTrainingSessionsParams{
		UserID:      userID,
		ScenarioID:  scenarioID,
		Status:      optionalTrainingStatus(filter.Status),
		Mode:        optionalTrainingMode(filter.Mode),
		Channel:     optionalTrainingChannel(filter.Channel),
		LimitCount:  filter.Limit,
		OffsetCount: filter.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.TrainingSession, 0, len(rows))
	for _, row := range rows {
		items = append(items, *sessionFromListRow(row))
	}
	return items, nil
}

func (r *Repository) UpdateState(ctx context.Context, sessionID string, state []byte, status entity.TrainingStatus) (*entity.TrainingSession, error) {
	pgID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).UpdateTrainingState(ctx, sqlc.UpdateTrainingStateParams{ID: pgID, State: state, Status: string(status)})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTrainingSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return sessionFromSQL(row, ""), nil
}

func (r *Repository) Complete(ctx context.Context, sessionID string, score float64, analysis []byte) (*entity.TrainingSession, error) {
	pgID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CompleteTrainingSession(ctx, sqlc.CompleteTrainingSessionParams{ID: pgID, Score: pgtype.Float8{Float64: score, Valid: true}, Analysis: analysis})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTrainingSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return sessionFromSQL(row, ""), nil
}

func (r *Repository) SetAnalysis(ctx context.Context, sessionID string, analysis []byte) (*entity.TrainingSession, error) {
	pgID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).SetTrainingAnalysis(ctx, sqlc.SetTrainingAnalysisParams{ID: pgID, Analysis: analysis})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTrainingSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return sessionFromSQL(row, ""), nil
}

func (r *Repository) CompleteLinkedAssignment(ctx context.Context, sessionID string) error {
	pgID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return entity.ErrInvalidRequest
	}
	return r.getQueries(ctx).CompleteLinkedAssignment(ctx, pgID)
}

func (r *Repository) CreateMessage(ctx context.Context, params entity.CreateTrainingMessage) (*entity.TrainingMessage, error) {
	id := params.ID
	if id == "" {
		newID, err := entity.NewUUID()
		if err != nil {
			return nil, err
		}
		id = newID
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	sessionID, err := pgxconv.StringToUUID(params.SessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateTrainingMessage(ctx, sqlc.CreateTrainingMessageParams{
		ID:                  pgID,
		SessionID:           sessionID,
		Author:              params.Author,
		Content:             params.Content,
		SemanticActionCode:  optionalText(params.SemanticActionCode),
		SemanticAnnotations: nullableJSON(params.SemanticAnnotations),
		ExpectedActionCode:  optionalText(params.ExpectedActionCode),
		Classification:      nullableJSON(params.Classification),
	})
	if err != nil {
		return nil, err
	}
	return messageFromSQL(row), nil
}

func (r *Repository) ListMessages(ctx context.Context, sessionID string) ([]entity.TrainingMessage, error) {
	pgID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListTrainingMessages(ctx, pgID)
	if err != nil {
		return nil, err
	}
	messages := make([]entity.TrainingMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, *messageFromSQL(row))
	}
	return messages, nil
}

func (r *Repository) GetMessage(ctx context.Context, sessionID, messageID string) (*entity.TrainingMessage, error) {
	pgSessionID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	pgMessageID, err := pgxconv.StringToUUID(messageID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetTrainingMessage(ctx, sqlc.GetTrainingMessageParams{SessionID: pgSessionID, ID: pgMessageID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTrainingMessageNotFound
	}
	if err != nil {
		return nil, err
	}
	return messageFromSQL(row), nil
}

func (r *Repository) UpdateMessageClassification(ctx context.Context, sessionID, messageID string, classification []byte) (*entity.TrainingMessage, error) {
	pgSessionID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	pgMessageID, err := pgxconv.StringToUUID(messageID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).UpdateTrainingMessageClassification(ctx, sqlc.UpdateTrainingMessageClassificationParams{SessionID: pgSessionID, ID: pgMessageID, Classification: classification})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTrainingMessageNotFound
	}
	if err != nil {
		return nil, err
	}
	return messageFromSQL(row), nil
}

func (r *Repository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func sessionFromSQL(row sqlc.LearningTrainingSession, title string) *entity.TrainingSession {
	return &entity.TrainingSession{
		ID:             pgxconv.UUIDToString(row.ID),
		UserID:         pgxconv.UUIDToString(row.UserID),
		ScenarioID:     pgxconv.UUIDToString(row.ScenarioID),
		ScenarioTitle:  title,
		ScenarioRoleID: row.ScenarioRoleID,
		Mode:           entity.TrainingMode(row.Mode),
		Channel:        entity.TrainingChannel(row.Channel),
		Status:         entity.TrainingStatus(row.Status),
		AssignmentID:   pgxconv.PgUUIDtoString(row.AssignmentID),
		Snapshot:       row.Snapshot,
		State:          row.State,
		Analysis:       row.Analysis,
		Score:          floatScoreToInt(row.Score),
		CreatedAt:      row.CreatedAt.Time,
		CompletedAt:    pgxconv.TimestamptzToTime(row.CompletedAt),
	}
}

func sessionFromGetRow(row sqlc.GetTrainingSessionRow) *entity.TrainingSession {
	base := sqlc.LearningTrainingSession{ID: row.ID, UserID: row.UserID, ScenarioID: row.ScenarioID, ScenarioRoleID: row.ScenarioRoleID, Mode: row.Mode, Channel: row.Channel, Status: row.Status, AssignmentID: row.AssignmentID, Snapshot: row.Snapshot, State: row.State, Analysis: row.Analysis, Score: row.Score, CreatedAt: row.CreatedAt, CompletedAt: row.CompletedAt}
	return sessionFromSQL(base, row.ScenarioTitle)
}

func sessionFromListRow(row sqlc.ListTrainingSessionsRow) *entity.TrainingSession {
	base := sqlc.LearningTrainingSession{ID: row.ID, UserID: row.UserID, ScenarioID: row.ScenarioID, ScenarioRoleID: row.ScenarioRoleID, Mode: row.Mode, Channel: row.Channel, Status: row.Status, AssignmentID: row.AssignmentID, Snapshot: row.Snapshot, State: row.State, Analysis: row.Analysis, Score: row.Score, CreatedAt: row.CreatedAt, CompletedAt: row.CompletedAt}
	return sessionFromSQL(base, row.ScenarioTitle)
}

func messageFromSQL(row sqlc.LearningTrainingMessage) *entity.TrainingMessage {
	return &entity.TrainingMessage{
		ID:                  pgxconv.UUIDToString(row.ID),
		SessionID:           pgxconv.UUIDToString(row.SessionID),
		Author:              row.Author,
		Content:             row.Content,
		SemanticActionCode:  pgxconv.TextToStringPtr(row.SemanticActionCode),
		SemanticAnnotations: row.SemanticAnnotations,
		ExpectedActionCode:  pgxconv.TextToStringPtr(row.ExpectedActionCode),
		Classification:      row.Classification,
		CreatedAt:           row.CreatedAt.Time,
	}
}

func optionalUUID(value *string) (pgtype.UUID, error) {
	if value == nil || *value == "" {
		return pgtype.UUID{}, nil
	}
	return pgxconv.StringToUUID(*value)
}

func optionalText(value *string) pgtype.Text {
	if value == nil || *value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalTrainingStatus(value *entity.TrainingStatus) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	raw := string(*value)
	return optionalText(&raw)
}

func optionalTrainingMode(value *entity.TrainingMode) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	raw := string(*value)
	return optionalText(&raw)
}

func optionalTrainingChannel(value *entity.TrainingChannel) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	raw := string(*value)
	return optionalText(&raw)
}

func nullableJSON(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return value
}

func floatScoreToInt(value pgtype.Float8) *int {
	if !value.Valid {
		return nil
	}
	v := int(math.Floor(value.Float64))
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return &v
}
