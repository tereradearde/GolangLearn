package postgres

import (
	"context"

	dom "github.com/example/learngo/internal/domain/lesson"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LessonModel struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CourseID         uuid.UUID  `gorm:"type:uuid;index;not null"`
	ModuleID         uuid.UUID  `gorm:"type:uuid;index;not null;default:'00000000-0000-0000-0000-000000000000'"`
	Slug             string     `gorm:"size:255;index;not null;default:''"`
	Title            string     `gorm:"size:255;not null"`
	Type             string     `gorm:"size:16;not null;default:'text'"`
	Content          string     `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
	DurationMinutes  int        `gorm:"not null;default:0"`
	Order            int        `gorm:"not null;column:sort_order"`
	IsFree           bool       `gorm:"not null;default:false"`
	NextLessonID     *uuid.UUID `gorm:"type:uuid;default:null"`
	PreviousLessonID *uuid.UUID `gorm:"type:uuid;default:null"`
}

func (LessonModel) TableName() string { return "lessons" }

func lessonToModel(l dom.Lesson) LessonModel {
	moduleID := l.ModuleID
	if moduleID == uuid.Nil {
		moduleID = l.SectionID // для обратной совместимости
	}
	return LessonModel{
		ID:               l.ID,
		CourseID:         l.CourseID,
		ModuleID:         moduleID,
		Slug:             l.Slug,
		Title:            l.Title,
		Type:             l.Type,
		Content:          string(l.Content),
		DurationMinutes:  l.DurationMinutes,
		Order:            l.Order,
		IsFree:           l.IsFree,
		NextLessonID:     l.NextLessonID,
		PreviousLessonID: l.PreviousLessonID,
	}
}

func lessonToDomain(m LessonModel) dom.Lesson {
	return dom.Lesson{
		ID:               m.ID,
		CourseID:         m.CourseID,
		ModuleID:         m.ModuleID,
		SectionID:        m.ModuleID, // для обратной совместимости
		Slug:             m.Slug,
		Title:            m.Title,
		Type:             m.Type,
		Content:          []byte(m.Content),
		DurationMinutes:  m.DurationMinutes,
		Order:            m.Order,
		IsFree:           m.IsFree,
		NextLessonID:     m.NextLessonID,
		PreviousLessonID: m.PreviousLessonID,
	}
}

type LessonRepository struct{ db *gorm.DB }

func NewLessonRepository(db *gorm.DB) *LessonRepository { return &LessonRepository{db: db} }
func (r *LessonRepository) AutoMigrate() error          { return r.db.AutoMigrate(&LessonModel{}) }

func (r *LessonRepository) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Lesson, error) {
	var rows []LessonModel
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.Lesson, 0, len(rows))
	for _, row := range rows {
		out = append(out, lessonToDomain(row))
	}
	return out, nil
}

func (r *LessonRepository) ListBySection(ctx context.Context, sectionID uuid.UUID) ([]dom.Lesson, error) {
	var rows []LessonModel
	if err := r.db.WithContext(ctx).Where("module_id = ?", sectionID).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.Lesson, 0, len(rows))
	for _, row := range rows {
		out = append(out, lessonToDomain(row))
	}
	return out, nil
}

func (r *LessonRepository) Create(ctx context.Context, l dom.Lesson) (dom.Lesson, error) {
	m := lessonToModel(l)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return dom.Lesson{}, err
	}
	return lessonToDomain(m), nil
}

func (r *LessonRepository) Get(ctx context.Context, id uuid.UUID) (dom.Lesson, error) {
	var m LessonModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.Lesson{}, nil
		}
		return dom.Lesson{}, err
	}
	return lessonToDomain(m), nil
}

func (r *LessonRepository) Update(ctx context.Context, id uuid.UUID, title, content string, order int, sectionID uuid.UUID) (dom.Lesson, error) {
	var m LessonModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.Lesson{}, nil
		}
		return dom.Lesson{}, err
	}
	m.Title = title
	m.Content = content
	m.Order = order
	if sectionID != uuid.Nil {
		m.ModuleID = sectionID
	}
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return dom.Lesson{}, err
	}
	return lessonToDomain(m), nil
}

func (r *LessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&LessonModel{}, "id = ?", id).Error
}
