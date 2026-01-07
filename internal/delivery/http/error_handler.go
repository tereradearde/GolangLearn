package httpdelivery

import (
	"net/http"

	"github.com/example/learngo/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CtxRequestID = "request_id"

// RequestIDMiddleware добавляет request_id к каждому запросу
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(CtxRequestID, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// GetRequestID получает request_id из контекста
func GetRequestID(c *gin.Context) string {
	requestID, exists := c.Get(CtxRequestID)
	if !exists {
		return uuid.New().String()
	}
	return requestID.(string)
}

// ErrorResponse отправляет стандартизированный ответ об ошибке
func ErrorResponse(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	requestID := GetRequestID(c)
	err := errors.NewAPIError(code, message, requestID)
	if details != nil {
		err.WithDetails(details)
	}
	c.JSON(statusCode, err)
}

// ValidationError отправляет ошибку валидации
func ValidationError(c *gin.Context, message string, details map[string]interface{}) {
	ErrorResponse(c, http.StatusBadRequest, errors.ErrCodeValidation, message, details)
}

// UnauthorizedError отправляет ошибку авторизации
func UnauthorizedError(c *gin.Context, message string) {
	if message == "" {
		message = errors.MsgUnauthorized
	}
	ErrorResponse(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized, message, nil)
}

// ForbiddenError отправляет ошибку доступа
func ForbiddenError(c *gin.Context, message string) {
	if message == "" {
		message = errors.MsgForbidden
	}
	ErrorResponse(c, http.StatusForbidden, errors.ErrCodeForbidden, message, nil)
}

// NotFoundError отправляет ошибку "не найдено"
func NotFoundError(c *gin.Context, resource string) {
	message := errors.MsgNotFound
	if resource != "" {
		message = resource + " not found"
	}
	ErrorResponse(c, http.StatusNotFound, errors.ErrCodeNotFound, message, nil)
}

// InternalError отправляет ошибку внутреннего сервера
func InternalError(c *gin.Context, message string, err error) {
	if message == "" {
		message = errors.MsgInternalError
	}
	details := make(map[string]interface{})
	if err != nil {
		details["error"] = err.Error()
	}
	ErrorResponse(c, http.StatusInternalServerError, errors.ErrCodeInternal, message, details)
}

// RateLimitError отправляет ошибку rate limit
func RateLimitError(c *gin.Context, message string, details map[string]interface{}) {
	if message == "" {
		message = errors.MsgRateLimitExceeded
	}
	ErrorResponse(c, http.StatusTooManyRequests, errors.ErrCodeRateLimit, message, details)
}

// ServiceUnavailableError отправляет ошибку недоступности сервиса
func ServiceUnavailableError(c *gin.Context, message string) {
	if message == "" {
		message = errors.MsgServiceUnavailable
	}
	ErrorResponse(c, http.StatusServiceUnavailable, errors.ErrCodeServiceUnavailable, message, nil)
}

// BadRequestError отправляет ошибку некорректного запроса
func BadRequestError(c *gin.Context, message string, details map[string]interface{}) {
	if message == "" {
		message = errors.MsgBadRequest
	}
	ErrorResponse(c, http.StatusBadRequest, errors.ErrCodeBadRequest, message, details)
}

// ErrorHandlerMiddleware обрабатывает панику и возвращает стандартизированную ошибку
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Проверяем, есть ли ошибки
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID := GetRequestID(c)

			apiErr := errors.NewAPIError(
				errors.ErrCodeInternal,
				errors.MsgInternalError,
				requestID,
			).WithDetail("error", err.Error())

			c.JSON(http.StatusInternalServerError, apiErr)
			c.Abort()
		}
	}
}

// RecoveryMiddleware обрабатывает панику
func RecoveryMiddleware() gin.RecoveryFunc {
	return func(c *gin.Context, recovered interface{}) {
		requestID := GetRequestID(c)

		apiErr := errors.NewAPIError(
			errors.ErrCodeInternal,
			"Internal server error",
			requestID,
		).WithDetail("panic", recovered)

		c.JSON(http.StatusInternalServerError, apiErr)
		c.Abort()
	}
}
