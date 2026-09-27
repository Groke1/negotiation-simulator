package arena

import (
	"context"
	"encoding/json"
	"strings"
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
	CreateMessage(ctx context.Context, params entity.CreateArenaMessage) (*entity.ArenaMessage, error)
	ListMessages(ctx context.Context, sessionID, viewerID string) ([]entity.ArenaMessage, error)
	GetMessage(ctx context.Context, sessionID, messageID string) (*entity.ArenaMessage, error)
	CreateMessageClassification(ctx context.Context, params entity.ClassifyArenaMessage, classification json.RawMessage) (*entity.ArenaMessageClassification, error)
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
	if !validArenaChannel(params.Channel) {
		return nil, entity.ErrInvalidRequest
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
	updated, err := s.repository.UpdateJoin(ctx, session.ID, joined, status, startedAt)
	if err != nil {
		return nil, err
	}
	return s.decorateSessionScenario(ctx, updated, claims.UserID)
}

func (s *Service) GetSession(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.ArenaSession, error) {
	session, err := s.getSessionForAccess(ctx, claims, sessionID)
	if err != nil {
		return nil, err
	}
	return s.decorateSessionScenario(ctx, session, claims.UserID)
}

func (s *Service) getSessionForAccess(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.ArenaSession, error) {
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
	if _, err := s.getSessionForAccess(ctx, claims, sessionID); err != nil {
		return nil, err
	}
	return s.repository.ListMessages(ctx, sessionID, claims.UserID)
}

func (s *Service) SendMessage(ctx context.Context, claims entity.AccessClaims, params entity.CreateArenaMessage) (*entity.ArenaMessage, error) {
	session, err := s.getSessionForAccess(ctx, claims, params.SessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.ArenaSessionInProgress {
		return nil, entity.ErrInvalidArenaState
	}
	if sessionParticipantRole(session.Participants, claims.UserID) == "" {
		return nil, entity.ErrArenaAccessDenied
	}
	if !contentMatchesArenaChannel(params.Content, session.Channel) {
		return nil, entity.ErrInvalidRequest
	}
	if !semanticMarkingMatchesArenaChannel(params.SemanticActionCode, params.SemanticAnnotations, session.Channel) {
		return nil, entity.ErrInvalidRequest
	}
	params.SenderID = claims.UserID
	return s.repository.CreateMessage(ctx, params)
}

func (s *Service) ClassifyMessage(ctx context.Context, claims entity.AccessClaims, params entity.ClassifyArenaMessage) (*entity.ArenaMessageClassification, error) {
	session, err := s.getSessionForAccess(ctx, claims, params.SessionID)
	if err != nil {
		return nil, err
	}
	if sessionParticipantRole(session.Participants, claims.UserID) == "" {
		return nil, entity.ErrArenaAccessDenied
	}
	msg, err := s.repository.GetMessage(ctx, session.ID, params.MessageID)
	if err != nil {
		return nil, err
	}
	if msg.SenderID == claims.UserID {
		return nil, entity.ErrMessageNotClassifiable
	}
	if !semanticMarkingMatchesArenaChannel(params.SemanticActionCode, params.SemanticAnnotations, session.Channel) {
		return nil, entity.ErrInvalidRequest
	}
	if !arenaMessageHasExpectedMarking(msg, session.Channel) {
		return nil, entity.ErrMessageNotClassifiable
	}
	params.ClassifierID = claims.UserID
	raw := arenaClassificationJSON(params, *msg, session.Channel)
	return s.repository.CreateMessageClassification(ctx, params, raw)
}

func (s *Service) Complete(ctx context.Context, claims entity.AccessClaims, sessionID string) error {
	session, err := s.getSessionForAccess(ctx, claims, sessionID)
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
	if session.Status != entity.ArenaSessionCompleted {
		return nil, entity.ErrInvalidArenaState
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
	session, err = s.decorateSessionScenario(ctx, session, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	messages, err := s.repository.ListMessages(ctx, sessionID, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	feedback, err := s.repository.ListFeedback(ctx, sessionID)
	if err != nil {
		return nil, nil, nil, err
	}
	return session, messages, feedback, nil
}

func (s *Service) decorateSessionScenario(ctx context.Context, session *entity.ArenaSession, viewerUserID string) (*entity.ArenaSession, error) {
	scenario, err := s.scenarioRepository.GetScenarioByID(ctx, session.ScenarioID)
	if err != nil {
		return nil, err
	}
	doc := entity.ScenarioDocumentFromScenario(*scenario)
	public := entity.PublicScenarioInfoFromDocument(doc)
	session.Scenario = &public
	if session.ScenarioTitle == "" {
		session.ScenarioTitle = doc.Title
	}
	roleID := sessionParticipantRole(session.Participants, viewerUserID)
	for _, role := range doc.Roles {
		role := role
		if role.ID == roleID {
			session.MyRole = &role
			continue
		}
		if session.OpponentRole == nil {
			session.OpponentRole = &entity.PublicScenarioRole{
				ID:          role.ID,
				Name:        role.Name,
				Description: role.Description,
				Goal:        role.Goal,
				Interests:   role.Interests,
				Constraints: role.Constraints,
			}
		}
	}
	return session, nil
}

func validArenaChannel(channel string) bool {
	switch channel {
	case "chat", "email", "voice_message":
		return true
	default:
		return false
	}
}

func contentMatchesArenaChannel(raw json.RawMessage, channel string) bool {
	if len(raw) == 0 || !json.Valid(raw) {
		return false
	}
	var payload struct {
		Type       string `json:"type"`
		Text       string `json:"text"`
		Subject    string `json:"subject"`
		Body       string `json:"body"`
		MediaID    string `json:"media_id"`
		Transcript string `json:"transcript"`
		SpokenText string `json:"spoken_text"`
	}
	if json.Unmarshal(raw, &payload) != nil || payload.Type != channel {
		return false
	}
	switch channel {
	case "chat":
		return strings.TrimSpace(payload.Text) != ""
	case "email":
		return strings.TrimSpace(payload.Subject) != "" && strings.TrimSpace(payload.Body) != ""
	case "voice_message":
		return strings.TrimSpace(payload.MediaID) != ""
	default:
		return false
	}
}

func semanticMarkingMatchesArenaChannel(code *string, annotations json.RawMessage, channel string) bool {
	switch channel {
	case "chat", "voice_message":
		return code != nil && strings.TrimSpace(*code) != ""
	case "email":
		return len(annotations) > 0 && json.Valid(annotations) && string(annotations) != "null" && string(annotations) != "[]"
	default:
		return false
	}
}

func arenaMessageHasExpectedMarking(message *entity.ArenaMessage, channel string) bool {
	return semanticMarkingMatchesArenaChannel(message.SemanticActionCode, message.SemanticAnnotations, channel)
}

func arenaClassificationJSON(params entity.ClassifyArenaMessage, message entity.ArenaMessage, channel string) json.RawMessage {
	correct := false
	switch channel {
	case "chat", "voice_message":
		correct = params.SemanticActionCode != nil && message.SemanticActionCode != nil && strings.TrimSpace(*params.SemanticActionCode) == strings.TrimSpace(*message.SemanticActionCode)
	case "email":
		correct = equalJSON(params.SemanticAnnotations, message.SemanticAnnotations)
	}
	raw, _ := json.Marshal(entity.ArenaMessageClassification{
		MessageID:                   message.ID,
		SelectedSemanticActionCode:  params.SemanticActionCode,
		ExpectedSemanticActionCode:  message.SemanticActionCode,
		SelectedSemanticAnnotations: nullableRaw(params.SemanticAnnotations),
		ExpectedSemanticAnnotations: nullableRaw(message.SemanticAnnotations),
		Correct:                     correct,
	})
	return raw
}

func equalJSON(left, right json.RawMessage) bool {
	var l any
	var r any
	if json.Unmarshal(left, &l) != nil || json.Unmarshal(right, &r) != nil {
		return false
	}
	lraw, _ := json.Marshal(l)
	rraw, _ := json.Marshal(r)
	return string(lraw) == string(rraw)
}

func nullableRaw(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return raw
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
