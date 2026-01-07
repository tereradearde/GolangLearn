package httpdelivery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	aidom "github.com/example/learngo/internal/domain/ai"
	aiuc "github.com/example/learngo/internal/usecase/ai"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AIHandler struct {
	svc    aiuc.Service
	logger interface {
		Error(msg string, args ...interface{})
	}
}

func NewAIHandler(s aiuc.Service, logger interface {
	Error(msg string, args ...interface{})
}) *AIHandler {
	return &AIHandler{svc: s, logger: logger}
}

// Chat обрабатывает POST /api/ai/chat с SSE streaming
func (h *AIHandler) Chat(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		UnauthorizedError(c, "")
		return
	}

	var req struct {
		Messages []aidom.ChatMessage `json:"messages" binding:"required"`
		Context  aidom.ChatContext   `json:"context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Настраиваем SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Отключаем буферизацию в nginx

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	messageID := uuid.New().String()

	// Отправляем начальное сообщение
	fmt.Fprintf(c.Writer, "data: %s\n\n", mustJSON(map[string]interface{}{
		"type": "start",
		"id":   messageID,
	}))
	c.Writer.Flush()

	// Обработчик для каждого чанка
	tokensUsed, err := h.svc.ChatStream(ctx, userID.String(), req.Messages, req.Context, func(content string) error {
		// Отправляем токен
		data := map[string]interface{}{
			"type":    "token",
			"content": content,
		}
		fmt.Fprintf(c.Writer, "data: %s\n\n", mustJSON(data))
		c.Writer.Flush()
		return nil
	})

	if err != nil {
		h.logger.Error("ai chat stream error", "error", err)
		fmt.Fprintf(c.Writer, "data: %s\n\n", mustJSON(map[string]interface{}{
			"type":    "error",
			"message": "Ошибка при генерации ответа",
		}))
		c.Writer.Flush()
		return
	}

	// Отправляем финальное сообщение
	fmt.Fprintf(c.Writer, "data: %s\n\n", mustJSON(map[string]interface{}{
		"type": "done",
		"id":   messageID,
		"usage": map[string]interface{}{
			"tokens": tokensUsed,
		},
	}))
	c.Writer.Flush()

	// Закрываем поток
	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
}

// CodeReview обрабатывает POST /api/ai/code-review
func (h *AIHandler) CodeReview(c *gin.Context) {
	_, ok := UserIDFromContext(c)
	if !ok {
		UnauthorizedError(c, "")
		return
	}

	var req aidom.CodeReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	resp, err := h.svc.CodeReview(ctx, req)
	if err != nil {
		h.logger.Error("ai code review error", "error", err)
		InternalError(c, "Failed to review code", err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ExplainError обрабатывает POST /api/ai/explain-error
func (h *AIHandler) ExplainError(c *gin.Context) {
	_, ok := UserIDFromContext(c)
	if !ok {
		UnauthorizedError(c, "")
		return
	}

	var req aidom.ExplainErrorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	resp, err := h.svc.ExplainError(ctx, req)
	if err != nil {
		h.logger.Error("ai explain error", "error", err)
		InternalError(c, "Failed to explain error", err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Hints обрабатывает POST /api/ai/hints
func (h *AIHandler) Hints(c *gin.Context) {
	_, ok := UserIDFromContext(c)
	if !ok {
		UnauthorizedError(c, "")
		return
	}

	var req aidom.HintsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	resp, err := h.svc.GetHints(ctx, req)
	if err != nil {
		h.logger.Error("ai hints error", "error", err)
		InternalError(c, "Failed to get hints", err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func mustJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
