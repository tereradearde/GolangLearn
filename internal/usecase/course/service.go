package course

import (
	"context"
	"errors"
	"time"

	dom "github.com/example/learngo/internal/domain/course"
	"github.com/example/learngo/pkg/utils"
	"github.com/google/uuid"
)

// ErrNotFound возвращается, когда курс не найден.
var ErrNotFound = errors.New("course not found")

// Service интерфейс бизнес-логики.
type Service interface {
	ListCourses(ctx context.Context) ([]dom.Course, error)
	CreateCourse(ctx context.Context, title, description string, opts ...func(*dom.Course)) (dom.Course, error)
	GetCourse(ctx context.Context, id uuid.UUID) (dom.Course, error)
	GetCourseBySlug(ctx context.Context, slug string) (dom.Course, error)
	UpdateCourse(ctx context.Context, id uuid.UUID, updated dom.Course) (dom.Course, error)
	DeleteCourse(ctx context.Context, id uuid.UUID) error
	SearchCourses(ctx context.Context, f dom.ListFilter) (dom.ListResult, error)
}

// service реализация бизнес-логики.
type service struct {
	repo   dom.Repository
	logger *utils.Logger
}

// NewService конструктор сервиса курсов.
func NewService(repo dom.Repository, logger *utils.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) ListCourses(ctx context.Context) ([]dom.Course, error) {
	start := time.Now()
	courses, err := s.repo.List(ctx)
	s.logger.Debug("usecase: list courses", "duration_ms", time.Since(start).Milliseconds())
	return courses, err
}

func (s *service) CreateCourse(ctx context.Context, title, description string, opts ...func(*dom.Course)) (dom.Course, error) {
	start := time.Now()
	course := dom.Course{ID: uuid.New(), Title: title, Description: description}
	for _, apply := range opts {
		apply(&course)
	}
	created, err := s.repo.Create(ctx, course)
	s.logger.Debug("usecase: create course", "duration_ms", time.Since(start).Milliseconds())
	return created, err
}

func (s *service) GetCourse(ctx context.Context, id uuid.UUID) (dom.Course, error) {
	start := time.Now()
	course, err := s.repo.Get(ctx, id)
	s.logger.Debug("usecase: get course", "duration_ms", time.Since(start).Milliseconds())
	if err != nil {
		return dom.Course{}, err
	}
	if course.ID == uuid.Nil {
		return dom.Course{}, ErrNotFound
	}
	return course, nil
}

func (s *service) GetCourseBySlug(ctx context.Context, slug string) (dom.Course, error) {
	start := time.Now()
	course, err := s.repo.GetBySlug(ctx, slug)
	s.logger.Debug("usecase: get course by slug", "duration_ms", time.Since(start).Milliseconds())
	if err != nil {
		return dom.Course{}, err
	}
	if course.ID == uuid.Nil {
		return dom.Course{}, ErrNotFound
	}
	return course, nil
}

func (s *service) UpdateCourse(ctx context.Context, id uuid.UUID, updated dom.Course) (dom.Course, error) {
	updated.ID = id
	updated, err := s.repo.Update(ctx, id, updated)
	if err != nil {
		return dom.Course{}, err
	}
	if updated.ID == uuid.Nil {
		return dom.Course{}, ErrNotFound
	}
	return updated, nil
}

func (s *service) DeleteCourse(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) SearchCourses(ctx context.Context, f dom.ListFilter) (dom.ListResult, error) {
	start := time.Now()
	res, err := s.repo.Search(ctx, f)
	s.logger.Debug("usecase: search courses", "duration_ms", time.Since(start).Milliseconds())
	return res, err
}
