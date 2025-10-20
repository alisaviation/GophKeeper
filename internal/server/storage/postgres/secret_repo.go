package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

type secretRepository struct {
	db *pgxpool.Pool
}

// NewSecretRepository создает новый экземпляр SecretRepository для PostgreSQL
func NewSecretRepository(db *pgxpool.Pool) interfaces.SecretRepository {
	return &secretRepository{db: db}
}

// Create создает новый секрет
func (r *secretRepository) Create(ctx context.Context, secret *domain.Secret) error {
	query := `
		INSERT INTO secrets (id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
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

	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// GetByID возвращает секрет по ID и UserID
func (r *secretRepository) GetByID(ctx context.Context, id, userID string) (*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted
		FROM secrets
		WHERE id = $1 AND user_id = $2 AND NOT is_deleted
	`

	var secret domain.Secret
	var secretType string

	err := r.db.QueryRow(ctx, query, id, userID).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, interfaces.ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to get secret by id: %w", err)
	}

	secret.Type = domain.SecretType(secretType)
	return &secret, nil
}

// ListByUser возвращает все секреты пользователя
func (r *secretRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted
		FROM secrets
		WHERE user_id = $1 AND NOT is_deleted
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user secrets: %w", err)
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
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}

		secret.Type = domain.SecretType(secretType)
		secrets = append(secrets, &secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secrets: %w", err)
	}

	return secrets, nil
}

// ListByUserAndType возвращает секреты пользователя определенного типа
func (r *secretRepository) ListByUserAndType(ctx context.Context, userID string, secretType domain.SecretType) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted
		FROM secrets
		WHERE user_id = $1 AND type = $2 AND NOT is_deleted
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, string(secretType))
	if err != nil {
		return nil, fmt.Errorf("failed to list user secrets by type: %w", err)
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
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}

		secret.Type = domain.SecretType(st)
		secrets = append(secrets, &secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secrets: %w", err)
	}

	return secrets, nil
}

// Update обновляет секрет с проверкой версии
func (r *secretRepository) Update(ctx context.Context, secret *domain.Secret) error {
	query := `
		UPDATE secrets 
		SET type = $1, name = $2, encrypted_data = $3, encrypted_meta = $4, 
		    version = version + 1, updated_at = $5, is_deleted = $6
		WHERE id = $7 AND user_id = $8 AND version = $9
	`

	result, err := r.db.Exec(ctx, query,
		string(secret.Type),
		secret.Name,
		secret.EncryptedData,
		secret.EncryptedMeta,
		time.Now(),
		secret.IsDeleted,
		secret.ID,
		secret.UserID,
		secret.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		existing, err := r.GetByID(ctx, secret.ID, secret.UserID)
		if err != nil {
			return interfaces.ErrSecretNotFound
		}
		if existing.Version != secret.Version {
			return interfaces.ErrVersionConflict
		}
		return interfaces.ErrSecretNotFound
	}

	return nil
}

// Delete удаляет секрет
func (r *secretRepository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM secrets WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return interfaces.ErrSecretNotFound
	}

	return nil
}

// SoftDelete помечает секрет как удаленный
func (r *secretRepository) SoftDelete(ctx context.Context, id, userID string) error {
	query := `
		UPDATE secrets 
		SET is_deleted = true, updated_at = $1
		WHERE id = $2 AND user_id = $3 AND NOT is_deleted
	`

	result, err := r.db.Exec(ctx, query, time.Now(), id, userID)
	if err != nil {
		return fmt.Errorf("failed to soft delete secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return interfaces.ErrSecretNotFound
	}

	return nil
}

// GetUserSecretsVersion возвращает текущую версию секретов пользователя
func (r *secretRepository) GetUserSecretsVersion(ctx context.Context, userID string) (int64, error) {
	query := `
		SELECT current_version 
		FROM user_secrets_version 
		WHERE user_id = $1
	`

	var version int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil // Пользователь без секретов
		}
		return 0, fmt.Errorf("failed to get user secrets version: %w", err)
	}

	return version, nil
}

// GetChangedSecrets возвращает секреты, измененные после указанной версии
func (r *secretRepository) GetChangedSecrets(ctx context.Context, userID string, lastSyncVersion int64) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, encrypted_meta, version, created_at, updated_at, is_deleted
		FROM secrets
		WHERE user_id = $1 AND version > $2
		ORDER BY version ASC
	`

	rows, err := r.db.Query(ctx, query, userID, lastSyncVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed secrets: %w", err)
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
			return nil, fmt.Errorf("failed to scan changed secret: %w", err)
		}

		secret.Type = domain.SecretType(secretType)
		secrets = append(secrets, &secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating changed secrets: %w", err)
	}

	return secrets, nil
}
