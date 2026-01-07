package main

import (
	"context"
	"time"

	adom "github.com/example/learngo/internal/domain/assignment"
	cdom "github.com/example/learngo/internal/domain/course"
	ldom "github.com/example/learngo/internal/domain/lesson"
	udom "github.com/example/learngo/internal/domain/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// seedDev наполняет данными при пустой БД (или in-memory) в dev-режиме.
func seedDev(ctx context.Context, cfg *ConfigLike, repos *Repositories, auth AuthServiceLike, logger *LoggerLike) {
	if cfg.SeedDemo {
		// создать админа, если указан email/password
		if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
			// проверить наличие пользователя
			if u, _ := repos.User.GetByEmail(ctx, cfg.AdminEmail); u.ID == uuid.Nil {
				// Создаём администратора с корректным хешем пароля
				_ = createUser(ctx, repos, cfg.AdminEmail, cfg.AdminPassword)
			}
		}
		// если нет курсов — создать демо-курс с уроком и заданием
		courses, _ := repos.Course.List(ctx)
		if len(courses) == 0 {
			c := cdom.Course{ID: uuid.New(), Title: "Демо курс Go", Description: "Быстрый старт"}
			_, _ = repos.Course.Create(ctx, c)
			l := ldom.Lesson{ID: uuid.New(), CourseID: c.ID, Title: "Введение", Content: []byte("[]"), Order: 1}
			_, _ = repos.Lesson.Create(ctx, l)
			a := adom.Assignment{ID: uuid.New(), LessonID: l.ID, Title: "Hello", Prompt: "Напечатайте Hello", StarterCode: "package main\nimport \"fmt\"\nfunc main(){}", Tests: "[]", Order: 1}
			_, _ = repos.Assignment.Create(ctx, a)
			_ = auth // reserved for future when services needed
			_ = logger
			_ = time.Now()
		}
	}
}

// createUser — упрощённо, без хеширования (оставим usecase для реальных путей).
func createUser(ctx context.Context, repos *Repositories, email, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := udom.User{ID: uuid.New(), Email: email, PasswordHash: string(hashed), Role: udom.RoleAdmin, CreatedAt: time.Now().UTC()}
	_, err := repos.User.Create(ctx, u)
	return err
}

// Вспомогательные фасады для сидера, чтобы не тянуть все пакеты в main.
type ConfigLike struct {
	SeedDemo      bool
	AdminEmail    string
	AdminPassword string
}

type LoggerLike interface{}

type Repositories struct {
	Course interface {
		List(ctx context.Context) ([]cdom.Course, error)
		Create(ctx context.Context, c cdom.Course) (cdom.Course, error)
	}
	Lesson interface {
		Create(ctx context.Context, l ldom.Lesson) (ldom.Lesson, error)
	}
	Assignment interface {
		Create(ctx context.Context, a adom.Assignment) (adom.Assignment, error)
	}
	User interface {
		GetByEmail(ctx context.Context, email string) (udom.User, error)
		Create(ctx context.Context, u udom.User) (udom.User, error)
	}
}

type ServicesLike struct{}

type AuthServiceLike interface {
	Register(ctx context.Context, email, password, name string) (string, string, udom.User, error)
}
