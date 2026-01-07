package lesson

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ListByCourse(ctx context.Context, courseID uuid.UUID) ([]Lesson, error)
	ListBySection(ctx context.Context, sectionID uuid.UUID) ([]Lesson, error)
	Create(ctx context.Context, lesson Lesson) (Lesson, error)
	Get(ctx context.Context, id uuid.UUID) (Lesson, error)
	Update(ctx context.Context, id uuid.UUID, title, content string, order int, sectionID uuid.UUID) (Lesson, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
