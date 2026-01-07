package httpdelivery

import (
	"net/http"

	progressuc "github.com/example/learngo/internal/usecase/progress"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProgressHandler struct {
	svc progressuc.Service
}

func NewProgressHandler(s progressuc.Service) *ProgressHandler {
	return &ProgressHandler{svc: s}
}

// GetCourseProgress обрабатывает GET /api/users/:userId/progress/:courseId
func (h *ProgressHandler) GetCourseProgress(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	// Проверяем, что пользователь запрашивает свой прогресс
	authUserID, ok := UserIDFromContext(c)
	if !ok || authUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	progress, err := h.svc.GetCourseProgress(c.Request.Context(), userID, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// UpsertLessonProgress обрабатывает POST /api/lessons/:lessonId/progress
func (h *ProgressHandler) UpsertLessonProgress(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	var req struct {
		Code             string `json:"code"`
		Completed        bool   `json:"completed"`
		TimeSpentMinutes int    `json:"time_spent_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: получить courseID из урока
	// Пока используем пустой UUID, нужно будет добавить метод GetLesson в сервис
	courseID := uuid.Nil

	progress, err := h.svc.UpsertLessonProgress(c.Request.Context(), userID, courseID, lessonID, req.Code, req.Completed, req.TimeSpentMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"progress": gin.H{
			"lesson_id":      progress.LessonID.String(),
			"completed":      progress.Completed,
			"completed_at":   progress.CompletedAt,
			"next_lesson_id": nil, // TODO: вычислить следующий урок
		},
	})
}
