package dashboard

import (
	"context"
	"errors"
	"sync"

	achievementdom "github.com/example/learngo/internal/domain/achievement"
	coursedom "github.com/example/learngo/internal/domain/course"
	dashboarddom "github.com/example/learngo/internal/domain/dashboard"
	enrollmentdom "github.com/example/learngo/internal/domain/enrollment"
	lessondom "github.com/example/learngo/internal/domain/lesson"
	progressdom "github.com/example/learngo/internal/domain/progress"
	userdom "github.com/example/learngo/internal/domain/user"
	"github.com/google/uuid"
)

// Service интерфейс сервиса дашборда
type Service interface {
	GetDashboard(ctx context.Context, userID uuid.UUID) (dashboarddom.Dashboard, error)
}

type service struct {
	userRepo        userdom.Repository
	courseRepo      coursedom.Repository
	lessonRepo      lessondom.Repository
	progressRepo    progressdom.Repository
	enrollmentRepo  enrollmentdom.Repository
	achievementRepo achievementdom.Repository
}

func NewService(
	userRepo userdom.Repository,
	courseRepo coursedom.Repository,
	lessonRepo lessondom.Repository,
	progressRepo progressdom.Repository,
	enrollmentRepo enrollmentdom.Repository,
	achievementRepo achievementdom.Repository,
) Service {
	return &service{
		userRepo:        userRepo,
		courseRepo:      courseRepo,
		lessonRepo:      lessonRepo,
		progressRepo:    progressRepo,
		enrollmentRepo:  enrollmentRepo,
		achievementRepo: achievementRepo,
	}
}

func (s *service) GetDashboard(ctx context.Context, userID uuid.UUID) (dashboarddom.Dashboard, error) {
	// Параллельная загрузка данных для производительности
	var (
		user         userdom.User
		enrollments  []enrollmentdom.Enrollment
		achievements []achievementdom.UserAchievement
		userErr      error
		enrollErr    error
		achErr       error
	)

	var wg sync.WaitGroup
	wg.Add(3)

	// Загружаем пользователя
	go func() {
		defer wg.Done()
		user, userErr = s.userRepo.GetByID(ctx, userID)
	}()

	// Загружаем записи на курсы
	go func() {
		defer wg.Done()
		enrollments, enrollErr = s.enrollmentRepo.ListByUser(ctx, userID)
	}()

	// Загружаем достижения
	go func() {
		defer wg.Done()
		if s.achievementRepo != nil {
			achievements, achErr = s.achievementRepo.ListUserAchievements(ctx, userID)
		}
	}()

	wg.Wait()

	// Проверяем критические ошибки
	if userErr != nil {
		return dashboarddom.Dashboard{}, userErr
	}
	if user.ID == uuid.Nil {
		return dashboarddom.Dashboard{}, errors.New("user not found")
	}

	// Обрабатываем ошибки с fallback
	if enrollErr != nil {
		enrollments = []enrollmentdom.Enrollment{}
	}
	if achErr != nil {
		achievements = []achievementdom.UserAchievement{}
	}

	// Собираем статистику
	stats := s.calculateStats(ctx, userID, enrollments)

	// Получаем активные курсы
	activeCourses := s.getActiveCourses(ctx, userID, enrollments)

	// Получаем достижения с деталями
	achievementsList := s.getAchievementsWithDetails(ctx, achievements)

	// Получаем последние активности
	recentActivity := s.getRecentActivity(ctx, userID)

	return dashboarddom.Dashboard{
		User: dashboarddom.UserInfo{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
		},
		Stats:          stats,
		ActiveCourses:  activeCourses,
		Achievements:   achievementsList,
		RecentActivity: recentActivity,
	}, nil
}

func (s *service) calculateStats(ctx context.Context, userID uuid.UUID, enrollments []enrollmentdom.Enrollment) dashboarddom.UserStats {
	stats := dashboarddom.UserStats{
		CoursesEnrolled: len(enrollments),
	}

	// Подсчитываем завершенные курсы и уроки
	totalLessonsCompleted := 0
	totalTimeMinutes := 0
	completedCourses := 0

	for _, enrollment := range enrollments {
		progress, err := s.progressRepo.GetCourseProgress(ctx, userID, enrollment.CourseID)
		if err != nil {
			continue // Пропускаем при ошибке
		}

		totalLessonsCompleted += progress.CompletedLessons
		totalTimeMinutes += progress.TimeSpentMinutes

		if progress.ProgressPercentage >= 100 {
			completedCourses++
		}
	}

	stats.TotalLessonsCompleted = totalLessonsCompleted
	stats.TotalTimeHours = totalTimeMinutes / 60
	stats.CoursesCompleted = completedCourses

	// TODO: Вычислить streaks из истории активности
	// Пока используем заглушки
	stats.CurrentStreak = 0
	stats.LongestStreak = 0

	return stats
}

