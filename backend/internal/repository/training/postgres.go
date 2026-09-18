package training

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/training/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type trainingRepository struct{ queries *sqlc.Queries }

func NewTrainingRepository(qdb sqlc.DBTX) *trainingRepository {
	return &trainingRepository{queries: sqlc.New(qdb)}
}

func (r *trainingRepository) CreateSession(ctx context.Context, session entity.TrainingSession) (*entity.TrainingSession, error) {
	params := sqlc.CreateTrainingSessionParams{
		ID: uuid(session.ID), UserID: uuid(session.UserID), ScenarioID: uuid(session.ScenarioID),
		ScenarioVersionID: uuidPtr(session.ScenarioVersionID), ScenarioSnapshot: session.ScenarioSnapshot,
		Mode: sqlc.TrainingSessionMode(session.Mode), Channel: sqlc.TrainingChannel(session.Channel),
		IsPreview: session.IsPreview, CurrentNodeID: pgxconv.StringPtrToText(session.CurrentNodeID),
	}
	row, err := r.queriesFor(ctx).CreateTrainingSession(ctx, params)
	if err != nil {
		return nil, err
	}
	return trainingSessionFromSQL(row), nil
}

func (r *trainingRepository) GetSession(ctx context.Context, id string) (*entity.TrainingSession, error) {
	row, err := r.queriesFor(ctx).GetTrainingSessionByID(ctx, uuid(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTrainingSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return trainingSessionFromSQL(row), nil
}

func (r *trainingRepository) UpdateSessionState(ctx context.Context, id string, currentNodeID *string, status entity.TrainingStatus) (*entity.TrainingSession, error) {
	row, err := r.queriesFor(ctx).UpdateTrainingSessionState(ctx, sqlc.UpdateTrainingSessionStateParams{ID: uuid(id), CurrentNodeID: pgxconv.StringPtrToText(currentNodeID), Status: toSQLStatus(status)})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTrainingSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return trainingSessionFromSQL(row), nil
}

func (r *trainingRepository) CompleteSession(ctx context.Context, id string, outcome, analysis []byte) (*entity.TrainingSession, error) {
	row, err := r.queriesFor(ctx).CompleteTrainingSession(ctx, sqlc.CompleteTrainingSessionParams{ID: uuid(id), Outcome: outcome, Analysis: analysis})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrInvalidTrainingState
	}
	if err != nil {
		return nil, err
	}
	return trainingSessionFromSQL(row), nil
}

func (r *trainingRepository) CreateEvent(ctx context.Context, event entity.TrainingEvent) (*entity.TrainingEvent, error) {
	row, err := r.queriesFor(ctx).CreateTrainingEvent(ctx, sqlc.CreateTrainingEventParams{
		ID: uuid(event.ID), SessionID: uuid(event.SessionID), ClientEventID: uuid(event.ClientEventID),
		SequenceNumber: event.SequenceNumber, NodeID: event.NodeID, ActionID: event.ActionID,
		SemanticAction: event.SemanticAction, NextNodeID: pgxconv.StringPtrToText(event.NextNodeID),
	})
	if err != nil {
		return nil, err
	}
	return trainingEventFromSQL(row), nil
}

func (r *trainingRepository) CountEvents(ctx context.Context, sessionID string) (int32, error) {
	return r.queriesFor(ctx).CountTrainingEvents(ctx, uuid(sessionID))
}

func (r *trainingRepository) ListEvents(ctx context.Context, sessionID string) ([]entity.TrainingEvent, error) {
	rows, err := r.queriesFor(ctx).ListTrainingEvents(ctx, uuid(sessionID))
	if err != nil {
		return nil, err
	}
	items := make([]entity.TrainingEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, *trainingEventFromSQL(row))
	}
	return items, nil
}

