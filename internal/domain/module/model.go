package module

import (
	"context"

	"github.com/google/uuid"
)

// Module — логический раздел курса.
type Module struct {
	ID         uuid.UUID `json:"id"`
	CourseID   uuid.UUID `json:"courseId"`
	Title      string    `json:"title"`
	OrderIndex int       `json:"orderIndex"`
}

type Repository interface {
	ListByCourse(ctx context.Context, courseID uuid.UUID) ([]Module, error)
	Create(ctx context.Context, m Module) (Module, error)
	Update(ctx context.Context, id uuid.UUID, title string, orderIndex int) (Module, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
