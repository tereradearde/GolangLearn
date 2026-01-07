package enrollment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Enrollment struct {
	UserID    uuid.UUID `json:"userId"`
	CourseID  uuid.UUID `json:"courseId"`
	Status    string    `json:"status"` // enrolled|purchased
	CreatedAt time.Time `json:"createdAt"`
}

type Repository interface {
	Upsert(ctx context.Context, e Enrollment) error
	IsEnrolled(ctx context.Context, userID, courseID uuid.UUID) (bool, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Enrollment, error)
}
