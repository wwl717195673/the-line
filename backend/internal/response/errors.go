package response

import "net/http"

const (
	CodeValidationError = "VALIDATION_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeForbidden       = "FORBIDDEN"
	CodeConflict        = "CONFLICT"
	CodeInvalidState    = "INVALID_STATE"
	CodeInternalError   = "INTERNAL_ERROR"
)

type AppError struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details"`
	HTTPStatus int            `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(httpStatus int, code string, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    map[string]any{},
		HTTPStatus: httpStatus,
	}
}

func Validation(message string) *AppError {
	return NewAppError(http.StatusBadRequest, CodeValidationError, message)
}

func NotFound(message string) *AppError {
	return NewAppError(http.StatusNotFound, CodeNotFound, message)
}

func Forbidden(message string) *AppError {
	return NewAppError(http.StatusForbidden, CodeForbidden, message)
}

func Conflict(message string) *AppError {
	return NewAppError(http.StatusConflict, CodeConflict, message)
}

func InvalidState(message string) *AppError {
	return NewAppError(http.StatusConflict, CodeInvalidState, message)
}

func Internal(message string) *AppError {
	return NewAppError(http.StatusInternalServerError, CodeInternalError, message)
}
