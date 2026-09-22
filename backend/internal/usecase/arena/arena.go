package arena

import (
	"context"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type repository interface {
	CreateAssignment(ctx context.Context, params entity.CreateArenaAssignment) (*entity.ArenaAssignment, error)
	GetAssignment(ctx context.Context, id string) (*entity.ArenaAssignment, error)
	ListAssignments(ctx context.Context, curatorID string, limit, offset int32) ([]entity.ArenaAssignment, error)
	ListUserAssignments(ctx context.Context, userID string, limit, offset int32) ([]entity.ArenaAssignment, error)
	CreateSession(ctx context.Context, assignmentID string) (*entity.ArenaSession, error)
	GetSession(ctx context.Context, id string) (*entity.ArenaSession, error)
	UpdateJoin(ctx context.Context, sessionID string, joined []entity.ArenaSessionParticipant, status entity.ArenaSessionStatus, startedAt *time.Time) (*entity.ArenaSession, error)
	CompleteSession(ctx context.Context, sessionID string, assignmentID string) error
	CreateMessage(ctx context.Context, sessionID, senderID, text, code string) (*entity.ArenaMessage, error)
	ListMessages(ctx context.Context, sessionID string) ([]entity.ArenaMessage, error)
	CreateFeedback(ctx context.Context, feedback entity.ArenaFeedback) (*entity.ArenaFeedback, error)
	ListFeedback(ctx context.Context, sessionID string) ([]entity.ArenaFeedback, error)
	ListUserHistory(ctx context.Context, userID string, limit, offset int32) ([]entity.UserArenaHistoryItem, error)
}

type userRepository interface {
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
}

type scenarioRepository interface {
	GetScenarioByID(ctx context.Context, id string) (*entity.Scenario, error)
}

type profileRecalculator interface {
	RecalculateAfterArenaFeedback(ctx context.Context, userID, sourceID string, ratings map[string]int) ([]entity.ProfileClusterChange, error)
}

type transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

type Service struct {
	repository         repository
	userRepository     userRepository
	scenarioRepository scenarioRepository
	profile            profileRecalculator
	transactor         transactor
}

func NewService(repository repository, userRepository userRepository, scenarioRepository scenarioRepository, profile profileRecalculator, transactor transactor) *Service {
	return &Service{repository: repository, userRepository: userRepository, scenarioRepository: scenarioRepository, profile: profile, transactor: transactor}
}

func (s *Service) ListAssignments(ctx context.Context, claims entity.AccessClaims, limit, offset int32) ([]entity.ArenaAssignment, error) {
	if err := requireCurator(claims); err != nil {
		return nil, err
	}
	return s.repository.ListAssignments(ctx, claims.UserID, limit, offset)
}

func (s *Service) CreateAssignment(ctx context.Context, claims entity.AccessClaims, params entity.CreateArenaAssignment) (*entity.ArenaAssignment, error) {
	if err := requireCurator(claims); err != nil {
		return nil, err
	}
	if len(params.Participants) != 2 {
		return nil, entity.ErrArenaFeedbackParticipantInvalid
	}
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, params.ScenarioID)
	if err != nil {
		return nil, err
	}
	if scenario.Status != entity.ScenarioStatusPublished {
		return nil, entity.ErrScenarioNotAvailable
	}
	params.CreatedBy = claims.UserID
	if params.Channel == "" {
		params.Channel = "chat"
	}
	assignment, err := s.repository.CreateAssignment(ctx, params)
	if err != nil {
		return nil, err
	}
	assignment.ScenarioTitle = scenario.Title
	return assignment, nil
}

func (s *Service) GetAssignment(ctx context.Context, claims entity.AccessClaims, id string) (*entity.ArenaAssignment, error) {
	assignment, err := s.repository.GetAssignment(ctx, id)
	if err != nil {
		return nil, err
	}
	if assignment.CreatedBy == claims.UserID || assignmentParticipantRole(assignment.Participants, claims.UserID) != "" {
		return assignment, nil
	}
	return nil, entity.ErrArenaAccessDenied
}

func (s *Service) Join(ctx context.Context, claims entity.AccessClaims, assignmentID string) (*entity.ArenaSession, error) {
	assignment, err := s.repository.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	roleID := assignmentParticipantRole(assignment.Participants, claims.UserID)
	if roleID == "" {
		return nil, entity.ErrArenaJoinNotAllowed
	}
	sessionID := ""
	if assignment.SessionID == nil {
		session, err := s.repository.CreateSession(ctx, assignment.ID)
		if err != nil {
			return nil, err
		}
		sessionID = session.ID
	} else {
		sessionID = *assignment.SessionID
	}
	session, err := s.repository.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	joined := upsertJoined(session.Joined, entity.ArenaSessionParticipant{UserID: claims.UserID, ScenarioRoleID: roleID, JoinedAt: &now})
	status := entity.ArenaSessionWaiting
	var startedAt *time.Time
	if len(joined) >= len(session.Participants) {
		status = entity.ArenaSessionInProgress
		startedAt = &now
	}
	return s.repository.UpdateJoin(ctx, session.ID, joined, status, startedAt)
}

func (s *Service) GetSession(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.ArenaSession, error) {
	session, err := s.repository.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.CreatedBy == claims.UserID || sessionParticipantRole(session.Participants, claims.UserID) != "" {
		return session, nil
	}
	return nil, entity.ErrArenaAccessDenied
}

