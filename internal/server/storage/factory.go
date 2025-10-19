package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

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

// Config представляет конфигурацию хранилища
type Config struct {
	Type     Type             `yaml:"type"`
	Postgres *postgres.Config `yaml:"postgres,omitempty"`
}

// NewStorage создает новое хранилище в зависимости от конфигурации
func NewStorage(config Config) (interfaces.Storage, error) {
	switch config.Type {
	case TypePostgreSQL:
		if config.Postgres == nil {
			return nil, fmt.Errorf("postgres config is required for postgres storage type")
		}

		poolConfig, err := pgxpool.ParseConfig(config.Postgres.DSN())
		if err != nil {
			return nil, fmt.Errorf("failed to parse postgres DSN: %w", err)
		}

		poolConfig.MaxConns = int32(config.Postgres.MaxOpenConns)
		poolConfig.MinConns = int32(config.Postgres.MaxIdleConns)
		poolConfig.MaxConnLifetime = config.Postgres.ConnMaxLifetime
		poolConfig.MaxConnIdleTime = config.Postgres.ConnMaxIdleTime

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
