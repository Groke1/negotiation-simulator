package training

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	usecase "sourcecraft.dev/lct-hackaton-2026.arena/internal/usecase/training"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	ListSessions(ctx context.Context, claims entity.AccessClaims, filter entity.TrainingSessionFilter) ([]entity.TrainingSession, error)
	StartSession(ctx context.Context, claims entity.AccessClaims, params entity.StartTrainingSession) (*entity.TrainingSessionDetails, error)
	GetSession(ctx context.Context, claims entity.AccessClaims, sessionID string) (*entity.TrainingSessionDetails, error)
	SubmitAction(ctx context.Context, claims entity.AccessClaims, sessionID, actionID, selectedCode, clientEventID string) (*entity.TrainingActionResult, error)
	SendMessage(ctx context.Context, claims entity.AccessClaims, sessionID string, messageID string, content json.RawMessage, semanticActionCode *string, annotations json.RawMessage) (*entity.TrainingMessageExchange, error)
	ClassifyMessage(ctx context.Context, claims entity.AccessClaims, sessionID, messageID, selectedCode string) (*entity.TrainingMessage, error)
	GetAnalysis(ctx context.Context, claims entity.AccessClaims, sessionID string) (json.RawMessage, error)
}

type Controller struct {
	logger  *zap.Logger
	service Service
}

func NewController(logger *zap.Logger, service Service) *Controller {
	return &Controller{logger: logger, service: service}
}

func (c *Controller) InitRoutes(r chi.Router) {
	r.Get("/training/sessions", c.ListSessions)
	r.Post("/training/sessions", c.StartSession)
	r.Get("/training/sessions/{session_id}", c.GetSession)
	r.Post("/training/sessions/{session_id}/actions", c.SubmitAction)
	r.Post("/training/sessions/{session_id}/messages", c.SendMessage)
	r.Post("/training/sessions/{session_id}/messages/{message_id}/classification", c.ClassifyMessage)
	r.Get("/training/sessions/{session_id}/analysis", c.GetAnalysis)
}

func (c *Controller) ListSessions(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	limit, offset, ok := readLimitOffset(w, r)
	if !ok {
		return
	}
	filter := entity.TrainingSessionFilter{Limit: limit, Offset: offset}
	if raw := strings.TrimSpace(r.URL.Query().Get("scenario_id")); raw != "" {
		id, err := entity.NormalizeUUID(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
			return
		}
		filter.ScenarioID = &id
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		status := entity.TrainingStatus(raw)
		filter.Status = &status
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("mode")); raw != "" {
		mode := entity.TrainingMode(raw)
		filter.Mode = &mode
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("channel")); raw != "" {
		channel := entity.TrainingChannel(raw)
		filter.Channel = &channel
	}
	items, err := c.service.ListSessions(r.Context(), claims, filter)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := trainingSessionListResponse{Sessions: make([]trainingSessionListItem, 0, len(items))}
	for _, item := range items {
		resp.Sessions = append(resp.Sessions, sessionToListItem(item))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) StartSession(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	var req startTrainingSessionRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validStartRequest(req) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	scenarioID, err := entity.NormalizeUUID(req.ScenarioID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	details, err := c.service.StartSession(r.Context(), claims, entity.StartTrainingSession{ScenarioID: scenarioID, ScenarioRoleID: strings.TrimSpace(req.ScenarioRoleID), Mode: req.Mode, Channel: req.Channel, AssignmentID: req.AssignmentID})
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, detailsToResponse(*details))
}

func (c *Controller) GetSession(w http.ResponseWriter, r *http.Request) {
	claims, sessionID, ok := claimsAndSessionID(w, r)
	if !ok {
		return
	}
	details, err := c.service.GetSession(r.Context(), claims, sessionID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, detailsToResponse(*details))
}

func (c *Controller) SubmitAction(w http.ResponseWriter, r *http.Request) {
	claims, sessionID, ok := claimsAndSessionID(w, r)
	if !ok {
		return
	}
	var req trainingActionRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.ActionID) == "" || strings.TrimSpace(req.SemanticActionCode) == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	result, err := c.service.SubmitAction(r.Context(), claims, sessionID, strings.TrimSpace(req.ActionID), strings.TrimSpace(req.SemanticActionCode), strings.TrimSpace(req.ClientEventID))
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	next := (*deterministicCurrentNode)(nil)
	if details, err := c.service.GetSession(r.Context(), claims, sessionID); err == nil {
		doc := scenarioDocument(details.Session.Snapshot)
		next = currentNodeFromState(result.NextState, &doc)
	}
	pkghttpx.WriteJSON(w, http.StatusOK, trainingActionResponse{EventID: result.EventID, Classification: result.Classification, NextNode: next, Status: result.Status, IsEnd: result.Status == entity.TrainingStatusCompleted})
}

