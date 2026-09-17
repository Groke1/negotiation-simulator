package entity

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInternal       = errors.New("internal error")

	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")

	ErrRefreshTokenMissing = errors.New("refresh token missing")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")
	ErrRefreshTokenReused  = errors.New("refresh token reused")

	ErrAccessTokenMissing = errors.New("access token missing")
	ErrAccessTokenInvalid = errors.New("access token invalid")
	ErrAccessTokenExpired = errors.New("access token expired")

	ErrOrganizationAlreadyExists = errors.New("organization already exists")
	ErrOrganizationNotFound      = errors.New("organization not found")
	ErrOrganizationAlreadySet    = errors.New("organization already set")

	ErrUserNotFound        = errors.New("user not found")
	ErrRoleAlreadyAssigned = errors.New("role already assigned")
	ErrRoleNotAssigned     = errors.New("role not assigned")

	ErrMethodistRoleRequired         = errors.New("methodist role required")
	ErrOrganizationRequired          = errors.New("organization required")
	ErrScenarioNotFound              = errors.New("scenario not found")
	ErrScenarioNotOwned              = errors.New("scenario not owned")
	ErrInvalidScenarioStatus         = errors.New("invalid scenario status")
	ErrInvalidScenarioJSON           = errors.New("invalid scenario json")
	ErrScenarioVersionAlreadyExists  = errors.New("scenario version already exists")
	ErrScenarioNotValidated          = errors.New("scenario not validated")
	ErrScenarioValidationFailed      = errors.New("scenario validation failed")
	ErrScenarioVersionNotFound       = errors.New("scenario version not found")
	ErrScenarioGenerationRateLimited = errors.New("scenario generation rate limited")
	ErrScenarioGenerationFailed      = errors.New("scenario generation failed")
)
