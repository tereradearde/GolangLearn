package errors

import (
	"time"
)

// APIError стандартная структура ошибки API
type APIError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id"`
}

// NewAPIError создает новую ошибку API
func NewAPIError(code, message string, requestID string) *APIError {
	return &APIError{
		Code:      code,
		Message:   message,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	}
}

// WithDetail добавляет деталь к ошибке
func (e *APIError) WithDetail(key string, value interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails добавляет несколько деталей
func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// Predefined error codes
const (
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeRateLimit          = "RATE_LIMIT_EXCEEDED"
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Predefined error messages
const (
	MsgValidationError    = "Validation error"
	MsgUnauthorized       = "Unauthorized"
	MsgForbidden          = "Forbidden"
	MsgNotFound           = "Resource not found"
	MsgInternalError      = "Internal server error"
	MsgRateLimitExceeded  = "Rate limit exceeded"
	MsgBadRequest         = "Bad request"
	MsgServiceUnavailable = "Service unavailable"
)
