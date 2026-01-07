package memory

import (
	"context"
	"sync"

	dom "github.com/example/learngo/internal/domain/enrollment"
	"github.com/google/uuid"
)

type InMemoryEnrollmentRepository struct {
	mu sync.RWMutex
	// ключ: userID:courseID
	m map[string]dom.Enrollment
}

func NewInMemoryEnrollmentRepository() *InMemoryEnrollmentRepository {
	return &InMemoryEnrollmentRepository{m: make(map[string]dom.Enrollment)}
}

func key(u, c uuid.UUID) string { return u.String() + ":" + c.String() }

func (r *InMemoryEnrollmentRepository) Upsert(ctx context.Context, e dom.Enrollment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[key(e.UserID, e.CourseID)] = e
	return nil
}

func (r *InMemoryEnrollmentRepository) IsEnrolled(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.m[key(userID, courseID)]
	return ok, nil
}

func (r *InMemoryEnrollmentRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]dom.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]dom.Enrollment, 0)
	for _, v := range r.m {
		if v.UserID == userID {
			res = append(res, v)
		}
	}
	return res, nil
}
