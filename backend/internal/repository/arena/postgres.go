package arena

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/arena/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type arenaRepository struct{ queries *sqlc.Queries }

func NewArenaRepository(qdb sqlc.DBTX) *arenaRepository {
	return &arenaRepository{queries: sqlc.New(qdb)}
}

func (r *arenaRepository) CreateAssignment(ctx context.Context, assignment entity.ArenaAssignment) (*entity.ArenaAssignment, error) {
	row, err := r.queriesFor(ctx).CreateArenaAssignment(ctx, sqlc.CreateArenaAssignmentParams{ID: uuid(assignment.ID), OrganizationID: uuid(assignment.OrganizationID), ScenarioID: uuid(assignment.ScenarioID), ScenarioVersionID: uuid(assignment.ScenarioVersionID), ScenarioTitle: assignment.ScenarioTitle, ScenarioVersion: assignment.ScenarioVersion, CreatedBy: uuid(assignment.CreatedBy), ScheduledAt: pgxconv.TimeToRequiredTimestamptz(assignment.ScheduledAt)})
	if err != nil {
		return nil, err
	}
	for _, participant := range assignment.Participants {
		if err := r.queriesFor(ctx).CreateArenaAssignmentParticipant(ctx, sqlc.CreateArenaAssignmentParticipantParams{AssignmentID: row.ID, UserID: uuid(participant.UserID), Role: sqlc.ArenaParticipantRole(participant.Role)}); err != nil {
			return nil, err
		}
	}
	assignment.ID = pgxconv.UUIDToString(row.ID)
	assignment.Status = entity.ArenaAssignmentStatus(row.Status)
	assignment.CreatedAt = row.CreatedAt.Time
	return &assignment, nil
}

