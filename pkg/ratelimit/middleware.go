package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Store интерфейс для хранения данных о rate limit
type Store interface {
	Get(key string) (*LimitInfo, error)
	Increment(key string, window time.Duration) (*LimitInfo, error)
}

// LimitInfo информация о текущем лимите
type LimitInfo struct {
	Count     int
	Limit     int
	ResetTime time.Time
}

// InMemoryStore простая in-memory реализация для dev
type InMemoryStore struct {
	mu    sync.RWMutex
	data  map[string]*LimitInfo
	limit int
}

// NewInMemoryStore создает новый in-memory store
func NewInMemoryStore(limit int) *InMemoryStore {
	return &InMemoryStore{
		data:  make(map[string]*LimitInfo),
		limit: limit,
	}
}

// Get получает информацию о лимите
func (s *InMemoryStore) Get(key string) (*LimitInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.data[key]
	if !exists {
		return &LimitInfo{
			Count:     0,
			Limit:     s.limit,
			ResetTime: time.Now().Add(time.Hour),
		}, nil
	}

	// Проверяем, не истек ли период
	if time.Now().After(info.ResetTime) {
		return &LimitInfo{
			Count:     0,
			Limit:     s.limit,
			ResetTime: time.Now().Add(time.Hour),
		}, nil
	}

	return info, nil
}

// Increment увеличивает счетчик
func (s *InMemoryStore) Increment(key string, window time.Duration) (*LimitInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.data[key]
	now := time.Now()

	if !exists || now.After(info.ResetTime) {
		// Создаем новый период
		info = &LimitInfo{
			Count:     1,
			Limit:     s.limit,
			ResetTime: now.Add(window),
		}
		s.data[key] = info
		return info, nil
	}

	// Увеличиваем счетчик
	info.Count++
	s.data[key] = info

	return info, nil
}

// Middleware создает middleware для rate limiting
func Middleware(store Store, limit int, window time.Duration, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)

		// Получаем текущую информацию
		info, err := store.Get(key)
		if err != nil {
			c.Next()
			return
		}

		// Увеличиваем счетчик
		info, err = store.Increment(key, window)
		if err != nil {
			c.Next()
			return
		}

		// Устанавливаем заголовки
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, info.Limit-info.Count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))

		// Проверяем лимит
		if info.Count > info.Limit {
			// Используем стандартный формат ошибок
			requestID := c.GetString("request_id")
			if requestID == "" {
				requestID = "unknown"
			}

			apiErr := map[string]interface{}{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "Rate limit exceeded",
				"details": map[string]interface{}{
					"limit":  info.Limit,
					"window": window.String(),
				},
				"timestamp":  time.Now().UTC(),
				"request_id": requestID,
			}
			c.JSON(429, apiErr)
			c.Abort()
			return
		}

		c.Next()
	}
}

// KeyByIP возвращает ключ на основе IP адреса
func KeyByIP(c *gin.Context) string {
	return c.ClientIP()
}

// KeyByUserID возвращает ключ на основе user ID (если авторизован)
func KeyByUserID(c *gin.Context) string {
	// Пытаемся получить userId из контекста (устанавливается в AuthRequired middleware)
	userID, exists := c.Get("userId")
	if exists && userID != nil {
		return fmt.Sprintf("user:%v", userID)
	}
	// Fallback на IP если не авторизован
	return c.ClientIP()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
