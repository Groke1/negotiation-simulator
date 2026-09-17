package entity

import "time"

type JWT struct {
	AccessToken  string
	RefreshToken string
}

type AccessClaims struct {
	UserID     string
	ActiveRole ProductRole
	SessionID  string
	ExpiresAt  time.Time
}

type RefreshToken struct {
	ID         string
	UserID     string
	TokenHash  string
	ActiveRole ProductRole
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	UsedAt     *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
