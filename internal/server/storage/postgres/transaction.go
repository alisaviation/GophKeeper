package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

// postgresTransaction реализует Transaction для PostgreSQL
type postgresTransaction struct {
	tx pgx.Tx
}

// NewPostgresTransaction создает новую транзакцию
func NewPostgresTransaction(tx pgx.Tx) interfaces.Transaction {
	return &postgresTransaction{tx: tx}
}

// Commit подтверждает транзакцию
func (t *postgresTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback откатывает транзакцию
func (t *postgresTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// UserRepository возвращает UserRepository в контексте транзакции
func (t *postgresTransaction) UserRepository() interfaces.UserRepository {
	// todo создать адаптер для tx
	return &userRepository{db: nil} // Заглушка
}

// SecretRepository возвращает SecretRepository в контексте транзакции
func (t *postgresTransaction) SecretRepository() interfaces.SecretRepository {
	// todo создать адаптер для tx
	return &secretRepository{db: nil} // Заглушка
}
