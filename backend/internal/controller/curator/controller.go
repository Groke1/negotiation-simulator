package curator

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	ListTeam(ctx context.Context, claims entity.AccessClaims, limit, offset int32) ([]entity.CuratorTeamMember, error)
	ListTeamSessions(ctx context.Context, claims entity.AccessClaims, limit, offset int32) ([]entity.TeamSessionListItem, error)
	ListAssignments(ctx context.Context, claims entity.AccessClaims, filter entity.AssignmentFilter) ([]entity.Assignment, error)
	CreateAssignment(ctx context.Context, claims entity.AccessClaims, learnerID, scenarioID string) (*entity.Assignment, error)
	TeamAnalytics(ctx context.Context, claims entity.AccessClaims) (*entity.TeamAnalytics, error)
}

type Controller struct {
	logger  *zap.Logger
	service Service
}

func NewController(logger *zap.Logger, service Service) *Controller {
	return &Controller{logger: logger, service: service}
}

func (c *Controller) InitRoutes(r chi.Router) {
	r.Get("/curator/team", c.ListTeam)
	r.Get("/curator/team/sessions", c.ListTeamSessions)
	r.Get("/assignments", c.ListAssignments)
	r.Post("/assignments", c.CreateAssignment)
	r.Get("/curator/team/analytics", c.TeamAnalytics)
}

func (c *Controller) ListTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	limit, offset, ok := readLimitOffset(w, r)
	if !ok {
		return
	}
	team, err := c.service.ListTeam(r.Context(), claims, limit, offset)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, curatorTeamResponse{Team: team})
}

func (c *Controller) ListTeamSessions(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	limit, offset, ok := readLimitOffset(w, r)
	if !ok {
		return
	}
	sessions, err := c.service.ListTeamSessions(r.Context(), claims, limit, offset)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, teamSessionListResponse{Sessions: sessions})
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
	filter := entity.AssignmentFilter{Limit: limit, Offset: offset}
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		status := entity.AssignmentStatus(raw)
		filter.Status = &status
	}
	assignments, err := c.service.ListAssignments(r.Context(), claims, filter)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := assignmentListResponse{Assignments: make([]assignmentResponse, 0, len(assignments))}
	for _, assignment := range assignments {
		resp.Assignments = append(resp.Assignments, assignmentToResponse(assignment))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	var req createAssignmentRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	learnerID, err := entity.NormalizeUUID(req.LearnerID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	scenarioID, err := entity.NormalizeUUID(req.ScenarioID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	assignment, err := c.service.CreateAssignment(r.Context(), claims, learnerID, scenarioID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, assignmentToResponse(*assignment))
}

func (c *Controller) TeamAnalytics(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	analytics, err := c.service.TeamAnalytics(r.Context(), claims)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, analytics)
}

func (c *Controller) writeError(w http.ResponseWriter, r *http.Request, err error) {
	code := httpx.ErrorCodeFromError(err)
	status := http.StatusInternalServerError
	switch code {
	case httpx.ErrInvalidRequest:
		status = http.StatusBadRequest
	case httpx.ErrCuratorRoleRequired, httpx.ErrAssignmentAccessDenied:
		status = http.StatusForbidden
	case httpx.ErrUserNotFound, httpx.ErrScenarioNotFound:
		status = http.StatusNotFound
	case httpx.ErrAssignmentConflict, httpx.ErrScenarioNotAvailable:
		status = http.StatusConflict
	}
	if status == http.StatusInternalServerError {
		c.logger.Error("curator request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
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

func assignmentToResponse(assignment entity.Assignment) assignmentResponse {
	return assignmentResponse{AssignmentID: assignment.ID, CuratorID: assignment.CuratorID, LearnerID: assignment.LearnerID, ScenarioID: assignment.ScenarioID, ScenarioTitle: assignment.ScenarioTitle, Status: assignment.Status, CreatedAt: assignment.CreatedAt, CompletedAt: assignment.CompletedAt}
}
