package httpdelivery

import (
	"net/http"
	"strconv"

	coursedom "github.com/example/learngo/internal/domain/course"
	courseuc "github.com/example/learngo/internal/usecase/course"
	enrolluc "github.com/example/learngo/internal/usecase/enrollment"
	lessonuc "github.com/example/learngo/internal/usecase/lesson"
	moduleuc "github.com/example/learngo/internal/usecase/module"
	"github.com/example/learngo/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CourseHandler HTTP-обработчики для курса.
type CourseHandler struct {
	service       courseuc.Service
	enrollmentSvc enrolluc.Service
	lessonSvc     lessonuc.Service
	moduleSvc     moduleuc.Service
	logger        *utils.Logger
}

func NewCourseHandler(service courseuc.Service, logger *utils.Logger) *CourseHandler {
	// временно без enrollmentSvc; будет проставлен в router при создании, если доступен
	return &CourseHandler{service: service, logger: logger}
}

func (h *CourseHandler) List(c *gin.Context) {
	// Параметры согласно документации: language, difficulty, page, limit
	language := c.Query("language")
	difficulty := c.Query("difficulty")
	page := parseIntDefault(c.Query("page"), 1)
	limit := parseIntDefault(c.Query("limit"), 20)
	if limit > 100 {
		limit = 100 // max 100 согласно документации
	}

	// Используем SearchCourses для фильтрации и пагинации
	filter := coursedom.ListFilter{
		Language:   language,
		Difficulty: difficulty,
		Page:       page,
		PageSize:   limit,
	}

	res, err := h.service.SearchCourses(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("search courses failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Формируем ответ согласно документации
	totalPages := int(res.Total) / limit
	if int(res.Total)%limit > 0 {
		totalPages++
	}

	// Преобразуем курсы в формат API
	courses := make([]gin.H, 0, len(res.Items))
	for _, course := range res.Items {
		courses = append(courses, h.courseToAPIResponse(course))
	}

	c.JSON(http.StatusOK, gin.H{
		"courses": courses,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       res.Total,
			"total_pages": totalPages,
		},
	})
}

// courseToAPIResponse преобразует доменную модель курса в формат API
func (h *CourseHandler) courseToAPIResponse(course coursedom.Course) gin.H {
	// Вычисляем price из PriceCents если Price не задан
	price := course.Price
	if price == nil && course.PriceCents > 0 {
		p := float64(course.PriceCents) / 100.0
		price = &p
	}

	return gin.H{
		"id":             course.ID.String(),
		"slug":           course.Slug,
		"title":          course.Title,
		"description":    course.Description,
		"language":       course.Language,
		"difficulty":     course.Difficulty,
		"duration_hours": course.DurationHours,
		"lessons_count":  course.LessonsCount,
		"students_count": course.StudentsCount,
		"rating":         course.Rating,
		"thumbnail_url":  course.ThumbnailURL,
		"is_free":        course.IsFree,
		"price":          price,
	}
}

type createCourseRequest struct {
	Title         string   `json:"title" binding:"required,min=3"`
	Description   string   `json:"description" binding:"required,min=3"`
	Difficulty    string   `json:"difficulty" binding:"omitempty,oneof=beginner intermediate advanced"`
	DurationHours int      `json:"duration_hours"`
	Tags          []string `json:"tags"`
	ThumbnailURL  string   `json:"thumbnail_url"`
	ImageURL      string   `json:"imageUrl"`
	Objectives    []string `json:"objectives"`
	Requirements  []string `json:"requirements"`
	IsFree        bool     `json:"is_free"`
	Price         *float64 `json:"price"`
}

func (h *CourseHandler) Create(c *gin.Context) {
	var req createCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// собираем доп. поля через опции
	course, err := h.service.CreateCourse(
		c.Request.Context(), req.Title, req.Description,
		func(ca *coursedom.Course) {
			ca.Difficulty = req.Difficulty
			ca.DurationHours = req.DurationHours
			ca.Tags = req.Tags
			ca.ThumbnailURL = req.ThumbnailURL
			if ca.ThumbnailURL == "" {
				ca.ThumbnailURL = req.ImageURL
			}
			ca.ImageURL = req.ImageURL
			ca.Objectives = req.Objectives
			ca.Requirements = req.Requirements
			ca.IsFree = req.IsFree
			ca.Price = req.Price
		},
	)
	if err != nil {
		h.logger.Error("create course failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, course)
}

func (h *CourseHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	crs, err := h.service.GetCourse(c.Request.Context(), id)
	if err != nil {
		if err == courseuc.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		h.logger.Error("get course failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Получаем модули и уроки для детального ответа
	ctx := c.Request.Context()
	modules := []gin.H{}
	if h.moduleSvc != nil {
		mods, _ := h.moduleSvc.ListByCourse(ctx, id)
		lessons, _ := h.lessonSvc.ListByCourse(ctx, id)

		// Группируем уроки по модулям
		lessonsByModule := make(map[uuid.UUID][]gin.H)
		for _, lesson := range lessons {
			moduleID := lesson.SectionID
			lessonsByModule[moduleID] = append(lessonsByModule[moduleID], gin.H{
				"id":               lesson.ID.String(),
				"title":            lesson.Title,
				"slug":             lesson.Slug,
				"duration_minutes": 0, // TODO: добавить в модель урока
				"order":            lesson.Order,
				"is_free":          false, // TODO: добавить в модель урока
			})
		}

		// Формируем модули с уроками
		for _, mod := range mods {
			moduleLessons := lessonsByModule[mod.ID]
			if moduleLessons == nil {
				moduleLessons = []gin.H{}
			}
			modules = append(modules, gin.H{
				"id":      mod.ID.String(),
				"title":   mod.Title,
				"order":   mod.OrderIndex,
				"lessons": moduleLessons,
			})
		}
	}

	// Формируем ответ согласно документации
	response := gin.H{
		"id":                crs.ID.String(),
		"slug":              crs.Slug,
		"title":             crs.Title,
		"description":       crs.Description,
		"language":          crs.Language,
		"difficulty":        crs.Difficulty,
		"duration_hours":    crs.DurationHours,
		"lessons_count":     crs.LessonsCount,
		"students_count":    crs.StudentsCount,
		"rating":            crs.Rating,
		"requirements":      crs.Requirements,
		"learning_outcomes": crs.Objectives,
		"modules":           modules,
	}

	c.JSON(http.StatusOK, response)
}

// GetBySlug возвращает курс по слагу
func (h *CourseHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug"})
		return
	}
	crs, err := h.service.GetCourseBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == courseuc.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		h.logger.Error("get course by slug failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if uid, ok := UserIDFromContext(c); ok && h.enrollmentSvc != nil {
		// найдём ID курса
		id := crs.ID
		status := "none"
		if list, err := h.enrollmentSvc.ListByUser(c.Request.Context(), uid); err == nil {
			for _, e := range list {
				if e.CourseID == id {
					if e.Status != "" {
						status = e.Status
					} else {
						status = "enrolled"
					}
					break
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"course": crs, "enrollmentStatus": status})
		return
	}
	c.JSON(http.StatusOK, crs)
}

type updateCourseRequest struct {
	Title         string   `json:"title" binding:"required,min=3"`
	Description   string   `json:"description" binding:"required,min=3"`
	Difficulty    string   `json:"difficulty" binding:"omitempty,oneof=beginner intermediate advanced"`
	DurationHours int      `json:"duration_hours"`
	Tags          []string `json:"tags"`
	ThumbnailURL  string   `json:"thumbnail_url"`
	ImageURL      string   `json:"imageUrl"`
	Objectives    []string `json:"objectives"`
	Requirements  []string `json:"requirements"`
	IsFree        bool     `json:"is_free"`
	Price         *float64 `json:"price"`
}

func (h *CourseHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.UpdateCourse(c.Request.Context(), id, coursedom.Course{
		Title:         req.Title,
		Description:   req.Description,
		Difficulty:    req.Difficulty,
		DurationHours: req.DurationHours,
		Tags:          req.Tags,
		ThumbnailURL:  req.ThumbnailURL,
		ImageURL:      req.ImageURL,
		Objectives:    req.Objectives,
		Requirements:  req.Requirements,
		IsFree:        req.IsFree,
		Price:         req.Price,
	})
	if err != nil {
		if err == courseuc.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		h.logger.Error("update course failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *CourseHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteCourse(c.Request.Context(), id); err != nil {
		h.logger.Error("delete course failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
