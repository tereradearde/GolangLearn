package httpdelivery

import (
	"net/http"
	"os"

	achievementuc "github.com/example/learngo/internal/usecase/achievement"
	aiuc "github.com/example/learngo/internal/usecase/ai"
	assignuc "github.com/example/learngo/internal/usecase/assignment"
	authuc "github.com/example/learngo/internal/usecase/auth"
	codeexecuc "github.com/example/learngo/internal/usecase/codeexec"
	"github.com/example/learngo/internal/usecase/course"
	dashboarduc "github.com/example/learngo/internal/usecase/dashboard"
	enrolluc "github.com/example/learngo/internal/usecase/enrollment"
	lessonuc "github.com/example/learngo/internal/usecase/lesson"
	moduleuc "github.com/example/learngo/internal/usecase/module"
	progressuc "github.com/example/learngo/internal/usecase/progress"
	sectionuc "github.com/example/learngo/internal/usecase/section"
	"github.com/example/learngo/pkg/observability"
	"github.com/example/learngo/pkg/storage"
	"github.com/example/learngo/pkg/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Router оборачивает Gin и настраивает маршруты HTTP-API.
type Router struct{ engine *gin.Engine }

// NewRouter конструирует HTTP-роутер и регистрирует обработчики.
func NewRouter(logger *utils.Logger, courseService course.Service, authService authuc.Service, jwt *utils.JWTManager, cfg *utils.Config, lessonService lessonuc.Service, assignmentService assignuc.Service, progressService progressuc.Service, enrollmentService enrolluc.Service, sectionService sectionuc.Service, moduleService moduleuc.Service, achievementService achievementuc.Service, dashboardService dashboarduc.Service, aiService aiuc.Service, codeExecService codeexecuc.Service) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.Use(gin.CustomRecovery(RecoveryMiddleware()))
	r.Use(ErrorHandlerMiddleware())
	observability.InitMetrics()
	r.Use(observability.MetricsMiddleware())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	// Prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// OpenAPI static docs (файл может быть рядом с бинарём в Docker)
	r.GET("/api/docs/openapi.yaml", func(c *gin.Context) {
		if _, err := os.Stat("api/openapi.yaml"); err == nil {
			c.File("api/openapi.yaml")
			return
		}
		if _, err := os.Stat("openapi.yaml"); err == nil {
			c.File("openapi.yaml")
			return
		}
		c.AbortWithStatus(http.StatusNotFound)
	})
	r.GET("/api/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!doctype html><html><head><title>API Docs</title></head><body>
<redoc spec-url="/api/docs/openapi.yaml"></redoc>
<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body></html>`)
	})

	// Swagger UI (альтернатива ReDoc)
	r.GET("/api/docs/swagger", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!doctype html><html><head>
<title>Swagger UI</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
</head><body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
<script>
  window.ui = SwaggerUIBundle({ url: '/api/docs/openapi.yaml', dom_id: '#swagger-ui' });
  </script>
</body></html>`)
	})

	h := NewCourseHandler(courseService, logger)
	// внедряем сервисы для статуса и деталей
	h.enrollmentSvc = enrollmentService
	h.lessonSvc = lessonService
	h.moduleSvc = moduleService
	authHandler := NewAuthHandler(authService, logger)
	lh := NewLessonHandler(lessonService, logger)
	var sh *SectionHandler
	if sectionService != nil {
		sh = NewSectionHandler(sectionService)
	}
	var mh *ModuleHandler
	if moduleService != nil {
		mh = NewModuleHandler(moduleService)
	}
	// enrollments
	eh := NewEnrollmentHandler(enrollmentService, logger)
	ah := NewAssignmentHandler(assignmentService, logger)
	ph := NewProgressHandler(progressService)
	var achHandler *AchievementHandler
	if achievementService != nil {
		achHandler = NewAchievementHandler(achievementService)
	}
	var dashboardHandler *DashboardHandler
	if dashboardService != nil {
		dashboardHandler = NewDashboardHandler(dashboardService)
	}
	var aiHandler *AIHandler
	if aiService != nil {
		aiHandler = NewAIHandler(aiService, logger)
	}
	var codeHandler *CodeHandler
	if codeExecService != nil {
		codeHandler = NewCodeHandler(codeExecService)
	}

	api := r.Group("/api")
	{
		// Глобальный rate limit
		api.Use(globalRateLimiter(cfg))

		// Подключаем rate limits для auth
		attachRateLimits(api, cfg)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)
		courses := api.Group("/courses")
		{
			courses.GET("", h.List)
			courses.POST("", AuthRequired(jwt), RequireRoles("admin", "teacher"), h.Create)
			courses.GET(":id", h.Get)
			// SEO-friendly: курс по слагу
			courses.GET("slug/:slug", h.GetBySlug)
			courses.PUT(":id", AuthRequired(jwt), RequireRoles("admin", "teacher"), h.Update)
			courses.DELETE(":id", AuthRequired(jwt), RequireRoles("admin", "teacher"), h.Delete)
			// nested sections & lessons
			if sh != nil {
				courses.GET(":id/sections", sh.ListByCourse)
				courses.POST(":id/sections", AuthRequired(jwt), RequireRoles("admin", "teacher"), sh.Create)
			}
			if mh != nil {
				courses.GET(":id/modules", mh.ListByCourse)
				courses.POST(":id/modules", AuthRequired(jwt), RequireRoles("admin", "teacher"), mh.Create)
			}
			courses.GET(":id/lessons", lh.ListByCourse)
			courses.POST(":id/lessons", AuthRequired(jwt), RequireRoles("admin", "teacher"), lh.Create)
			// lessons by section
			api.GET("/sections/:id/lessons", lh.ListBySection)
			api.POST("/sections/:id/lessons", AuthRequired(jwt), RequireRoles("admin", "teacher"), lh.Create)
			if mh != nil {
				api.GET("/modules/:id/lessons", lh.ListBySection) // временно используем тот же метод (по id)
			}
		}
		// Новые эндпоинты прогресса согласно документации
		api.GET("/users/:userId/progress/:courseId", AuthRequired(jwt), ph.GetCourseProgress)
		api.POST("/lessons/:lessonId/progress", AuthRequired(jwt), ph.UpsertLessonProgress)

		// achievements
		if achHandler != nil {
			api.GET("/users/:userId/achievements", AuthRequired(jwt), achHandler.GetUserAchievements)
		}

		// dashboard
		if dashboardHandler != nil {
			api.GET("/users/:userId/dashboard", AuthRequired(jwt), dashboardHandler.GetDashboard)
		}

		// AI endpoints с отдельным rate limit
		if aiHandler != nil {
			aiGroup := api.Group("/ai")
			aiGroup.Use(aiRateLimiter(cfg))
			{
				aiGroup.POST("/chat", AuthRequired(jwt), aiHandler.Chat)
				aiGroup.POST("/code-review", AuthRequired(jwt), aiHandler.CodeReview)
				aiGroup.POST("/explain-error", AuthRequired(jwt), aiHandler.ExplainError)
				aiGroup.POST("/hints", AuthRequired(jwt), aiHandler.Hints)
			}
		}

		// Code execution с отдельным rate limit
		if codeHandler != nil {
			api.POST("/code/execute", AuthRequired(jwt), codeExecRateLimiter(cfg), codeHandler.Execute)
		}

		// enrollments
		api.POST("/enrollments", AuthRequired(jwt), RequireRoles("user", "admin", "teacher"), eh.Enroll)
		// lesson and assignments
		api.GET("/lessons/:id", lh.Get)
		api.PUT("/lessons/:id", AuthRequired(jwt), RequireRoles("admin", "teacher"), lh.Update)
		api.DELETE("/lessons/:id", AuthRequired(jwt), RequireRoles("admin", "teacher"), lh.Delete)
		if sh != nil {
			api.PUT("/section/:id", AuthRequired(jwt), RequireRoles("admin", "teacher"), sh.Update)
			api.DELETE("/section/:id", AuthRequired(jwt), RequireRoles("admin", "teacher"), sh.Delete)
		}
		if mh != nil {
			api.PUT("/module/:id", AuthRequired(jwt), RequireRoles("admin", "teacher"), mh.Update)
			api.DELETE("/module/:id", AuthRequired(jwt), RequireRoles("admin", "teacher"), mh.Delete)
		}
		api.GET("/lessons/:id/assignments", ah.ListByLesson)
		api.POST("/lessons/:id/assignments", AuthRequired(jwt), RequireRoles("admin", "teacher"), ah.Create)
		api.PUT("/assignments/:id", AuthRequired(jwt), RequireRoles("admin", "teacher"), ah.Update)
		api.DELETE("/assignments/:id", AuthRequired(jwt), RequireRoles("admin", "teacher"), ah.Delete)

		// S3 presign upload (для админки и загрузок обложек)
		api.POST("/uploads/presign", AuthRequired(jwt), func(c *gin.Context) {
			type req struct {
				Prefix      string `json:"prefix"`
				Ext         string `json:"ext"`
				ContentType string `json:"contentType"`
			}
			var rbody req
			_ = c.ShouldBindJSON(&rbody)
			s3, err := storage.NewS3Client(storage.S3Config{Endpoint: cfg.S3Endpoint, AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey, Bucket: cfg.S3Bucket, UseSSL: true, BaseURL: cfg.S3BaseURL})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "s3 init failed"})
				return
			}
			if err := s3.EnsureBucket(c); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "bucket"})
				return
			}
			key := s3.GenerateKey(rbody.Prefix, rbody.Ext)
			url, err := s3.PresignPut(c, key, rbody.ContentType, 15*60*1e9)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "presign"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"uploadUrl": url, "objectKey": key, "publicUrl": s3.ObjectURL(key)})
		})
		api.GET("/me", AuthRequired(jwt), func(c *gin.Context) {
			uid, _ := UserIDFromContext(c)
			role := c.GetString(CtxRole)
			c.JSON(http.StatusOK, gin.H{"userId": uid, "role": role})
		})
	}

	return &Router{engine: r}
}

// Run запускает HTTP-сервер.
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
