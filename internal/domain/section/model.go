package section

import (
	"context"

	"github.com/google/uuid"
)

// Section представляет главу/модуль курса.
type Section struct {
	ID       uuid.UUID `json:"id"`
	CourseID uuid.UUID `json:"courseId"`
	Title    string    `json:"title"`
	Order    int       `json:"order"`
}

// Repository контракт хранилища разделов.
type Repository interface {
	ListByCourse(ctx context.Context, courseID uuid.UUID) ([]Section, error)
	Create(ctx context.Context, s Section) (Section, error)
	Update(ctx context.Context, id uuid.UUID, title string, order int) (Section, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
