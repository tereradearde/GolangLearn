package achievement

import (
	"context"

	dom "github.com/example/learngo/internal/domain/achievement"
	"github.com/google/uuid"
)

type Service interface {
	GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]dom.UserAchievement, error)
	CheckAndUnlockAchievements(ctx context.Context, userID uuid.UUID, eventType string, eventData map[string]interface{}) error
}

type service struct {
	repo dom.Repository
}

func NewService(repo dom.Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]dom.UserAchievement, error) {
	return s.repo.ListUserAchievements(ctx, userID)
}

func (s *service) CheckAndUnlockAchievements(ctx context.Context, userID uuid.UUID, eventType string, eventData map[string]interface{}) error {
	// Получаем все достижения
	allAchievements, err := s.repo.ListAll(ctx)
	if err != nil {
		return err
	}

	// Проверяем каждое достижение на соответствие критериям
	for _, achievement := range allAchievements {
		// Проверяем, не разблокировано ли уже
		existing, _ := s.repo.GetUserAchievement(ctx, userID, achievement.ID)
		if existing.ID != uuid.Nil {
			continue
		}

		// TODO: Парсим criteria и проверяем соответствие eventType и eventData
		// Пока простая логика: если eventType соответствует достижению
		if s.matchesCriteria(achievement, eventType, eventData) {
			_, err := s.repo.UnlockAchievement(ctx, userID, achievement.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *service) matchesCriteria(achievement dom.Achievement, eventType string, eventData map[string]interface{}) bool {
	// Простая логика проверки
	// TODO: Реализовать полноценную проверку по criteria JSON
	switch achievement.ID {
	case "achievement-first-lesson":
		return eventType == "lesson_completed"
	default:
		return false
	}
}
