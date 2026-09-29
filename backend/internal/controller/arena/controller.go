package arena

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	ListAssignments(ctx context.Context, claims entity.AccessClaims, limit, offset int32) ([]entity.ArenaAssignment, error)
	CreateAssignment(ctx context.Context, claims entity.AccessClaims, params entity.CreateArenaAssignment) (*entity.ArenaAssignment, error)
	GetAssignment(ctx context.Context, claims entity.AccessClaims, id string) (*entity.ArenaAssignment, error)
	Join(ctx context.Context, claims entity.AccessClaims, assignmentID string) (*entity.ArenaSession, error)
	GetSession(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.ArenaSession, error)
	ListMessages(ctx context.Context, claims entity.AccessClaims, sessionID string) ([]entity.ArenaMessage, error)
	SendMessage(ctx context.Context, claims entity.AccessClaims, params entity.CreateArenaMessage) (*entity.ArenaMessage, error)
	ClassifyMessage(ctx context.Context, claims entity.AccessClaims, params entity.ClassifyArenaMessage) (*entity.ArenaMessageClassification, error)
	Complete(ctx context.Context, claims entity.AccessClaims, sessionID string) error
	ListFeedback(ctx context.Context, claims entity.AccessClaims, sessionID string) ([]entity.ArenaFeedback, error)
	CreateFeedback(ctx context.Context, claims entity.AccessClaims, sessionID, participantID, comment string, ratings []entity.ArenaFeedbackRating) (*entity.ArenaFeedback, error)
	ListUserAssignments(ctx context.Context, claims entity.AccessClaims, userID string, limit, offset int32) ([]entity.ArenaAssignment, error)
	ListUserHistory(ctx context.Context, claims entity.AccessClaims, userID string, limit, offset int32) ([]entity.UserArenaHistoryItem, error)
	UserHistoryDetails(ctx context.Context, claims entity.AccessClaims, userID, sessionID string) (*entity.ArenaSession, []entity.ArenaMessage, []entity.ArenaFeedback, error)
}

type Controller struct {
	logger   *zap.Logger
	service  Service
	verifier ctrlmw.AccessTokenVerifier
	hub      *hub
}

func NewController(logger *zap.Logger, service Service, verifier ctrlmw.AccessTokenVerifier) *Controller {
	return &Controller{logger: logger, service: service, verifier: verifier, hub: newHub()}
}

func (c *Controller) InitRoutes(r chi.Router) {
	r.Get("/arena/assignments", c.ListAssignments)
	r.Post("/arena/assignments", c.CreateAssignment)
	r.Get("/arena/assignments/{arena_assignment_id}", c.GetAssignment)
	r.Post("/arena/assignments/{arena_assignment_id}/join", c.Join)
	r.Get("/arena/sessions/{arena_session_id}", c.GetSession)
	r.Get("/arena/sessions/{arena_session_id}/presence", c.GetPresence)
	r.Post("/arena/sessions/{arena_session_id}/typing", c.UpdateTyping)
	r.Get("/arena/sessions/{arena_session_id}/messages", c.ListMessages)
	r.Post("/arena/sessions/{arena_session_id}/messages", c.SendMessage)
	r.Post("/arena/sessions/{arena_session_id}/messages/{message_id}/classification", c.ClassifyMessage)
	r.Post("/arena/sessions/{arena_session_id}/complete", c.Complete)
	r.Get("/arena/sessions/{arena_session_id}/feedback", c.ListFeedback)
	r.Post("/arena/sessions/{arena_session_id}/feedback", c.CreateFeedback)
	r.Get("/users/{user_id}/arena-assignments", c.UserAssignments)
	r.Get("/users/{user_id}/arena-history", c.UserHistory)
	r.Get("/users/{user_id}/arena-history/{arena_session_id}", c.UserHistoryDetails)
}

func (c *Controller) InitWebSocketRoutes(r chi.Router) {
	r.Get("/ws/arena/sessions/{arena_session_id}", c.WebSocket)
}

func (c *Controller) ListAssignments(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	limit, offset, ok := readLimitOffset(w, r)
	if !ok {
		return
	}
	items, err := c.service.ListAssignments(r.Context(), claims, limit, offset)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, assignmentListResponse{Arenas: items})
}

