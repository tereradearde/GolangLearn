package httpdelivery

import (
	"time"

	"github.com/example/learngo/pkg/ratelimit"
	"github.com/example/learngo/pkg/utils"
	"github.com/gin-gonic/gin"
)

// attachRateLimits добавляет лимиты на наиболее чувствительные роуты.
func attachRateLimits(api *gin.RouterGroup, cfg *utils.Config) {
	// Auth endpoints: 10 requests per minute
	authStore := ratelimit.NewInMemoryStore(10)
	authGroup := api.Group("/auth")
	authGroup.Use(ratelimit.Middleware(authStore, 10, time.Minute, ratelimit.KeyByIP))
}

// aiRateLimiter возвращает middleware для AI эндпоинтов
func aiRateLimiter(cfg *utils.Config) gin.HandlerFunc {
	// AI endpoints: configurable per hour
	store := ratelimit.NewInMemoryStore(cfg.RateLimitAI)
	return ratelimit.Middleware(store, cfg.RateLimitAI, time.Hour, ratelimit.KeyByUserID)
}

// codeExecRateLimiter возвращает middleware для code execution
func codeExecRateLimiter(cfg *utils.Config) gin.HandlerFunc {
	// Code execution: configurable per hour
	store := ratelimit.NewInMemoryStore(cfg.RateLimitExecute)
	return ratelimit.Middleware(store, cfg.RateLimitExecute, time.Hour, ratelimit.KeyByUserID)
}

// globalRateLimiter возвращает middleware для глобального лимита
func globalRateLimiter(cfg *utils.Config) gin.HandlerFunc {
	// Global limit: configurable per hour
	store := ratelimit.NewInMemoryStore(cfg.RateLimitGlobal)
	return ratelimit.Middleware(store, cfg.RateLimitGlobal, time.Hour, ratelimit.KeyByIP)
}
