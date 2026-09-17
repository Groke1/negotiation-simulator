package organizations

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	CreateOrganization(ctx context.Context, params entity.CreateOrganization) (*entity.Organization, error)
	ListOrganizations(ctx context.Context, params entity.ListOrganizations) ([]entity.Organization, error)
}

type Controller struct {
	logger  *zap.Logger
	service Service
}

type createOrganizationRequest struct {
	Name string `json:"name"`
}

type organizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type organizationListResponse struct {
	Organizations []organizationResponse `json:"organizations"`
}

func NewController(logger *zap.Logger, service Service) *Controller {
	return &Controller{logger: logger, service: service}
}

func (c *Controller) InitRoutes(r chi.Router) {
	r.Get("/organizations", c.ListOrganizations)
}

func (c *Controller) InitInternalRoutes(r chi.Router) {
	r.Post("/organizations", c.CreateOrganization)
}

func (c *Controller) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	params, ok := readListParams(r)
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	items, err := c.service.ListOrganizations(r.Context(), params)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	resp := organizationListResponse{Organizations: make([]organizationResponse, 0, len(items))}
	for _, org := range items {
		resp.Organizations = append(resp.Organizations, organizationToResponse(org))
	}
	pkghttpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req createOrganizationRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !validOrganizationName(req.Name) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	org, err := c.service.CreateOrganization(r.Context(), entity.CreateOrganization{Name: strings.TrimSpace(req.Name)})
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, organizationToResponse(*org))
}

func (c *Controller) writeError(w http.ResponseWriter, r *http.Request, err error) {
	code := httpx.ErrorCodeFromError(err)
	status := http.StatusInternalServerError
	switch code {
	case httpx.ErrOrganizationAlreadyExists:
		status = http.StatusConflict
	case httpx.ErrInvalidRequest:
		status = http.StatusBadRequest
	}
	if status == http.StatusInternalServerError {
		c.logger.Error("organizations request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
	}
	httpx.WriteError(w, status, code)
}

func readListParams(r *http.Request) (entity.ListOrganizations, bool) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if len(query) > 100 {
		return entity.ListOrganizations{}, false
	}
	limit := int32(20)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || value < 1 || value > 100 {
			return entity.ListOrganizations{}, false
		}
		limit = int32(value)
	}
	offset := int32(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || value < 0 {
			return entity.ListOrganizations{}, false
		}
		offset = int32(value)
	}
	return entity.ListOrganizations{Query: query, Limit: limit, Offset: offset}, true
}

func validOrganizationName(name string) bool {
	trimmed := strings.TrimSpace(name)
	return len(trimmed) >= 1 && len(trimmed) <= 200
}

func organizationToResponse(org entity.Organization) organizationResponse {
	return organizationResponse{ID: org.ID, Name: org.Name}
}