func (c *Controller) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	var req createAssignmentRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.ScenarioID) == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	scenarioID, err := entity.NormalizeUUID(req.ScenarioID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	item, err := c.service.CreateAssignment(r.Context(), claims, entity.CreateArenaAssignment{ScenarioID: scenarioID, Participants: req.Participants, Channel: req.Channel, ScheduledAt: req.ScheduledAt})
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, createAssignmentResponse{ArenaAssignmentID: item.ID, Status: item.Status})
}

func (c *Controller) GetAssignment(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_assignment_id")
	if !ok {
		return
	}
	item, err := c.service.GetAssignment(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, item)
}

func (c *Controller) Join(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_assignment_id")
	if !ok {
		return
	}
	session, err := c.service.Join(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	presence := c.presenceResponse(session)
	pkghttpx.WriteJSON(w, http.StatusOK, joinArenaResponse{ArenaSessionID: session.ID, Status: session.Status, Participants: presence.Participants, Presence: presence, WebsocketPath: "/ws/arena/sessions/" + session.ID, Scenario: session.Scenario, Preparation: arenaPreparationFromSession(session, claims.UserID)})
}

func (c *Controller) GetSession(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	session, err := c.service.GetSession(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	session.Participants = c.presenceResponse(session).Participants
	pkghttpx.WriteJSON(w, http.StatusOK, session)
}

func (c *Controller) GetPresence(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	session, err := c.service.GetSession(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, c.presenceResponse(session))
}

func (c *Controller) UpdateTyping(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	var req typingRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	session, err := c.service.GetSession(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	participant, ok := participantByUserID(session.Participants, claims.UserID)
	if !ok {
		c.writeError(w, r, entity.ErrArenaAccessDenied)
		return
	}
	if session.Status == entity.ArenaSessionCompleted || session.Status == entity.ArenaSessionCancelled {
		c.writeError(w, r, entity.ErrInvalidArenaState)
		return
	}
	c.hub.setTyping(id, claims.UserID, req.IsTyping)
	c.hub.broadcast(id, typingUpdatedPayload(id, participant, req.IsTyping))
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) ListMessages(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	messages, err := c.service.ListMessages(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, messageListResponse{Messages: messages})
}

func (c *Controller) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	var req sendMessageRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	params, ok := arenaMessageParams(id, req.MessageID, req.Content, req.SemanticActionCode, req.SemanticAnnotations)
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	msg, err := c.service.SendMessage(r.Context(), claims, params)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, msg)
}

func (c *Controller) ClassifyMessage(w http.ResponseWriter, r *http.Request) {
	claims, sessionID, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	messageID, err := entity.NormalizeUUID(chi.URLParam(r, "message_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	var req classifyMessageRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	params := entity.ClassifyArenaMessage{SessionID: sessionID, MessageID: messageID, SemanticActionCode: trimOptionalString(req.SemanticActionCode), SemanticAnnotations: req.SemanticAnnotations}
	classification, err := c.service.ClassifyMessage(r.Context(), claims, params)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, classification)
}

func (c *Controller) Complete(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	if err := c.service.Complete(r.Context(), claims, id); err != nil {
		c.writeError(w, r, err)
		return
	}
	c.hub.broadcast(id, map[string]any{"type": "arena.completed", "arena_session_id": id, "completed_by": claims.UserID})
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) ListFeedback(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	feedback, err := c.service.ListFeedback(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, feedbackListResponse{Feedback: feedback})
}

func (c *Controller) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndID(w, r, "arena_session_id")
	if !ok {
		return
	}
	var req feedbackRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.ParticipantID) == "" || strings.TrimSpace(req.Comment) == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	participantID, err := entity.NormalizeUUID(req.ParticipantID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	feedback, err := c.service.CreateFeedback(r.Context(), claims, id, participantID, req.Comment, req.Ratings)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, feedback)
}

func (c *Controller) UserAssignments(w http.ResponseWriter, r *http.Request) {
	claims, userID, ok := claimsAndID(w, r, "user_id")
	if !ok {
		return
	}
	limit, offset, ok := readLimitOffset(w, r)
	if !ok {
		return
	}
	items, err := c.service.ListUserAssignments(r.Context(), claims, userID, limit, offset)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, userAssignmentsResponse{Arenas: items})
}

func (c *Controller) UserHistory(w http.ResponseWriter, r *http.Request) {
	claims, userID, ok := claimsAndID(w, r, "user_id")
	if !ok {
		return
	}
	limit, offset, ok := readLimitOffset(w, r)
	if !ok {
		return
	}
	items, err := c.service.ListUserHistory(r.Context(), claims, userID, limit, offset)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, userHistoryResponse{Arenas: items})
}