func (r *trainingRepository) CreateMessage(ctx context.Context, message entity.TrainingMessage) (*entity.TrainingMessage, error) {
	row, err := r.queriesFor(ctx).CreateTrainingMessage(ctx, sqlc.CreateTrainingMessageParams{ID: uuid(message.ID), SessionID: uuid(message.SessionID), Author: message.Author, Text: message.Text})
	if err != nil {
		return nil, err
	}
	return trainingMessageFromSQL(row), nil
}

func (r *trainingRepository) ListMessages(ctx context.Context, sessionID string) ([]entity.TrainingMessage, error) {
	rows, err := r.queriesFor(ctx).ListTrainingMessages(ctx, uuid(sessionID))
	if err != nil {
		return nil, err
	}
	items := make([]entity.TrainingMessage, 0, len(rows))
	for _, row := range rows {
		items = append(items, *trainingMessageFromSQL(row))
	}
	return items, nil
}

func (r *trainingRepository) queriesFor(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func uuid(value string) pgtype.UUID { result, _ := pgxconv.StringToUUID(value); return result }
func uuidPtr(value *string) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return uuid(*value)
}
func toSQLStatus(status entity.TrainingStatus) sqlc.TrainingSessionStatus {
	switch status {
	case entity.TrainingStatusReadyToComplete:
		return sqlc.TrainingSessionStatusREADYTOCOMPLETE
	case entity.TrainingStatusCompleted:
		return sqlc.TrainingSessionStatusCOMPLETED
	case entity.TrainingStatusAbandoned:
		return sqlc.TrainingSessionStatusABANDONED
	default:
		return sqlc.TrainingSessionStatusINPROGRESS
	}
}
func fromSQLStatus(status sqlc.TrainingSessionStatus) entity.TrainingStatus {
	switch status {
	case sqlc.TrainingSessionStatusREADYTOCOMPLETE:
		return entity.TrainingStatusReadyToComplete
	case sqlc.TrainingSessionStatusCOMPLETED:
		return entity.TrainingStatusCompleted
	case sqlc.TrainingSessionStatusABANDONED:
		return entity.TrainingStatusAbandoned
	default:
		return entity.TrainingStatusInProgress
	}
}
func trainingSessionFromSQL(row sqlc.TrainingSession) *entity.TrainingSession {
	var snapshot struct {
		Version string `json:"version"`
	}
	_ = json.Unmarshal(row.ScenarioSnapshot, &snapshot)
	return &entity.TrainingSession{ID: pgxconv.UUIDToString(row.ID), UserID: pgxconv.UUIDToString(row.UserID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioVersionID: pgxconv.PgUUIDtoString(row.ScenarioVersionID), ScenarioVersion: snapshot.Version, ScenarioSnapshot: row.ScenarioSnapshot, Mode: entity.TrainingMode(row.Mode), Channel: string(row.Channel), IsPreview: row.IsPreview, Status: fromSQLStatus(row.Status), CurrentNodeID: pgxconv.TextToStringPtr(row.CurrentNodeID), Outcome: row.Outcome, Analysis: row.Analysis, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time, CompletedAt: pgxconv.TimestamptzToTime(row.CompletedAt)}
}
func trainingEventFromSQL(row sqlc.TrainingSessionEvent) *entity.TrainingEvent {
	return &entity.TrainingEvent{ID: pgxconv.UUIDToString(row.ID), SessionID: pgxconv.UUIDToString(row.SessionID), ClientEventID: pgxconv.UUIDToString(row.ClientEventID), SequenceNumber: row.SequenceNumber, NodeID: row.NodeID, ActionID: row.ActionID, SemanticAction: row.SemanticAction, NextNodeID: pgxconv.TextToStringPtr(row.NextNodeID), CreatedAt: row.CreatedAt.Time}
}
func trainingMessageFromSQL(row sqlc.TrainingSessionMessage) *entity.TrainingMessage {
	return &entity.TrainingMessage{ID: pgxconv.UUIDToString(row.ID), SessionID: pgxconv.UUIDToString(row.SessionID), Author: row.Author, Text: row.Text, CreatedAt: row.CreatedAt.Time}
}
