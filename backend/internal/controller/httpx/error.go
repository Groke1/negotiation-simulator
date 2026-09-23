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

	ErrUserNotFound         ErrorCode = "USER_NOT_FOUND"
	ErrRoleAlreadyAssigned  ErrorCode = "ROLE_ALREADY_ASSIGNED"
	ErrRoleNotAssigned      ErrorCode = "ROLE_NOT_ASSIGNED"
	ErrUserDataAccessDenied ErrorCode = "USER_DATA_ACCESS_DENIED"
	ErrMediaTooLarge        ErrorCode = "MEDIA_TOO_LARGE"
	ErrUnsupportedMediaType ErrorCode = "UNSUPPORTED_MEDIA_TYPE"
	ErrMediaAccessDenied    ErrorCode = "MEDIA_ACCESS_DENIED"
	ErrMediaNotFound        ErrorCode = "MEDIA_NOT_FOUND"

	ErrTrainingSessionNotFound           ErrorCode = "TRAINING_SESSION_NOT_FOUND"
	ErrTrainingAccessDenied              ErrorCode = "TRAINING_ACCESS_DENIED"
	ErrScenarioNotAvailable              ErrorCode = "SCENARIO_NOT_AVAILABLE"
	ErrTrainingConfigurationNotSupported ErrorCode = "TRAINING_CONFIGURATION_NOT_SUPPORTED"
	ErrInvalidTrainingState              ErrorCode = "INVALID_TRAINING_STATE"
	ErrActionNotAvailable                ErrorCode = "ACTION_NOT_AVAILABLE"
	ErrTrainingMessageChannelMismatch    ErrorCode = "TRAINING_MESSAGE_CHANNEL_MISMATCH"
	ErrTrainingMessageNotFound           ErrorCode = "TRAINING_MESSAGE_NOT_FOUND"
	ErrMessageAlreadyClassified          ErrorCode = "MESSAGE_ALREADY_CLASSIFIED"
	ErrMessageNotClassifiable            ErrorCode = "MESSAGE_NOT_CLASSIFIABLE"

	ErrCuratorRoleRequired    ErrorCode = "CURATOR_ROLE_REQUIRED"
	ErrAssignmentAccessDenied ErrorCode = "ASSIGNMENT_ACCESS_DENIED"
	ErrAssignmentConflict     ErrorCode = "ASSIGNMENT_CONFLICT"

	ErrArenaAssignmentNotFound         ErrorCode = "ARENA_ASSIGNMENT_NOT_FOUND"
	ErrArenaSessionNotFound            ErrorCode = "ARENA_SESSION_NOT_FOUND"
	ErrArenaAccessDenied               ErrorCode = "ARENA_ACCESS_DENIED"
	ErrArenaJoinNotAllowed             ErrorCode = "ARENA_JOIN_NOT_ALLOWED"
	ErrInvalidArenaState               ErrorCode = "INVALID_ARENA_STATE"
	ErrArenaFeedbackForbidden          ErrorCode = "ARENA_FEEDBACK_FORBIDDEN"
	ErrArenaFeedbackConflict           ErrorCode = "ARENA_FEEDBACK_CONFLICT"
	ErrArenaFeedbackParticipantInvalid ErrorCode = "ARENA_FEEDBACK_PARTICIPANT_INVALID"

	ErrMethodistRoleRequired         ErrorCode = "METHODIST_ROLE_REQUIRED"
	ErrOrganizationRequired          ErrorCode = "ORGANIZATION_REQUIRED"
	ErrScenarioNotFound              ErrorCode = "SCENARIO_NOT_FOUND"
	ErrScenarioNotOwned              ErrorCode = "SCENARIO_OWNERSHIP_REQUIRED"
	ErrScenarioNotDraft              ErrorCode = "SCENARIO_NOT_DRAFT"
	ErrInvalidScenarioStatus         ErrorCode = "INVALID_SCENARIO_STATUS"
	ErrInvalidScenarioJSON           ErrorCode = "INVALID_SCENARIO_JSON"
	ErrScenarioValidationFailed      ErrorCode = "SCENARIO_VALIDATION_FAILED"
	ErrBlockNotFound                 ErrorCode = "BLOCK_NOT_FOUND"
	ErrAIGenerationContextIncomplete ErrorCode = "AI_GENERATION_CONTEXT_INCOMPLETE"
	ErrAIGenerationFailed            ErrorCode = "AI_GENERATION_FAILED"
	ErrAIInvalidStructuredOutput     ErrorCode = "AI_INVALID_STRUCTURED_OUTPUT"
	ErrAIProviderUnavailable         ErrorCode = "AI_PROVIDER_UNAVAILABLE"
	ErrAIGenerationTimeout           ErrorCode = "AI_GENERATION_TIMEOUT"
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
	case ErrUserDataAccessDenied:
		return "Access to this user data is denied"
	case ErrMediaTooLarge:
		return "Media file is too large"
	case ErrUnsupportedMediaType:
		return "Media type is not supported"
	case ErrMediaAccessDenied:
		return "Access to this media is denied"
	case ErrMediaNotFound:
		return "Media not found"
	case ErrTrainingSessionNotFound:
		return "Training session not found"
	case ErrTrainingAccessDenied:
		return "Access to this training session is denied"
	case ErrScenarioNotAvailable:
		return "Scenario is not available"
	case ErrTrainingConfigurationNotSupported:
		return "Training configuration is not supported"
	case ErrInvalidTrainingState:
		return "Training state does not allow this operation"
	case ErrActionNotAvailable:
		return "Requested action is not available"
	case ErrTrainingMessageChannelMismatch:
		return "Message content does not match session channel"
	case ErrTrainingMessageNotFound:
		return "Training message not found"
	case ErrMessageAlreadyClassified:
		return "Message is already classified"
	case ErrMessageNotClassifiable:
		return "Message cannot be classified"
	case ErrCuratorRoleRequired:
		return "Active role must be curator"
	case ErrAssignmentAccessDenied:
		return "Access to this assignment is denied"
	case ErrAssignmentConflict:
		return "Assignment already exists"
	case ErrArenaAssignmentNotFound:
		return "Arena assignment not found"
	case ErrArenaSessionNotFound:
		return "Arena session not found"
	case ErrArenaAccessDenied:
		return "Access to this arena is denied"
	case ErrArenaJoinNotAllowed:
		return "Cannot join this arena"
	case ErrInvalidArenaState:
		return "Arena state does not allow this operation"
	case ErrArenaFeedbackForbidden:
		return "Feedback is forbidden"
	case ErrArenaFeedbackConflict:
		return "Feedback already exists"
	case ErrArenaFeedbackParticipantInvalid:
		return "Feedback participant is invalid"
	case ErrMethodistRoleRequired:
		return "Active role must be methodist"
	case ErrOrganizationRequired:
		return "User must belong to an organization"
	case ErrScenarioNotFound:
		return "Scenario not found"
	case ErrScenarioNotOwned:
		return "Only the scenario owner can modify this scenario"
	case ErrScenarioNotDraft:
		return "AI generation is available only for draft scenarios"
	case ErrInvalidScenarioStatus:
		return "Scenario status does not allow this operation"
	case ErrInvalidScenarioJSON:
		return "Scenario JSON is invalid"
	case ErrScenarioValidationFailed:
		return "Scenario validation failed"
	case ErrBlockNotFound:
		return "Block not found"
	case ErrAIGenerationContextIncomplete:
		return "Scenario does not contain enough data for this generation"
	case ErrAIGenerationFailed:
		return "AI generation failed"
	case ErrAIInvalidStructuredOutput:
		return "AI returned invalid structured output"
	case ErrAIProviderUnavailable:
		return "AI provider is unavailable"
	case ErrAIGenerationTimeout:
		return "AI generation timed out"
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
	case errors.Is(err, entity.ErrUserDataAccessDenied):
		return ErrUserDataAccessDenied
	case errors.Is(err, entity.ErrMediaTooLarge):
		return ErrMediaTooLarge
	case errors.Is(err, entity.ErrUnsupportedMediaType):
		return ErrUnsupportedMediaType
	case errors.Is(err, entity.ErrMediaAccessDenied):
		return ErrMediaAccessDenied
	case errors.Is(err, entity.ErrMediaNotFound):
		return ErrMediaNotFound
	case errors.Is(err, entity.ErrTrainingSessionNotFound):
		return ErrTrainingSessionNotFound
	case errors.Is(err, entity.ErrTrainingAccessDenied):
		return ErrTrainingAccessDenied
	case errors.Is(err, entity.ErrScenarioNotAvailable):
		return ErrScenarioNotAvailable
	case errors.Is(err, entity.ErrTrainingConfigurationNotSupported):
		return ErrTrainingConfigurationNotSupported
	case errors.Is(err, entity.ErrInvalidTrainingState):
		return ErrInvalidTrainingState
	case errors.Is(err, entity.ErrActionNotAvailable):
		return ErrActionNotAvailable
	case errors.Is(err, entity.ErrTrainingMessageChannelMismatch):
		return ErrTrainingMessageChannelMismatch
	case errors.Is(err, entity.ErrTrainingMessageNotFound):
		return ErrTrainingMessageNotFound
	case errors.Is(err, entity.ErrMessageAlreadyClassified):
		return ErrMessageAlreadyClassified
	case errors.Is(err, entity.ErrMessageNotClassifiable):
		return ErrMessageNotClassifiable
	case errors.Is(err, entity.ErrCuratorRoleRequired):
		return ErrCuratorRoleRequired
	case errors.Is(err, entity.ErrAssignmentAccessDenied):
		return ErrAssignmentAccessDenied
	case errors.Is(err, entity.ErrAssignmentConflict):
		return ErrAssignmentConflict
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
	case errors.Is(err, entity.ErrArenaFeedbackForbidden):
		return ErrArenaFeedbackForbidden
	case errors.Is(err, entity.ErrArenaFeedbackConflict):
		return ErrArenaFeedbackConflict
	case errors.Is(err, entity.ErrArenaFeedbackParticipantInvalid):
		return ErrArenaFeedbackParticipantInvalid
	case errors.Is(err, entity.ErrMethodistRoleRequired):
		return ErrMethodistRoleRequired
	case errors.Is(err, entity.ErrOrganizationRequired):
		return ErrOrganizationRequired
	case errors.Is(err, entity.ErrScenarioNotFound):
		return ErrScenarioNotFound
	case errors.Is(err, entity.ErrScenarioNotOwned):
		return ErrScenarioNotOwned
	case errors.Is(err, entity.ErrScenarioNotDraft):
		return ErrScenarioNotDraft
	case errors.Is(err, entity.ErrInvalidScenarioStatus):
		return ErrInvalidScenarioStatus
	case errors.Is(err, entity.ErrInvalidScenarioJSON):
		return ErrInvalidScenarioJSON
	case errors.Is(err, entity.ErrScenarioValidationFailed):
		return ErrScenarioValidationFailed
	case errors.Is(err, entity.ErrBlockNotFound):
		return ErrBlockNotFound
	case errors.Is(err, entity.ErrAIGenerationContextIncomplete):
		return ErrAIGenerationContextIncomplete
	case errors.Is(err, entity.ErrAIGenerationFailed):
		return ErrAIGenerationFailed
	case errors.Is(err, entity.ErrAIInvalidStructuredOutput):
		return ErrAIInvalidStructuredOutput
	case errors.Is(err, entity.ErrAIProviderUnavailable):
		return ErrAIProviderUnavailable
	case errors.Is(err, entity.ErrAIGenerationTimeout):
		return ErrAIGenerationTimeout
	case errors.Is(err, entity.ErrInvalidRequest):
		return ErrInvalidRequest
	default:
		return ErrInternal
	}
}