func (s *service) getActiveCourses(ctx context.Context, userID uuid.UUID, enrollments []enrollmentdom.Enrollment) []dashboarddom.ActiveCourse {
	activeCourses := make([]dashboarddom.ActiveCourse, 0, len(enrollments))

	for _, enrollment := range enrollments {
		// Получаем прогресс курса
		progress, err := s.progressRepo.GetCourseProgress(ctx, userID, enrollment.CourseID)
		if err != nil {
			continue
		}

		// Пропускаем завершенные курсы
		if progress.ProgressPercentage >= 100 {
			continue
		}

		// Получаем информацию о курсе
		course, err := s.courseRepo.Get(ctx, enrollment.CourseID)
		if err != nil || course.ID == uuid.Nil {
			continue
		}

		// Получаем текущий урок
		currentLessonTitle := ""
		if progress.CurrentLessonID != nil {
			lesson, err := s.lessonRepo.Get(ctx, *progress.CurrentLessonID)
			if err == nil && lesson.ID != uuid.Nil {
				currentLessonTitle = lesson.Title
			}
		}

		activeCourses = append(activeCourses, dashboarddom.ActiveCourse{
			CourseID:           course.ID,
			Title:              course.Title,
			Language:           course.Language,
			ProgressPercentage: progress.ProgressPercentage,
			LastAccessedAt:     progress.LastAccessedAt,
			CurrentLessonTitle: currentLessonTitle,
		})
	}

	return activeCourses
}

func (s *service) getAchievementsWithDetails(ctx context.Context, userAchievements []achievementdom.UserAchievement) []dashboarddom.Achievement {
	achievements := make([]dashboarddom.Achievement, 0, len(userAchievements))

	for _, ua := range userAchievements {
		ach, err := s.achievementRepo.GetByID(ctx, ua.AchievementID)
		if err != nil || ach.ID == "" {
			continue
		}

		achievements = append(achievements, dashboarddom.Achievement{
			ID:          ach.ID,
			Title:       ach.Title,
			Description: ach.Description,
			IconURL:     ach.IconURL,
			Unlocked:    true,
			UnlockedAt:  &ua.UnlockedAt,
		})
	}

	return achievements
}

func (s *service) getRecentActivity(ctx context.Context, userID uuid.UUID) []dashboarddom.RecentActivity {
	// Получаем последние завершенные уроки из прогресса
	// TODO: Создать отдельную таблицу активности для более детальной информации
	activities := make([]dashboarddom.RecentActivity, 0)

	// Получаем все enrollments для поиска активности
	enrollments, err := s.enrollmentRepo.ListByUser(ctx, userID)
	if err != nil {
		return activities
	}

	// Собираем последние завершенные уроки
	for _, enrollment := range enrollments {
		progress, err := s.progressRepo.GetCourseProgress(ctx, userID, enrollment.CourseID)
		if err != nil {
			continue
		}

		course, err := s.courseRepo.Get(ctx, enrollment.CourseID)
		if err != nil {
			continue
		}

		// Берем последние 5 завершенных уроков
		completedLessons := 0
		for _, lp := range progress.LessonsProgress {
			if lp.Completed && lp.CompletedAt != nil {
				lesson, err := s.lessonRepo.Get(ctx, lp.LessonID)
				if err == nil && lesson.ID != uuid.Nil {
					activities = append(activities, dashboarddom.RecentActivity{
						Type:        "lesson_completed",
						LessonTitle: lesson.Title,
						CourseTitle: course.Title,
						Timestamp:   *lp.CompletedAt,
					})
					completedLessons++
					if completedLessons >= 5 {
						break
					}
				}
			}
		}
	}

	// Сортируем по времени (новые первыми)
	// Простая сортировка
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[i].Timestamp.Before(activities[j].Timestamp) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	// Ограничиваем 10 последними
	if len(activities) > 10 {
		activities = activities[:10]
	}

	return activities
}
