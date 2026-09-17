package entity

import "time"

type Organization struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type CreateOrganization struct {
	Name string
}

type ListOrganizations struct {
	Query  string
	Limit  int32
	Offset int32
}

type SetUserOrganization struct {
	UserID         string
	OrganizationID string
}
