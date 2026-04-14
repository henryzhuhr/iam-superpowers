package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) StatusCode() int {
	switch e.Code {
	case ErrValidation, ErrBadRequest:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	case ErrTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

const (
	ErrValidation      = "VALIDATION_ERROR"
	ErrBadRequest      = "BAD_REQUEST"
	ErrUnauthorized    = "UNAUTHORIZED"
	ErrForbidden       = "FORBIDDEN"
	ErrNotFound        = "NOT_FOUND"
	ErrConflict        = "CONFLICT"
	ErrTooManyRequests = "TOO_MANY_REQUESTS"
	ErrInternal        = "INTERNAL_ERROR"
)

func NewValidationError(msg string) *APIError {
	return &APIError{Code: ErrValidation, Message: msg}
}

func NewUnauthorizedError(msg string) *APIError {
	return &APIError{Code: ErrUnauthorized, Message: msg}
}

func NewForbiddenError(msg string) *APIError {
	return &APIError{Code: ErrForbidden, Message: msg}
}

func NewNotFoundError(msg string) *APIError {
	return &APIError{Code: ErrNotFound, Message: msg}
}

func NewConflictError(msg string) *APIError {
	return &APIError{Code: ErrConflict, Message: msg}
}

func NewInternalError(msg string) *APIError {
	return &APIError{Code: ErrInternal, Message: msg}
}

func Respond(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{"data": data})
}

func RespondError(c *gin.Context, err *APIError) {
	c.JSON(err.StatusCode(), gin.H{
		"code":    err.Code,
		"message": err.Message,
		"details": err.Details,
	})
}
