package progress

import (
	"time"

	"github.com/google/uuid"
)

// LessonProgress модель прогресса пользователя по уроку.
type LessonProgress struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	CourseID         uuid.UUID  `json:"course_id"`
	LessonID         uuid.UUID  `json:"lesson_id"`
	Completed        bool       `json:"completed"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CodeSubmitted    string     `json:"code_submitted"`
	Attempts         int        `json:"attempts"`
	TimeSpentMinutes int        `json:"time_spent_minutes"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CourseProgress модель агрегированного прогресса пользователя по курсу.
type CourseProgress struct {
	ID                 uuid.UUID        `json:"id"`
	UserID             uuid.UUID        `json:"user_id"`
	CourseID           uuid.UUID        `json:"course_id"`
	EnrolledAt         time.Time        `json:"enrolled_at"`
	LastAccessedAt     time.Time        `json:"last_accessed_at"`
	ProgressPercentage int              `json:"progress_percentage"`
	CompletedLessons   int              `json:"completed_lessons"`
	TotalLessons       int              `json:"total_lessons"`
	CurrentLessonID    *uuid.UUID       `json:"current_lesson_id,omitempty"`
	TimeSpentMinutes   int              `json:"time_spent_minutes"`
	LessonsProgress    []LessonProgress `json:"lessons_progress"`
}
