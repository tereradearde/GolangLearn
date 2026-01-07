package httpdelivery

import (
	"encoding/json"
	"net/http"

	lessondom "github.com/example/learngo/internal/domain/lesson"
	lessonuc "github.com/example/learngo/internal/usecase/lesson"
	"github.com/example/learngo/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LessonHandler struct {
	svc    lessonuc.Service
	logger *utils.Logger
}

func NewLessonHandler(s lessonuc.Service, logger *utils.Logger) *LessonHandler {
	return &LessonHandler{svc: s, logger: logger}
}

func (h *LessonHandler) ListByCourse(c *gin.Context) {
	cid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid courseId"})
		return
	}
	list, err := h.svc.ListByCourse(c.Request.Context(), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *LessonHandler) ListBySection(c *gin.Context) {
	sid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sectionId"})
		return
	}
	list, err := h.svc.ListBySection(c.Request.Context(), sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *LessonHandler) Create(c *gin.Context) {
	cid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid courseId"})
		return
	}
	var req struct {
		Title, Content string
		Order          int
		SectionID      string `json:"sectionId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.SectionID != "" {
		if sid, e := uuid.Parse(req.SectionID); e == nil {
			l, err := h.svc.CreateInSection(c.Request.Context(), cid, sid, req.Title, req.Content, req.Order)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
			c.JSON(http.StatusCreated, l)
			return
		}
	}
	l, err := h.svc.Create(c.Request.Context(), cid, req.Title, req.Content, req.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, l)
}

func (h *LessonHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	l, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if l.ID == uuid.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Парсим Content из JSON
	var content lessondom.LessonContent
	if len(l.Content) > 0 {
		if err := json.Unmarshal(l.Content, &content); err != nil {
			h.logger.Error("failed to parse lesson content", "error", err)
			// Используем пустую структуру если не удалось распарсить
			content = lessondom.LessonContent{}
		}
	}

	// Формируем ответ согласно документации
	response := gin.H{
		"id":               l.ID.String(),
		"course_id":        l.CourseID.String(),
		"module_id":        l.ModuleID.String(),
		"title":            l.Title,
		"slug":             l.Slug,
		"content":          content,
		"duration_minutes": l.DurationMinutes,
		"order":            l.Order,
	}

	if l.NextLessonID != nil {
		response["next_lesson_id"] = l.NextLessonID.String()
	}
	if l.PreviousLessonID != nil {
		response["previous_lesson_id"] = l.PreviousLessonID.String()
	}

	c.JSON(http.StatusOK, response)
}

func (h *LessonHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Title, Content string
		Order          int
		SectionID      string `json:"sectionId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var sid uuid.UUID
	if req.SectionID != "" {
		if parsed, e := uuid.Parse(req.SectionID); e == nil {
			sid = parsed
		}
	}
	l, err := h.svc.Update(c.Request.Context(), id, req.Title, req.Content, req.Order, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if l.ID == uuid.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *LessonHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}
