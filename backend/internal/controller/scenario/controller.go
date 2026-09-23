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
	ListCompetencyClusters(ctx context.Context) ([]entity.CompetencyCluster, error)
	ListSemanticActions(ctx context.Context) ([]entity.SemanticActionDefinition, error)
	ListScenarios(ctx context.Context, claims entity.AccessClaims, params entity.ListScenarios) ([]entity.Scenario, error)
	CreateScenario(ctx context.Context, claims entity.AccessClaims, params entity.CreateScenario) (*entity.Scenario, error)
	GetScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.Scenario, error)
	UpdateScenario(ctx context.Context, claims entity.AccessClaims, params entity.UpdateScenario) error
	ValidateScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.ScenarioValidation, error)
	PublishScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.Scenario, *entity.ScenarioValidation, error)
	ArchiveScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) error
	CloneScenario(ctx context.Context, claims entity.AccessClaims, scenarioID string) (*entity.Scenario, error)
	AddFavorite(ctx context.Context, claims entity.AccessClaims, scenarioID string) error
	RemoveFavorite(ctx context.Context, claims entity.AccessClaims, scenarioID string) error
	GetPersonalStats(ctx context.Context, userID, scenarioID string) (*entity.ScenarioPersonalStats, error)
	ExpandContext(ctx context.Context, claims entity.AccessClaims, scenarioID string, options entity.AIGenerationOptions) (string, error)
	GenerateDraft(ctx context.Context, claims entity.AccessClaims, scenarioID string, options entity.AIGenerationOptions) (*entity.ScenarioDraftProposal, *entity.ScenarioValidation, error)
	GenerateBlock(ctx context.Context, claims entity.AccessClaims, scenarioID, blockID string, options entity.AIGenerationOptions) (*entity.BlockContentProposal, *entity.ScenarioValidation, error)
	GenerateDialogue(ctx context.Context, claims entity.AccessClaims, scenarioID, blockID, learnerRoleID string, options entity.AIGenerationOptions) (*entity.DeterministicDialogue, *entity.ScenarioValidation, error)
}

type Controller struct {
	logger  *zap.Logger
	service Service
}

func NewController(logger *zap.Logger, service Service) *Controller {
	return &Controller{logger: logger, service: service}
}

func (c *Controller) InitRoutes(router chi.Router) {
	router.Get("/competency-clusters", c.ListCompetencyClusters)
	router.Get("/semantic-actions", c.ListSemanticActions)
	router.Route("/scenarios", func(r chi.Router) {
		r.Get("/", c.ListScenarios)
		r.Post("/", c.CreateScenario)
		r.Route("/{scenario_id}", func(r chi.Router) {
			r.Get("/", c.GetScenario)
			r.Put("/", c.UpdateScenario)
			r.Post("/validate", c.ValidateScenario)
			r.Post("/publish", c.PublishScenario)
			r.Post("/archive", c.ArchiveScenario)
			r.Post("/clone", c.CloneScenario)
			r.Put("/favorite", c.AddFavorite)
			r.Delete("/favorite", c.RemoveFavorite)
			r.Post("/ai/expand-context", c.AIExpandContext)
			r.Post("/ai/generate-draft", c.AIGenerateDraft)
			r.Post("/blocks/{block_id}/ai/generate", c.AIGenerateBlock)
			r.Post("/blocks/{block_id}/ai/generate-dialogue", c.AIGenerateDialogue)
		})
	})
}

