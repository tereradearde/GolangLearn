package achievement

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Achievement, error)
	ListAll(ctx context.Context) ([]Achievement, error)

	GetUserAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (UserAchievement, error)
	ListUserAchievements(ctx context.Context, userID uuid.UUID) ([]UserAchievement, error)
	UnlockAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (UserAchievement, error)
}
