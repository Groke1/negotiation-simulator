package user

import (
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type organizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type currentUserResponse struct {
	ID           string                `json:"id"`
	Username     string                `json:"username"`
	Email        string                `json:"email"`
	Roles        []entity.ProductRole  `json:"roles"`
	ActiveRole   entity.ProductRole    `json:"active_role"`
	Organization *organizationResponse `json:"organization"`
}

type setOrganizationRequest struct {
	OrganizationID string `json:"organization_id"`
}

type switchRoleRequest struct {
	Role entity.ProductRole `json:"role"`
}

type grantRoleRequest struct {
	Role entity.ProductRole `json:"role"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type rolesResponse struct {
	Items      []entity.ProductRole `json:"items"`
	ActiveRole entity.ProductRole   `json:"active_role"`
}

type organizationUserResponse struct {
	ID        string               `json:"id"`
	Username  string               `json:"username"`
	Email     string               `json:"email"`
	Roles     []entity.ProductRole `json:"roles"`
	CreatedAt time.Time            `json:"created_at"`
}

type organizationUserListResponse struct {
	Users []organizationUserResponse `json:"users"`
}
