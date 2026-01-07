package memory

import (
	"context"
	"sync"
	"time"

	dom "github.com/example/learngo/internal/domain/user"
	"github.com/google/uuid"
)

// InMemoryUserRepository простая потокобезопасная in-memory реализация репозитория пользователей.
type InMemoryUserRepository struct {
	mu      sync.RWMutex
	byID    map[uuid.UUID]dom.User
	byEmail map[string]uuid.UUID
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{byID: make(map[uuid.UUID]dom.User), byEmail: make(map[string]uuid.UUID)}
}

func (r *InMemoryUserRepository) Create(ctx context.Context, u dom.User) (dom.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u.ID
	return u, nil
}

func (r *InMemoryUserRepository) GetByEmail(ctx context.Context, email string) (dom.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if id, ok := r.byEmail[email]; ok {
		if u, ok2 := r.byID[id]; ok2 {
			return u, nil
		}
	}
	return dom.User{}, nil
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id uuid.UUID) (dom.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.byID[id]; ok {
		return u, nil
	}
	return dom.User{}, nil
}

func (r *InMemoryUserRepository) Update(ctx context.Context, id uuid.UUID, u dom.User) (dom.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[id]; !ok {
		return dom.User{}, nil
	}
	u.ID = id
	r.byID[id] = u
	r.byEmail[u.Email] = id
	return u, nil
}

func (r *InMemoryUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.byID[id]; ok {
		now := time.Now().UTC()
		u.LastLoginAt = &now
		r.byID[id] = u
	}
	return nil
}
