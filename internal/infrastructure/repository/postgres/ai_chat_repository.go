package postgres

import (
	"context"
	"encoding/json"
	"time"

	aidom "github.com/example/learngo/internal/domain/ai"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AIChatHistoryModel модель для истории чата ИИ
type AIChatHistoryModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;index;not null"`
	LessonID   string    `gorm:"type:varchar(100);index"`
	Messages   string    `gorm:"type:jsonb;not null;default:'[]'::jsonb"`
	TokensUsed int       `gorm:"not null;default:0"`
	CreatedAt  time.Time `gorm:"not null"`
}

func (AIChatHistoryModel) TableName() string { return "ai_chat_history" }

func aiChatHistoryToModel(h aidom.ChatHistory) (AIChatHistoryModel, error) {
	messagesJSON, err := json.Marshal(h.Messages)
	if err != nil {
		return AIChatHistoryModel{}, err
	}

	id, err := uuid.Parse(h.ID)
	if err != nil {
		id = uuid.New()
	}

	userID, err := uuid.Parse(h.UserID)
	if err != nil {
		return AIChatHistoryModel{}, err
	}

	return AIChatHistoryModel{
		ID:         id,
		UserID:     userID,
		LessonID:   h.LessonID,
		Messages:   string(messagesJSON),
		TokensUsed: h.TokensUsed,
		CreatedAt:  h.CreatedAt,
	}, nil
}

func aiChatHistoryToDomain(m AIChatHistoryModel) (aidom.ChatHistory, error) {
	var messages []aidom.ChatMessage
	if err := json.Unmarshal([]byte(m.Messages), &messages); err != nil {
		return aidom.ChatHistory{}, err
	}

	return aidom.ChatHistory{
		ID:         m.ID.String(),
		UserID:     m.UserID.String(),
		LessonID:   m.LessonID,
		Messages:   messages,
		TokensUsed: m.TokensUsed,
		CreatedAt:  m.CreatedAt,
	}, nil
}

// AIChatHistoryRepository репозиторий для истории чата ИИ
type AIChatHistoryRepository struct {
	db *gorm.DB
}

func NewAIChatHistoryRepository(db *gorm.DB) *AIChatHistoryRepository {
	return &AIChatHistoryRepository{db: db}
}

func (r *AIChatHistoryRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&AIChatHistoryModel{})
}

// Create создает новую запись истории чата
func (r *AIChatHistoryRepository) Create(ctx context.Context, h aidom.ChatHistory) (aidom.ChatHistory, error) {
	m, err := aiChatHistoryToModel(h)
	if err != nil {
		return aidom.ChatHistory{}, err
	}

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return aidom.ChatHistory{}, err
	}

	return aiChatHistoryToDomain(m)
}

// ListByUserID возвращает историю чата пользователя
func (r *AIChatHistoryRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]aidom.ChatHistory, error) {
	var rows []AIChatHistoryModel
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]aidom.ChatHistory, 0, len(rows))
	for _, row := range rows {
		domain, err := aiChatHistoryToDomain(row)
		if err != nil {
			continue
		}
		result = append(result, domain)
	}

	return result, nil
}

// ListByLessonID возвращает историю чата для урока
func (r *AIChatHistoryRepository) ListByLessonID(ctx context.Context, lessonID string, limit int) ([]aidom.ChatHistory, error) {
	var rows []AIChatHistoryModel
	query := r.db.WithContext(ctx).Where("lesson_id = ?", lessonID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]aidom.ChatHistory, 0, len(rows))
	for _, row := range rows {
		domain, err := aiChatHistoryToDomain(row)
		if err != nil {
			continue
		}
		result = append(result, domain)
	}

	return result, nil
}

// Delete удаляет запись истории
func (r *AIChatHistoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&AIChatHistoryModel{}, "id = ?", id).Error
}
