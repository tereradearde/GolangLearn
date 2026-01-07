package httpdelivery

import (
	"net/http"

	codedom "github.com/example/learngo/internal/domain/code"
	codeexecuc "github.com/example/learngo/internal/usecase/codeexec"
	"github.com/gin-gonic/gin"
)

type CodeHandler struct {
	svc codeexecuc.Service
}

func NewCodeHandler(s codeexecuc.Service) *CodeHandler {
	return &CodeHandler{svc: s}
}

// Execute обрабатывает POST /api/code/execute
func (h *CodeHandler) Execute(c *gin.Context) {
	var req codedom.ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Валидация языка
	validLanguages := map[string]bool{
		"python":     true,
		"javascript": true,
		"java":       true,
		"go":         true,
		"cpp":        true,
	}
	if !validLanguages[req.Language] {
		ValidationError(c, "Invalid language", map[string]interface{}{
			"language":            req.Language,
			"supported_languages": []string{"python", "javascript", "java", "go", "cpp"},
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.svc.Execute(ctx, req)
	if err != nil {
		InternalError(c, "Failed to execute code", err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
