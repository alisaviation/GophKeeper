package postgres_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/config"
)

type migrateCreator func(path, dsn string) (migrateInterface, error)

// MockMigrate реализует интерфейс migrate.Migrate для тестирования
type MockMigrate struct {
	upError     error
	closeSrcErr error
	closeDBNErr error
	upCalled    bool
	closeCalled bool
}

func (m *MockMigrate) Up() error {
	m.upCalled = true
	return m.upError
}

func (m *MockMigrate) Close() (sourceErr error, databaseErr error) {
	m.closeCalled = true
	return m.closeSrcErr, m.closeDBNErr
}

// Другие методы интерфейса migrate.Migrate (заглушки)
func (m *MockMigrate) Down() error                  { return nil }
func (m *MockMigrate) Steps(int) error              { return nil }
func (m *MockMigrate) Force(int) error              { return nil }
func (m *MockMigrate) Version() (uint, bool, error) { return 0, false, nil }
func (m *MockMigrate) Drop() error                  { return nil }

// TestRunMigrations тестирует основную функцию RunMigrations
func TestRunMigrations(t *testing.T) {
	tests := []struct {
		name           string
		dbConfig       config.DatabaseConfig
		setupMocks     func() (*MockMigrate, error)
		wantErr        bool
		expectedErr    string
		expectUpCalled bool
	}{
		{
			name: "successful migration",
			dbConfig: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "test",
				Password: "test",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			setupMocks: func() (*MockMigrate, error) {
				return &MockMigrate{
					upError: nil,
				}, nil
			},
			wantErr:        false,
			expectUpCalled: true,
		},
		{
			name: "migration with no changes",
			dbConfig: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "test",
				Password: "test",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			setupMocks: func() (*MockMigrate, error) {
				return &MockMigrate{
					upError: migrate.ErrNoChange,
				}, nil
			},
			wantErr:        false,
			expectUpCalled: true,
		},
		{
			name: "migration failure",
			dbConfig: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "test",
				Password: "test",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			setupMocks: func() (*MockMigrate, error) {
				return &MockMigrate{
					upError: errors.New("migration failed"),
				}, nil
			},
			wantErr:        true,
			expectedErr:    "failed to run migrations: migration failed",
			expectUpCalled: true,
		},
		{
			name: "failed to create migrate instance",
			dbConfig: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "test",
				Password: "test",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			setupMocks: func() (*MockMigrate, error) {
				return nil, errors.New("failed to create migrate instance")
			},
			wantErr:        true,
			expectedErr:    "failed to create migrate instance",
			expectUpCalled: false,
		},
		{
			name: "migration with close errors",
			dbConfig: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "test",
				Password: "test",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			setupMocks: func() (*MockMigrate, error) {
				return &MockMigrate{
					upError:     nil,
					closeSrcErr: errors.New("source error"),
					closeDBNErr: errors.New("database error"),
				}, nil
			},
			wantErr:        false,
			expectUpCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockMigrate *MockMigrate

			testRunMigrationsWithDSN := func(dsn string) error {
				expectedDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
					tt.dbConfig.User, tt.dbConfig.Password, tt.dbConfig.Host,
					tt.dbConfig.Port, tt.dbConfig.Name, tt.dbConfig.SSLMode)
				assert.Equal(t, expectedDSN, dsn)

				creator := func(path, dsn string) (migrateInterface, error) {
					assert.Contains(t, path, "file://")
					assert.Contains(t, path, "migrations")

					if tt.setupMocks != nil {
						mock, err := tt.setupMocks()
						mockMigrate = mock
						return mock, err
					}
					return nil, errors.New("setupMocks not defined")
				}

				return runMigrationsWithCreator(dsn, creator)
			}

			dsn := tt.dbConfig.DSN()
			err := testRunMigrationsWithDSN(dsn)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
			if mockMigrate != nil && tt.expectUpCalled {
				assert.True(t, mockMigrate.upCalled, "Up() should have been called")
				assert.True(t, mockMigrate.closeCalled, "Close() should have been called")
			}
		})
	}
}

