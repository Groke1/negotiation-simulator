package scenario

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
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	ListScenarios(ctx context.Context, claims entity.AccessClaims, params entity.ListScenarios) ([]entity.Scenario, error)
	CreateScenario(ctx context.Context, claims entity.AccessClaims, params entity.CreateScenario) (*entity.Scenario, error)
	GenerateScenario(ctx context.Context, claims entity.AccessClaims, prompt string) (*entity.Scenario, error)
	GetScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.Scenario, error)
	UpdateScenario(ctx context.Context, claims entity.AccessClaims, params entity.UpdateScenario) error
	ValidateScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.ScenarioValidation, error)
	PublishScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.ScenarioVersion, error)
	ArchiveScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) error
	ListScenarioVersions(ctx context.Context, claims entity.AccessClaims, scenarioID string, limit, offset int32) ([]entity.ScenarioVersion, error)
	GetScenarioVersion(ctx context.Context, claims entity.AccessClaims, scenarioID, version string) (*entity.ScenarioVersion, error)
}

type Controller struct {
	logger  *zap.Logger
	service Service
}

func NewController(logger *zap.Logger, service Service) *Controller {
	return &Controller{logger: logger, service: service}
}

func (c *Controller) InitRoutes(router chi.Router) {
	router.Route("/scenarios", func(r chi.Router) {
		r.Get("/", c.ListScenarios)
		r.Post("/", c.CreateScenario)
		r.Post("/generate", c.GenerateScenario)
		r.Route("/{scenario_id}", func(r chi.Router) {
			r.Get("/", c.GetScenario)
			r.Put("/", c.UpdateScenario)
			r.Post("/validate", c.ValidateScenario)
			r.Post("/publish", c.PublishScenario)
			r.Post("/archive", c.ArchiveScenario)
			r.Get("/versions", c.ListScenarioVersions)
			r.Get("/versions/{version}", c.GetScenarioVersion)
		})
	})
}

func (c *Controller) ListScenarios(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	limit, offset, ok := readLimitOffset(w, r)
	if !ok {
		return
	}
	var status *entity.ScenarioStatus
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		parsed := entity.ScenarioStatus(raw)
		if !parsed.Valid() {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
			return
		}
		status = &parsed
	}
	items, err := c.service.ListScenarios(r.Context(), claims, entity.ListScenarios{UserID: claims.UserID, Status: status, Limit: limit, Offset: offset})
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := scenarioListResponse{Scenarios: make([]scenarioSummaryResponse, 0, len(items))}
	for _, item := range items {
		resp.Scenarios = append(resp.Scenarios, scenarioToSummary(item))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) CreateScenario(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	var req scenarioRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validScenarioRequest(req) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	scenario, err := c.service.CreateScenario(r.Context(), claims, createScenarioFromRequest(claims.UserID, req))
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, scenarioCreatedResponse{ID: scenario.ID})
}

func (c *Controller) GenerateScenario(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	var req generateScenarioRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || len(strings.TrimSpace(req.Prompt)) < 20 || len(req.Prompt) > 20000 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	scenario, err := c.service.GenerateScenario(r.Context(), claims, req.Prompt)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	generationID, err := entity.NewUUID()
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, generateScenarioResponse{ScenarioDraft: scenarioToResponse(*scenario), AIGenerationID: generationID, Provider: "ollama"})
}

func (c *Controller) GetScenario(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	scenario, err := c.service.GetScenario(r.Context(), claims, scenarioID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, scenarioToResponse(*scenario))
}

