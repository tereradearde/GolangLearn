package section

import (
	"context"

	dom "github.com/example/learngo/internal/domain/section"
	"github.com/example/learngo/pkg/utils"
	"github.com/google/uuid"
)

type Service interface {
	ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Section, error)
	Create(ctx context.Context, courseID uuid.UUID, title string, order int) (dom.Section, error)
	Update(ctx context.Context, id uuid.UUID, title string, order int) (dom.Section, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo   dom.Repository
	logger *utils.Logger
}

func NewService(repo dom.Repository, logger *utils.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Section, error) {
	return s.repo.ListByCourse(ctx, courseID)
}

func (s *service) Create(ctx context.Context, courseID uuid.UUID, title string, order int) (dom.Section, error) {
	if order <= 0 {
		order = 1
	}
	return s.repo.Create(ctx, dom.Section{ID: uuid.New(), CourseID: courseID, Title: title, Order: order})
}

func (s *service) Update(ctx context.Context, id uuid.UUID, title string, order int) (dom.Section, error) {
	return s.repo.Update(ctx, id, title, order)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error { return s.repo.Delete(ctx, id) }
