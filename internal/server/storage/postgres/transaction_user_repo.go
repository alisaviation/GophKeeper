package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

// txUserRepository реализует UserRepository для транзакций
type txUserRepository struct {
	tx pgx.Tx
}

// NewTxUserRepository создает новый UserRepository для транзакций
func NewTxUserRepository(tx pgx.Tx) interfaces.UserRepository {
	return &txUserRepository{tx: tx}
}

// Create создает нового пользователя
func (r *txUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, login, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.tx.Exec(ctx, query,
		user.ID,
		user.Login,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetByLogin получает пользователя по логину
func (r *txUserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	query := `
		SELECT id, login, password_hash, created_at, updated_at
		FROM users
		WHERE login = $1
	`

	var user domain.User
	err := r.tx.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByID получает пользователя по ID
func (r *txUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, login, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.tx.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update обновляет данные пользователя
func (r *txUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET login = $1, password_hash = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.tx.Exec(ctx, query,
		user.Login,
		user.PasswordHash,
		user.UpdatedAt,
		user.ID,
	)

	return err
}

// Delete удаляет пользователя
func (r *txUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}
