package curator

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	sqlc "sourcecraft.dev/lct-hackaton-2026.arena/internal/repository/curator/sqlc"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/db"
	"sourcecraft.dev/lct-hackaton-2026.arena/pkg/pgxconv"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(qdb sqlc.DBTX) *Repository {
	return &Repository{queries: sqlc.New(qdb)}
}

func (r *Repository) CreateAssignment(ctx context.Context, params entity.CreateAssignment) (*entity.Assignment, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	pgID, err := pgxconv.StringToUUID(id)
	if err != nil {
		return nil, err
	}
	curatorID, err := pgxconv.StringToUUID(params.CuratorID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	learnerID, err := pgxconv.StringToUUID(params.LearnerID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	scenarioID, err := pgxconv.StringToUUID(params.ScenarioID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).CreateAssignment(ctx, sqlc.CreateAssignmentParams{ID: pgID, CuratorID: curatorID, LearnerID: learnerID, ScenarioID: scenarioID})
	if err != nil {
		if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "ux_assignments_pending" {
				return nil, entity.ErrAssignmentConflict
			}
			if pgErr.Code == "23503" {
				return nil, entity.ErrUserNotFound
			}
		}
		return nil, err
	}
	return assignmentFromSQL(row, ""), nil
}

func (r *Repository) ListAssignments(ctx context.Context, filter entity.AssignmentFilter) ([]entity.Assignment, error) {
	curatorID, err := pgxconv.StringToUUID(filter.CuratorID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListAssignments(ctx, sqlc.ListAssignmentsParams{CuratorID: curatorID, Status: assignmentStatusText(filter.Status), LimitCount: filter.Limit, OffsetCount: filter.Offset})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Assignment, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.Assignment{ID: pgxconv.UUIDToString(row.ID), CuratorID: pgxconv.UUIDToString(row.CuratorID), LearnerID: pgxconv.UUIDToString(row.LearnerID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioTitle: row.ScenarioTitle, Status: entity.AssignmentStatus(row.Status), CreatedAt: row.CreatedAt.Time, CompletedAt: pgxconv.TimestamptzToTime(row.CompletedAt)})
	}
	return items, nil
}

func (r *Repository) ListTeam(ctx context.Context, curatorID string, limit, offset int32) ([]entity.CuratorTeamMember, error) {
	pgCuratorID, err := pgxconv.StringToUUID(curatorID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListCuratorLearners(ctx, sqlc.ListCuratorLearnersParams{CuratorID: pgCuratorID, LimitCount: limit, OffsetCount: offset})
	if err != nil {
		return nil, err
	}
	members := make([]entity.CuratorTeamMember, 0, len(rows))
	for _, row := range rows {
		members = append(members, entity.CuratorTeamMember{UserID: pgxconv.UUIDToString(row.ID), Name: row.Username, SessionsCount: int(row.SessionsCount), LastSessionAt: pgxconv.TimestamptzToTime(row.LastSessionAt), AverageScore: scoreOrNil(row.AverageScore)})
	}
	return members, nil
}

func (r *Repository) ListTeamSessions(ctx context.Context, curatorID string, limit, offset int32) ([]entity.TeamSessionListItem, error) {
	pgCuratorID, err := pgxconv.StringToUUID(curatorID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	rows, err := r.getQueries(ctx).ListTeamTrainingSessions(ctx, sqlc.ListTeamTrainingSessionsParams{CuratorID: pgCuratorID, LimitCount: limit, OffsetCount: offset})
	if err != nil {
		return nil, err
	}
	items := make([]entity.TeamSessionListItem, 0, len(rows))
	for _, row := range rows {
		userIDs := make([]string, 0, len(row.UserIds))
		for _, id := range row.UserIds {
			userIDs = append(userIDs, pgxconv.UUIDToString(id))
		}
		startedAt := row.StartedAt.Time
		items = append(items, entity.TeamSessionListItem{Type: row.Type, SessionID: pgxconv.UUIDToString(row.SessionID), UserIDs: userIDs, ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioTitle: row.ScenarioTitle, Status: row.Status, Score: scoreOrNil(row.Score), StartedAt: &startedAt, CompletedAt: pgxconv.TimestamptzToTime(row.CompletedAt)})
	}
	return items, nil
}

func (r *Repository) TeamAnalytics(ctx context.Context, curatorID string) (*entity.TeamAnalytics, error) {
	pgCuratorID, err := pgxconv.StringToUUID(curatorID)
	if err != nil {
		return nil, entity.ErrInvalidRequest
	}
	row, err := r.getQueries(ctx).TeamAnalytics(ctx, pgCuratorID)
	if err != nil {
		return nil, err
	}
	return &entity.TeamAnalytics{UsersCount: int(row.UsersCount), TrainingSessionsCount: int(row.TrainingSessionsCount), ArenaSessionsCount: int(row.ArenaSessionsCount), PendingAssignmentsCount: int(row.PendingAssignmentsCount), Clusters: []entity.TeamAnalyticsCluster{}}, nil
}

func (r *Repository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := db.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func assignmentFromSQL(row sqlc.LearningAssignment, title string) *entity.Assignment {
	return &entity.Assignment{ID: pgxconv.UUIDToString(row.ID), CuratorID: pgxconv.UUIDToString(row.CuratorID), LearnerID: pgxconv.UUIDToString(row.LearnerID), ScenarioID: pgxconv.UUIDToString(row.ScenarioID), ScenarioTitle: title, Status: entity.AssignmentStatus(row.Status), CreatedAt: row.CreatedAt.Time, CompletedAt: pgxconv.TimestamptzToTime(row.CompletedAt)}
}

func assignmentStatusText(value *entity.AssignmentStatus) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*value), Valid: true}
}

func scoreOrNil(value int32) *int {
	if value < 0 {
		return nil
	}
	v := int(value)
	return &v
}