func (c *Controller) ListCompetencyClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := c.service.ListCompetencyClusters(r.Context())
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := competencyClusterListResponse{Clusters: make([]competencyClusterResponse, 0, len(clusters))}
	for _, cluster := range clusters {
		resp.Clusters = append(resp.Clusters, competencyClusterResponse{ID: cluster.ID, Code: cluster.Code, Name: cluster.Name, Description: cluster.Description})
	}
	pkghttpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) ListSemanticActions(w http.ResponseWriter, r *http.Request) {
	actions, err := c.service.ListSemanticActions(r.Context())
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := semanticActionListResponse{Actions: make([]semanticActionResponse, 0, len(actions))}
	for _, action := range actions {
		resp.Actions = append(resp.Actions, semanticActionResponse{Code: action.Code, Name: action.Name, Description: action.Description})
	}
	pkghttpx.WriteJSON(w, http.StatusOK, resp)
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
	params := entity.ListScenarios{UserID: claims.UserID, Limit: limit, Offset: offset}
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		parsed := entity.ScenarioStatus(raw)
		if !parsed.Valid() {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
			return
		}
		params.Status = &parsed
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("difficulty")); raw != "" {
		parsed := entity.ScenarioDifficulty(raw)
		if !parsed.Valid() {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
			return
		}
		params.Difficulty = &parsed
	}
	params.Tag = optionalQueryString(r, "tag")
	params.Runtime = optionalQueryString(r, "runtime")
	params.Channel = optionalQueryString(r, "channel")
	var okBool bool
	if params.Favorite, okBool = optionalQueryBool(w, r, "favorite"); !okBool {
		return
	}
	if params.AssignedByCurator, okBool = optionalQueryBool(w, r, "assigned_by_curator"); !okBool {
		return
	}
	if params.Completed, okBool = optionalQueryBool(w, r, "completed"); !okBool {
		return
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("owned_by_me")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
			return
		}
		params.OwnedByMe = value
	}
	items, err := c.service.ListScenarios(r.Context(), claims, params)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := scenarioListResponse{Scenarios: make([]scenarioSummaryResponse, 0, len(items))}
	for _, item := range items {
		stats, err := c.service.GetPersonalStats(r.Context(), claims.UserID, item.ID)
		if err != nil {
			c.writeError(w, r, err)
			return
		}
		resp.Scenarios = append(resp.Scenarios, scenarioToSummary(item, claims.UserID, stats))
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
	pkghttpx.WriteJSON(w, http.StatusCreated, scenarioToResponse(*scenario, claims.UserID))
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
	pkghttpx.WriteJSON(w, http.StatusOK, scenarioToResponse(*scenario, claims.UserID))
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
	scenario, err := c.service.GetScenario(r.Context(), claims, scenarioID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, scenarioToResponse(*scenario, claims.UserID))
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
	published, validation, err := c.service.PublishScenario(r.Context(), claims, scenarioID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	warnings := make([]entity.ScenarioValidationIssue, 0)
	for _, issue := range validation.Issues {
		if issue.Severity == "warning" {
			warnings = append(warnings, issue)
		}
	}
	publishedAt := published.UpdatedAt
	if published.PublishedAt != nil {
		publishedAt = *published.PublishedAt
	}
	pkghttpx.WriteJSON(w, http.StatusOK, publishScenarioResponse{ScenarioID: scenarioID, Status: string(entity.ScenarioStatusPublished), PublishedAt: publishedAt, Warnings: warnings})
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

func (c *Controller) CloneScenario(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	scenario, err := c.service.CloneScenario(r.Context(), claims, scenarioID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, scenarioToResponse(*scenario, claims.UserID))
}

func (c *Controller) AddFavorite(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	if err := c.service.AddFavorite(r.Context(), claims, scenarioID); err != nil {
		c.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	if err := c.service.RemoveFavorite(r.Context(), claims, scenarioID); err != nil {
		c.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) AIExpandContext(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	options, ok := readAIOptions(w, r)
	if !ok {
		return
	}
	expanded, err := c.service.ExpandContext(r.Context(), claims, scenarioID, options)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, expandContextResponse{ExpandedContext: expanded})
}

func (c *Controller) AIGenerateDraft(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return
	}
	options, ok := readAIOptions(w, r)
	if !ok {
		return
	}
	proposal, validation, err := c.service.GenerateDraft(r.Context(), claims, scenarioID, options)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, generateScenarioDraftResponse{Proposal: scenarioDraftProposalResponse{ExpandedContext: proposal.ExpandedContext, Blocks: proposal.Blocks}, Validation: validation})
}

func (c *Controller) AIGenerateBlock(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, blockID, ok := claimsScenarioAndBlockID(w, r)
	if !ok {
		return
	}
	options, ok := readAIOptions(w, r)
	if !ok {
		return
	}
	proposal, validation, err := c.service.GenerateBlock(r.Context(), claims, scenarioID, blockID, options)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, generateBlockResponse{BlockID: blockID, Proposal: *proposal, Validation: validation})
}

