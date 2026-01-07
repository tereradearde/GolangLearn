package main

import (
	"context"
	"log"
	"time"

	httpdelivery "github.com/example/learngo/internal/delivery/http"
	achievementdomain "github.com/example/learngo/internal/domain/achievement"
	assignmentdomain "github.com/example/learngo/internal/domain/assignment"
	coursedomain "github.com/example/learngo/internal/domain/course"
	enrollmentdomain "github.com/example/learngo/internal/domain/enrollment"
	lessondomain "github.com/example/learngo/internal/domain/lesson"
	moduledomain "github.com/example/learngo/internal/domain/module"
	progressdomain "github.com/example/learngo/internal/domain/progress"
	sectiondomain "github.com/example/learngo/internal/domain/section"
	userdomain "github.com/example/learngo/internal/domain/user"
	"github.com/example/learngo/internal/infrastructure/db"
	memoryrepo "github.com/example/learngo/internal/infrastructure/repository/memory"
	postgresrepo "github.com/example/learngo/internal/infrastructure/repository/postgres"
	achievementuc "github.com/example/learngo/internal/usecase/achievement"
	aiuc "github.com/example/learngo/internal/usecase/ai"
	assignuc "github.com/example/learngo/internal/usecase/assignment"
	authuc "github.com/example/learngo/internal/usecase/auth"
	codeexecuc "github.com/example/learngo/internal/usecase/codeexec"
	courseuc "github.com/example/learngo/internal/usecase/course"
	dashboarduc "github.com/example/learngo/internal/usecase/dashboard"
	"github.com/example/learngo/internal/usecase/enrollment"
	lessonuc "github.com/example/learngo/internal/usecase/lesson"
	moduleuc "github.com/example/learngo/internal/usecase/module"
	progressuc "github.com/example/learngo/internal/usecase/progress"
	sectionsvc "github.com/example/learngo/internal/usecase/section"
	"github.com/example/learngo/pkg/ai"
	"github.com/example/learngo/pkg/codeexec"
	"github.com/example/learngo/pkg/utils"
)

