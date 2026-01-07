package postgres

import (
	"context"

	dom "github.com/example/learngo/internal/domain/section"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SectionModel struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	CourseID uuid.UUID `gorm:"type:uuid;index;not null"`
	Title    string    `gorm:"size:255;not null"`
	Order    int       `gorm:"not null;column:sort_order"`
}

func (SectionModel) TableName() string { return "sections" }

func toSectionModel(s dom.Section) SectionModel {
	return SectionModel{ID: s.ID, CourseID: s.CourseID, Title: s.Title, Order: s.Order}
}
func toSectionDomain(m SectionModel) dom.Section {
	return dom.Section{ID: m.ID, CourseID: m.CourseID, Title: m.Title, Order: m.Order}
}

type SectionRepository struct{ db *gorm.DB }

func NewSectionRepository(db *gorm.DB) *SectionRepository { return &SectionRepository{db: db} }
func (r *SectionRepository) AutoMigrate() error           { return r.db.AutoMigrate(&SectionModel{}) }

func (r *SectionRepository) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Section, error) {
	var rows []SectionModel
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.Section, 0, len(rows))
	for _, m := range rows {
		out = append(out, toSectionDomain(m))
	}
	return out, nil
}

func (r *SectionRepository) Create(ctx context.Context, s dom.Section) (dom.Section, error) {
	m := toSectionModel(s)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return dom.Section{}, err
	}
	return toSectionDomain(m), nil
}

func (r *SectionRepository) Update(ctx context.Context, id uuid.UUID, title string, order int) (dom.Section, error) {
	var m SectionModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return dom.Section{}, err
	}
	m.Title, m.Order = title, order
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return dom.Section{}, err
	}
	return toSectionDomain(m), nil
}

func (r *SectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&SectionModel{}, "id = ?", id).Error
}
