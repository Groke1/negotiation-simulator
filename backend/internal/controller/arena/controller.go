package arena

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	CreateAssignment(context.Context, entity.AccessClaims, entity.CreateArenaAssignment) (*entity.ArenaAssignment, error)
	GetAssignment(context.Context, entity.AccessClaims, string) (*entity.ArenaAssignment, error)
	UserAssignments(context.Context, entity.AccessClaims, string) ([]entity.ArenaAssignment, error)
	OrganizationAssignments(context.Context, entity.AccessClaims) ([]entity.ArenaAssignment, error)
	Join(context.Context, entity.AccessClaims, string) (*entity.ArenaSession, error)
	GetSession(context.Context, entity.AccessClaims, string) (*entity.ArenaSession, error)
	Messages(context.Context, entity.AccessClaims, string) ([]entity.ArenaMessage, error)
	Complete(context.Context, entity.AccessClaims, string) error
	SaveMessage(context.Context, entity.AccessClaims, string, string, string) (*entity.ArenaMessage, error)
	Feedback(context.Context, entity.AccessClaims, string) ([]entity.ArenaFeedback, error)
	CreateFeedback(context.Context, entity.AccessClaims, entity.CreateArenaFeedback) (*entity.ArenaFeedback, error)
	UserHistory(context.Context, entity.AccessClaims, string) ([]entity.ArenaSession, []bool, error)
}
type Controller struct {
	logger  *zap.Logger
	service Service
	hub     *hub
}

func NewController(logger *zap.Logger, service Service) *Controller {
	return &Controller{logger: logger, service: service, hub: newHub()}
}
func (c *Controller) InitRoutes(r chi.Router) {
	r.Route("/arena", func(r chi.Router) {
		r.Post("/assignments", c.CreateAssignment)
		r.Get("/assignments", c.OrganizationAssignments)
		r.Route("/assignments/{arena_assignment_id}", func(r chi.Router) { r.Get("/", c.GetAssignment); r.Post("/join", c.Join) })
		r.Route("/sessions/{arena_session_id}", func(r chi.Router) {
			r.Get("/", c.GetSession)
			r.Get("/messages", c.Messages)
			r.Post("/complete", c.Complete)
			r.Get("/feedback", c.Feedback)
			r.Post("/feedback", c.CreateFeedback)
		})
	})
	r.Get("/users/{user_id}/arena-assignments", c.UserAssignments)
	r.Get("/users/{user_id}/arena-history", c.UserHistory)
	r.Get("/users/{user_id}/arena-history/{arena_session_id}", c.HistoryDetails)
}