func (c *Controller) UserHistoryDetails(w http.ResponseWriter, r *http.Request) {
	claims, userID, ok := claimsAndID(w, r, "user_id")
	if !ok {
		return
	}
	sessionID, err := entity.NormalizeUUID(chi.URLParam(r, "arena_session_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	session, messages, feedback, err := c.service.UserHistoryDetails(r.Context(), claims, userID, sessionID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, sessionDetailsResponse{Session: *session, Scenario: session.Scenario, Messages: messages, Feedback: feedback, Media: []any{}})
}

func (c *Controller) WebSocket(w http.ResponseWriter, r *http.Request) {
	token := tokenFromWebSocketRequest(r)
	claims, err := c.verifier.VerifyAccessToken(r.Context(), token)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeFromError(err))
		return
	}
	sessionID, err := entity.NormalizeUUID(chi.URLParam(r, "arena_session_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	session, err := c.service.GetSession(r.Context(), *claims, sessionID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	participant, ok := participantByUserID(session.Participants, claims.UserID)
	if !ok {
		c.writeError(w, r, entity.ErrArenaAccessDenied)
		return
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		Subprotocols:    []string{"arena-v1"},
		CheckOrigin:     checkWebSocketOrigin,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	client, becameOnline := c.hub.add(sessionID, conn, participant)
	_ = client.writeJSON(presenceSnapshotPayload(sessionID, c.presenceResponse(session).Participants))
	if becameOnline {
		c.hub.broadcast(sessionID, participantOnlinePayload(sessionID, participant))
	}
	defer func() {
		wasTyping, becameOffline := c.hub.remove(sessionID, client)
		if wasTyping {
			c.hub.broadcast(sessionID, typingUpdatedPayload(sessionID, participant, false))
		}
		if becameOffline {
			c.hub.broadcast(sessionID, participantOfflinePayload(sessionID, participant))
		}
	}()
	for {
		var req wsMessageSend
		if err := conn.ReadJSON(&req); err != nil {
			return
		}
		if req.Type == "typing.update" {
			if req.IsTyping == nil {
				_ = client.writeJSON(map[string]any{"type": "error", "code": string(httpx.ErrInvalidRequest), "message": httpx.ErrInvalidRequest.Message()})
				continue
			}
			c.hub.setTyping(sessionID, claims.UserID, *req.IsTyping)
			c.hub.broadcast(sessionID, typingUpdatedPayload(sessionID, participant, *req.IsTyping))
			continue
		}
		if req.Type != "message.send" {
			_ = client.writeJSON(map[string]any{"type": "error", "code": string(httpx.ErrInvalidRequest), "message": httpx.ErrInvalidRequest.Message()})
			continue
		}
		content := req.Content
		if len(content) == 0 && strings.TrimSpace(req.Text) != "" {
			content = chatContent(strings.TrimSpace(req.Text))
		}
		params, ok := arenaMessageParams(sessionID, req.MessageID, content, req.SemanticActionCode, req.SemanticAnnotations)
		if !ok {
			_ = client.writeJSON(map[string]any{"type": "error", "code": string(httpx.ErrInvalidRequest), "message": httpx.ErrInvalidRequest.Message()})
			continue
		}
		msg, err := c.service.SendMessage(r.Context(), *claims, params)
		if err != nil {
			code := httpx.ErrorCodeFromError(err)
			_ = client.writeJSON(map[string]any{"type": "error", "code": string(code), "message": code.Message()})
			continue
		}
		c.hub.broadcast(sessionID, map[string]any{"type": "message.created", "message": msg})
	}
}

func arenaMessageParams(sessionID, messageID string, content json.RawMessage, semanticActionCode *string, semanticAnnotations json.RawMessage) (entity.CreateArenaMessage, bool) {
	id, err := entity.NormalizeUUID(strings.TrimSpace(messageID))
	if err != nil || len(content) == 0 || !json.Valid(content) {
		return entity.CreateArenaMessage{}, false
	}
	code := trimOptionalString(semanticActionCode)
	return entity.CreateArenaMessage{ID: id, SessionID: sessionID, Content: content, SemanticActionCode: code, SemanticAnnotations: semanticAnnotations}, true
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func chatContent(text string) json.RawMessage {
	raw, _ := json.Marshal(map[string]string{"type": "chat", "text": text})
	return raw
}

func arenaPreparationFromSession(session *entity.ArenaSession, userID string) arenaPreparation {
	if session.Scenario == nil {
		return arenaPreparation{ScenarioID: session.ScenarioID, Title: session.ScenarioTitle}
	}
	my := entity.ScenarioRole{}
	if session.MyRole != nil {
		my = *session.MyRole
	}
	opponent := entity.PublicScenarioRole{}
	if session.OpponentRole != nil {
		opponent = *session.OpponentRole
	}
	return arenaPreparation{
		ScenarioID:      session.Scenario.ScenarioID,
		Title:           session.Scenario.Title,
		Description:     session.Scenario.Description,
		ExpandedContext: session.Scenario.ExpandedContext,
		MyRole:          scenarioRolePrivate{ID: my.ID, Name: my.Name, Description: my.Description, Goal: my.Goal, Interests: my.Interests, Constraints: my.Constraints, PrivateInformation: my.PrivateInformation},
		OpponentRole:    scenarioRoleVisible{ID: opponent.ID, Name: opponent.Name, Description: opponent.Description, Goal: opponent.Goal, Interests: opponent.Interests, Constraints: opponent.Constraints},
		Roles:           session.Scenario.Roles,
		Blocks:          session.Scenario.Blocks,
	}
}

func sessionParticipantRole(participants []entity.ArenaSessionParticipant, userID string) string {
	for _, participant := range participants {
		if participant.UserID == userID {
			return participant.ScenarioRoleID
		}
	}
	return ""
}

func participantByUserID(participants []entity.ArenaSessionParticipant, userID string) (entity.ArenaSessionParticipant, bool) {
	for _, participant := range participants {
		if participant.UserID == userID {
			return participant, true
		}
	}
	return entity.ArenaSessionParticipant{}, false
}

func (c *Controller) presenceResponse(session *entity.ArenaSession) arenaPresenceResponse {
	online := c.hub.onlineUsers(session.ID)
	participants := make([]entity.ArenaSessionParticipant, 0, len(session.Participants))
	for _, participant := range session.Participants {
		participant.Online = online[participant.UserID]
		participants = append(participants, participant)
	}
	return arenaPresenceResponse{Participants: participants}
}

func presenceSnapshotPayload(sessionID string, participants []entity.ArenaSessionParticipant) map[string]any {
	return map[string]any{"type": "presence.snapshot", "arena_session_id": sessionID, "participants": participants}
}

func participantOnlinePayload(sessionID string, participant entity.ArenaSessionParticipant) map[string]any {
	participant.Online = true
	return map[string]any{"type": "participant.online", "arena_session_id": sessionID, "participant": participant}
}

func participantOfflinePayload(sessionID string, participant entity.ArenaSessionParticipant) map[string]any {
	participant.Online = false
	return map[string]any{"type": "participant.offline", "arena_session_id": sessionID, "participant": participant}
}

func typingUpdatedPayload(sessionID string, participant entity.ArenaSessionParticipant, isTyping bool) map[string]any {
	return map[string]any{"type": "typing.updated", "arena_session_id": sessionID, "user_id": participant.UserID, "username": participant.Username, "scenario_role_id": participant.ScenarioRoleID, "is_typing": isTyping}
}

func (c *Controller) writeError(w http.ResponseWriter, r *http.Request, err error) {
	code := httpx.ErrorCodeFromError(err)
	status := http.StatusInternalServerError
	switch code {
	case httpx.ErrInvalidRequest:
		status = http.StatusBadRequest
	case httpx.ErrCuratorRoleRequired, httpx.ErrArenaAccessDenied, httpx.ErrArenaFeedbackForbidden:
		status = http.StatusForbidden
	case httpx.ErrArenaAssignmentNotFound, httpx.ErrArenaSessionNotFound, httpx.ErrArenaMessageNotFound, httpx.ErrScenarioNotFound, httpx.ErrUserNotFound:
		status = http.StatusNotFound
	case httpx.ErrArenaJoinNotAllowed, httpx.ErrArenaWaitingForParticipant, httpx.ErrInvalidArenaState, httpx.ErrArenaFeedbackConflict, httpx.ErrScenarioNotAvailable, httpx.ErrMessageAlreadyClassified:
		status = http.StatusConflict
	case httpx.ErrArenaFeedbackParticipantInvalid, httpx.ErrMessageNotClassifiable:
		status = http.StatusUnprocessableEntity
	}
	if status == http.StatusInternalServerError {
		c.logger.Error("arena request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
	}
	httpx.WriteError(w, status, code)
}

func claimsFromRequest(w http.ResponseWriter, r *http.Request) (entity.AccessClaims, bool) {
	claims, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
		return entity.AccessClaims{}, false
	}
	return claims, true
}

func claimsAndID(w http.ResponseWriter, r *http.Request, name string) (entity.AccessClaims, string, bool) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return entity.AccessClaims{}, "", false
	}
	id, err := entity.NormalizeUUID(chi.URLParam(r, name))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return entity.AccessClaims{}, "", false
	}
	return claims, id, true
}

func readLimitOffset(w http.ResponseWriter, r *http.Request) (int32, int32, bool) {
	limit := int32(20)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || value < 1 || value > 100 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
			return 0, 0, false
		}
		limit = int32(value)
	}
	offset := int32(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || value < 0 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
			return 0, 0, false
		}
		offset = int32(value)
	}
	return limit, offset, true
}

