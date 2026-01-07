package httpdelivery

import (
	"context"
	"net/http"
	"time"

	dashboarduc "github.com/example/learngo/internal/usecase/dashboard"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DashboardHandler struct {
	svc dashboarduc.Service
}

func NewDashboardHandler(s dashboarduc.Service) *DashboardHandler {
	return &DashboardHandler{svc: s}
}

// GetDashboard обрабатывает GET /api/users/:userId/dashboard
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Проверяем, что пользователь запрашивает свой дашборд
	authUserID, ok := UserIDFromContext(c)
	if !ok || authUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Устанавливаем таймаут для запроса (отказоустойчивость)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dashboard, err := h.svc.GetDashboard(ctx, userID)
	if err != nil {
		if err == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "request timeout"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
