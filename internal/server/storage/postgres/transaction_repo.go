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

func (r *txUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

// txSecretRepository реализует SecretRepository для транзакций
type txSecretRepository struct {
	tx pgx.Tx
}

// NewTxSecretRepository создает новый SecretRepository для транзакций
func NewTxSecretRepository(tx pgx.Tx) interfaces.SecretRepository {
	return &txSecretRepository{tx: tx}
}

func (r *txSecretRepository) Create(ctx context.Context, secret *domain.Secret) error {
	query := `
		INSERT INTO secrets (id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.tx.Exec(ctx, query,
		secret.ID,
		secret.UserID,
		string(secret.Type),
		secret.Name,
		secret.EncryptedData,
		secret.EncryptedMeta,
		secret.Version,
		secret.CreatedAt,
		secret.UpdatedAt,
		secret.IsDeleted,
	)

	return err
}

func (r *txSecretRepository) GetByID(ctx context.Context, id, userID string) (*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted
		FROM secrets
		WHERE id = $1 AND user_id = $2 AND NOT is_deleted
	`

	var secret domain.Secret
	var secretType string

	err := r.tx.QueryRow(ctx, query, id, userID).Scan(
		&secret.ID,
		&secret.UserID,
		&secretType,
		&secret.Name,
		&secret.EncryptedData,
		&secret.EncryptedMeta,
		&secret.Version,
		&secret.CreatedAt,
		&secret.UpdatedAt,
		&secret.IsDeleted,
	)

	if err != nil {
		return nil, err
	}

	secret.Type = domain.SecretType(secretType)
	return &secret, nil
}

func (r *txSecretRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted
		FROM secrets
		WHERE user_id = $1 AND NOT is_deleted
		ORDER BY created_at DESC
	`

	rows, err := r.tx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []*domain.Secret
	for rows.Next() {
		var secret domain.Secret
		var secretType string

		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secretType,
			&secret.Name,
			&secret.EncryptedData,
			&secret.EncryptedMeta,
			&secret.Version,
			&secret.CreatedAt,
			&secret.UpdatedAt,
			&secret.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		secret.Type = domain.SecretType(secretType)
		secrets = append(secrets, &secret)
	}

	return secrets, nil
}

func (r *txSecretRepository) ListByUserAndType(ctx context.Context, userID string, secretType domain.SecretType) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted
		FROM secrets
		WHERE user_id = $1 AND type = $2 AND NOT is_deleted
		ORDER BY created_at DESC
	`

	rows, err := r.tx.Query(ctx, query, userID, string(secretType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []*domain.Secret
	for rows.Next() {
		var secret domain.Secret
		var st string

		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&st,
			&secret.Name,
			&secret.EncryptedData,
			&secret.EncryptedMeta,
			&secret.Version,
			&secret.CreatedAt,
			&secret.UpdatedAt,
			&secret.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		secret.Type = domain.SecretType(st)
		secrets = append(secrets, &secret)
	}

	return secrets, nil
}

func (r *txSecretRepository) Update(ctx context.Context, secret *domain.Secret) error {
	query := `
		UPDATE secrets 
		SET type = $1, name = $2, encrypted_data = $3, encrypted_meta = $4, 
		    version = version + 1, updated_at = $5, is_deleted = $6
		WHERE id = $7 AND user_id = $8 AND version = $9
	`

	result, err := r.tx.Exec(ctx, query,
		string(secret.Type),
		secret.Name,
		secret.EncryptedData,
		secret.EncryptedMeta,
		secret.UpdatedAt,
		secret.IsDeleted,
		secret.ID,
		secret.UserID,
		secret.Version,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrVersionConflict
	}

	return nil
}

func (r *txSecretRepository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM secrets WHERE id = $1 AND user_id = $2`
	_, err := r.tx.Exec(ctx, query, id, userID)
	return err
}

func (r *txSecretRepository) SoftDelete(ctx context.Context, id, userID string) error {
	query := `
		UPDATE secrets 
		SET is_deleted = true, updated_at = $1
		WHERE id = $2 AND user_id = $3 AND NOT is_deleted
	`

	result, err := r.tx.Exec(ctx, query, domain.Now(), id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrSecretNotFound
	}

	return nil
}

func (r *txSecretRepository) GetUserSecretsVersion(ctx context.Context, userID string) (int64, error) {
	query := `
		SELECT COALESCE(MAX(version), 0) 
		FROM secrets 
		WHERE user_id = $1 AND NOT is_deleted
	`

	var version int64
	err := r.tx.QueryRow(ctx, query, userID).Scan(&version)
	return version, err
}

func (r *txSecretRepository) GetChangedSecrets(ctx context.Context, userID string, lastSyncVersion int64) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted
		FROM secrets
		WHERE user_id = $1 AND version > $2
		ORDER BY version ASC
	`

	rows, err := r.tx.Query(ctx, query, userID, lastSyncVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []*domain.Secret
	for rows.Next() {
		var secret domain.Secret
		var secretType string

		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secretType,
			&secret.Name,
			&secret.EncryptedData,
			&secret.EncryptedMeta,
			&secret.Version,
			&secret.CreatedAt,
			&secret.UpdatedAt,
			&secret.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		secret.Type = domain.SecretType(secretType)
		secrets = append(secrets, &secret)
	}

	return secrets, nil
}