func (c *Controller) AIGenerateDialogue(w http.ResponseWriter, r *http.Request) {
	claims, scenarioID, blockID, ok := claimsScenarioAndBlockID(w, r)
	if !ok {
		return
	}
	var req generateDialogueRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.LearnerRoleID) == "" || len(req.Instructions) > 5000 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	dialogue, validation, err := c.service.GenerateDialogue(r.Context(), claims, scenarioID, blockID, strings.TrimSpace(req.LearnerRoleID), entity.AIGenerationOptions{Instructions: strings.TrimSpace(req.Instructions)})
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, generateDialogueResponse{BlockID: blockID, LearnerRoleID: strings.TrimSpace(req.LearnerRoleID), Proposal: *dialogue, Validation: validation})
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
	case httpx.ErrScenarioNotFound:
		status = http.StatusNotFound
	case httpx.ErrBlockNotFound:
		status = http.StatusNotFound
	case httpx.ErrOrganizationRequired, httpx.ErrInvalidScenarioStatus, httpx.ErrScenarioNotDraft:
		status = http.StatusConflict
	case httpx.ErrScenarioValidationFailed:
		status = http.StatusUnprocessableEntity
	case httpx.ErrAIGenerationContextIncomplete:
		status = http.StatusUnprocessableEntity
	case httpx.ErrAIGenerationFailed, httpx.ErrAIInvalidStructuredOutput:
		status = http.StatusBadGateway
	case httpx.ErrAIProviderUnavailable:
		status = http.StatusServiceUnavailable
	case httpx.ErrAIGenerationTimeout:
		status = http.StatusGatewayTimeout
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

func claimsScenarioAndBlockID(w http.ResponseWriter, r *http.Request) (entity.AccessClaims, string, string, bool) {
	claims, scenarioID, ok := claimsAndScenarioID(w, r)
	if !ok {
		return entity.AccessClaims{}, "", "", false
	}
	blockID, err := entity.NormalizeUUID(chi.URLParam(r, "block_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return entity.AccessClaims{}, "", "", false
	}
	return claims, scenarioID, blockID, true
}

func readAIOptions(w http.ResponseWriter, r *http.Request) (entity.AIGenerationOptions, bool) {
	if r.Body == nil || r.ContentLength == 0 {
		return entity.AIGenerationOptions{}, true
	}
	var req aiGenerationOptionsRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || len(req.Instructions) > 5000 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return entity.AIGenerationOptions{}, false
	}
	return entity.AIGenerationOptions{Instructions: strings.TrimSpace(req.Instructions)}, true
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
	return validString(req.Title, 1, 200)
}

func validUpdateScenarioRequest(req updateScenarioRequest) bool {
	return validString(req.Title, 1, 200) &&
		validString(req.Description, 1, 4000) &&
		req.Difficulty.Valid() &&
		len(req.SupportedRuntimes) > 0 &&
		len(req.SupportedChannels) > 0 &&
		rawObjectOrArray(req.Roles) &&
		rawObjectOrArray(req.Blocks)
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
	return entity.CreateScenario{
		UserID:               userID,
		Title:                strings.TrimSpace(req.Title),
		ModeSupport:          []string{"deterministic", "virtual_partner"},
		Context:              entity.ScenarioContextJSON("", "", nil),
		Roles:                json.RawMessage("[]"),
		Goals:                json.RawMessage("{}"),
		Interests:            json.RawMessage("{}"),
		Constraints:          json.RawMessage("{}"),
		Parameters:           json.RawMessage("[]"),
		PrivateInformation:   json.RawMessage("{}"),
		Difficulty:           entity.ScenarioDifficultyEasy,
		Competencies:         []string{},
		ObservationRules:     json.RawMessage("[]"),
		CompletionConditions: json.RawMessage("[]"),
		Channels:             []string{"chat"},
		Nodes:                json.RawMessage("[]"),
	}
}

