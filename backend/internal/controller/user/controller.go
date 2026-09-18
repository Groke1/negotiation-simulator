package user

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	GetCurrentUser(ctx context.Context, claims entity.AccessClaims) (*entity.CurrentUser, error)
	SetOrganization(ctx context.Context, params entity.SetUserOrganization) error
	ListRoles(ctx context.Context, userID string) ([]entity.ProductRole, error)
	SwitchActiveRole(ctx context.Context, params entity.SwitchRole) (*entity.JWT, error)
	GrantRole(ctx context.Context, params entity.GrantRole) error
	ListOrganizationUsers(ctx context.Context, claims entity.AccessClaims) ([]entity.OrganizationUser, error)
}

type Controller struct {
	logger  *zap.Logger
	service Service
}

func NewController(logger *zap.Logger, service Service) *Controller {
	return &Controller{logger: logger, service: service}
}

func (c *Controller) InitRoutes(r chi.Router) {
	r.Get("/users/me", c.GetCurrentUser)
	r.Put("/users/me/organization", c.SetOrganization)
	r.Get("/roles", c.ListRoles)
	r.Put("/roles/active", c.SwitchActiveRole)
	r.Post("/roles", c.GrantOwnRole)
	r.Get("/organizations/me/users", c.ListOrganizationUsers)
}

func (c *Controller) GrantOwnRole(w http.ResponseWriter, r *http.Request) {
	claims, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
		return
	}
	var req grantRoleRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !req.Role.Valid() {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	if err := c.service.GrantRole(r.Context(), entity.GrantRole{UserID: claims.UserID, Role: req.Role}); err != nil {
		c.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) ListOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	claims, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
		return
	}
	users, err := c.service.ListOrganizationUsers(r.Context(), claims)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	response := organizationUserListResponse{Users: make([]organizationUserResponse, 0, len(users))}
	for _, user := range users {
		response.Users = append(response.Users, organizationUserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Roles:     user.Roles,
			CreatedAt: user.CreatedAt,
		})
	}
	pkghttpx.WriteJSON(w, http.StatusOK, response)
}

func (c *Controller) InitInternalRoutes(r chi.Router) {
	r.Post("/users/{user_id}/roles", c.GrantRole)
}

func (c *Controller) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
		return
	}
	currentUser, err := c.service.GetCurrentUser(r.Context(), claims)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, currentUserToResponse(*currentUser))
}

func (c *Controller) SetOrganization(w http.ResponseWriter, r *http.Request) {
	claims, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
		return
	}
	var req setOrganizationRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	organizationID, err := entity.NormalizeUUID(req.OrganizationID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	if err := c.service.SetOrganization(r.Context(), entity.SetUserOrganization{UserID: claims.UserID, OrganizationID: organizationID}); err != nil {
		c.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) ListRoles(w http.ResponseWriter, r *http.Request) {
	claims, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
		return
	}
	roles, err := c.service.ListRoles(r.Context(), claims.UserID)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, rolesResponse{Items: roles, ActiveRole: claims.ActiveRole})
}

func (c *Controller) SwitchActiveRole(w http.ResponseWriter, r *http.Request) {
	claims, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
		return
	}
	var req switchRoleRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !req.Role.Valid() {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	tokens, err := c.service.SwitchActiveRole(r.Context(), entity.SwitchRole{UserID: claims.UserID, SessionID: claims.SessionID, Role: req.Role})
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, accessTokenResponse{AccessToken: tokens.AccessToken})
}

func (c *Controller) GrantRole(w http.ResponseWriter, r *http.Request) {
	userID, err := entity.NormalizeUUID(chi.URLParam(r, "user_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	var req grantRoleRequest
	if err := pkghttpx.DecodeJSON(w, r, &req); err != nil || !req.Role.Valid() {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	if err := c.service.GrantRole(r.Context(), entity.GrantRole{UserID: userID, Role: req.Role}); err != nil {
		c.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) writeError(w http.ResponseWriter, r *http.Request, err error) {
	code := httpx.ErrorCodeFromError(err)
	status := http.StatusInternalServerError
	switch code {
	case httpx.ErrInvalidRequest, httpx.ErrOrganizationAlreadySet:
		status = http.StatusBadRequest
	case httpx.ErrAccessTokenInvalid, httpx.ErrAccessTokenExpired, httpx.ErrAccessTokenMissing:
		status = http.StatusUnauthorized
	case httpx.ErrRoleNotAssigned:
		status = http.StatusForbidden
	case httpx.ErrOrganizationNotFound, httpx.ErrUserNotFound:
		status = http.StatusNotFound
	case httpx.ErrRoleAlreadyAssigned:
		status = http.StatusConflict
	}
	if status == http.StatusInternalServerError {
		c.logger.Error("users request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
	}
	httpx.WriteError(w, status, code)
}

func currentUserToResponse(user entity.CurrentUser) currentUserResponse {
	var org *organizationResponse
	if user.Organization != nil {
		org = &organizationResponse{ID: user.Organization.ID, Name: user.Organization.Name}
	}
	return currentUserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Roles:        user.Roles,
		ActiveRole:   user.ActiveRole,
		Organization: org,
	}
}
