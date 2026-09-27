package arena

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/arena/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(qdb sqlc.DBTX) *Repository {
	return &Repository{queries: sqlc.New(qdb)}
}

func (r *Repository) CreateAssignment(ctx context.Context, params entity.CreateArenaAssignment) (*entity.ArenaAssignment, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, err
	}
	scenarioID, err := pgxconv.StringToUUID(params.ScenarioID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	createdBy, err := pgxconv.StringToUUID(params.CreatedBy)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateArenaAssignment(ctx, sqlc.CreateArenaAssignmentParams{ID: pgID, ScenarioID: scenarioID, CreatedBy: createdBy, Participants: entity.ArenaParticipantsJSON(params.Participants), Channel: params.Channel, ScheduledAt: pgxconv.TimeToRequiredTimestamptz(params.ScheduledAt)})
	if err != nil {
		return nil, err
	}
	return assignmentFromSQL(row, ""), nil
}

func (r *Repository) GetAssignment(ctx context.Context, id string) (*entity.ArenaAssignment, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetArenaAssignment(ctx, pgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrArenaAssignmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return assignmentFromGet(row), nil
}

func (r *Repository) ListAssignments(ctx context.Context, curatorID string, limit, offset int32) ([]entity.ArenaAssignment, error) {
	pgID, err := pgxconv.StringToUUID(curatorID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListArenaAssignments(ctx, sqlc.ListArenaAssignmentsParams{CuratorID: pgID, LimitCount: limit, OffsetCount: offset})
	if err != nil {
		return nil, err
	}
	items := make([]entity.ArenaAssignment, 0, len(rows))
	for _, row := range rows {
		items = append(items, *assignmentFromList(row))
	}
	return items, nil
}

func (r *Repository) ListUserAssignments(ctx context.Context, userID string, limit, offset int32) ([]entity.ArenaAssignment, error) {
	pgID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListUserArenaAssignments(ctx, sqlc.ListUserArenaAssignmentsParams{ParticipantFilter: participantFilter(userID), LimitCount: limit, OffsetCount: offset})
	if err != nil {
		_ = pgID
		return nil, err
	}
	items := make([]entity.ArenaAssignment, 0, len(rows))
	for _, row := range rows {
		items = append(items, *assignmentFromUserList(row))
	}
	return items, nil
}

func (r *Repository) CreateSession(ctx context.Context, assignmentID string) (*entity.ArenaSession, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, err
	}
	pgAssignmentID, err := pgxconv.StringToUUID(assignmentID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateArenaSession(ctx, sqlc.CreateArenaSessionParams{ID: pgID, AssignmentID: pgAssignmentID})
	if err != nil {
		return nil, err
	}
	if _, err := r.getQueries(ctx).SetArenaAssignmentSession(ctx, sqlc.SetArenaAssignmentSessionParams{ID: pgAssignmentID, SessionID: row.ID, Status: string(entity.ArenaAssignmentInProgress)}); err != nil {
		return nil, err
	}
	return r.GetSession(ctx, pgxconv.UUIDToString(row.ID))
}

func (r *Repository) GetSession(ctx context.Context, id string) (*entity.ArenaSession, error) {
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetArenaSession(ctx, pgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrArenaSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return sessionFromGet(row), nil
}

func (r *Repository) UpdateJoin(ctx context.Context, sessionID string, joined []entity.ArenaSessionParticipant, status entity.ArenaSessionStatus, startedAt *time.Time) (*entity.ArenaSession, error) {
	pgID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	raw, _ := json.Marshal(joined)
	row, err := r.getQueries(ctx).UpdateArenaSessionJoin(ctx, sqlc.UpdateArenaSessionJoinParams{ID: pgID, Joined: raw, Status: string(status), StartedAt: pgxconv.TimeToTimestamptz(startedAt)})
	if err != nil {
		return nil, err
	}
	return r.GetSession(ctx, pgxconv.UUIDToString(row.ID))
}

func (r *Repository) CompleteSession(ctx context.Context, sessionID string, assignmentID string) error {
	pgSessionID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return entity.ErrInvalidRequest
	}
	pgAssignmentID, err := pgxconv.StringToUUID(assignmentID)
	if err != nil {
		return entity.ErrInvalidRequest
	}
	if _, err := r.getQueries(ctx).CompleteArenaSession(ctx, pgSessionID); err != nil {
		return err
	}
	return r.getQueries(ctx).CompleteArenaAssignment(ctx, pgAssignmentID)
}

func (r *Repository) CreateMessage(ctx context.Context, params entity.CreateArenaMessage) (*entity.ArenaMessage, error) {
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
	pgSessionID, err := pgxconv.StringToUUID(params.SessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	pgSenderID, err := pgxconv.StringToUUID(params.SenderID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateArenaMessage(ctx, sqlc.CreateArenaMessageParams{
		ID:                  pgID,
		SessionID:           pgSessionID,
		SenderID:            pgSenderID,
		Content:             params.Content,
		SemanticActionCode:  optionalText(params.SemanticActionCode),
		SemanticAnnotations: nullableJSON(params.SemanticAnnotations),
	})
	if err != nil {
		return nil, err
	}
	return messageFromSQL(row), nil
}

func (r *Repository) ListMessages(ctx context.Context, sessionID, viewerID string) ([]entity.ArenaMessage, error) {
	pgID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	viewer, err := pgxconv.StringToUUID(viewerID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListArenaMessages(ctx, sqlc.ListArenaMessagesParams{SessionID: pgID, ViewerID: viewer})
	if err != nil {
		return nil, err
	}
	items := make([]entity.ArenaMessage, 0, len(rows))
	for _, row := range rows {
		items = append(items, *messageFromList(row))
	}
	return items, nil
}

func (r *Repository) GetMessage(ctx context.Context, sessionID, messageID string) (*entity.ArenaMessage, error) {
	pgSessionID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	pgMessageID, err := pgxconv.StringToUUID(messageID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).GetArenaMessage(ctx, sqlc.GetArenaMessageParams{ID: pgMessageID, SessionID: pgSessionID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrArenaMessageNotFound
	}
	if err != nil {
		return nil, err
	}
	return messageFromSQL(row), nil
}

func (r *Repository) CreateMessageClassification(ctx context.Context, params entity.ClassifyArenaMessage, classification json.RawMessage) (*entity.ArenaMessageClassification, error) {
	messageID, err := pgxconv.StringToUUID(params.MessageID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	classifierID, err := pgxconv.StringToUUID(params.ClassifierID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateArenaMessageClassification(ctx, sqlc.CreateArenaMessageClassificationParams{
		MessageID:                   messageID,
		ClassifierID:                classifierID,
		SelectedSemanticActionCode:  optionalText(params.SemanticActionCode),
		SelectedSemanticAnnotations: nullableJSON(params.SemanticAnnotations),
		Classification:              classification,
	})
	if err != nil {
		if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, entity.ErrMessageAlreadyClassified
		}
		return nil, err
	}
	return classificationFromSQL(row), nil
}

func (r *Repository) CreateFeedback(ctx context.Context, feedback entity.ArenaFeedback) (*entity.ArenaFeedback, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	pgID, _ := pgxconv.StringToUUID(id)
	sessionID, err := pgxconv.StringToUUID(feedback.SessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	participantID, err := pgxconv.StringToUUID(feedback.ParticipantID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	curatorID, err := pgxconv.StringToUUID(feedback.CuratorID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	ratings, _ := json.Marshal(feedback.Ratings)
	change, _ := json.Marshal(feedback.ProfileChange)
	row, err := r.getQueries(ctx).CreateArenaFeedback(ctx, sqlc.CreateArenaFeedbackParams{ID: pgID, SessionID: sessionID, ParticipantID: participantID, CuratorID: curatorID, Comment: feedback.Comment, Ratings: ratings, ProfileChange: nullableJSON(change)})
	if err != nil {
		if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, entity.ErrArenaFeedbackConflict
		}
		return nil, err
	}
	return feedbackFromSQL(row), nil
}

func (r *Repository) ListFeedback(ctx context.Context, sessionID string) ([]entity.ArenaFeedback, error) {
	pgID, err := pgxconv.StringToUUID(sessionID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListArenaFeedback(ctx, pgID)
	if err != nil {
		return nil, err
	}
	items := make([]entity.ArenaFeedback, 0, len(rows))
	for _, row := range rows {
		items = append(items, *feedbackFromSQL(row))
	}
	return items, nil
}

func (r *Repository) ListUserHistory(ctx context.Context, userID string, limit, offset int32) ([]entity.UserArenaHistoryItem, error) {
	pgID, err := pgxconv.StringToUUID(userID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListUserArenaHistory(ctx, sqlc.ListUserArenaHistoryParams{UserID: pgID, ParticipantFilter: participantFilter(userID), LimitCount: limit, OffsetCount: offset})
	if err != nil {
		return nil, err
	}
	items := make([]entity.UserArenaHistoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.UserArenaHistoryItem{ArenaSessionID: pgxconv.UUIDToString(row.ArenaSessionID), ArenaAssignmentID: pgxconv.UUIDToString(row.ArenaAssignmentID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioTitle: row.ScenarioTitle, ScheduledAt: row.ScheduledAt.Time, StartedAt: pgxconv.TimestamptzToTime(row.StartedAt), CompletedAt: pgxconv.TimestamptzToTime(row.CompletedAt), Status: entity.ArenaSessionStatus(row.Status), HasFeedback: row.HasFeedback})
	}
	return items, nil
}

func (r *Repository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func assignmentFromSQL(row sqlc.LearningArenaAssignment, title string) *entity.ArenaAssignment {
	participants := []entity.ArenaParticipant{}
	_ = json.Unmarshal(row.Participants, &participants)
	return &entity.ArenaAssignment{ID: pgxconv.UUIDToString(row.ID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioTitle: title, Participants: participants, Channel: row.Channel, ScheduledAt: row.ScheduledAt.Time, Status: entity.ArenaAssignmentStatus(row.Status), SessionID: pgxconv.PgUUIDtoString(row.SessionID), CreatedBy: pgxconv.UUIDToString(row.CreatedBy), CreatedAt: row.CreatedAt.Time}
}

func assignmentFromGet(row sqlc.GetArenaAssignmentRow) *entity.ArenaAssignment {
	base := sqlc.LearningArenaAssignment{ID: row.ID, ScenarioID: row.ScenarioID, CreatedBy: row.CreatedBy, Participants: row.Participants, Channel: row.Channel, ScheduledAt: row.ScheduledAt, Status: row.Status, SessionID: row.SessionID, CreatedAt: row.CreatedAt}
	return assignmentFromSQL(base, row.ScenarioTitle)
}

func assignmentFromList(row sqlc.ListArenaAssignmentsRow) *entity.ArenaAssignment {
	base := sqlc.LearningArenaAssignment{ID: row.ID, ScenarioID: row.ScenarioID, CreatedBy: row.CreatedBy, Participants: row.Participants, Channel: row.Channel, ScheduledAt: row.ScheduledAt, Status: row.Status, SessionID: row.SessionID, CreatedAt: row.CreatedAt}
	return assignmentFromSQL(base, row.ScenarioTitle)
}

func assignmentFromUserList(row sqlc.ListUserArenaAssignmentsRow) *entity.ArenaAssignment {
	base := sqlc.LearningArenaAssignment{ID: row.ID, ScenarioID: row.ScenarioID, CreatedBy: row.CreatedBy, Participants: row.Participants, Channel: row.Channel, ScheduledAt: row.ScheduledAt, Status: row.Status, SessionID: row.SessionID, CreatedAt: row.CreatedAt}
	return assignmentFromSQL(base, row.ScenarioTitle)
}

func sessionFromGet(row sqlc.GetArenaSessionRow) *entity.ArenaSession {
	participants := []entity.ArenaParticipant{}
	joined := []entity.ArenaSessionParticipant{}
	_ = json.Unmarshal(row.Participants, &participants)
	_ = json.Unmarshal(row.Joined, &joined)
	sessionParticipants := make([]entity.ArenaSessionParticipant, 0, len(participants))
	for _, participant := range participants {
		item := entity.ArenaSessionParticipant{UserID: participant.UserID, ScenarioRoleID: participant.ScenarioRoleID}
		for _, joinedParticipant := range joined {
			if joinedParticipant.UserID == participant.UserID {
				item.JoinedAt = joinedParticipant.JoinedAt
				break
			}
		}
		sessionParticipants = append(sessionParticipants, item)
	}
	return &entity.ArenaSession{ID: pgxconv.UUIDToString(row.ID), AssignmentID: pgxconv.UUIDToString(row.AssignmentID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioTitle: row.ScenarioTitle, Channel: row.Channel, Status: entity.ArenaSessionStatus(row.Status), ScheduledAt: row.ScheduledAt.Time, Participants: sessionParticipants, Joined: joined, CreatedBy: pgxconv.UUIDToString(row.CreatedBy), StartedAt: pgxconv.TimestamptzToTime(row.StartedAt), CompletedAt: pgxconv.TimestamptzToTime(row.CompletedAt)}
}

func messageFromSQL(row sqlc.LearningArenaMessage) *entity.ArenaMessage {
	return &entity.ArenaMessage{ID: pgxconv.UUIDToString(row.ID), SessionID: pgxconv.UUIDToString(row.SessionID), SenderID: pgxconv.UUIDToString(row.SenderID), Content: row.Content, SemanticActionCode: pgTextToString(row.SemanticActionCode), SemanticAnnotations: row.SemanticAnnotations, CreatedAt: row.CreatedAt.Time}
}

func messageFromList(row sqlc.ListArenaMessagesRow) *entity.ArenaMessage {
	return &entity.ArenaMessage{ID: pgxconv.UUIDToString(row.ID), SessionID: pgxconv.UUIDToString(row.SessionID), SenderID: pgxconv.UUIDToString(row.SenderID), Content: row.Content, SemanticActionCode: pgTextToString(row.SemanticActionCode), SemanticAnnotations: row.SemanticAnnotations, Classification: nullableRaw(row.Classification), CreatedAt: row.CreatedAt.Time}
}

func classificationFromSQL(row sqlc.LearningArenaMessageClassification) *entity.ArenaMessageClassification {
	classification := entity.ArenaMessageClassification{}
	_ = json.Unmarshal(row.Classification, &classification)
	if classification.MessageID == "" {
		classification.MessageID = pgxconv.UUIDToString(row.MessageID)
	}
	return &classification
}

func feedbackFromSQL(row sqlc.LearningArenaFeedback) *entity.ArenaFeedback {
	ratings := []entity.ArenaFeedbackRating{}
	changes := []entity.ProfileClusterChange{}
	_ = json.Unmarshal(row.Ratings, &ratings)
	_ = json.Unmarshal(row.ProfileChange, &changes)
	return &entity.ArenaFeedback{ID: pgxconv.UUIDToString(row.ID), SessionID: pgxconv.UUIDToString(row.SessionID), ParticipantID: pgxconv.UUIDToString(row.ParticipantID), CuratorID: pgxconv.UUIDToString(row.CuratorID), Comment: row.Comment, Ratings: ratings, ProfileChange: changes, CreatedAt: row.CreatedAt.Time}
}

func participantFilter(userID string) []byte {
	raw, _ := json.Marshal([]map[string]string{{"user_id": userID}})
	return raw
}

func nullableJSON(value []byte) []byte {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return value
}

func nullableRaw(value []byte) json.RawMessage {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return value
}

func optionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func pgTextToString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	out := value.String
	return &out
}
