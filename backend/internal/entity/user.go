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

// ActiveRole is empty until a user selects their first product role.  Keeping
// this state explicit is important: registration must not silently grant a
// learner permission.
const NoActiveRole ProductRole = ""

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

type OrganizationUser struct {
	ID        string        `json:"id"`
	Username  string        `json:"username"`
	Email     string        `json:"email"`
	Roles     []ProductRole `json:"roles"`
	CreatedAt time.Time     `json:"created_at"`
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

type ProfileCluster struct {
	ClusterID string     `json:"cluster_id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Score     int        `json:"score"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type ProfileHistoryEvent struct {
	ID        string                 `json:"-"`
	Source    string                 `json:"source"`
	SourceID  string                 `json:"source_id"`
	Changes   []ProfileClusterChange `json:"changes"`
	CreatedAt time.Time              `json:"created_at"`
}

type ProfileClusterChange struct {
	ClusterID string `json:"cluster_id"`
	Before    int    `json:"before"`
	After     int    `json:"after"`
	Delta     int    `json:"delta"`
}