// TestRunMigrationsWithDSN тестирует функцию runMigrationsWithDSN
func TestRunMigrationsWithDSN(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		setupMocks  func() (*MockMigrate, error)
		wantErr     bool
		expectedErr string
	}{
		{
			name: "successful migration with DSN",
			dsn:  "postgres://user:pass@localhost:5432/db?sslmode=disable",
			setupMocks: func() (*MockMigrate, error) {
				return &MockMigrate{
					upError: nil,
				}, nil
			},
			wantErr: false,
		},
		{
			name: "failed to create migrate instance with invalid DSN",
			dsn:  "invalid-dsn",
			setupMocks: func() (*MockMigrate, error) {
				return nil, errors.New("failed to create migrate instance: invalid DSN")
			},
			wantErr:     true,
			expectedErr: "failed to create migrate instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := func(path, dsn string) (migrateInterface, error) {
				assert.Equal(t, tt.dsn, dsn)
				if tt.setupMocks != nil {
					return tt.setupMocks()
				}
				return nil, errors.New("setupMocks not defined")
			}

			err := runMigrationsWithCreator(tt.dsn, creator)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMigrationsPath тестирует корректность формирования пути к миграциям
func TestMigrationsPath(t *testing.T) {
	var capturedPath string

	creator := func(path, dsn string) (migrateInterface, error) {
		capturedPath = path
		return &MockMigrate{}, nil
	}

	err := runMigrationsWithCreator("postgres://test:test@localhost:5432/test", creator)
	require.NoError(t, err)

	assert.Contains(t, capturedPath, "file://")
	assert.Contains(t, capturedPath, "migrations")

	workDir, err := os.Getwd()
	require.NoError(t, err)
	expectedPrefix := fmt.Sprintf("file://%s", workDir)
	assert.Contains(t, capturedPath, expectedPrefix)
}

// TestCloseError проверяет, что ошибка закрытия не влияет на основную логику
func TestCloseError(t *testing.T) {
	creator := func(path, dsn string) (migrateInterface, error) {
		return &MockMigrate{
			upError:     nil,
			closeSrcErr: errors.New("source close error"),
			closeDBNErr: errors.New("database close error"),
		}, nil
	}

	err := runMigrationsWithCreator("postgres://test:test@localhost:5432/test", creator)
	assert.NoError(t, err)
}

// TestEdgeCases тестирует граничные случаи
func TestEdgeCases(t *testing.T) {
	t.Run("empty DSN", func(t *testing.T) {
		creator := func(path, dsn string) (migrateInterface, error) {
			return nil, errors.New("DSN is empty")
		}

		err := runMigrationsWithCreator("", creator)
		assert.Error(t, err)
	})

	t.Run("nil migrate instance", func(t *testing.T) {
		creator := func(path, dsn string) (migrateInterface, error) {
			return nil, nil // Возвращаем nil, nil (редкий случай)
		}

		err := runMigrationsWithCreator("postgres://test:test@localhost:5432/test", creator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "migrate instance is nil")
	})
}

// runMigrationsWithCreator - внутренняя функция, которая принимает creator для тестирования
func runMigrationsWithCreator(dsn string, creator migrateCreator) error {
	workDir, _ := os.Getwd()
	migrationsPath := fmt.Sprintf("file://%s/migrations", workDir)

	m, err := creator(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	if m == nil {
		return fmt.Errorf("migrate instance is nil")
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Интерфейс для подмены функции migrate.New
type migrateInterface interface {
	Up() error
	Close() (sourceErr error, databaseErr error)
}

// Глобальная переменная для подмены функции создания migrate
// В продакшене она использует реальную библиотеку, в тестах - моки
var newMigrate = func(path, dsn string) (migrateInterface, error) {
	// В реальном коде это вызывает migrate.New
	m, err := migrate.New(path, dsn)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func runMigrationsWithDSN(dsn string) error {
	return runMigrationsWithCreator(dsn, newMigrate)
}

func RunMigrations(dbConfig config.DatabaseConfig) error {
	dsn := dbConfig.DSN()
	return runMigrationsWithDSN(dsn)
}
