package user

import (
	"context"
	"errors"
	"time"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type userRepository interface {
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	GetUserRoles(ctx context.Context, userID string) ([]entity.ProductRole, error)
	AddUserRole(ctx context.Context, userID string, role entity.ProductRole) error
	SetUserOrganization(ctx context.Context, userID, organizationID string) (*entity.User, error)
	ListProfile(ctx context.Context, userID string) ([]entity.ProfileCluster, error)
	ListProfileHistory(ctx context.Context, userID string, from, to *time.Time, limit, offset int32) ([]entity.ProfileHistoryEvent, error)
	ListOrganizationUsers(ctx context.Context, requesterID string, limit, offset int32) ([]entity.OrganizationUser, error)
}

func (s *userService) GetProfile(ctx context.Context, claims entity.AccessClaims, userID string) ([]entity.ProfileCluster, error) {
	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if claims.UserID != userID {
		requester, err := s.userRepository.GetUserByID(ctx, claims.UserID)
		if err != nil {
			return nil, err
		}
		if claims.ActiveRole != entity.RoleCurator || requester.OrganizationID == nil || user.OrganizationID == nil || *requester.OrganizationID != *user.OrganizationID {
			return nil, entity.ErrUserDataAccessDenied
		}
	}
	return s.userRepository.ListProfile(ctx, userID)
}

func (s *userService) GetProfileHistory(ctx context.Context, claims entity.AccessClaims, userID string, from, to *time.Time, limit, offset int32) ([]entity.ProfileHistoryEvent, error) {
	if _, err := s.GetProfile(ctx, claims, userID); err != nil {
		return nil, err
	}
	return s.userRepository.ListProfileHistory(ctx, userID, from, to, limit, offset)
}

func (s *userService) ListOrganizationUsers(ctx context.Context, claims entity.AccessClaims, limit, offset int32) ([]entity.OrganizationUser, error) {
	if claims.ActiveRole != entity.RoleCurator {
		return nil, entity.ErrCuratorRoleRequired
	}
	user, err := s.userRepository.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.OrganizationID == nil {
		return nil, entity.ErrOrganizationRequired
	}
	return s.userRepository.ListOrganizationUsers(ctx, claims.UserID, limit, offset)
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
