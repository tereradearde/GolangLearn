package utils

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager отвечает за выпуск и валидацию JWT.
type JWTManager struct {
	secret        []byte
	refreshSecret []byte
	ttl           time.Duration
	refreshTTL    time.Duration
}

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string, ttlMinutes int, refreshSecret string, refreshTTLDays int) *JWTManager {
	if ttlMinutes <= 0 {
		ttlMinutes = 60
	}
	if refreshTTLDays <= 0 {
		refreshTTLDays = 7
	}
	return &JWTManager{
		secret:        []byte(secret),
		refreshSecret: []byte(refreshSecret),
		ttl:           time.Duration(ttlMinutes) * time.Minute,
		refreshTTL:    time.Duration(refreshTTLDays) * 24 * time.Hour,
	}
}

func (m *JWTManager) Generate(userID uuid.UUID, role string) (string, error) {
	return m.generateWithSecret(userID, role, m.secret, m.ttl)
}

func (m *JWTManager) GenerateRefresh(userID uuid.UUID, role string) (string, error) {
	return m.generateWithSecret(userID, role, m.refreshSecret, m.refreshTTL)
}

func (m *JWTManager) generateWithSecret(userID uuid.UUID, role string, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (m *JWTManager) Verify(tokenString string) (*Claims, error) {
	return m.verifyWithSecret(tokenString, m.secret)
}

func (m *JWTManager) VerifyRefresh(tokenString string) (*Claims, error) {
	return m.verifyWithSecret(tokenString, m.refreshSecret)
}

func (m *JWTManager) verifyWithSecret(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
