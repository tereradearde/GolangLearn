package postgres

import (
	"context"
	"time"

	dom "github.com/example/learngo/internal/domain/enrollment"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EnrollmentModel struct {
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid"`
	CourseID  uuid.UUID `gorm:"primaryKey;type:uuid"`
	Status    string
	CreatedAt time.Time
}

type EnrollmentRepository struct{ db *gorm.DB }

func NewEnrollmentRepository(db *gorm.DB) *EnrollmentRepository { return &EnrollmentRepository{db: db} }

func (r *EnrollmentRepository) AutoMigrate() error { return r.db.AutoMigrate(&EnrollmentModel{}) }

func (r *EnrollmentRepository) Upsert(ctx context.Context, e dom.Enrollment) error {
	m := EnrollmentModel{UserID: e.UserID, CourseID: e.CourseID, Status: e.Status, CreatedAt: e.CreatedAt}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "course_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status"}),
	}).Create(&m).Error
}

func (r *EnrollmentRepository) IsEnrolled(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&EnrollmentModel{}).Where("user_id = ? AND course_id = ?", userID, courseID).Count(&cnt).Error
	return cnt > 0, err
}

func (r *EnrollmentRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]dom.Enrollment, error) {
	var rows []EnrollmentModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]dom.Enrollment, 0, len(rows))
	for _, m := range rows {
		res = append(res, dom.Enrollment{UserID: m.UserID, CourseID: m.CourseID, Status: m.Status, CreatedAt: m.CreatedAt})
	}
	return res, nil
}

// no extra helpers
