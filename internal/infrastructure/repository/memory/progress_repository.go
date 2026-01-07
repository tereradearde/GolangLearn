package memory

import (
	"context"
	"sync"
	"time"

	dom "github.com/example/learngo/internal/domain/progress"
	"github.com/google/uuid"
)

type InMemoryProgressRepository struct {
	mu sync.RWMutex
	// ключ: userID + ":" + courseID
	byKey map[string]dom.CourseProgress
	// ключ: userID + ":" + lessonID для lesson progress
	lessonByKey map[string]dom.LessonProgress
}

func NewInMemoryProgressRepository() *InMemoryProgressRepository {
	return &InMemoryProgressRepository{
		byKey:       make(map[string]dom.CourseProgress),
		lessonByKey: make(map[string]dom.LessonProgress),
	}
}

func makeKey(userID, courseID uuid.UUID) string { return userID.String() + ":" + courseID.String() }
func makeLessonKey(userID, lessonID uuid.UUID) string {
	return userID.String() + ":" + lessonID.String()
}

func (r *InMemoryProgressRepository) Upsert(ctx context.Context, p dom.CourseProgress) (dom.CourseProgress, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := makeKey(p.UserID, p.CourseID)
	// сохраняем как есть; ID можно не использовать в памяти
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	r.byKey[key] = p
	return p, nil
}

func (r *InMemoryProgressRepository) GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (dom.CourseProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := makeKey(userID, courseID)
	if v, ok := r.byKey[key]; ok {
		return v, nil
	}
	return dom.CourseProgress{}, nil
}

func (r *InMemoryProgressRepository) GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (dom.CourseProgress, error) {
	return r.GetByUserAndCourse(ctx, userID, courseID)
}

func (r *InMemoryProgressRepository) UpsertLessonProgress(ctx context.Context, p dom.LessonProgress) (dom.LessonProgress, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := makeLessonKey(p.UserID, p.LessonID)
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	r.lessonByKey[key] = p
	return p, nil
}

func (r *InMemoryProgressRepository) GetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID) (dom.LessonProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := makeLessonKey(userID, lessonID)
	if v, ok := r.lessonByKey[key]; ok {
		return v, nil
	}
	return dom.LessonProgress{}, nil
}

func (r *InMemoryProgressRepository) ListLessonProgressByCourse(ctx context.Context, userID, courseID uuid.UUID) ([]dom.LessonProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]dom.LessonProgress, 0)
	for _, p := range r.lessonByKey {
		if p.UserID == userID && p.CourseID == courseID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *InMemoryProgressRepository) UpdateCourseProgressLastAccessed(ctx context.Context, userID, courseID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := makeKey(userID, courseID)
	if p, ok := r.byKey[key]; ok {
		p.LastAccessedAt = time.Now().UTC()
		r.byKey[key] = p
	} else {
		// Создаем новый прогресс если не существует
		p = dom.CourseProgress{
			UserID:         userID,
			CourseID:       courseID,
			LastAccessedAt: time.Now().UTC(),
			EnrolledAt:     time.Now().UTC(),
		}
		r.byKey[key] = p
	}
	return nil
}
