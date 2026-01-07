package utils

import (
	"fmt"

	env "github.com/caarlos0/env/v11"
)

// Config конфигурация приложения, читается из переменных окружения.
type Config struct {
	Env               string   `env:"APP_ENV" envDefault:"dev"`
	HTTPPort          string   `env:"HTTP_PORT" envDefault:":8080"`
	CORSOrigins       []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`
	DBDsn             string   `env:"DB_DSN"`
	JWTSecret         string   `env:"JWT_SECRET" envDefault:"dev-secret-change"`
	JWTTTLMin         int      `env:"JWT_TTL_MIN" envDefault:"60"`
	JWTRefreshSecret  string   `env:"JWT_REFRESH_SECRET" envDefault:"dev-refresh-secret-change"`
	JWTRefreshTTLDays int      `env:"JWT_REFRESH_TTL_DAYS" envDefault:"7"`
	AdminEmail        string   `env:"ADMIN_EMAIL"`
	AdminPassword     string   `env:"ADMIN_PASSWORD"`
	SeedDemo          bool     `env:"SEED_DEMO" envDefault:"false"`
	RabbitURL         string   `env:"RABBIT_URL" envDefault:"amqp://guest:guest@rabbitmq:5672/"`
	RunnerQueue       string   `env:"RUNNER_QUEUE" envDefault:"runner.jobs"`
	RedisAddr         string   `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisPassword     string   `env:"REDIS_PASSWORD" envDefault:""`
	ResultTTLMin      int      `env:"RESULT_TTL_MIN" envDefault:"10"`
	S3Endpoint        string   `env:"S3_ENDPOINT" envDefault:"s3.twcstorage.ru"`
	S3AccessKey       string   `env:"S3_ACCESS_KEY"`
	S3SecretKey       string   `env:"S3_SECRET_KEY"`
	S3Bucket          string   `env:"S3_BUCKET"`
	S3BaseURL         string   `env:"S3_BASE_URL" envDefault:"https://s3.twcstorage.ru"`

	// OpenAI / AI Provider
	OpenAIAPIKey      string  `env:"OPENAI_API_KEY"`
	OpenAIModel       string  `env:"OPENAI_MODEL" envDefault:"gpt-4o"`
	OpenAIMaxTokens   int     `env:"OPENAI_MAX_TOKENS" envDefault:"500"`
	OpenAITemperature float64 `env:"OPENAI_TEMPERATURE" envDefault:"0.7"`
	OpenAIBaseURL     string  `env:"OPENAI_BASE_URL" envDefault:"https://api.openai.com/v1"`

	// Code Execution
	CodeExecutionTimeout     int    `env:"CODE_EXECUTION_TIMEOUT" envDefault:"5000"`     // milliseconds
	CodeExecutionMemoryLimit int    `env:"CODE_EXECUTION_MEMORY_LIMIT" envDefault:"128"` // MB
	Judge0APIURL             string `env:"JUDGE0_API_URL" envDefault:"https://judge0.com/api/v1"`
	Judge0APIKey             string `env:"JUDGE0_API_KEY"`

	// Rate Limiting
	RateLimitAI      int `env:"RATE_LIMIT_AI" envDefault:"60"`       // requests per hour
	RateLimitExecute int `env:"RATE_LIMIT_EXECUTE" envDefault:"100"` // requests per hour
	RateLimitGlobal  int `env:"RATE_LIMIT_GLOBAL" envDefault:"1000"` // requests per hour
}

// LoadConfig загружает конфигурацию из окружения.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	return &cfg, nil
}
