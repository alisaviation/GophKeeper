package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

// postgresStorage реализует Storage для PostgreSQL
type postgresStorage struct {
	db      *pgxpool.Pool
	users   interfaces.UserRepository
	secrets interfaces.SecretRepository
}

// NewStorage создает новый экземпляр Storage для PostgreSQL
func NewStorage(db *pgxpool.Pool) interfaces.Storage {
	return &postgresStorage{
		db:      db,
		users:   NewUserRepository(db),
		secrets: NewSecretRepository(db),
	}
}

// UserRepository возвращает репозиторий пользователей
func (s *postgresStorage) UserRepository() interfaces.UserRepository {
	return s.users
}

// SecretRepository возвращает репозиторий секретов
func (s *postgresStorage) SecretRepository() interfaces.SecretRepository {
	return s.secrets
}

// TransactionManager возвращает менеджер транзакций
func (s *postgresStorage) TransactionManager() interfaces.TransactionManager {
	return s
}

// BeginTx начинает новую транзакцию
func (s *postgresStorage) BeginTx(ctx context.Context) (interfaces.Transaction, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return NewPostgresTransaction(tx), nil
}

// Close закрывает соединение с базой данных
func (s *postgresStorage) Close() error {
	s.db.Close()
	return nil
}

// Ping проверяет соединение с базой данных
func (s *postgresStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
