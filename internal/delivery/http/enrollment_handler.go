package httpdelivery

import (
	"net/http"

	"github.com/example/learngo/internal/usecase/enrollment"
	"github.com/example/learngo/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EnrollmentHandler struct {
	svc    enrollment.Service
	logger *utils.Logger
}

func NewEnrollmentHandler(svc enrollment.Service, logger *utils.Logger) *EnrollmentHandler {
	return &EnrollmentHandler{svc: svc, logger: logger}
}

type enrollRequest struct {
	CourseID  string `json:"courseId" binding:"required"`
	Purchased bool   `json:"purchased"`
}

func (h *EnrollmentHandler) Enroll(c *gin.Context) {
	var req enrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid courseId"})
		return
	}
	if err := h.svc.Enroll(c.Request.Context(), uid, courseID, req.Purchased); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "enroll failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}
