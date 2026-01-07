package postgres

import (
	"context"

	dom "github.com/example/learngo/internal/domain/assignment"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssignmentModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	LessonID    uuid.UUID `gorm:"type:uuid;index;not null"`
	Title       string    `gorm:"size:255;not null"`
	Prompt      string    `gorm:"type:text;not null"`
	StarterCode string    `gorm:"type:text;not null"`
	Tests       string    `gorm:"type:text;not null"`
	Order       int       `gorm:"not null;column:sort_order"`
}

func (AssignmentModel) TableName() string { return "assignments" }

func assignmentToModel(a dom.Assignment) AssignmentModel {
	return AssignmentModel{ID: a.ID, LessonID: a.LessonID, Title: a.Title, Prompt: a.Prompt, StarterCode: a.StarterCode, Tests: a.Tests, Order: a.Order}
}
func assignmentToDomain(m AssignmentModel) dom.Assignment {
	return dom.Assignment{ID: m.ID, LessonID: m.LessonID, Title: m.Title, Prompt: m.Prompt, StarterCode: m.StarterCode, Tests: m.Tests, Order: m.Order}
}

type AssignmentRepository struct{ db *gorm.DB }

func NewAssignmentRepository(db *gorm.DB) *AssignmentRepository { return &AssignmentRepository{db: db} }
func (r *AssignmentRepository) AutoMigrate() error              { return r.db.AutoMigrate(&AssignmentModel{}) }

func (r *AssignmentRepository) ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]dom.Assignment, error) {
	var rows []AssignmentModel
	if err := r.db.WithContext(ctx).Where("lesson_id = ?", lessonID).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.Assignment, 0, len(rows))
	for _, row := range rows {
		out = append(out, assignmentToDomain(row))
	}
	return out, nil
}

func (r *AssignmentRepository) Create(ctx context.Context, a dom.Assignment) (dom.Assignment, error) {
	m := assignmentToModel(a)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return dom.Assignment{}, err
	}
	return assignmentToDomain(m), nil
}

func (r *AssignmentRepository) Get(ctx context.Context, id uuid.UUID) (dom.Assignment, error) {
	var m AssignmentModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.Assignment{}, nil
		}
		return dom.Assignment{}, err
	}
	return assignmentToDomain(m), nil
}

func (r *AssignmentRepository) Update(ctx context.Context, id uuid.UUID, title, prompt, starterCode, tests string, order int) (dom.Assignment, error) {
	var m AssignmentModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.Assignment{}, nil
		}
		return dom.Assignment{}, err
	}
	m.Title = title
	m.Prompt = prompt
	m.StarterCode = starterCode
	m.Tests = tests
	m.Order = order
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return dom.Assignment{}, err
	}
	return assignmentToDomain(m), nil
}

func (r *AssignmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&AssignmentModel{}, "id = ?", id).Error
}