func main() {
	cfg, err := utils.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := utils.NewLogger(cfg.Env)

	// Инициализация зависимостей (infrastructure -> usecase -> delivery)
	var (
		courseRepo      coursedomain.Repository
		lessonRepo      lessondomain.Repository
		sectionRepo     sectiondomain.Repository
		moduleRepo      moduledomain.Repository
		assignmentRepo  assignmentdomain.Repository
		userRepo        userdomain.Repository
		progressRepo    progressdomain.Repository
		enrollmentRepo  enrollmentdomain.Repository
		achievementRepo achievementdomain.Repository
	)

	var pdbOpened bool
	if cfg.DBDsn != "" {
		if pdb, err := db.OpenPostgres(cfg.DBDsn); err == nil {
			pdbOpened = true
			cr := postgresrepo.NewCourseRepository(pdb)
			_ = cr.AutoMigrate()
			courseRepo = cr
			lr := postgresrepo.NewLessonRepository(pdb)
			_ = lr.AutoMigrate()
			lessonRepo = lr
			sr := postgresrepo.NewSectionRepository(pdb)
			_ = sr.AutoMigrate()
			sectionRepo = sr
			mr := postgresrepo.NewModuleRepository(pdb)
			_ = mr.AutoMigrate()
			moduleRepo = mr
			ar := postgresrepo.NewAssignmentRepository(pdb)
			_ = ar.AutoMigrate()
			assignmentRepo = ar
			ur := postgresrepo.NewUserRepository(pdb)
			_ = ur.AutoMigrate()
			userRepo = ur
			pr := postgresrepo.NewProgressRepository(pdb)
			_ = pr.AutoMigrate()
			progressRepo = pr
			er := postgresrepo.NewEnrollmentRepository(pdb)
			_ = er.AutoMigrate()
			enrollmentRepo = er
			achr := postgresrepo.NewAchievementRepository(pdb)
			_ = achr.AutoMigrate()
			achievementRepo = achr

			// AI Chat History repository
			aiChatRepo := postgresrepo.NewAIChatHistoryRepository(pdb)
			_ = aiChatRepo.AutoMigrate()
		} else {
			logger.Error("postgres connect failed, fallback to memory", "error", err)
		}
	}
	if !pdbOpened {
		courseRepo = memoryrepo.NewInMemoryCourseRepository()
		lessonRepo = memoryrepo.NewInMemoryLessonRepository()
		assignmentRepo = memoryrepo.NewInMemoryAssignmentRepository()
		userRepo = memoryrepo.NewInMemoryUserRepository()
		progressRepo = memoryrepo.NewInMemoryProgressRepository()
		enrollmentRepo = memoryrepo.NewInMemoryEnrollmentRepository()
	}

	// Use cases
	courseService := courseuc.NewService(courseRepo, logger)
	lessonService := lessonuc.NewService(lessonRepo, logger)
	assignmentService := assignuc.NewService(assignmentRepo, logger)
	jwtManager := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTTTLMin, cfg.JWTRefreshSecret, cfg.JWTRefreshTTLDays)
	authService := authuc.NewService(userRepo, jwtManager)
	var progressService progressuc.Service
	if progressRepo != nil {
		progressService = progressuc.NewService(progressRepo)
	}

	// Dev-seed
	seedDev(
		context.Background(),
		&ConfigLike{SeedDemo: cfg.SeedDemo, AdminEmail: cfg.AdminEmail, AdminPassword: cfg.AdminPassword},
		&Repositories{Course: courseRepo, Lesson: lessonRepo, Assignment: assignmentRepo, User: userRepo},
		authService,
		nil,
	)

	// Передаём enrollmentRepo через контекст Router'у через добавление параметра — упростим: внедрим через package-level? Лучше расширить сигнатуру
	// Enrollment use case
	enrollService := enrollment.NewService(enrollmentRepo, logger)
	// Section/Module use case (может быть nil)
	var sectionService sectionsvc.Service
	if sectionRepo != nil {
		sectionService = sectionsvc.NewService(sectionRepo, logger)
	}
	var moduleService moduleuc.Service
	if moduleRepo != nil {
		moduleService = moduleuc.NewService(moduleRepo, logger)
	}
	var achievementService achievementuc.Service
	if achievementRepo != nil {
		achievementService = achievementuc.NewService(achievementRepo)
	}

	// Dashboard service
	var dashboardService dashboarduc.Service
	if userRepo != nil && courseRepo != nil && lessonRepo != nil && progressRepo != nil && enrollmentRepo != nil {
		dashboardService = dashboarduc.NewService(
			userRepo,
			courseRepo,
			lessonRepo,
			progressRepo,
			enrollmentRepo,
			achievementRepo,
		)
	}

	// AI service
	var aiService aiuc.Service
	if cfg.OpenAIAPIKey != "" {
		openAIClient := ai.NewClient(
			cfg.OpenAIAPIKey,
			cfg.OpenAIBaseURL,
			cfg.OpenAIModel,
			cfg.OpenAIMaxTokens,
			cfg.OpenAITemperature,
		)
		if err := openAIClient.ValidateConfig(); err == nil {
			aiService = aiuc.NewService(openAIClient, logger)
		} else {
			logger.Error("openai client config invalid", "error", err)
		}
	}

	// Code execution service
	var codeExecService codeexecuc.Service
	if cfg.Judge0APIURL != "" {
		judge0Client := codeexec.NewClient(cfg.Judge0APIURL, cfg.Judge0APIKey)
		timeout := time.Duration(cfg.CodeExecutionTimeout) * time.Millisecond
		memoryLimitKB := cfg.CodeExecutionMemoryLimit * 1024 // MB to KB
		codeExecService = codeexecuc.NewService(judge0Client, logger, timeout, memoryLimitKB)
	} else {
		logger.Warn("judge0 not configured, code execution will be limited")
	}

	router := httpdelivery.NewRouter(logger, courseService, authService, jwtManager, cfg, lessonService, assignmentService, progressService, enrollService, sectionService, moduleService, achievementService, dashboardService, aiService, codeExecService)
	logger.Info("starting http server", "port", cfg.HTTPPort)
	if err := router.Run(cfg.HTTPPort); err != nil {
		logger.Error("http server stopped with error", "error", err)
	}
}
