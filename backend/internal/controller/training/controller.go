package training

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	Start(context.Context, entity.AccessClaims, entity.StartTrainingSession) (*entity.TrainingSession, *entity.TrainingNode, []entity.TrainingMessage, error)
	Get(context.Context, entity.AccessClaims, string) (*entity.TrainingSession, *entity.TrainingNode, []entity.TrainingMessage, error)
	SelectAction(context.Context, entity.AccessClaims, entity.SelectTrainingAction) (*entity.TrainingActionResult, error)
	SendMessage(context.Context, entity.AccessClaims, entity.SendTrainingMessage) (*entity.TrainingMessageResult, error)
	Complete(context.Context, entity.AccessClaims, string) (*entity.TrainingCompletion, error)
	Analysis(context.Context, entity.AccessClaims, string) (json.RawMessage, error)
}
type Controller struct {
	logger  *zap.Logger
	service Service
}

func NewController(logger *zap.Logger, service Service) *Controller {
	return &Controller{logger: logger, service: service}
}
func (c *Controller) InitRoutes(r chi.Router) {
	r.Route("/training/sessions", func(r chi.Router) {
		r.Post("/", c.Start)
		r.Route("/{session_id}", func(r chi.Router) {
			r.Get("/", c.Get)
			r.Post("/actions", c.Action)
			r.Post("/messages", c.Message)
			r.Post("/complete", c.Complete)
			r.Get("/analysis", c.Analysis)
		})
	})
}
func (c *Controller) Start(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	var req startRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validUUID(req.ScenarioID) || !entity.TrainingMode(req.Mode).Valid() || req.Channel != "chat" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	session, node, messages, err := c.service.Start(r.Context(), claims, entity.StartTrainingSession{ScenarioID: req.ScenarioID, Mode: entity.TrainingMode(req.Mode), Channel: req.Channel, Preview: req.Preview})
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, response(session, node, messages))
}
func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := sessionID(w, r)
	if !ok {
		return
	}
	session, node, messages, err := c.service.Get(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, response(session, node, messages))
}
func (c *Controller) Action(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := sessionID(w, r)
	if !ok {
		return
	}
	var req actionRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.ActionID) == "" || !validUUID(req.ClientEventID) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	result, err := c.service.SelectAction(r.Context(), claims, entity.SelectTrainingAction{SessionID: id, ActionID: strings.TrimSpace(req.ActionID), ClientEventID: req.ClientEventID})
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, actionResponse{EventID: result.Event.ID, NextNode: node(result.NextNode), Status: string(result.Status)})
}
func (c *Controller) Message(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := sessionID(w, r)
	if !ok {
		return
	}
	var req messageRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || len(strings.TrimSpace(req.Text)) == 0 || len(req.Text) > 5000 || !validUUID(req.MessageID) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	result, err := c.service.SendMessage(r.Context(), claims, entity.SendTrainingMessage{SessionID: id, Text: strings.TrimSpace(req.Text), MessageID: req.MessageID})
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, trainingMessageResponse{LearnerMessage: message(result.LearnerMessage), PartnerMessage: message(result.PartnerMessage), DetectedSemanticActions: result.SemanticActions})
}
func (c *Controller) Complete(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := sessionID(w, r)
	if !ok {
		return
	}
	result, err := c.service.Complete(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, completeResponse{Outcome: result.Outcome})
}
func (c *Controller) Analysis(w http.ResponseWriter, r *http.Request) {
	claims, ok := claims(w, r)
	if !ok {
		return
	}
	id, ok := sessionID(w, r)
	if !ok {
		return
	}
	result, err := c.service.Analysis(r.Context(), claims, id)
	if err != nil {
		c.error(w, r, err)
		return
	}
	var output any
	if json.Unmarshal(result, &output) != nil {
		c.error(w, r, entity.ErrInternal)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, output)
}
func (c *Controller) error(w http.ResponseWriter, r *http.Request, err error) {
	status := httpx.StatusFromError(err)
	code := httpx.ErrorCodeFromError(err)
	if status >= 500 {
		c.logger.Error("training request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
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
func sessionID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := entity.NormalizeUUID(chi.URLParam(r, "session_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return "", false
	}
	return id, true
}
func validUUID(value string) bool { _, err := entity.NormalizeUUID(value); return err == nil }
func node(value *entity.TrainingNode) any {
	if value == nil {
		return nil
	}
	return nodeResponse{NodeID: value.ID, Speaker: value.Speaker, Text: value.Text, AvailableActions: value.AvailableActions}
}
func message(value entity.TrainingMessage) messageResponse {
	return messageResponse{MessageID: value.ID, Author: value.Author, Text: value.Text, CreatedAt: value.CreatedAt}
}
func response(session *entity.TrainingSession, current *entity.TrainingNode, messages []entity.TrainingMessage) sessionResponse {
	output := make([]messageResponse, 0, len(messages))
	for _, item := range messages {
		output = append(output, message(item))
	}
	return sessionResponse{SessionID: session.ID, ScenarioID: session.ScenarioID, ScenarioVersion: session.ScenarioVersion, Mode: string(session.Mode), Channel: session.Channel, Status: string(session.Status), CurrentNode: node(current), Messages: output}
}
