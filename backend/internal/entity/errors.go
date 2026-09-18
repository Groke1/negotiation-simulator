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

	ErrCuratorRoleRequired        = errors.New("curator role required")
	ErrTrainingSessionNotFound    = errors.New("training session not found")
	ErrTrainingAccessDenied       = errors.New("training access denied")
	ErrTrainingModeNotSupported   = errors.New("training mode not supported")
	ErrInvalidTrainingState       = errors.New("invalid training state")
	ErrTrainingActionNotAvailable = errors.New("training action not available")
	ErrArenaAssignmentNotFound    = errors.New("arena assignment not found")
	ErrArenaSessionNotFound       = errors.New("arena session not found")
	ErrArenaAccessDenied          = errors.New("arena access denied")
	ErrArenaJoinNotAllowed        = errors.New("arena join not allowed")
	ErrInvalidArenaState          = errors.New("invalid arena state")
	ErrArenaParticipantInvalid    = errors.New("invalid arena participants")
	ErrArenaScenarioNotSupported  = errors.New("scenario does not support arena")
	ErrArenaFeedbackParticipant   = errors.New("feedback participant is invalid")
	ErrLLMUnavailable             = errors.New("llm provider unavailable")
	ErrLLMResponseInvalid         = errors.New("llm response is invalid")
)
