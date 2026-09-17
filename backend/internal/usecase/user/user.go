package user

import (
	"context"
	"errors"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type userRepository interface {
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	GetUserRoles(ctx context.Context, userID string) ([]entity.ProductRole, error)
	AddUserRole(ctx context.Context, userID string, role entity.ProductRole) error
	SetUserOrganization(ctx context.Context, userID, organizationID string) (*entity.User, error)
}

type organizationRepository interface {
	GetOrganizationByID(ctx context.Context, id string) (*entity.Organization, error)
}

type authRepository interface {
	UpdateRefreshTokenActiveRole(ctx context.Context, sessionID, userID string, role entity.ProductRole) (*entity.RefreshToken, error)
}

type accessTokenIssuer interface {
	IssueAccessToken(ctx context.Context, userID string, role entity.ProductRole, sessionID string) (string, error)
}

type userService struct {
	userRepository         userRepository
	organizationRepository organizationRepository
	authRepository         authRepository
	accessTokenIssuer      accessTokenIssuer
}

func NewUserService(userRepository userRepository, organizationRepository organizationRepository, authRepository authRepository, issuer accessTokenIssuer) *userService {
	return &userService{
		userRepository:         userRepository,
		organizationRepository: organizationRepository,
		authRepository:         authRepository,
		accessTokenIssuer:      issuer,
	}
}

func (s *userService) GetCurrentUser(ctx context.Context, claims entity.AccessClaims) (*entity.CurrentUser, error) {
	user, err := s.userRepository.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	roles, err := s.userRepository.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	var org *entity.Organization
	if user.OrganizationID != nil {
		org, err = s.organizationRepository.GetOrganizationByID(ctx, *user.OrganizationID)
		if err != nil {
			return nil, err
		}
	}
	return &entity.CurrentUser{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Roles:        roles,
		ActiveRole:   claims.ActiveRole,
		Organization: org,
	}, nil
}

func (s *userService) SetOrganization(ctx context.Context, params entity.SetUserOrganization) error {
	user, err := s.userRepository.GetUserByID(ctx, params.UserID)
	if err != nil {
		return err
	}
	if user.OrganizationID != nil {
		return entity.ErrOrganizationAlreadySet
	}
	if _, err := s.organizationRepository.GetOrganizationByID(ctx, params.OrganizationID); err != nil {
		return err
	}
	_, err = s.userRepository.SetUserOrganization(ctx, params.UserID, params.OrganizationID)
	return err
}

func (s *userService) ListRoles(ctx context.Context, userID string) ([]entity.ProductRole, error) {
	return s.userRepository.GetUserRoles(ctx, userID)
}

func (s *userService) SwitchActiveRole(ctx context.Context, params entity.SwitchRole) (*entity.JWT, error) {
	roles, err := s.userRepository.GetUserRoles(ctx, params.UserID)
	if err != nil {
		return nil, err
	}
	if !roleAssigned(params.Role, roles) {
		return nil, entity.ErrRoleNotAssigned
	}
	if _, err := s.authRepository.UpdateRefreshTokenActiveRole(ctx, params.SessionID, params.UserID, params.Role); err != nil {
		return nil, err
	}
	accessToken, err := s.accessTokenIssuer.IssueAccessToken(ctx, params.UserID, params.Role, params.SessionID)
	if err != nil {
		return nil, err
	}
	return &entity.JWT{AccessToken: accessToken}, nil
}

func (s *userService) GrantRole(ctx context.Context, params entity.GrantRole) error {
	if _, err := s.userRepository.GetUserByID(ctx, params.UserID); err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return entity.ErrUserNotFound
		}
		return err
	}
	return s.userRepository.AddUserRole(ctx, params.UserID, params.Role)
}

func roleAssigned(role entity.ProductRole, roles []entity.ProductRole) bool {
	for _, item := range roles {
		if item == role {
			return true
		}
	}
	return false
}