func (c *Controller) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims, sessionID, ok := claimsAndSessionID(w, r)
	if !ok {
		return
	}
	var req trainingMessageRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.MessageID) == "" || !json.Valid(req.Content) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	code := usecase.TrimSemanticActionCode(req.SemanticActionCode)
	exchange, err := c.service.SendMessage(r.Context(), claims, sessionID, strings.TrimSpace(req.MessageID), req.Content, code, req.SemanticAnnotations)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, trainingMessageResponse{LearnerMessage: messageToResponse(exchange.LearnerMessage), PartnerMessage: messageToResponse(exchange.PartnerMessage), Status: exchange.Status, IsEnd: exchange.Status == entity.TrainingStatusCompleted})
}

func (c *Controller) ClassifyMessage(w http.ResponseWriter, r *http.Request) {
	claims, sessionID, ok := claimsAndSessionID(w, r)
	if !ok {
		return
	}
	messageID, err := entity.NormalizeUUID(chi.URLParam(r, "message_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	var req classifyMessageRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.SemanticActionCode) == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	msg, err := c.service.ClassifyMessage(r.Context(), claims, sessionID, messageID, strings.TrimSpace(req.SemanticActionCode))
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := classifyMessageResponse{MessageID: msg.ID}
	_ = json.Unmarshal(msg.Classification, &resp)
	pkghttpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	claims, sessionID, ok := claimsAndSessionID(w, r)
	if !ok {
		return
	}
	analysis, err := c.service.GetAnalysis(r.Context(), claims, sessionID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, json.RawMessage(analysis))
}

func (c *Controller) writeError(w http.ResponseWriter, r *http.Request, err error) {
	code := httpx.ErrorCodeFromError(err)
	status := http.StatusInternalServerError
	switch code {
	case httpx.ErrInvalidRequest:
		status = http.StatusBadRequest
	case httpx.ErrTrainingAccessDenied:
		status = http.StatusForbidden
	case httpx.ErrScenarioNotFound, httpx.ErrTrainingSessionNotFound, httpx.ErrTrainingMessageNotFound:
		status = http.StatusNotFound
	case httpx.ErrScenarioNotAvailable, httpx.ErrInvalidTrainingState, httpx.ErrMessageAlreadyClassified:
		status = http.StatusConflict
	case httpx.ErrTrainingConfigurationNotSupported, httpx.ErrActionNotAvailable, httpx.ErrTrainingMessageChannelMismatch, httpx.ErrMessageNotClassifiable:
		status = http.StatusUnprocessableEntity
	}
	if status == http.StatusInternalServerError {
		c.logger.Error("training request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
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

func claimsAndSessionID(w http.ResponseWriter, r *http.Request) (entity.AccessClaims, string, bool) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return entity.AccessClaims{}, "", false
	}
	id, err := entity.NormalizeUUID(chi.URLParam(r, "session_id"))
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

func validStartRequest(req startTrainingSessionRequest) bool {
	return strings.TrimSpace(req.ScenarioID) != "" && strings.TrimSpace(req.ScenarioRoleID) != "" && (req.Mode == entity.TrainingModeDeterministic || req.Mode == entity.TrainingModeVirtualPartner) && (req.Channel == entity.TrainingChannelChat || req.Channel == entity.TrainingChannelEmail || req.Channel == entity.TrainingChannelVoiceMessage)
}

func detailsToResponse(details entity.TrainingSessionDetails) trainingSessionResponse {
	doc := scenarioDocument(details.Session.Snapshot)
	return trainingSessionResponse{
		SessionID:      details.Session.ID,
		ScenarioID:     details.Session.ScenarioID,
		ScenarioTitle:  scenarioTitle(details.Session, doc),
		ScenarioRoleID: details.Session.ScenarioRoleID,
		AssignmentID:   details.Session.AssignmentID,
		Mode:           details.Session.Mode,
		Channel:        details.Session.Channel,
		Status:         details.Session.Status,
		IsEnd:          details.Session.Status == entity.TrainingStatusCompleted,
		Preparation:    preparation(doc, details.Session.ScenarioRoleID),
		CurrentNode:    currentNodeFromState(details.Session.State, &doc),
		Messages:       messagesToResponse(details.Messages),
		Score:          details.Session.Score,
		StartedAt:      details.Session.CreatedAt,
		CompletedAt:    details.Session.CompletedAt,
	}
}

func sessionToListItem(session entity.TrainingSession) trainingSessionListItem {
	return trainingSessionListItem{SessionID: session.ID, ScenarioID: session.ScenarioID, ScenarioTitle: session.ScenarioTitle, ScenarioRoleID: session.ScenarioRoleID, Mode: session.Mode, Channel: session.Channel, Status: session.Status, Score: session.Score, StartedAt: session.CreatedAt, CompletedAt: session.CompletedAt}
}

func messagesToResponse(messages []entity.TrainingMessage) []trainingMessage {
	resp := make([]trainingMessage, 0, len(messages))
	for _, msg := range messages {
		resp = append(resp, messageToResponse(msg))
	}
	return resp
}

func messageToResponse(msg entity.TrainingMessage) trainingMessage {
	return trainingMessage{MessageID: msg.ID, Author: msg.Author, Content: msg.Content, SemanticActionCode: msg.SemanticActionCode, SemanticAnnotations: emptyJSONArray(msg.SemanticAnnotations), Classification: nullableJSON(msg.Classification), CreatedAt: msg.CreatedAt}
}

func scenarioDocument(raw json.RawMessage) entity.ScenarioDocument {
	var doc entity.ScenarioDocument
	_ = json.Unmarshal(raw, &doc)
	return doc
}

func preparation(doc entity.ScenarioDocument, roleID string) trainingPreparation {
	me := entity.ScenarioRole{}
	partner := entity.ScenarioRole{}
	for _, role := range doc.Roles {
		if role.ID == roleID {
			me = role
		} else if partner.ID == "" {
			partner = role
		}
	}
	return trainingPreparation{ScenarioID: doc.ScenarioID, Title: doc.Title, Description: doc.Description, ExpandedContext: doc.ExpandedContext, MyRole: scenarioRolePrivate(me), PartnerRole: scenarioRoleVisible{ID: partner.ID, Name: partner.Name, Description: partner.Description}}
}

func currentNodeFromState(raw json.RawMessage, doc *entity.ScenarioDocument) *deterministicCurrentNode {
	if doc == nil {
		return nil
	}
	var state struct {
		BlockID       string `json:"block_id"`
		CurrentNodeID string `json:"current_node_id"`
	}
	if json.Unmarshal(raw, &state) != nil || state.BlockID == "" || state.CurrentNodeID == "" {
		return nil
	}
	for _, block := range doc.Blocks {
		if block.ID != state.BlockID {
			continue
		}
		if block.DeterministicDialogue == nil {
			return nil
		}
		for _, node := range block.DeterministicDialogue.Nodes {
			if node.ID != state.CurrentNodeID {
				continue
			}
			actions := make([]deterministicAvailableAction, 0, len(node.ReplyOptions))
			for _, action := range node.ReplyOptions {
				actions = append(actions, deterministicAvailableAction{ActionID: action.ID, Text: action.Text})
			}
			var incoming *partnerDeterministicMessage
			if node.IncomingMessage != nil {
				incoming = &partnerDeterministicMessage{MessageID: state.CurrentNodeID, Text: node.IncomingMessage.Text}
			}
			return &deterministicCurrentNode{NodeID: node.ID, BlockID: block.ID, IncomingMessage: incoming, AvailableActions: actions}
		}
	}
	return nil
}

func scenarioTitle(session entity.TrainingSession, doc entity.ScenarioDocument) string {
	if session.ScenarioTitle != "" {
		return session.ScenarioTitle
	}
	return doc.Title
}

func emptyJSONArray(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("[]")
	}
	return raw
}

func nullableJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
