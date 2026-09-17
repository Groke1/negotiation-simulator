package httpx

import (
	"errors"
	"net/http"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type ErrorCode string

const (
	ErrInvalidRequest ErrorCode = "INVALID_REQUEST"
	ErrInternal       ErrorCode = "INTERNAL_ERROR"

	ErrUsernameAlreadyExists ErrorCode = "USERNAME_ALREADY_EXISTS"
	ErrEmailAlreadyExists    ErrorCode = "EMAIL_ALREADY_EXISTS"
	ErrInvalidCredentials    ErrorCode = "INVALID_CREDENTIALS"

	ErrRefreshTokenMissing ErrorCode = "REFRESH_TOKEN_MISSING"
	ErrRefreshTokenInvalid ErrorCode = "REFRESH_TOKEN_INVALID"
	ErrRefreshTokenExpired ErrorCode = "REFRESH_TOKEN_EXPIRED"
	ErrRefreshTokenRevoked ErrorCode = "REFRESH_TOKEN_REVOKED"
	ErrRefreshTokenReused  ErrorCode = "REFRESH_TOKEN_REUSED"

	ErrAccessTokenMissing ErrorCode = "ACCESS_TOKEN_MISSING"
	ErrAccessTokenInvalid ErrorCode = "ACCESS_TOKEN_INVALID"
	ErrAccessTokenExpired ErrorCode = "ACCESS_TOKEN_EXPIRED"

	ErrOrganizationAlreadyExists ErrorCode = "ORGANIZATION_ALREADY_EXISTS"
	ErrOrganizationNotFound      ErrorCode = "ORGANIZATION_NOT_FOUND"
	ErrOrganizationAlreadySet    ErrorCode = "ORGANIZATION_ALREADY_SET"

	ErrUserNotFound        ErrorCode = "USER_NOT_FOUND"
	ErrRoleAlreadyAssigned ErrorCode = "ROLE_ALREADY_ASSIGNED"
	ErrRoleNotAssigned     ErrorCode = "ROLE_NOT_ASSIGNED"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e ErrorCode) Message() string {
	switch e {
	case ErrInvalidRequest:
		return "Request validation failed"
	case ErrInternal:
		return "Internal server error"
	case ErrUsernameAlreadyExists:
		return "A user with this username already exists"
	case ErrEmailAlreadyExists:
		return "A user with this email already exists"
	case ErrInvalidCredentials:
		return "Invalid email or password"
	case ErrRefreshTokenMissing:
		return "Refresh token is required"
	case ErrRefreshTokenInvalid:
		return "Refresh token is invalid"
	case ErrRefreshTokenExpired:
		return "Refresh token has expired"
	case ErrRefreshTokenRevoked:
		return "Refresh token has been revoked"
	case ErrRefreshTokenReused:
		return "Refresh token has already been used"
	case ErrAccessTokenMissing:
		return "Access token is required"
	case ErrAccessTokenInvalid:
		return "Access token is invalid"
	case ErrAccessTokenExpired:
		return "Access token has expired"
	case ErrOrganizationAlreadyExists:
		return "An organization with this name already exists"
	case ErrOrganizationNotFound:
		return "Organization not found"
	case ErrOrganizationAlreadySet:
		return "User already belongs to an organization"
	case ErrUserNotFound:
		return "User not found"
	case ErrRoleAlreadyAssigned:
		return "Requested role is already assigned to the user"
	case ErrRoleNotAssigned:
		return "Requested role is not assigned to the user"
	default:
		return "Unknown error"
	}
}

func WriteError(w http.ResponseWriter, status int, code ErrorCode) {
	pkghttpx.WriteJSON(w, status, ErrorResponse{Code: string(code), Message: code.Message()})
}

func ErrorCodeFromError(err error) ErrorCode {
	switch {
	case errors.Is(err, entity.ErrUsernameAlreadyExists):
		return ErrUsernameAlreadyExists
	case errors.Is(err, entity.ErrEmailAlreadyExists):
		return ErrEmailAlreadyExists
	case errors.Is(err, entity.ErrInvalidCredentials):
		return ErrInvalidCredentials
	case errors.Is(err, entity.ErrRefreshTokenMissing):
		return ErrRefreshTokenMissing
	case errors.Is(err, entity.ErrRefreshTokenInvalid):
		return ErrRefreshTokenInvalid
	case errors.Is(err, entity.ErrRefreshTokenExpired):
		return ErrRefreshTokenExpired
	case errors.Is(err, entity.ErrRefreshTokenRevoked):
		return ErrRefreshTokenRevoked
	case errors.Is(err, entity.ErrRefreshTokenReused):
		return ErrRefreshTokenReused
	case errors.Is(err, entity.ErrAccessTokenMissing):
		return ErrAccessTokenMissing
	case errors.Is(err, entity.ErrAccessTokenInvalid):
		return ErrAccessTokenInvalid
	case errors.Is(err, entity.ErrAccessTokenExpired):
		return ErrAccessTokenExpired
	case errors.Is(err, entity.ErrOrganizationAlreadyExists):
		return ErrOrganizationAlreadyExists
	case errors.Is(err, entity.ErrOrganizationNotFound):
		return ErrOrganizationNotFound
	case errors.Is(err, entity.ErrOrganizationAlreadySet):
		return ErrOrganizationAlreadySet
	case errors.Is(err, entity.ErrUserNotFound):
		return ErrUserNotFound
	case errors.Is(err, entity.ErrRoleAlreadyAssigned):
		return ErrRoleAlreadyAssigned
	case errors.Is(err, entity.ErrRoleNotAssigned):
		return ErrRoleNotAssigned
	case errors.Is(err, entity.ErrInvalidRequest):
		return ErrInvalidRequest
	default:
		return ErrInternal
	}
}