func (s *Service) ListMessages(ctx context.Context, claims entity.AccessClaims, sessionID string) ([]entity.ArenaMessage, error) {
	if _, err := s.GetSession(ctx, claims, sessionID); err != nil {
		return nil, err
	}
	return s.repository.ListMessages(ctx, sessionID)
}

func (s *Service) SendMessage(ctx context.Context, claims entity.AccessClaims, sessionID, text, code string) (*entity.ArenaMessage, error) {
	session, err := s.GetSession(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.ArenaSessionInProgress {
		return nil, entity.ErrInvalidArenaState
	}
	if sessionParticipantRole(session.Participants, claims.UserID) == "" {
		return nil, entity.ErrArenaAccessDenied
	}
	return s.repository.CreateMessage(ctx, sessionID, claims.UserID, text, code)
}

func (s *Service) Complete(ctx context.Context, claims entity.AccessClaims, sessionID string) error {
	session, err := s.GetSession(ctx, claims, sessionID)
	if err != nil {
		return err
	}
	if session.Status != entity.ArenaSessionInProgress {
		return entity.ErrInvalidArenaState
	}
	return s.repository.CompleteSession(ctx, session.ID, session.AssignmentID)
}

func (s *Service) ListFeedback(ctx context.Context, claims entity.AccessClaims, sessionID string) ([]entity.ArenaFeedback, error) {
	if _, err := s.GetSession(ctx, claims, sessionID); err != nil {
		return nil, err
	}
	return s.repository.ListFeedback(ctx, sessionID)
}

func (s *Service) CreateFeedback(ctx context.Context, claims entity.AccessClaims, sessionID, participantID, comment string, ratings []entity.ArenaFeedbackRating) (*entity.ArenaFeedback, error) {
	if err := requireCurator(claims); err != nil {
		return nil, entity.ErrArenaFeedbackForbidden
	}
	session, err := s.repository.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.CreatedBy != claims.UserID {
		return nil, entity.ErrArenaFeedbackForbidden
	}
	if sessionParticipantRole(session.Participants, participantID) == "" {
		return nil, entity.ErrArenaFeedbackParticipantInvalid
	}
	var feedback *entity.ArenaFeedback
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		ratingMap := map[string]int{}
		for _, rating := range ratings {
			ratingMap[rating.ClusterID] = rating.Score
		}
		changes, err := s.profile.RecalculateAfterArenaFeedback(txCtx, participantID, session.ID, ratingMap)
		if err != nil {
			return err
		}
		feedback, err = s.repository.CreateFeedback(txCtx, entity.ArenaFeedback{SessionID: session.ID, ParticipantID: participantID, CuratorID: claims.UserID, Comment: comment, Ratings: ratings, ProfileChange: changes})
		return err
	})
	if err != nil {
		return nil, err
	}
	return feedback, nil
}

func (s *Service) ListUserAssignments(ctx context.Context, claims entity.AccessClaims, userID string, limit, offset int32) ([]entity.ArenaAssignment, error) {
	if claims.UserID != userID && claims.ActiveRole != entity.RoleCurator {
		return nil, entity.ErrArenaAccessDenied
	}
	return s.repository.ListUserAssignments(ctx, userID, limit, offset)
}

func (s *Service) ListUserHistory(ctx context.Context, claims entity.AccessClaims, userID string, limit, offset int32) ([]entity.UserArenaHistoryItem, error) {
	if claims.UserID != userID && claims.ActiveRole != entity.RoleCurator {
		return nil, entity.ErrArenaAccessDenied
	}
	return s.repository.ListUserHistory(ctx, userID, limit, offset)
}

func (s *Service) UserHistoryDetails(ctx context.Context, claims entity.AccessClaims, userID, sessionID string) (*entity.ArenaSession, []entity.ArenaMessage, []entity.ArenaFeedback, error) {
	if claims.UserID != userID && claims.ActiveRole != entity.RoleCurator {
		return nil, nil, nil, entity.ErrArenaAccessDenied
	}
	session, err := s.repository.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, nil, err
	}
	if sessionParticipantRole(session.Participants, userID) == "" {
		return nil, nil, nil, entity.ErrArenaAccessDenied
	}
	messages, err := s.repository.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, nil, nil, err
	}
	feedback, err := s.repository.ListFeedback(ctx, sessionID)
	if err != nil {
		return nil, nil, nil, err
	}
	return session, messages, feedback, nil
}

func requireCurator(claims entity.AccessClaims) error {
	if claims.ActiveRole != entity.RoleCurator {
		return entity.ErrCuratorRoleRequired
	}
	return nil
}

func assignmentParticipantRole(participants []entity.ArenaParticipant, userID string) string {
	for _, participant := range participants {
		if participant.UserID == userID {
			return participant.ScenarioRoleID
		}
	}
	return ""
}

func sessionParticipantRole(participants []entity.ArenaSessionParticipant, userID string) string {
	for _, participant := range participants {
		if participant.UserID == userID {
			return participant.ScenarioRoleID
		}
	}
	return ""
}

func upsertJoined(items []entity.ArenaSessionParticipant, next entity.ArenaSessionParticipant) []entity.ArenaSessionParticipant {
	for i := range items {
		if items[i].UserID == next.UserID {
			return items
		}
	}
	return append(items, next)
}
