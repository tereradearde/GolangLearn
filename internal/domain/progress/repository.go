package progress

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// LessonProgress методы
	UpsertLessonProgress(ctx context.Context, p LessonProgress) (LessonProgress, error)
	GetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID) (LessonProgress, error)
	ListLessonProgressByCourse(ctx context.Context, userID, courseID uuid.UUID) ([]LessonProgress, error)

	// CourseProgress методы (агрегированные)
	GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (CourseProgress, error)
	UpdateCourseProgressLastAccessed(ctx context.Context, userID, courseID uuid.UUID) error
}
