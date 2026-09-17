package entity

import "time"

type ProductRole string

const (
	RoleLearner   ProductRole = "learner"
	RoleMethodist ProductRole = "methodist"
	RoleCurator   ProductRole = "curator"
)

func (r ProductRole) Valid() bool {
	switch r {
	case RoleLearner, RoleMethodist, RoleCurator:
		return true
	default:
		return false
	}
}

type User struct {
	ID             string
	OrganizationID *string
	Username       string
	Email          string
	PasswordHash   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RegisterUser struct {
	Username string
	Email    string
	Password string
}

type LoginUser struct {
	Email    string
	Password string
}

type CurrentUser struct {
	ID           string
	Username     string
	Email        string
	Roles        []ProductRole
	ActiveRole   ProductRole
	Organization *Organization
}

type GrantRole struct {
	UserID string
	Role   ProductRole
}

type SwitchRole struct {
	UserID    string
	SessionID string
	Role      ProductRole
}