func updateScenarioFromRequest(scenarioID, userID string, req updateScenarioRequest) entity.UpdateScenario {
	title := strings.TrimSpace(req.Title)
	context := entity.ScenarioContextJSON(req.Description, req.ExpandedContext, req.Tags)
	difficulty := req.Difficulty
	return entity.UpdateScenario{
		ScenarioID:   scenarioID,
		UserID:       userID,
		Title:        &title,
		ModeSupport:  &req.SupportedRuntimes,
		Context:      &context,
		Roles:        &req.Roles,
		Difficulty:   &difficulty,
		Competencies: &req.TargetClusterIDs,
		Channels:     &req.SupportedChannels,
		Nodes:        &req.Blocks,
	}
}

func optionalQueryString(r *http.Request, key string) *string {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil
	}
	return &raw
}

func optionalQueryBool(w http.ResponseWriter, r *http.Request, key string) (*bool, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, true
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return nil, false
	}
	return &value, true
}

func scenarioToSummary(s entity.Scenario, userID string, stats *entity.ScenarioPersonalStats) scenarioSummaryResponse {
	doc := entity.ScenarioDocumentFromScenario(s)
	resp := scenarioSummaryResponse{
		ScenarioID:          s.ID,
		Title:               s.Title,
		Description:         doc.Description,
		Status:              s.Status,
		Difficulty:          string(s.Difficulty),
		Tags:                doc.Tags,
		SupportedRuntimes:   s.ModeSupport,
		SupportedChannels:   s.Channels,
		CreatedBy:           stringPtrValue(s.CreatedBy),
		IsOwnedByMe:         s.CreatedBy != nil && *s.CreatedBy == userID,
		UpdatedAt:           s.UpdatedAt,
		IsFavorite:          stats != nil && stats.IsFavorite,
		IsAssignedByCurator: stats != nil && stats.IsAssignedByCurator,
		IsCompleted:         stats != nil && stats.IsCompleted,
	}
	if stats != nil {
		resp.AttemptsCount = stats.AttemptsCount
		if stats.BestAttempt != nil {
			resp.BestAttempt = &bestTrainingAttempt{SessionID: stats.BestAttempt.SessionID, Score: stats.BestAttempt.Score, CompletedAt: stats.BestAttempt.CompletedAt}
		}
	}
	return resp
}

func scenarioToResponse(s entity.Scenario, userID string) scenarioResponse {
	doc := entity.ScenarioDocumentFromScenario(s)
	return scenarioResponse{
		ScenarioID:        doc.ScenarioID,
		Title:             doc.Title,
		Description:       doc.Description,
		ExpandedContext:   doc.ExpandedContext,
		Status:            doc.Status,
		Difficulty:        doc.Difficulty,
		Tags:              doc.Tags,
		SupportedRuntimes: doc.SupportedRuntimes,
		SupportedChannels: doc.SupportedChannels,
		TargetClusterIDs:  doc.TargetClusterIDs,
		Roles:             doc.Roles,
		Blocks:            doc.Blocks,
		CreatedBy:         stringPtrValue(s.CreatedBy),
		IsOwnedByMe:       s.CreatedBy != nil && *s.CreatedBy == userID,
		CreatedAt:         doc.CreatedAt,
		UpdatedAt:         doc.UpdatedAt,
		PublishedAt:       doc.PublishedAt,
	}
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
