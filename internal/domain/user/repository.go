package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository контракт хранилища пользователей.
type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	Update(ctx context.Context, id uuid.UUID, updated User) (User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}
