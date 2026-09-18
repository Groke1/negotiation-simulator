package arena

import (
	"context"
	"encoding/json"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type arenaRepository interface {
	CreateAssignment(context.Context, entity.ArenaAssignment) (*entity.ArenaAssignment, error)
	GetAssignment(context.Context, string) (*entity.ArenaAssignment, error)
	ListUserAssignments(context.Context, string) ([]entity.ArenaAssignment, error)
	ListOrganizationAssignments(context.Context, string) ([]entity.ArenaAssignment, error)
	CreateSession(context.Context, string) (*entity.ArenaSession, error)
	GetSession(context.Context, string) (*entity.ArenaSession, error)
	AddSessionParticipant(context.Context, string, entity.ArenaParticipant) error
	CountSessionParticipants(context.Context, string) (int32, error)
	MarkSessionInProgress(context.Context, string) error
	CompleteSession(context.Context, string) error
	CreateMessage(context.Context, entity.ArenaMessage) (*entity.ArenaMessage, error)
	ListMessages(context.Context, string) ([]entity.ArenaMessage, error)
	CreateFeedback(context.Context, entity.ArenaFeedback) (*entity.ArenaFeedback, error)
	ListFeedback(context.Context, string) ([]entity.ArenaFeedback, error)
	ListUserHistory(context.Context, string) ([]entity.ArenaSession, []bool, error)
}
type scenarioRepository interface {
	GetScenarioByID(context.Context, string) (*entity.Scenario, error)
	GetScenarioVersion(context.Context, string, string) (*entity.ScenarioVersion, error)
}
type userRepository interface {
	GetUserByID(context.Context, string) (*entity.User, error)
}
type transactor interface {
	WithTx(context.Context, func(context.Context) error) error
}

type Service struct {
	arena     arenaRepository
	scenarios scenarioRepository
	users     userRepository
	tx        transactor
}

func NewService(arena arenaRepository, scenarios scenarioRepository, users userRepository, tx transactor) *Service {
	return &Service{arena: arena, scenarios: scenarios, users: users, tx: tx}
}

func (s *Service) CreateAssignment(ctx context.Context, claims entity.AccessClaims, input entity.CreateArenaAssignment) (*entity.ArenaAssignment, error) {
	if claims.ActiveRole != entity.RoleCurator {
		return nil, entity.ErrCuratorRoleRequired
	}
	if len(input.Participants) != 2 || input.Participants[0].UserID == input.Participants[1].UserID || input.Participants[0].Role == input.Participants[1].Role || !input.Participants[0].Role.Valid() || !input.Participants[1].Role.Valid() {
		return nil, entity.ErrArenaParticipantInvalid
	}
	curator, err := s.users.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if curator.OrganizationID == nil {
		return nil, entity.ErrOrganizationRequired
	}
	scenario, err := s.scenarios.GetScenarioByID(ctx, input.ScenarioID)
	if err != nil {
		return nil, err
	}
	if scenario.Status != entity.ScenarioStatusPublished || scenario.OrganizationID == nil || *scenario.OrganizationID != *curator.OrganizationID {
		return nil, entity.ErrScenarioNotFound
	}
	if !supports(scenario.ModeSupport, "arena") {
		return nil, entity.ErrArenaScenarioNotSupported
	}
	version, err := s.scenarios.GetScenarioVersion(ctx, scenario.ID, scenario.Version)
	if err != nil {
		return nil, err
	}
	for _, participant := range input.Participants {
		user, err := s.users.GetUserByID(ctx, participant.UserID)
		if err != nil {
			return nil, err
		}
		if user.OrganizationID == nil || *user.OrganizationID != *curator.OrganizationID {
			return nil, entity.ErrArenaParticipantInvalid
		}
	}
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	return s.arena.CreateAssignment(ctx, entity.ArenaAssignment{ID: id, OrganizationID: *curator.OrganizationID, ScenarioID: scenario.ID, ScenarioVersionID: version.ID, ScenarioTitle: scenario.Title, ScenarioVersion: version.Version, Participants: input.Participants, ScheduledAt: input.ScheduledAt, CreatedBy: claims.UserID})
}

func (s *Service) GetAssignment(ctx context.Context, claims entity.AccessClaims, assignmentID string) (*entity.ArenaAssignment, error) {
	assignment, err := s.arena.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if !s.assignmentAccess(ctx, assignment, claims) {
		return nil, entity.ErrArenaAccessDenied
	}
	return assignment, nil
}
func (s *Service) UserAssignments(ctx context.Context, claims entity.AccessClaims, userID string) ([]entity.ArenaAssignment, error) {
	if err := s.userAccess(ctx, claims, userID); err != nil {
		return nil, err
	}
	return s.arena.ListUserAssignments(ctx, userID)
}

func (s *Service) OrganizationAssignments(ctx context.Context, claims entity.AccessClaims) ([]entity.ArenaAssignment, error) {
	if claims.ActiveRole != entity.RoleCurator {
		return nil, entity.ErrCuratorRoleRequired
	}
	user, err := s.users.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.OrganizationID == nil {
		return nil, entity.ErrOrganizationRequired
	}
	return s.arena.ListOrganizationAssignments(ctx, *user.OrganizationID)
}

func (s *Service) Join(ctx context.Context, claims entity.AccessClaims, assignmentID string) (*entity.ArenaSession, error) {
	var session *entity.ArenaSession
	err := s.tx.WithTx(ctx, func(tx context.Context) error {
		assignment, err := s.arena.GetAssignment(tx, assignmentID)
		if err != nil {
			return err
		}
		participant, ok := participantFor(assignment.Participants, claims.UserID)
		if !ok {
			return entity.ErrArenaAccessDenied
		}
		if assignment.Status == entity.ArenaAssignmentCancelled || assignment.Status == entity.ArenaAssignmentCompleted {
			return entity.ErrArenaJoinNotAllowed
		}
		session, err = s.arena.CreateSession(tx, assignment.ID)
		if err != nil {
			return err
		}
		if err := s.arena.AddSessionParticipant(tx, session.ID, participant); err != nil {
			return err
		}
		count, err := s.arena.CountSessionParticipants(tx, session.ID)
		if err != nil {
			return err
		}
		if count == 2 {
			if err := s.arena.MarkSessionInProgress(tx, session.ID); err != nil {
				return err
			}
		}
		session, err = s.arena.GetSession(tx, session.ID)
		return err
	})
	return session, err
}

func (s *Service) GetSession(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.ArenaSession, error) {
	session, err := s.arena.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	assignment, err := s.arena.GetAssignment(ctx, session.AssignmentID)
	if err != nil {
		return nil, err
	}
	if !s.assignmentAccess(ctx, assignment, claims) {
		return nil, entity.ErrArenaAccessDenied
	}
	return session, nil
}
func (s *Service) Messages(ctx context.Context, claims entity.AccessClaims, sessionID string) ([]entity.ArenaMessage, error) {
	if _, err := s.GetSession(ctx, claims, sessionID); err != nil {
		return nil, err
	}
	return s.arena.ListMessages(ctx, sessionID)
}
func (s *Service) Complete(ctx context.Context, claims entity.AccessClaims, sessionID string) error {
	if _, err := s.GetSession(ctx, claims, sessionID); err != nil {
		return err
	}
	return s.arena.CompleteSession(ctx, sessionID)
}

func (s *Service) SaveMessage(ctx context.Context, claims entity.AccessClaims, sessionID, messageID, text string) (*entity.ArenaMessage, error) {
	session, err := s.GetSession(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.ArenaSessionProgress {
		return nil, entity.ErrInvalidArenaState
	}
	if strings.TrimSpace(text) == "" {
		return nil, entity.ErrInvalidRequest
	}
	return s.arena.CreateMessage(ctx, entity.ArenaMessage{ID: messageID, SessionID: sessionID, SenderID: claims.UserID, Text: strings.TrimSpace(text)})
}

func (s *Service) Feedback(ctx context.Context, claims entity.AccessClaims, sessionID string) ([]entity.ArenaFeedback, error) {
	if _, err := s.GetSession(ctx, claims, sessionID); err != nil {
		return nil, err
	}
	return s.arena.ListFeedback(ctx, sessionID)
}
func (s *Service) CreateFeedback(ctx context.Context, claims entity.AccessClaims, input entity.CreateArenaFeedback) (*entity.ArenaFeedback, error) {
	if claims.ActiveRole != entity.RoleCurator {
		return nil, entity.ErrCuratorRoleRequired
	}
	session, err := s.arena.GetSession(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.ArenaSessionComplete && session.Status != entity.ArenaSessionReviewed {
		return nil, entity.ErrInvalidArenaState
	}
	if _, ok := participantFor(session.Participants, input.ParticipantID); !ok {
		return nil, entity.ErrArenaFeedbackParticipant
	}
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	return s.arena.CreateFeedback(ctx, entity.ArenaFeedback{ID: id, SessionID: input.SessionID, ParticipantID: input.ParticipantID, ExpertID: claims.UserID, Comment: strings.TrimSpace(input.Comment), Ratings: nonNilJSON(input.Ratings)})
}
func (s *Service) UserHistory(ctx context.Context, claims entity.AccessClaims, userID string) ([]entity.ArenaSession, []bool, error) {
	if err := s.userAccess(ctx, claims, userID); err != nil {
		return nil, nil, err
	}
	return s.arena.ListUserHistory(ctx, userID)
}

func (s *Service) userAccess(ctx context.Context, claims entity.AccessClaims, userID string) error {
	if claims.UserID == userID {
		return nil
	}
	if claims.ActiveRole != entity.RoleCurator {
		return entity.ErrArenaAccessDenied
	}
	curator, err := s.users.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return err
	}
	target, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if curator.OrganizationID == nil || target.OrganizationID == nil || *curator.OrganizationID != *target.OrganizationID {
		return entity.ErrArenaAccessDenied
	}
	return nil
}

func (s *Service) assignmentAccess(ctx context.Context, assignment *entity.ArenaAssignment, claims entity.AccessClaims) bool {
	if assignment.CreatedBy == claims.UserID {
		return true
	}
	if _, ok := participantFor(assignment.Participants, claims.UserID); ok {
		return true
	}
	if claims.ActiveRole != entity.RoleCurator {
		return false
	}
	curator, err := s.users.GetUserByID(ctx, claims.UserID)
	return err == nil && curator.OrganizationID != nil && *curator.OrganizationID == assignment.OrganizationID
}
func supports(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func participantFor(participants []entity.ArenaParticipant, userID string) (entity.ArenaParticipant, bool) {
	for _, participant := range participants {
		if participant.UserID == userID {
			return participant, true
		}
	}
	return entity.ArenaParticipant{}, false
}
func nonNilJSON(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage("[]")
	}
	return value
}
