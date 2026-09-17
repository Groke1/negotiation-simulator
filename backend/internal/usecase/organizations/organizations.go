package organizations

import (
	"context"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type organizationRepository interface {
	CreateOrganization(ctx context.Context, org entity.Organization) (*entity.Organization, error)
	ListOrganizations(ctx context.Context, params entity.ListOrganizations) ([]entity.Organization, error)
}

type organizationService struct {
	repository organizationRepository
}

func NewOrganizationService(repository organizationRepository) *organizationService {
	return &organizationService{repository: repository}
}

func (s *organizationService) CreateOrganization(ctx context.Context, params entity.CreateOrganization) (*entity.Organization, error) {
	id, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	return s.repository.CreateOrganization(ctx, entity.Organization{ID: id, Name: strings.TrimSpace(params.Name)})
}

func (s *organizationService) ListOrganizations(ctx context.Context, params entity.ListOrganizations) ([]entity.Organization, error) {
	params.Query = strings.TrimSpace(params.Query)
	return s.repository.ListOrganizations(ctx, params)
}
