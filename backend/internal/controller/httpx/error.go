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

	ErrMethodistRoleRequired         ErrorCode = "METHODIST_ROLE_REQUIRED"
	ErrOrganizationRequired          ErrorCode = "ORGANIZATION_REQUIRED"
	ErrScenarioNotFound              ErrorCode = "SCENARIO_NOT_FOUND"
	ErrScenarioNotOwned              ErrorCode = "SCENARIO_NOT_OWNED"
	ErrInvalidScenarioStatus         ErrorCode = "INVALID_SCENARIO_STATUS"
	ErrInvalidScenarioJSON           ErrorCode = "INVALID_SCENARIO_JSON"
	ErrScenarioVersionAlreadyExists  ErrorCode = "SCENARIO_VERSION_ALREADY_EXISTS"
	ErrScenarioNotValidated          ErrorCode = "SCENARIO_NOT_VALIDATED"
	ErrScenarioValidationFailed      ErrorCode = "SCENARIO_VALIDATION_FAILED"
	ErrScenarioVersionNotFound       ErrorCode = "SCENARIO_VERSION_NOT_FOUND"
	ErrScenarioGenerationRateLimited ErrorCode = "SCENARIO_GENERATION_RATE_LIMITED"
	ErrScenarioGenerationFailed      ErrorCode = "SCENARIO_GENERATION_FAILED"
	ErrCuratorRoleRequired           ErrorCode = "CURATOR_ROLE_REQUIRED"
	ErrTrainingSessionNotFound       ErrorCode = "TRAINING_SESSION_NOT_FOUND"
	ErrTrainingAccessDenied          ErrorCode = "TRAINING_ACCESS_DENIED"
	ErrTrainingModeNotSupported      ErrorCode = "TRAINING_MODE_NOT_SUPPORTED"
	ErrInvalidTrainingState          ErrorCode = "INVALID_TRAINING_STATE"
	ErrTrainingActionNotAvailable    ErrorCode = "ACTION_NOT_AVAILABLE"
	ErrArenaAssignmentNotFound       ErrorCode = "ARENA_ASSIGNMENT_NOT_FOUND"
	ErrArenaSessionNotFound          ErrorCode = "ARENA_SESSION_NOT_FOUND"
	ErrArenaAccessDenied             ErrorCode = "ARENA_ACCESS_DENIED"
	ErrArenaJoinNotAllowed           ErrorCode = "ARENA_JOIN_NOT_ALLOWED"
	ErrInvalidArenaState             ErrorCode = "INVALID_ARENA_STATE"
	ErrArenaParticipantInvalid       ErrorCode = "ARENA_PARTICIPANTS_INVALID"
	ErrArenaScenarioNotSupported     ErrorCode = "ARENA_SCENARIO_NOT_SUPPORTED"
	ErrArenaFeedbackParticipant      ErrorCode = "ARENA_FEEDBACK_PARTICIPANT_INVALID"
	ErrLLMUnavailable                ErrorCode = "LLM_UNAVAILABLE"
	ErrLLMResponseInvalid            ErrorCode = "LLM_RESPONSE_INVALID"
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
	case ErrMethodistRoleRequired:
		return "Active role must be methodist"
	case ErrOrganizationRequired:
		return "User must belong to an organization"
	case ErrScenarioNotFound:
		return "Scenario not found"
	case ErrScenarioNotOwned:
		return "Scenario is owned by another user"
	case ErrInvalidScenarioStatus:
		return "Scenario status does not allow this operation"
	case ErrInvalidScenarioJSON:
		return "Scenario JSON is invalid"
	case ErrScenarioVersionAlreadyExists:
		return "Scenario version already exists"
	case ErrScenarioNotValidated:
		return "Scenario must be validated before publication"
	case ErrScenarioValidationFailed:
		return "Scenario validation failed"
	case ErrScenarioVersionNotFound:
		return "Scenario version not found"
	case ErrScenarioGenerationRateLimited:
		return "Scenario generation rate limit exceeded"
	case ErrScenarioGenerationFailed:
		return "Failed to generate a scenario"
	case ErrCuratorRoleRequired:
		return "Active role must be curator"
	case ErrTrainingSessionNotFound:
		return "Training session not found"
	case ErrTrainingAccessDenied:
		return "Access to this training session is denied"
	case ErrTrainingModeNotSupported:
		return "Training mode is not supported by this scenario"
	case ErrInvalidTrainingState:
		return "Training session state does not allow this operation"
	case ErrTrainingActionNotAvailable:
		return "Action is not available in the current node"
	case ErrArenaAssignmentNotFound:
		return "Arena assignment not found"
	case ErrArenaSessionNotFound:
		return "Arena session not found"
	case ErrArenaAccessDenied:
		return "Access to this arena is denied"
	case ErrArenaJoinNotAllowed:
		return "Arena cannot be joined at this time"
	case ErrInvalidArenaState:
		return "Arena session state does not allow this operation"
	case ErrArenaParticipantInvalid:
		return "Arena participants are invalid"
	case ErrArenaScenarioNotSupported:
		return "Scenario does not support arena"
	case ErrArenaFeedbackParticipant:
		return "Feedback participant is invalid"
	case ErrLLMUnavailable:
		return "LLM provider is unavailable"
	case ErrLLMResponseInvalid:
		return "LLM provider returned an invalid response"
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
	case errors.Is(err, entity.ErrMethodistRoleRequired):
		return ErrMethodistRoleRequired
	case errors.Is(err, entity.ErrOrganizationRequired):
		return ErrOrganizationRequired
	case errors.Is(err, entity.ErrScenarioNotFound):
		return ErrScenarioNotFound
	case errors.Is(err, entity.ErrScenarioNotOwned):
		return ErrScenarioNotOwned
	case errors.Is(err, entity.ErrInvalidScenarioStatus):
		return ErrInvalidScenarioStatus
	case errors.Is(err, entity.ErrInvalidScenarioJSON):
		return ErrInvalidScenarioJSON
	case errors.Is(err, entity.ErrScenarioVersionAlreadyExists):
		return ErrScenarioVersionAlreadyExists
	case errors.Is(err, entity.ErrScenarioNotValidated):
		return ErrScenarioNotValidated
	case errors.Is(err, entity.ErrScenarioValidationFailed):
		return ErrScenarioValidationFailed
	case errors.Is(err, entity.ErrScenarioVersionNotFound):
		return ErrScenarioVersionNotFound
	case errors.Is(err, entity.ErrScenarioGenerationRateLimited):
		return ErrScenarioGenerationRateLimited
	case errors.Is(err, entity.ErrScenarioGenerationFailed):
		return ErrScenarioGenerationFailed
	case errors.Is(err, entity.ErrCuratorRoleRequired):
		return ErrCuratorRoleRequired
	case errors.Is(err, entity.ErrTrainingSessionNotFound):
		return ErrTrainingSessionNotFound
	case errors.Is(err, entity.ErrTrainingAccessDenied):
		return ErrTrainingAccessDenied
	case errors.Is(err, entity.ErrTrainingModeNotSupported):
		return ErrTrainingModeNotSupported
	case errors.Is(err, entity.ErrInvalidTrainingState):
		return ErrInvalidTrainingState
	case errors.Is(err, entity.ErrTrainingActionNotAvailable):
		return ErrTrainingActionNotAvailable
	case errors.Is(err, entity.ErrArenaAssignmentNotFound):
		return ErrArenaAssignmentNotFound
	case errors.Is(err, entity.ErrArenaSessionNotFound):
		return ErrArenaSessionNotFound
	case errors.Is(err, entity.ErrArenaAccessDenied):
		return ErrArenaAccessDenied
	case errors.Is(err, entity.ErrArenaJoinNotAllowed):
		return ErrArenaJoinNotAllowed
	case errors.Is(err, entity.ErrInvalidArenaState):
		return ErrInvalidArenaState
	case errors.Is(err, entity.ErrArenaParticipantInvalid):
		return ErrArenaParticipantInvalid
	case errors.Is(err, entity.ErrArenaScenarioNotSupported):
		return ErrArenaScenarioNotSupported
	case errors.Is(err, entity.ErrArenaFeedbackParticipant):
		return ErrArenaFeedbackParticipant
	case errors.Is(err, entity.ErrLLMUnavailable):
		return ErrLLMUnavailable
	case errors.Is(err, entity.ErrLLMResponseInvalid):
		return ErrLLMResponseInvalid
	case errors.Is(err, entity.ErrInvalidRequest):
		return ErrInvalidRequest
	default:
		return ErrInternal
	}
}

