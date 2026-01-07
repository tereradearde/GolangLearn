package lesson

import (
	"context"
	"encoding/json"

	dom "github.com/example/learngo/internal/domain/lesson"
	"github.com/example/learngo/pkg/utils"
	"github.com/google/uuid"
)

type Service interface {
	ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Lesson, error)
	ListBySection(ctx context.Context, sectionID uuid.UUID) ([]dom.Lesson, error)
	Create(ctx context.Context, courseID uuid.UUID, title, content string, order int) (dom.Lesson, error)
	CreateInSection(ctx context.Context, courseID, sectionID uuid.UUID, title, content string, order int) (dom.Lesson, error)
	Get(ctx context.Context, id uuid.UUID) (dom.Lesson, error)
	Update(ctx context.Context, id uuid.UUID, title, content string, order int, sectionID uuid.UUID) (dom.Lesson, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo   dom.Repository
	logger *utils.Logger
}

func NewService(repo dom.Repository, logger *utils.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Lesson, error) {
	return s.repo.ListByCourse(ctx, courseID)
}

func (s *service) ListBySection(ctx context.Context, sectionID uuid.UUID) ([]dom.Lesson, error) {
	return s.repo.ListBySection(ctx, sectionID)
}

func (s *service) Create(ctx context.Context, courseID uuid.UUID, title, content string, order int) (dom.Lesson, error) {
	raw := json.RawMessage(content)
	l := dom.Lesson{ID: uuid.New(), CourseID: courseID, Title: title, Content: raw, Order: order}
	return s.repo.Create(ctx, l)
}

func (s *service) CreateInSection(ctx context.Context, courseID, sectionID uuid.UUID, title, content string, order int) (dom.Lesson, error) {
	raw := json.RawMessage(content)
	l := dom.Lesson{ID: uuid.New(), CourseID: courseID, SectionID: sectionID, Title: title, Content: raw, Order: order}
	return s.repo.Create(ctx, l)
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (dom.Lesson, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, title, content string, order int, sectionID uuid.UUID) (dom.Lesson, error) {
	return s.repo.Update(ctx, id, title, content, order, sectionID)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
