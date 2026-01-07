package postgres

import (
	"context"
	"time"

	dom "github.com/example/learngo/internal/domain/achievement"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementModel struct {
	ID          string    `gorm:"type:varchar(100);primaryKey"`
	Title       string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text;not null"`
	IconURL     string    `gorm:"size:512;not null;default:''"`
	Criteria    string    `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (AchievementModel) TableName() string { return "achievements" }

type UserAchievementModel struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID `gorm:"type:uuid;index;not null"`
	AchievementID string    `gorm:"type:varchar(100);index;not null"`
	UnlockedAt    time.Time `gorm:"not null"`
}

func (UserAchievementModel) TableName() string { return "user_achievements" }

func achievementToDomain(m AchievementModel) dom.Achievement {
	return dom.Achievement{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		IconURL:     m.IconURL,
		Criteria:    []byte(m.Criteria),
		CreatedAt:   m.CreatedAt,
	}
}

func userAchievementToDomain(m UserAchievementModel) dom.UserAchievement {
	return dom.UserAchievement{
		ID:            m.ID,
		UserID:        m.UserID,
		AchievementID: m.AchievementID,
		UnlockedAt:    m.UnlockedAt,
	}
}

type AchievementRepository struct{ db *gorm.DB }

func NewAchievementRepository(db *gorm.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func (r *AchievementRepository) AutoMigrate() error {
	if err := r.db.AutoMigrate(&AchievementModel{}); err != nil {
		return err
	}
	return r.db.AutoMigrate(&UserAchievementModel{})
}

func (r *AchievementRepository) GetByID(ctx context.Context, id string) (dom.Achievement, error) {
	var m AchievementModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.Achievement{}, nil
		}
		return dom.Achievement{}, err
	}
	return achievementToDomain(m), nil
}

func (r *AchievementRepository) ListAll(ctx context.Context) ([]dom.Achievement, error) {
	var rows []AchievementModel
	if err := r.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.Achievement, 0, len(rows))
	for _, row := range rows {
		out = append(out, achievementToDomain(row))
	}
	return out, nil
}

func (r *AchievementRepository) GetUserAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (dom.UserAchievement, error) {
	var m UserAchievementModel
	if err := r.db.WithContext(ctx).First(&m, "user_id = ? AND achievement_id = ?", userID, achievementID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.UserAchievement{}, nil
		}
		return dom.UserAchievement{}, err
	}
	return userAchievementToDomain(m), nil
}

func (r *AchievementRepository) ListUserAchievements(ctx context.Context, userID uuid.UUID) ([]dom.UserAchievement, error) {
	var rows []UserAchievementModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.UserAchievement, 0, len(rows))
	for _, row := range rows {
		out = append(out, userAchievementToDomain(row))
	}
	return out, nil
}

func (r *AchievementRepository) UnlockAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (dom.UserAchievement, error) {
	// Проверяем, не разблокировано ли уже
	existing, _ := r.GetUserAchievement(ctx, userID, achievementID)
	if existing.ID != uuid.Nil {
		return existing, nil
	}

	// Создаем новое достижение
	m := UserAchievementModel{
		ID:            uuid.New(),
		UserID:        userID,
		AchievementID: achievementID,
		UnlockedAt:    time.Now().UTC(),
	}

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return dom.UserAchievement{}, err
	}

	return userAchievementToDomain(m), nil
}
