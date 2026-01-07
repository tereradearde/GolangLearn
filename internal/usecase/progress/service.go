package progress

import (
	"context"
	"time"

	dom "github.com/example/learngo/internal/domain/progress"
	"github.com/google/uuid"
)

type Service interface {
	GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (dom.CourseProgress, error)
	UpsertLessonProgress(ctx context.Context, userID, courseID, lessonID uuid.UUID, code string, completed bool, timeSpentMinutes int) (dom.LessonProgress, error)
}

type service struct{ repo dom.Repository }

func NewService(r dom.Repository) Service {
	return &service{repo: r}
}

func (s *service) GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (dom.CourseProgress, error) {
	return s.repo.GetCourseProgress(ctx, userID, courseID)
}

func (s *service) UpsertLessonProgress(ctx context.Context, userID, courseID, lessonID uuid.UUID, code string, completed bool, timeSpentMinutes int) (dom.LessonProgress, error) {
	// Получаем существующий прогресс
	existing, _ := s.repo.GetLessonProgress(ctx, userID, lessonID)

	now := time.Now().UTC()
	attempts := existing.Attempts + 1

	var completedAt *time.Time
	if completed && !existing.Completed {
		completedAt = &now
	} else if existing.CompletedAt != nil {
		completedAt = existing.CompletedAt
	}

	progress := dom.LessonProgress{
		ID:               existing.ID,
		UserID:           userID,
		CourseID:         courseID,
		LessonID:         lessonID,
		Completed:        completed,
		CompletedAt:      completedAt,
		CodeSubmitted:    code,
		Attempts:         attempts,
		TimeSpentMinutes: existing.TimeSpentMinutes + timeSpentMinutes,
		CreatedAt:        existing.CreatedAt,
		UpdatedAt:        now,
	}

	if progress.CreatedAt.IsZero() {
		progress.CreatedAt = now
	}

	return s.repo.UpsertLessonProgress(ctx, progress)
}