func (c *Controller) UpdateScenario(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	var req updateScenarioRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validUpdateScenarioRequest(req) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	if err := c.service.UpdateScenario(r.Context(), claims, updateScenarioFromRequest(scenarioID, claims.UserID, req)); err != nil {
		c.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) ValidateScenario(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	validation, err := c.service.ValidateScenario(r.Context(), claims, scenarioID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, validation)
}

func (c *Controller) PublishScenario(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	version, err := c.service.PublishScenario(r.Context(), claims, scenarioID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, publishScenarioResponse{Version: version.Version, PublishedAt: version.PublishedAt})
}

func (c *Controller) ArchiveScenario(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	if err := c.service.ArchiveScenario(r.Context(), claims, scenarioID); err != nil {
		c.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) ListScenarioVersions(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	limit, offset, ok := readLimitOffset(w, r)
	if !ok {
		return
	}
	versions, err := c.service.ListScenarioVersions(r.Context(), claims, scenarioID, limit, offset)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := scenarioVersionListResponse{Versions: make([]scenarioVersionSummaryResponse, 0, len(versions))}
	for _, version := range versions {
		resp.Versions = append(resp.Versions, scenarioVersionSummaryResponse{Version: version.Version, PublishedAt: version.PublishedAt})
	}
	pkghttpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) GetScenarioVersion(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	version := strings.TrimSpace(chi.URLParam(r, "version"))
	if version == "" || len(version) > 32 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	scenarioVersion, err := c.service.GetScenarioVersion(r.Context(), claims, scenarioID, version)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, scenarioVersionResponse{Version: scenarioVersion.Version, Snapshot: scenarioVersion.Snapshot, PublishedAt: scenarioVersion.PublishedAt})
}

func (c *Controller) writeError(w http.ResponseWriter, r *http.Request, err error) {
	code := httpx.ErrorCodeFromError(err)
	status := http.StatusInternalServerError
	switch code {
	case httpx.ErrInvalidRequest, httpx.ErrInvalidScenarioJSON:
		status = http.StatusBadRequest
	case httpx.ErrAccessTokenMissing, httpx.ErrAccessTokenInvalid, httpx.ErrAccessTokenExpired:
		status = http.StatusUnauthorized
	case httpx.ErrMethodistRoleRequired, httpx.ErrScenarioNotOwned:
		status = http.StatusForbidden
	case httpx.ErrScenarioNotFound, httpx.ErrScenarioVersionNotFound:
		status = http.StatusNotFound
	case httpx.ErrOrganizationRequired, httpx.ErrScenarioVersionAlreadyExists, httpx.ErrInvalidScenarioStatus:
		status = http.StatusConflict
	case httpx.ErrScenarioNotValidated, httpx.ErrScenarioValidationFailed:
		status = http.StatusUnprocessableEntity
	case httpx.ErrScenarioGenerationRateLimited:
		status = http.StatusTooManyRequests
	case httpx.ErrScenarioGenerationFailed:
		status = http.StatusBadGateway
	}
	if status == http.StatusInternalServerError {
		c.logger.Error("scenario request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
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

func claimsAndScenarioID(w http.ResponseWriter, r *http.Request) (entity.AccessClaims, string, bool) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return entity.AccessClaims{}, "", false
	}
	scenarioID, err := entity.NormalizeUUID(chi.URLParam(r, "scenario_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return entity.AccessClaims{}, "", false
	}
	return claims, scenarioID, true
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

func validScenarioRequest(req scenarioRequest) bool {
	return validString(req.Version, 1, 32) && validString(req.Title, 1, 200) && req.Difficulty.Valid() && len(req.ModeSupport) > 0 && len(req.Channels) > 0 && rawObjectOrArray(req.Context) && rawObjectOrArray(req.Roles) && rawObjectOrArray(req.Goals) && rawObjectOrArray(req.Interests) && rawObjectOrArray(req.Constraints) && rawObjectOrArray(req.Parameters) && rawObjectOrArray(req.PrivateInformation) && rawObjectOrArray(req.ObservationRules) && rawObjectOrArray(req.CompletionConditions) && rawObjectOrArray(req.Nodes)
}

func validUpdateScenarioRequest(req updateScenarioRequest) bool {
	set := false
	if req.Version != nil {
		set = true
		if !validString(*req.Version, 1, 32) {
			return false
		}
	}
	if req.Title != nil {
		set = true
		if !validString(*req.Title, 1, 200) {
			return false
		}
	}
	if req.Difficulty != nil {
		set = true
		if !req.Difficulty.Valid() {
			return false
		}
	}
	jsonFields := []*json.RawMessage{req.Context, req.Roles, req.Goals, req.Interests, req.Constraints, req.Parameters, req.PrivateInformation, req.ObservationRules, req.CompletionConditions, req.Nodes}
	for _, field := range jsonFields {
		if field != nil {
			set = true
			if !rawObjectOrArray(*field) {
				return false
			}
		}
	}
	if req.ModeSupport != nil || req.Competencies != nil || req.Channels != nil {
		set = true
	}
	return set
}

func validString(value string, min, max int) bool {
	trimmed := strings.TrimSpace(value)
	return len(trimmed) >= min && len(trimmed) <= max
}

func rawObjectOrArray(raw json.RawMessage) bool {
	if len(raw) == 0 || !json.Valid(raw) {
		return false
	}
	trimmed := strings.TrimSpace(string(raw))
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}

func createScenarioFromRequest(userID string, req scenarioRequest) entity.CreateScenario {
	return entity.CreateScenario{UserID: userID, Version: strings.TrimSpace(req.Version), Title: strings.TrimSpace(req.Title), ModeSupport: req.ModeSupport, Context: req.Context, Roles: req.Roles, Goals: req.Goals, Interests: req.Interests, Constraints: req.Constraints, Parameters: req.Parameters, PrivateInformation: req.PrivateInformation, Difficulty: req.Difficulty, Competencies: req.Competencies, ObservationRules: req.ObservationRules, CompletionConditions: req.CompletionConditions, Channels: req.Channels, Nodes: req.Nodes}
}

func updateScenarioFromRequest(scenarioID, userID string, req updateScenarioRequest) entity.UpdateScenario {
	return entity.UpdateScenario{ScenarioID: scenarioID, UserID: userID, Version: req.Version, Title: req.Title, ModeSupport: req.ModeSupport, Context: req.Context, Roles: req.Roles, Goals: req.Goals, Interests: req.Interests, Constraints: req.Constraints, Parameters: req.Parameters, PrivateInformation: req.PrivateInformation, Difficulty: req.Difficulty, Competencies: req.Competencies, ObservationRules: req.ObservationRules, CompletionConditions: req.CompletionConditions, Channels: req.Channels, Nodes: req.Nodes}
}

func scenarioToSummary(s entity.Scenario) scenarioSummaryResponse {
	return scenarioSummaryResponse{ScenarioID: s.ID, Title: s.Title, Version: s.Version, Status: s.Status, Difficulty: string(s.Difficulty), Channels: s.Channels, UpdatedAt: s.UpdatedAt}
}

func scenarioToResponse(s entity.Scenario) scenarioResponse {
	return scenarioResponse{ScenarioID: s.ID, Version: s.Version, Title: s.Title, ModeSupport: s.ModeSupport, Context: s.Context, Roles: s.Roles, Goals: s.Goals, Interests: s.Interests, Constraints: s.Constraints, Parameters: s.Parameters, PrivateInformation: s.PrivateInformation, Difficulty: s.Difficulty, Competencies: s.Competencies, ObservationRules: s.ObservationRules, CompletionConditions: s.CompletionConditions, Channels: s.Channels, Nodes: s.Nodes}
}