func StatusFromError(err error) int {
	switch ErrorCodeFromError(err) {
	case ErrInvalidRequest, ErrInvalidScenarioJSON:
		return http.StatusBadRequest
	case ErrAccessTokenMissing, ErrAccessTokenInvalid, ErrAccessTokenExpired:
		return http.StatusUnauthorized
	case ErrMethodistRoleRequired, ErrCuratorRoleRequired, ErrScenarioNotOwned, ErrTrainingAccessDenied, ErrArenaAccessDenied:
		return http.StatusForbidden
	case ErrUserNotFound, ErrScenarioNotFound, ErrScenarioVersionNotFound, ErrTrainingSessionNotFound, ErrArenaAssignmentNotFound, ErrArenaSessionNotFound:
		return http.StatusNotFound
	case ErrOrganizationRequired, ErrOrganizationAlreadySet, ErrInvalidTrainingState, ErrInvalidArenaState, ErrArenaJoinNotAllowed:
		return http.StatusConflict
	case ErrTrainingModeNotSupported, ErrTrainingActionNotAvailable, ErrArenaParticipantInvalid, ErrArenaScenarioNotSupported, ErrArenaFeedbackParticipant:
		return http.StatusUnprocessableEntity
	case ErrLLMUnavailable, ErrLLMResponseInvalid:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
