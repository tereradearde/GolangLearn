package memory

import (
	"context"
	"sort"
	"sync"

	dom "github.com/example/learngo/internal/domain/assignment"
	"github.com/google/uuid"
)

type InMemoryAssignmentRepository struct {
	mu   sync.RWMutex
	byID map[uuid.UUID]dom.Assignment
}

func NewInMemoryAssignmentRepository() *InMemoryAssignmentRepository {
	return &InMemoryAssignmentRepository{byID: make(map[uuid.UUID]dom.Assignment)}
}

func (r *InMemoryAssignmentRepository) ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]dom.Assignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []dom.Assignment
	for _, a := range r.byID {
		if a.LessonID == lessonID {
			list = append(list, a)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Order < list[j].Order })
	return list, nil
}

func (r *InMemoryAssignmentRepository) Create(ctx context.Context, a dom.Assignment) (dom.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	r.byID[a.ID] = a
	return a, nil
}

func (r *InMemoryAssignmentRepository) Get(ctx context.Context, id uuid.UUID) (dom.Assignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if a, ok := r.byID[id]; ok {
		return a, nil
	}
	return dom.Assignment{}, nil
}

func (r *InMemoryAssignmentRepository) Update(ctx context.Context, id uuid.UUID, title, prompt, starterCode, tests string, order int) (dom.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.byID[id]
	if !ok {
		return dom.Assignment{}, nil
	}
	a.Title = title
	a.Prompt = prompt
	a.StarterCode = starterCode
	a.Tests = tests
	a.Order = order
	r.byID[id] = a
	return a, nil
}

func (r *InMemoryAssignmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byID, id)
	return nil
}
