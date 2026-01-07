package db

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenPostgres открывает соединение с Postgres через GORM.
func OpenPostgres(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DSN")
	}
	// Кастомный логгер: игнорируем record not found, уровень Warn
	gl := logger.New(stdlog.New(os.Stdout, "", stdlog.LstdFlags), logger.Config{
		SlowThreshold:             1 * time.Second,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
	cfg := &gorm.Config{Logger: gl}
	return gorm.Open(postgres.Open(dsn), cfg)
}
