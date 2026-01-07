package postgres

import (
	"context"
	"time"

	dom "github.com/example/learngo/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email        string     `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string     `gorm:"size:255;not null"`
	Name         string     `gorm:"size:100;not null"`
	AvatarURL    string     `gorm:"type:text"`
	Role         string     `gorm:"size:32;not null"`
	CreatedAt    time.Time  `gorm:"not null"`
	UpdatedAt    time.Time  `gorm:"not null"`
	LastLoginAt  *time.Time `gorm:"default:null"`
}

func (UserModel) TableName() string { return "users" }

func userToModel(u dom.User) UserModel {
	return UserModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		AvatarURL:    u.AvatarURL,
		Role:         string(u.Role),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		LastLoginAt:  u.LastLoginAt,
	}
}

func userToDomain(m UserModel) dom.User {
	return dom.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Name:         m.Name,
		AvatarURL:    m.AvatarURL,
		Role:         dom.Role(m.Role),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		LastLoginAt:  m.LastLoginAt,
	}
}

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) AutoMigrate() error { return r.db.AutoMigrate(&UserModel{}) }

func (r *UserRepository) Create(ctx context.Context, u dom.User) (dom.User, error) {
	m := userToModel(u)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return dom.User{}, err
	}
	return userToDomain(m), nil
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, updated dom.User) (dom.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.User{}, nil
		}
		return dom.User{}, err
	}
	m.Name = updated.Name
	m.AvatarURL = updated.AvatarURL
	m.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return dom.User{}, err
	}
	return userToDomain(m), nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": now,
		"updated_at":    now,
	}).Error
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (dom.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).First(&m, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.User{}, nil
		}
		return dom.User{}, err
	}
	return userToDomain(m), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (dom.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return dom.User{}, nil
		}
		return dom.User{}, err
	}
	return userToDomain(m), nil
}
