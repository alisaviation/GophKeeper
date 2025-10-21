package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alisaviation/GophKeeper/internal/config"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
	"github.com/alisaviation/GophKeeper/internal/server/storage/memory"
	"github.com/alisaviation/GophKeeper/internal/server/storage/postgres"
)

// Type представляет тип хранилища
type Type string

const (
	TypePostgreSQL Type = "postgres"
	TypeMemory     Type = "memory"
)

// RunMigrations применяет миграции к базе данных
func RunMigrations(dbConfig config.DatabaseConfig) error {
	return postgres.RunMigrations(dbConfig)
}

// Config представляет конфигурацию хранилища
type Config struct {
	Type     Type
	Database config.DatabaseConfig
}

// NewStorage создает новое хранилище в зависимости от конфигурации
func NewStorage(config Config) (interfaces.Storage, error) {
	switch config.Type {
	case TypePostgreSQL:
		dsn := config.Database.DSN()

		poolConfig, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DSN: %w", err)
		}

		poolConfig.MaxConns = int32(config.Database.MaxOpenConns)
		poolConfig.MinConns = int32(config.Database.MaxIdleConns)
		poolConfig.MaxConnLifetime = config.Database.ConnMaxLifetime
		poolConfig.MaxConnIdleTime = config.Database.ConnMaxIdleTime

		db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection pool: %w", err)
		}

		return postgres.NewStorage(db), nil

	case TypeMemory:
		return memory.NewStorage(), nil

	default:
		return nil, fmt.Errorf("unknown storage type: %s", config.Type)
	}
}
