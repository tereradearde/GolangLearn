package module

import (
	"context"

	dom "github.com/example/learngo/internal/domain/module"
	"github.com/example/learngo/pkg/utils"
	"github.com/google/uuid"
)

type Service interface {
	ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Module, error)
	Create(ctx context.Context, courseID uuid.UUID, title string, orderIndex int) (dom.Module, error)
	Update(ctx context.Context, id uuid.UUID, title string, orderIndex int) (dom.Module, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo   dom.Repository
	logger *utils.Logger
}

func NewService(repo dom.Repository, logger *utils.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Module, error) {
	return s.repo.ListByCourse(ctx, courseID)
}
func (s *service) Create(ctx context.Context, courseID uuid.UUID, title string, orderIndex int) (dom.Module, error) {
	if orderIndex <= 0 {
		orderIndex = 1
	}
	return s.repo.Create(ctx, dom.Module{ID: uuid.New(), CourseID: courseID, Title: title, OrderIndex: orderIndex})
}
func (s *service) Update(ctx context.Context, id uuid.UUID, title string, orderIndex int) (dom.Module, error) {
	return s.repo.Update(ctx, id, title, orderIndex)
}
func (s *service) Delete(ctx context.Context, id uuid.UUID) error { return s.repo.Delete(ctx, id) }