func (c *Controller) OrganizationAssignments(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	items, err := c.service.OrganizationAssignments(r.Context(), claims)
	if err != nil {
		c.error(w, r, err)
		return
	}
	out := make([]assignmentResponse, 0, len(items))
	for index := range items {
		out = append(out, assignmentView(&items[index]))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, map[string]any{"arenas": out})
}
func (c *Controller) InitWebsocketRoutes(r chi.Router) {
	r.Get("/ws/arena/sessions/{arena_session_id}", c.Websocket)
}
func (c *Controller) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	var req createAssignmentRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validUUID(req.ScenarioID) || len(req.Participants) != 2 || req.ScheduledAt.IsZero() {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	participants := make([]entity.ArenaParticipant, 0, 2)
	for _, item := range req.Participants {
		if !validUUID(item.UserID) || !entity.ArenaParticipantRole(item.Role).Valid() {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
			return
		}
		participants = append(participants, entity.ArenaParticipant{UserID: item.UserID, Role: entity.ArenaParticipantRole(item.Role)})
	}
	assignment, err := c.service.CreateAssignment(r.Context(), claims, entity.CreateArenaAssignment{ScenarioID: req.ScenarioID, Participants: participants, ScheduledAt: req.ScheduledAt})
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, map[string]any{"arena_assignment_id": assignment.ID, "status": assignment.Status})
}
func (c *Controller) GetAssignment(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "arena_assignment_id")
	if !ok {
		return
	}
	assignment, err := c.service.GetAssignment(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, assignmentView(assignment))
}
func (c *Controller) Join(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "arena_assignment_id")
	if !ok {
		return
	}
	session, err := c.service.Join(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, map[string]any{"arena_session_id": session.ID, "status": session.Status, "participants": participants(session.Participants), "websocket_path": "/ws/arena/sessions/" + session.ID})
}
func (c *Controller) GetSession(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "arena_session_id")
	if !ok {
		return
	}
	session, err := c.service.GetSession(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, sessionView(session))
}
func (c *Controller) Messages(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "arena_session_id")
	if !ok {
		return
	}
	messages, err := c.service.Messages(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	result := make([]messageResponse, 0, len(messages))
	for _, message := range messages {
		result = append(result, messageView(&message))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, map[string]any{"messages": result})
}
func (c *Controller) Complete(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "arena_session_id")
	if !ok {
		return
	}
	if err := c.service.Complete(r.Context(), claims, id); err != nil {
		c.error(w, r, err)
		return
	}
	c.hub.broadcast(id, map[string]any{"type": "arena.completed"})
	w.WriteHeader(http.StatusNoContent)
}
func (c *Controller) Feedback(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "arena_session_id")
	if !ok {
		return
	}
	feedback, err := c.service.Feedback(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	result := make([]feedbackResponse, 0, len(feedback))
	for _, item := range feedback {
		result = append(result, feedbackView(&item))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, map[string]any{"feedback": result})
}
func (c *Controller) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "arena_session_id")
	if !ok {
		return
	}
	var req feedbackRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validUUID(req.ParticipantID) || strings.TrimSpace(req.Comment) == "" || len(req.Comment) > 10000 || (len(req.Ratings) > 0 && !json.Valid(req.Ratings)) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	feedback, err := c.service.CreateFeedback(r.Context(), claims, entity.CreateArenaFeedback{SessionID: id, ParticipantID: req.ParticipantID, Comment: req.Comment, Ratings: req.Ratings})
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, feedbackView(feedback))
}
func (c *Controller) UserAssignments(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "user_id")
	if !ok {
		return
	}
	items, err := c.service.UserAssignments(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	out := make([]assignmentResponse, 0, len(items))
	for i := range items {
		out = append(out, assignmentView(&items[i]))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, map[string]any{"arenas": out})
}
func (c *Controller) UserHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "user_id")
	if !ok {
		return
	}
	items, hasFeedback, err := c.service.UserHistory(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for index := range items {
		item := items[index]
		out = append(out, map[string]any{"arena_session_id": item.ID, "arena_assignment_id": item.AssignmentID, "scenario_id": item.ScenarioID, "scenario_title": item.ScenarioTitle, "scheduled_at": item.ScheduledAt, "started_at": item.StartedAt, "completed_at": item.CompletedAt, "status": item.Status, "has_feedback": hasFeedback[index]})
	}
	pkghttpx.WriteJSON(w, http.StatusOK, map[string]any{"arenas": out})
}
func (c *Controller) HistoryDetails(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	userID, ok := pathID(w, r, "user_id")
	if !ok {
		return
	}
	sessionID, ok := pathID(w, r, "arena_session_id")
	if !ok {
		return
	}
	history, _, err := c.service.UserHistory(r.Context(), claims, userID)
	if err != nil {
		c.error(w, r, err)
		return
	}
	var session *entity.ArenaSession
	for i := range history {
		if history[i].ID == sessionID {
			session = &history[i]
			break
		}
	}
	if session == nil {
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrArenaSessionNotFound)
		return
	}
	messages, err := c.service.Messages(r.Context(), claims, sessionID)
	if err != nil {
		c.error(w, r, err)
		return
	}
	feedback, err := c.service.Feedback(r.Context(), claims, sessionID)
	if err != nil {
		c.error(w, r, err)
		return
	}
	messageOut := make([]messageResponse, 0, len(messages))
	for _, item := range messages {
		messageOut = append(messageOut, messageView(&item))
	}
	feedbackOut := make([]feedbackResponse, 0, len(feedback))
	for _, item := range feedback {
		feedbackOut = append(feedbackOut, feedbackView(&item))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, map[string]any{"session": sessionView(session), "messages": messageOut, "feedback": feedbackOut, "media": []any{}})
}
func (c *Controller) Websocket(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	sessionID, ok := pathID(w, r, "arena_session_id")
	if !ok {
		return
	}
	if _, err := c.service.GetSession(r.Context(), claims, sessionID); err != nil {
		c.error(w, r, err)
		return
	}
	upgrader := websocket.Upgrader{
		Subprotocols: []string{websocketAccessToken(r)},
		CheckOrigin:  checkWebSocketOrigin,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.logger.Error("arena websocket upgrade failed", zap.Error(err))
		return
	}
	client := c.hub.add(sessionID, claims.UserID, conn)
	defer c.hub.remove(client)
	c.hub.broadcast(sessionID, map[string]any{"type": "participant.joined", "user_id": claims.UserID})
	for {
		var incoming wsMessageSend
		if err := conn.ReadJSON(&incoming); err != nil {
			break
		}
		if incoming.Type != "message.send" || !validUUID(incoming.MessageID) || len(strings.TrimSpace(incoming.Text)) == 0 || len(incoming.Text) > 5000 {
			c.hub.send(client, map[string]any{"type": "error", "code": "INVALID_REQUEST", "message": "Request validation failed"})
			continue
		}
		message, err := c.service.SaveMessage(r.Context(), claims, sessionID, incoming.MessageID, incoming.Text)
		if err != nil {
			c.hub.send(client, map[string]any{"type": "error", "code": httpx.ErrorCodeFromError(err), "message": httpx.ErrorCodeFromError(err).Message()})
			continue
		}
		c.hub.broadcast(sessionID, map[string]any{"type": "message.created", "message": messageView(message)})
	}
	c.hub.broadcast(sessionID, map[string]any{"type": "participant.left", "user_id": claims.UserID})
}