func (r *arenaRepository) GetAssignment(ctx context.Context, id string) (*entity.ArenaAssignment, error) {
	row, err := r.queriesFor(ctx).GetArenaAssignment(ctx, uuid(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrArenaAssignmentNotFound
	}
	if err != nil {
		return nil, err
	}
	participants, err := r.assignmentParticipants(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return assignmentFromRow(row, participants), nil
}

func (r *arenaRepository) ListUserAssignments(ctx context.Context, userID string) ([]entity.ArenaAssignment, error) {
	rows, err := r.queriesFor(ctx).ListUserArenaAssignments(ctx, uuid(userID))
	if err != nil {
		return nil, err
	}
	items := make([]entity.ArenaAssignment, 0, len(rows))
	for _, row := range rows {
		participants, err := r.assignmentParticipants(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, *assignmentFromListRow(row, participants))
	}
	return items, nil
}

func (r *arenaRepository) ListOrganizationAssignments(ctx context.Context, organizationID string) ([]entity.ArenaAssignment, error) {
	rows, err := r.queriesFor(ctx).ListOrganizationArenaAssignments(ctx, uuid(organizationID))
	if err != nil {
		return nil, err
	}
	items := make([]entity.ArenaAssignment, 0, len(rows))
	for _, row := range rows {
		participants, err := r.assignmentParticipants(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, *assignmentFromOrganizationListRow(row, participants))
	}
	return items, nil
}

func (r *arenaRepository) CreateSession(ctx context.Context, assignmentID string) (*entity.ArenaSession, error) {
	row, err := r.queriesFor(ctx).CreateArenaSession(ctx, sqlc.CreateArenaSessionParams{ID: uuid(newID()), AssignmentID: uuid(assignmentID)})
	if err != nil {
		return nil, err
	}
	return r.GetSession(ctx, pgxconv.UUIDToString(row.ID))
}

func (r *arenaRepository) GetSession(ctx context.Context, id string) (*entity.ArenaSession, error) {
	row, err := r.queriesFor(ctx).GetArenaSession(ctx, uuid(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrArenaSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	participants, err := r.sessionParticipants(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return sessionFromRow(row, participants), nil
}

func (r *arenaRepository) AddSessionParticipant(ctx context.Context, sessionID string, participant entity.ArenaParticipant) error {
	return r.queriesFor(ctx).AddArenaSessionParticipant(ctx, sqlc.AddArenaSessionParticipantParams{SessionID: uuid(sessionID), UserID: uuid(participant.UserID), Role: sqlc.ArenaParticipantRole(participant.Role)})
}
func (r *arenaRepository) CountSessionParticipants(ctx context.Context, sessionID string) (int32, error) {
	return r.queriesFor(ctx).CountArenaSessionParticipants(ctx, uuid(sessionID))
}
func (r *arenaRepository) MarkSessionInProgress(ctx context.Context, sessionID string) error {
	_, err := r.queriesFor(ctx).SetArenaSessionInProgress(ctx, uuid(sessionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}
func (r *arenaRepository) CompleteSession(ctx context.Context, sessionID string) error {
	row, err := r.queriesFor(ctx).CompleteArenaSession(ctx, uuid(sessionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrInvalidArenaState
	}
	if err != nil {
		return err
	}
	return r.queriesFor(ctx).SetArenaAssignmentStatus(ctx, sqlc.SetArenaAssignmentStatusParams{ID: row.AssignmentID, Status: sqlc.ArenaAssignmentStatusCompleted})
}

func (r *arenaRepository) CreateMessage(ctx context.Context, message entity.ArenaMessage) (*entity.ArenaMessage, error) {
	row, err := r.queriesFor(ctx).CreateArenaMessage(ctx, sqlc.CreateArenaMessageParams{ID: uuid(message.ID), SessionID: uuid(message.SessionID), SenderID: uuid(message.SenderID), Text: message.Text})
	if err != nil {
		return nil, err
	}
	return arenaMessageFromSQL(row), nil
}
func (r *arenaRepository) ListMessages(ctx context.Context, sessionID string) ([]entity.ArenaMessage, error) {
	rows, err := r.queriesFor(ctx).ListArenaMessages(ctx, uuid(sessionID))
	if err != nil {
		return nil, err
	}
	result := make([]entity.ArenaMessage, 0, len(rows))
	for _, row := range rows {
		result = append(result, *arenaMessageFromSQL(row))
	}
	return result, nil
}
func (r *arenaRepository) CreateFeedback(ctx context.Context, feedback entity.ArenaFeedback) (*entity.ArenaFeedback, error) {
	row, err := r.queriesFor(ctx).CreateArenaFeedback(ctx, sqlc.CreateArenaFeedbackParams{ID: uuid(feedback.ID), SessionID: uuid(feedback.SessionID), ParticipantID: uuid(feedback.ParticipantID), ExpertID: uuid(feedback.ExpertID), Comment: feedback.Comment, Ratings: feedback.Ratings})
	if err != nil {
		return nil, err
	}
	return feedbackFromSQL(row), nil
}
func (r *arenaRepository) ListFeedback(ctx context.Context, sessionID string) ([]entity.ArenaFeedback, error) {
	rows, err := r.queriesFor(ctx).ListArenaFeedback(ctx, uuid(sessionID))
	if err != nil {
		return nil, err
	}
	result := make([]entity.ArenaFeedback, 0, len(rows))
	for _, row := range rows {
		result = append(result, *feedbackFromSQL(row))
	}
	return result, nil
}
func (r *arenaRepository) ListUserHistory(ctx context.Context, userID string) ([]entity.ArenaSession, []bool, error) {
	rows, err := r.queriesFor(ctx).ListUserArenaHistory(ctx, uuid(userID))
	if err != nil {
		return nil, nil, err
	}
	sessions := make([]entity.ArenaSession, 0, len(rows))
	feedback := make([]bool, 0, len(rows))
	for _, row := range rows {
		participants, err := r.sessionParticipants(ctx, row.ID)
		if err != nil {
			return nil, nil, err
		}
		sessions = append(sessions, entity.ArenaSession{ID: pgxconv.UUIDToString(row.ID), AssignmentID: pgxconv.UUIDToString(row.AssignmentID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioTitle: row.ScenarioTitle, ScenarioVersion: row.ScenarioVersion, Channel: row.Channel, Status: entity.ArenaSessionStatus(row.Status), ScheduledAt: row.ScheduledAt.Time, Participants: participants, StartedAt: pgxconv.TimestamptzToTime(row.StartedAt), CompletedAt: pgxconv.TimestamptzToTime(row.CompletedAt)})
		feedback = append(feedback, row.HasFeedback)
	}
	return sessions, feedback, nil
}

func (r *arenaRepository) assignmentParticipants(ctx context.Context, assignmentID pgtype.UUID) ([]entity.ArenaParticipant, error) {
	rows, err := r.queriesFor(ctx).ListArenaAssignmentParticipants(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	result := make([]entity.ArenaParticipant, 0, len(rows))
	for _, row := range rows {
		result = append(result, entity.ArenaParticipant{UserID: pgxconv.UUIDToString(row.UserID), Role: entity.ArenaParticipantRole(row.Role)})
	}
	return result, nil
}
func (r *arenaRepository) sessionParticipants(ctx context.Context, sessionID pgtype.UUID) ([]entity.ArenaParticipant, error) {
	rows, err := r.queriesFor(ctx).ListArenaSessionParticipants(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	result := make([]entity.ArenaParticipant, 0, len(rows))
	for _, row := range rows {
		result = append(result, entity.ArenaParticipant{UserID: pgxconv.UUIDToString(row.UserID), Role: entity.ArenaParticipantRole(row.Role), JoinedAt: pgxconv.TimestamptzToTime(row.JoinedAt)})
	}
	return result, nil
}
func (r *arenaRepository) queriesFor(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}
func uuid(value string) pgtype.UUID { id, _ := pgxconv.StringToUUID(value); return id }
func newID() string                 { id, _ := entity.NewUUID(); return id }
func assignmentFromRow(row sqlc.GetArenaAssignmentRow, participants []entity.ArenaParticipant) *entity.ArenaAssignment {
	return &entity.ArenaAssignment{ID: pgxconv.UUIDToString(row.ID), OrganizationID: pgxconv.UUIDToString(row.OrganizationID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioVersionID: pgxconv.UUIDToString(row.ScenarioVersionID), ScenarioTitle: row.ScenarioTitle, ScenarioVersion: row.ScenarioVersion, Participants: participants, ScheduledAt: row.ScheduledAt.Time, Status: entity.ArenaAssignmentStatus(row.Status), SessionID: pgxconv.PgUUIDtoString(row.SessionID), CreatedBy: pgxconv.UUIDToString(row.CreatedBy), CreatedAt: row.CreatedAt.Time}
}
func assignmentFromListRow(row sqlc.ListUserArenaAssignmentsRow, participants []entity.ArenaParticipant) *entity.ArenaAssignment {
	return &entity.ArenaAssignment{ID: pgxconv.UUIDToString(row.ID), OrganizationID: pgxconv.UUIDToString(row.OrganizationID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioVersionID: pgxconv.UUIDToString(row.ScenarioVersionID), ScenarioTitle: row.ScenarioTitle, ScenarioVersion: row.ScenarioVersion, Participants: participants, ScheduledAt: row.ScheduledAt.Time, Status: entity.ArenaAssignmentStatus(row.Status), SessionID: pgxconv.PgUUIDtoString(row.SessionID), CreatedBy: pgxconv.UUIDToString(row.CreatedBy), CreatedAt: row.CreatedAt.Time}
}
func assignmentFromOrganizationListRow(row sqlc.ListOrganizationArenaAssignmentsRow, participants []entity.ArenaParticipant) *entity.ArenaAssignment {
	return &entity.ArenaAssignment{ID: pgxconv.UUIDToString(row.ID), OrganizationID: pgxconv.UUIDToString(row.OrganizationID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioVersionID: pgxconv.UUIDToString(row.ScenarioVersionID), ScenarioTitle: row.ScenarioTitle, ScenarioVersion: row.ScenarioVersion, Participants: participants, ScheduledAt: row.ScheduledAt.Time, Status: entity.ArenaAssignmentStatus(row.Status), SessionID: pgxconv.PgUUIDtoString(row.SessionID), CreatedBy: pgxconv.UUIDToString(row.CreatedBy), CreatedAt: row.CreatedAt.Time}
}
func sessionFromRow(row sqlc.GetArenaSessionRow, participants []entity.ArenaParticipant) *entity.ArenaSession {
	return &entity.ArenaSession{ID: pgxconv.UUIDToString(row.ID), AssignmentID: pgxconv.UUIDToString(row.AssignmentID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioTitle: row.ScenarioTitle, ScenarioVersion: row.ScenarioVersion, Channel: row.Channel, Status: entity.ArenaSessionStatus(row.Status), ScheduledAt: row.ScheduledAt.Time, Participants: participants, StartedAt: pgxconv.TimestamptzToTime(row.StartedAt), CompletedAt: pgxconv.TimestamptzToTime(row.CompletedAt)}
}
func arenaMessageFromSQL(row sqlc.ArenaMessage) *entity.ArenaMessage {
	return &entity.ArenaMessage{ID: pgxconv.UUIDToString(row.ID), SessionID: pgxconv.UUIDToString(row.SessionID), SenderID: pgxconv.UUIDToString(row.SenderID), Text: row.Text, CreatedAt: row.CreatedAt.Time}
}
func feedbackFromSQL(row sqlc.ArenaFeedback) *entity.ArenaFeedback {
	return &entity.ArenaFeedback{ID: pgxconv.UUIDToString(row.ID), SessionID: pgxconv.UUIDToString(row.SessionID), ParticipantID: pgxconv.UUIDToString(row.ParticipantID), ExpertID: pgxconv.UUIDToString(row.ExpertID), Comment: row.Comment, Ratings: row.Ratings, CreatedAt: row.CreatedAt.Time}
}
