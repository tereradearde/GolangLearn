package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	dom "github.com/example/learngo/internal/domain/user"
	"github.com/example/learngo/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Service интерфейс аутентификации.
type Service interface {
	Register(ctx context.Context, email, password, name string) (accessToken, refreshToken string, user dom.User, err error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, user dom.User, err error)
	RefreshToken(ctx context.Context, refreshToken string) (accessToken string, err error)
}

type service struct {
	repo dom.Repository
	jwt  *utils.JWTManager
}

func NewService(repo dom.Repository, jwt *utils.JWTManager) Service {
	return &service{repo: repo, jwt: jwt}
}

func hashPassword(password string) (string, error) {
	// Используем cost factor 12 как указано в документации
	b, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(b), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func normalizeEmail(email string) string { return strings.TrimSpace(strings.ToLower(email)) }

func (s *service) Register(ctx context.Context, email, password, name string) (string, string, dom.User, error) {
	email = normalizeEmail(email)
	if len(email) < 5 || len(password) < 8 {
		return "", "", dom.User{}, errors.New("invalid email or password")
	}
	if len(name) < 2 || len(name) > 50 {
		return "", "", dom.User{}, errors.New("name must be between 2 and 50 characters")
	}
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", dom.User{}, err
	}
	if existing.ID != uuid.Nil {
		return "", "", dom.User{}, errors.New("email already in use")
	}
	hpw, err := hashPassword(password)
	if err != nil {
		return "", "", dom.User{}, err
	}
	now := time.Now().UTC()
	u := dom.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hpw,
		Name:         name,
		Role:         dom.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	created, err := s.repo.Create(ctx, u)
	if err != nil {
		return "", "", dom.User{}, err
	}
	accessToken, err := s.jwt.Generate(created.ID, string(created.Role))
	if err != nil {
		return "", "", dom.User{}, err
	}
	refreshToken, err := s.jwt.GenerateRefresh(created.ID, string(created.Role))
	if err != nil {
		return "", "", dom.User{}, err
	}
	return accessToken, refreshToken, created, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, string, dom.User, error) {
	email = normalizeEmail(email)
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", dom.User{}, err
	}
	if u.ID == uuid.Nil || !checkPassword(u.PasswordHash, password) {
		return "", "", dom.User{}, ErrInvalidCredentials
	}
	// Обновляем last_login_at
	_ = s.repo.UpdateLastLogin(ctx, u.ID)
	// Получаем обновленного пользователя
	u, _ = s.repo.GetByID(ctx, u.ID)

	accessToken, err := s.jwt.Generate(u.ID, string(u.Role))
	if err != nil {
		return "", "", dom.User{}, err
	}
	refreshToken, err := s.jwt.GenerateRefresh(u.ID, string(u.Role))
	if err != nil {
		return "", "", dom.User{}, err
	}
	return accessToken, refreshToken, u, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.jwt.VerifyRefresh(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	// Проверяем, что пользователь существует
	u, err := s.repo.GetByID(ctx, claims.UserID)
	if err != nil || u.ID == uuid.Nil {
		return "", errors.New("user not found")
	}
	// Генерируем новый access token
	accessToken, err := s.jwt.Generate(u.ID, string(u.Role))
	if err != nil {
		return "", err
	}
	return accessToken, nil
}
