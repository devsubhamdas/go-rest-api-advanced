package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Subham-Das-98/go-rest-api-advanced/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DB_PORT,
		cfg.DB_USER,
		cfg.DB_PASSWORD,
		cfg.DB_NAME,
		cfg.DB_SSLMODE,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("NewPostgres:: failed to get sql.DB \n%w", err)
	}

	// Connection pool configuration
	sqlDB.SetMaxOpenConns(3)
	sqlDB.SetMaxIdleConns(3)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Check database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("PingContext:: failed to ping database: \n%w", err)
	}

	slog.Info("database connected successfully...")
	return db, nil
}

func CloseConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("CloseConnection:: failed to get sql.DB: \n%w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("CloseConnection:: failed to close database: \n%w", err)
	}

	slog.Info("database connection closed successfully...")
	return nil
}
