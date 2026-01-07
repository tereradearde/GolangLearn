package achievement

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Achievement модель достижения
type Achievement struct {
	ID          string          `json:"id"`          // строковый ID (например, "achievement-first-lesson")
	Title       string          `json:"title"`       // название
	Description string          `json:"description"` // описание
	IconURL     string          `json:"icon_url"`    // URL иконки
	Criteria    json.RawMessage `json:"criteria"`    // JSON с критериями получения
	CreatedAt   time.Time       `json:"created_at"`
}

// UserAchievement модель достижения пользователя
type UserAchievement struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	AchievementID string    `json:"achievement_id"`
	UnlockedAt    time.Time `json:"unlocked_at"`
}