func websocketAccessToken(r *http.Request) string {
	for _, protocol := range websocket.Subprotocols(r) {
		if strings.Count(protocol, ".") == 2 {
			return protocol
		}
	}
	return ""
}

func checkWebSocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin == "" || origin == "http://localhost:5173"
}
func (c *Controller) error(w http.ResponseWriter, r *http.Request, err error) {
	status := httpx.StatusFromError(err)
	code := httpx.ErrorCodeFromError(err)
	if status >= 500 {
		c.logger.Error("arena request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
	}
	httpx.WriteError(w, status, code)
}
func claims(w http.ResponseWriter, r *http.Request) (entity.AccessClaims, bool) {
	value, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
	}
	return value, ok
}
func pathID(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	id, err := entity.NormalizeUUID(chi.URLParam(r, name))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return "", false
	}
	return id, true
}
func validUUID(value string) bool { _, err := entity.NormalizeUUID(value); return err == nil }
func participants(items []entity.ArenaParticipant) []participantResponse {
	out := make([]participantResponse, 0, len(items))
	for _, item := range items {
		out = append(out, participantResponse{UserID: item.UserID, Role: string(item.Role), JoinedAt: item.JoinedAt})
	}
	return out
}
func assignmentView(value *entity.ArenaAssignment) assignmentResponse {
	return assignmentResponse{ArenaAssignmentID: value.ID, ScenarioID: value.ScenarioID, ScenarioTitle: value.ScenarioTitle, ScenarioVersion: value.ScenarioVersion, Participants: participants(value.Participants), ScheduledAt: value.ScheduledAt, Status: string(value.Status), ArenaSessionID: value.SessionID, CreatedBy: value.CreatedBy, CreatedAt: value.CreatedAt}
}
func sessionView(value *entity.ArenaSession) sessionResponse {
	return sessionResponse{ArenaSessionID: value.ID, ArenaAssignmentID: value.AssignmentID, ScenarioID: value.ScenarioID, ScenarioTitle: value.ScenarioTitle, ScenarioVersion: value.ScenarioVersion, Channel: value.Channel, Status: string(value.Status), ScheduledAt: value.ScheduledAt, Participants: participants(value.Participants), StartedAt: value.StartedAt, CompletedAt: value.CompletedAt}
}
func messageView(value *entity.ArenaMessage) messageResponse {
	return messageResponse{MessageID: value.ID, SenderID: value.SenderID, Text: value.Text, CreatedAt: value.CreatedAt}
}
func feedbackView(value *entity.ArenaFeedback) feedbackResponse {
	return feedbackResponse{FeedbackID: value.ID, ArenaSessionID: value.SessionID, ParticipantID: value.ParticipantID, ExpertID: value.ExpertID, Comment: value.Comment, Ratings: value.Ratings, CreatedAt: value.CreatedAt}
}
