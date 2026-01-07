package dashboard

import (
	"time"

	"github.com/google/uuid"
)

// Dashboard модель дашборда пользователя
type Dashboard struct {
	User           UserInfo         `json:"user"`
	Stats          UserStats        `json:"stats"`
	ActiveCourses  []ActiveCourse   `json:"active_courses"`
	Achievements   []Achievement    `json:"achievements"`
	RecentActivity []RecentActivity `json:"recent_activity"`
}

// UserInfo информация о пользователе
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}

// UserStats статистика пользователя
type UserStats struct {
	CurrentStreak         int `json:"current_streak"`          // текущая серия дней
	LongestStreak         int `json:"longest_streak"`          // самая длинная серия
	TotalLessonsCompleted int `json:"total_lessons_completed"` // всего завершено уроков
	TotalTimeHours        int `json:"total_time_hours"`        // всего часов потрачено
	CoursesEnrolled       int `json:"courses_enrolled"`        // курсов записано
	CoursesCompleted      int `json:"courses_completed"`       // курсов завершено
}

// ActiveCourse активный курс пользователя
type ActiveCourse struct {
	CourseID           uuid.UUID `json:"course_id"`
	Title              string    `json:"title"`
	Language           string    `json:"language"`
	ProgressPercentage int       `json:"progress_percentage"`
	LastAccessedAt     time.Time `json:"last_accessed_at"`
	CurrentLessonTitle string    `json:"current_lesson_title"`
}

// Achievement достижение пользователя
type Achievement struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	IconURL     string     `json:"icon_url"`
	Unlocked    bool       `json:"unlocked"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
}

// RecentActivity последняя активность
type RecentActivity struct {
	Type        string    `json:"type"` // lesson_completed, course_started, etc.
	LessonTitle string    `json:"lesson_title,omitempty"`
	CourseTitle string    `json:"course_title,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}
