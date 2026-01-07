package enrollment

import (
	"context"
	"time"

	dom "github.com/example/learngo/internal/domain/enrollment"
	"github.com/example/learngo/pkg/utils"
	"github.com/google/uuid"
)

type Service interface {
	Enroll(ctx context.Context, userID, courseID uuid.UUID, purchased bool) error
	IsEnrolled(ctx context.Context, userID, courseID uuid.UUID) (bool, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]dom.Enrollment, error)
}

type service struct {
	repo   dom.Repository
	logger *utils.Logger
}

func NewService(repo dom.Repository, logger *utils.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) Enroll(ctx context.Context, userID, courseID uuid.UUID, purchased bool) error {
	status := "enrolled"
	if purchased {
		status = "purchased"
	}
	return s.repo.Upsert(ctx, dom.Enrollment{UserID: userID, CourseID: courseID, Status: status, CreatedAt: time.Now()})
}

func (s *service) IsEnrolled(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	return s.repo.IsEnrolled(ctx, userID, courseID)
}

func (s *service) ListByUser(ctx context.Context, userID uuid.UUID) ([]dom.Enrollment, error) {
	return s.repo.ListByUser(ctx, userID)
}
