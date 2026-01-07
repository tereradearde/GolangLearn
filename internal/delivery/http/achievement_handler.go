package httpdelivery

import (
	"net/http"

	achievementuc "github.com/example/learngo/internal/usecase/achievement"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AchievementHandler struct {
	svc achievementuc.Service
}

func NewAchievementHandler(s achievementuc.Service) *AchievementHandler {
	return &AchievementHandler{svc: s}
}

// GetUserAchievements обрабатывает GET /api/users/:userId/achievements
func (h *AchievementHandler) GetUserAchievements(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Проверяем, что пользователь запрашивает свои достижения
	authUserID, ok := UserIDFromContext(c)
	if !ok || authUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	achievements, err := h.svc.GetUserAchievements(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем в формат API
	result := make([]gin.H, 0, len(achievements))
	for _, ach := range achievements {
		result = append(result, gin.H{
			"id":          ach.AchievementID,
			"unlocked":    true,
			"unlocked_at": ach.UnlockedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"achievements": result})
}
