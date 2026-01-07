package memory

import (
	"context"
	"sort"
	"sync"

	dom "github.com/example/learngo/internal/domain/lesson"
	"github.com/google/uuid"
)

type InMemoryLessonRepository struct {
	mu   sync.RWMutex
	byID map[uuid.UUID]dom.Lesson
}

func NewInMemoryLessonRepository() *InMemoryLessonRepository {
	return &InMemoryLessonRepository{byID: make(map[uuid.UUID]dom.Lesson)}
}

func (r *InMemoryLessonRepository) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Lesson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []dom.Lesson
	for _, l := range r.byID {
		if l.CourseID == courseID {
			list = append(list, l)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Order < list[j].Order })
	return list, nil
}

func (r *InMemoryLessonRepository) ListBySection(ctx context.Context, sectionID uuid.UUID) ([]dom.Lesson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []dom.Lesson
	for _, l := range r.byID {
		if l.SectionID == sectionID {
			list = append(list, l)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Order < list[j].Order })
	return list, nil
}

func (r *InMemoryLessonRepository) Create(ctx context.Context, lesson dom.Lesson) (dom.Lesson, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if lesson.ID == uuid.Nil {
		lesson.ID = uuid.New()
	}
	r.byID[lesson.ID] = lesson
	return lesson, nil
}

func (r *InMemoryLessonRepository) Get(ctx context.Context, id uuid.UUID) (dom.Lesson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if l, ok := r.byID[id]; ok {
		return l, nil
	}
	return dom.Lesson{}, nil
}

func (r *InMemoryLessonRepository) Update(ctx context.Context, id uuid.UUID, title, content string, order int, sectionID uuid.UUID) (dom.Lesson, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.byID[id]
	if !ok {
		return dom.Lesson{}, nil
	}
	l.Title = title
	l.Content = []byte(content)
	l.Order = order
	if sectionID != uuid.Nil {
		l.SectionID = sectionID
	}
	r.byID[id] = l
	return l, nil
}

func (r *InMemoryLessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byID, id)
	return nil
}
