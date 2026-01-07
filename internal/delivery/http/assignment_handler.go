package httpdelivery

import (
	"net/http"

	assignuc "github.com/example/learngo/internal/usecase/assignment"
	"github.com/example/learngo/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AssignmentHandler struct {
	svc    assignuc.Service
	logger *utils.Logger
}

func NewAssignmentHandler(s assignuc.Service, logger *utils.Logger) *AssignmentHandler {
	return &AssignmentHandler{svc: s, logger: logger}
}

func (h *AssignmentHandler) ListByLesson(c *gin.Context) {
	lid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lessonId"})
		return
	}
	list, err := h.svc.ListByLesson(c.Request.Context(), lid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *AssignmentHandler) Create(c *gin.Context) {
	lid, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lessonId"})
		return
	}
	var req struct {
		Title, Prompt, StarterCode, Tests string
		Order                             int
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a, err := h.svc.Create(c.Request.Context(), lid, req.Title, req.Prompt, req.StarterCode, req.Tests, req.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *AssignmentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	a, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if a.ID == uuid.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AssignmentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Title, Prompt, StarterCode, Tests string
		Order                             int
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, req.Title, req.Prompt, req.StarterCode, req.Tests, req.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if a.ID == uuid.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AssignmentHandler) Delete(c *gin.Context) {
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
