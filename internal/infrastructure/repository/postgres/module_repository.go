package postgres

import (
	"context"

	dom "github.com/example/learngo/internal/domain/module"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModuleModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	CourseID   uuid.UUID `gorm:"type:uuid;index;not null"`
	Title      string    `gorm:"size:255;not null"`
	OrderIndex int       `gorm:"not null;column:order_index"`
}

func (ModuleModel) TableName() string { return "modules" }
func toModuleModel(m dom.Module) ModuleModel {
	return ModuleModel{ID: m.ID, CourseID: m.CourseID, Title: m.Title, OrderIndex: m.OrderIndex}
}
func toModuleDomain(m ModuleModel) dom.Module {
	return dom.Module{ID: m.ID, CourseID: m.CourseID, Title: m.Title, OrderIndex: m.OrderIndex}
}

type ModuleRepository struct{ db *gorm.DB }

func NewModuleRepository(db *gorm.DB) *ModuleRepository { return &ModuleRepository{db: db} }
func (r *ModuleRepository) AutoMigrate() error          { return r.db.AutoMigrate(&ModuleModel{}) }

func (r *ModuleRepository) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]dom.Module, error) {
	var rows []ModuleModel
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Order("order_index asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]dom.Module, 0, len(rows))
	for _, row := range rows {
		out = append(out, toModuleDomain(row))
	}
	return out, nil
}

func (r *ModuleRepository) Create(ctx context.Context, m dom.Module) (dom.Module, error) {
	mm := toModuleModel(m)
	if mm.ID == uuid.Nil {
		mm.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(&mm).Error; err != nil {
		return dom.Module{}, err
	}
	return toModuleDomain(mm), nil
}

func (r *ModuleRepository) Update(ctx context.Context, id uuid.UUID, title string, orderIndex int) (dom.Module, error) {
	var mm ModuleModel
	if err := r.db.WithContext(ctx).First(&mm, "id = ?", id).Error; err != nil {
		return dom.Module{}, err
	}
	mm.Title, mm.OrderIndex = title, orderIndex
	if err := r.db.WithContext(ctx).Save(&mm).Error; err != nil {
		return dom.Module{}, err
	}
	return toModuleDomain(mm), nil
}

func (r *ModuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&ModuleModel{}, "id = ?", id).Error
}
