package assignment

import (
	"context"

	dom "github.com/example/learngo/internal/domain/assignment"
	"github.com/example/learngo/pkg/utils"
	"github.com/google/uuid"
)

type Service interface {
	ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]dom.Assignment, error)
	Create(ctx context.Context, lessonID uuid.UUID, title, prompt, starterCode, tests string, order int) (dom.Assignment, error)
	Get(ctx context.Context, id uuid.UUID) (dom.Assignment, error)
	Update(ctx context.Context, id uuid.UUID, title, prompt, starterCode, tests string, order int) (dom.Assignment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo   dom.Repository
	logger *utils.Logger
}

func NewService(repo dom.Repository, logger *utils.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]dom.Assignment, error) {
	return s.repo.ListByLesson(ctx, lessonID)
}

func (s *service) Create(ctx context.Context, lessonID uuid.UUID, title, prompt, starterCode, tests string, order int) (dom.Assignment, error) {
	a := dom.Assignment{ID: uuid.New(), LessonID: lessonID, Title: title, Prompt: prompt, StarterCode: starterCode, Tests: tests, Order: order}
	return s.repo.Create(ctx, a)
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (dom.Assignment, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, title, prompt, starterCode, tests string, order int) (dom.Assignment, error) {
	return s.repo.Update(ctx, id, title, prompt, starterCode, tests, order)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