func tokenFromWebSocketRequest(r *http.Request) string {
	if header := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}
	for _, part := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		part = strings.TrimSpace(part)
		if part != "" && part != "arena-v1" {
			return part
		}
	}
	return ""
}

func checkWebSocketOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	originHost := strings.ToLower(parsed.Hostname())
	requestHost := strings.ToLower(hostnameOnly(r.Host))
	if originHost == requestHost {
		return true
	}
	switch originHost {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func hostnameOnly(host string) string {
	parsed, err := url.Parse("//" + host)
	if err != nil || parsed.Hostname() == "" {
		return host
	}
	return parsed.Hostname()
}

type hub struct {
	mu     sync.Mutex
	conns  map[string]map[*arenaClient]struct{}
	typing map[string]map[string]bool
}

func newHub() *hub {
	return &hub{conns: map[string]map[*arenaClient]struct{}{}, typing: map[string]map[string]bool{}}
}

type arenaClient struct {
	conn     *websocket.Conn
	userID   string
	username string
	roleID   string
	writeMu  sync.Mutex
}

func (c *arenaClient) writeJSON(payload any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(payload)
}

func (h *hub) add(sessionID string, conn *websocket.Conn, participant entity.ArenaSessionParticipant) (*arenaClient, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[sessionID] == nil {
		h.conns[sessionID] = map[*arenaClient]struct{}{}
	}
	wasOnline := h.userOnlineLocked(sessionID, participant.UserID)
	client := &arenaClient{conn: conn, userID: participant.UserID, username: participant.Username, roleID: participant.ScenarioRoleID}
	h.conns[sessionID][client] = struct{}{}
	return client, !wasOnline
}

func (h *hub) remove(sessionID string, client *arenaClient) (bool, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[sessionID], client)
	becameOffline := !h.userOnlineLocked(sessionID, client.userID)
	wasTyping := false
	if h.typing[sessionID] != nil {
		wasTyping = h.typing[sessionID][client.userID]
		delete(h.typing[sessionID], client.userID)
	}
	if len(h.conns[sessionID]) == 0 {
		delete(h.conns, sessionID)
		delete(h.typing, sessionID)
	}
	return wasTyping, becameOffline
}

func (h *hub) setTyping(sessionID, userID string, isTyping bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.typing[sessionID] == nil {
		h.typing[sessionID] = map[string]bool{}
	}
	if isTyping {
		h.typing[sessionID][userID] = true
		return
	}
	delete(h.typing[sessionID], userID)
}

func (h *hub) onlineUsers(sessionID string) map[string]bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	users := map[string]bool{}
	for client := range h.conns[sessionID] {
		users[client.userID] = true
	}
	return users
}

func (h *hub) broadcast(sessionID string, payload any) {
	h.mu.Lock()
	clients := make([]*arenaClient, 0, len(h.conns[sessionID]))
	for client := range h.conns[sessionID] {
		clients = append(clients, client)
	}
	h.mu.Unlock()
	for _, client := range clients {
		_ = client.writeJSON(payload)
	}
}

func (h *hub) userOnlineLocked(sessionID, userID string) bool {
	for client := range h.conns[sessionID] {
		if client.userID == userID {
			return true
		}
	}
	return false
}
