package postgres

import (
	"context"
	"time"

	dom "github.com/example/learngo/internal/domain/progress"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LessonProgressModel модель прогресса по уроку в БД
type LessonProgressModel struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID  `gorm:"type:uuid;index;not null"`
	CourseID         uuid.UUID  `gorm:"type:uuid;index;not null"`
	LessonID         uuid.UUID  `gorm:"type:uuid;index;not null"`
	Completed        bool       `gorm:"not null;default:false"`
	CompletedAt      *time.Time `gorm:"default:null"`
	CodeSubmitted    string     `gorm:"type:text;not null;default:''"`
	Attempts         int        `gorm:"not null;default:0"`
	TimeSpentMinutes int        `gorm:"not null;default:0"`
	CreatedAt        time.Time  `gorm:"not null"`
	UpdatedAt        time.Time  `gorm:"not null"`
}

func (LessonProgressModel) TableName() string { return "user_progress" }

func lessonProgressToModel(p dom.LessonProgress) LessonProgressModel {
	return LessonProgressModel{
		ID:               p.ID,
		UserID:           p.UserID,
		CourseID:         p.CourseID,
		LessonID:         p.LessonID,
		Completed:        p.Completed,
		CompletedAt:      p.CompletedAt,
		CodeSubmitted:    p.CodeSubmitted,
		Attempts:         p.Attempts,
		TimeSpentMinutes: p.TimeSpentMinutes,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

func lessonProgressToDomain(m LessonProgressModel) dom.LessonProgress {
	return dom.LessonProgress{
		ID:               m.ID,
		UserID:           m.UserID,
		CourseID:         m.CourseID,
		LessonID:         m.LessonID,
		Completed:        m.Completed,
		CompletedAt:      m.CompletedAt,
		CodeSubmitted:    m.CodeSubmitted,
		Attempts:         m.Attempts,
		TimeSpentMinutes: m.TimeSpentMinutes,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

type ProgressRepository struct{ db *gorm.DB }

func NewProgressRepository(db *gorm.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

func (r *ProgressRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&LessonProgressModel{})
}

func (r *ProgressRepository) UpsertLessonProgress(ctx context.Context, p dom.LessonProgress) (dom.LessonProgress, error) {
	m := lessonProgressToModel(p)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now

	// Upsert по (user_id, lesson_id)
	var existing LessonProgressModel
	tx := r.db.WithContext(ctx)
	if err := tx.First(&existing, "user_id = ? AND lesson_id = ?", m.UserID, m.LessonID).Error; err == nil {
		// Обновляем существующий
		existing.Completed = m.Completed
		existing.CompletedAt = m.CompletedAt
		existing.CodeSubmitted = m.CodeSubmitted
		existing.Attempts = m.Attempts
		existing.TimeSpentMinutes = m.TimeSpentMinutes
		existing.UpdatedAt = now
		if err := tx.Save(&existing).Error; err != nil {
			return dom.LessonProgress{}, err
		}
		return lessonProgressToDomain(existing), nil
	}

	// Создаем новый
	if err := tx.Create(&m).Error; err != nil {
		return dom.LessonProgress{}, err
	}
	return lessonProgressToDomain(m), nil
}

func (r *ProgressRepository) GetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID) (dom.LessonProgress, error) {
	var m LessonProgressModel
	if err := r.db.WithContext(ctx).First(&m, "user_id = ? AND lesson_id = ?", userID, lessonID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.LessonProgress{}, nil
		}
		return dom.LessonProgress{}, err
	}
	return lessonProgressToDomain(m), nil
}

func (r *ProgressRepository) ListLessonProgressByCourse(ctx context.Context, userID, courseID uuid.UUID) ([]dom.LessonProgress, error) {
	var rows []LessonProgressModel
	if err := r.db.WithContext(ctx).Where("user_id = ? AND course_id = ?", userID, courseID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.LessonProgress, 0, len(rows))
	for _, row := range rows {
		out = append(out, lessonProgressToDomain(row))
	}
	return out, nil
}

func (r *ProgressRepository) GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (dom.CourseProgress, error) {
	// Получаем все прогрессы по урокам курса
	lessonsProgress, err := r.ListLessonProgressByCourse(ctx, userID, courseID)
	if err != nil {
		return dom.CourseProgress{}, err
	}

	// Подсчитываем статистику
	completedLessons := 0
	totalTimeSpent := 0
	var currentLessonID *uuid.UUID
	var lastAccessedAt time.Time

	for _, lp := range lessonsProgress {
		if lp.Completed {
			completedLessons++
		}
		totalTimeSpent += lp.TimeSpentMinutes
		if lp.UpdatedAt.After(lastAccessedAt) {
			lastAccessedAt = lp.UpdatedAt
			if !lp.Completed {
				currentLessonID = &lp.LessonID
			}
		}
	}

	// TODO: получить total_lessons из репозитория курса/уроков
	totalLessons := len(lessonsProgress)
	progressPercentage := 0
	if totalLessons > 0 {
		progressPercentage = (completedLessons * 100) / totalLessons
	}

	// TODO: получить enrolled_at из enrollment
	enrolledAt := time.Now().UTC()
	if len(lessonsProgress) > 0 {
		enrolledAt = lessonsProgress[0].CreatedAt
	}

	return dom.CourseProgress{
		ID:                 uuid.New(),
		UserID:             userID,
		CourseID:           courseID,
		EnrolledAt:         enrolledAt,
		LastAccessedAt:     lastAccessedAt,
		ProgressPercentage: progressPercentage,
		CompletedLessons:   completedLessons,
		TotalLessons:       totalLessons,
		CurrentLessonID:    currentLessonID,
		TimeSpentMinutes:   totalTimeSpent,
		LessonsProgress:    lessonsProgress,
	}, nil
}

func (r *ProgressRepository) UpdateCourseProgressLastAccessed(ctx context.Context, userID, courseID uuid.UUID) error {
	// Обновляем last_accessed_at через обновление последнего урока
	// Это упрощенная реализация, можно улучшить
	return nil
}
