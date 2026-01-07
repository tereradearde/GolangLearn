package assignment

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]Assignment, error)
	Create(ctx context.Context, a Assignment) (Assignment, error)
	Get(ctx context.Context, id uuid.UUID) (Assignment, error)
	Update(ctx context.Context, id uuid.UUID, title, prompt, starterCode, tests string, order int) (Assignment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
