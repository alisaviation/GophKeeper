package postgres

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/alisaviation/GophKeeper/internal/config"
)

// RunMigrations применяет миграции к базе данных используя DatabaseConfig
func RunMigrations(dbConfig config.DatabaseConfig) error {
	dsn := dbConfig.DSN()
	return runMigrationsWithDSN(dsn)
}

func runMigrationsWithDSN(dsn string) error {
	workDir, _ := os.Getwd()
	migrationsPath := fmt.Sprintf("file://%s/migrations", workDir)

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
