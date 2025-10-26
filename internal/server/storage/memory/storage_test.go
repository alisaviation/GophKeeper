package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/memory"
)

func TestMemoryStorage_UserCRUD(t *testing.T) {
	storage := memory.NewStorage()
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New().String(),
		Login:        "testuser",
		PasswordHash: "hash123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := storage.UserRepository().Create(ctx, user)
	require.NoError(t, err)

	retrieved, err := storage.UserRepository().GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Login, retrieved.Login)

	retrievedByLogin, err := storage.UserRepository().GetByLogin(ctx, user.Login)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrievedByLogin.ID)

	user.PasswordHash = "newhash"
	err = storage.UserRepository().Update(ctx, user)
	require.NoError(t, err)

	updated, err := storage.UserRepository().GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "newhash", updated.PasswordHash)

	err = storage.UserRepository().Delete(ctx, user.ID)
	require.NoError(t, err)

	_, err = storage.UserRepository().GetByID(ctx, user.ID)
	require.Error(t, err)
}

func TestMemoryStorage_SecretCRUD(t *testing.T) {
	storage := memory.NewStorage()
	ctx := context.Background()

	userID := uuid.New().String()
	secret := &domain.Secret{
		ID:            uuid.New().String(),
		UserID:        userID,
		Type:          domain.LoginPassword,
		Name:          "Test Secret",
		EncryptedData: []byte("encrypted data"),
		EncryptedMeta: []byte("encrypted meta"),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
	}

	// Test Create
	err := storage.SecretRepository().Create(ctx, secret)
	require.NoError(t, err)

	// Test GetByID
	retrieved, err := storage.SecretRepository().GetByID(ctx, secret.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, secret.ID, retrieved.ID)
	assert.Equal(t, secret.Name, retrieved.Name)

	// Test ListByUser
	secrets, err := storage.SecretRepository().ListByUser(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, secret.ID, secrets[0].ID)

	// Test Update
	secret.Name = "Updated Secret"
	err = storage.SecretRepository().Update(ctx, secret)
	require.NoError(t, err)

	updated, err := storage.SecretRepository().GetByID(ctx, secret.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Secret", updated.Name)
	assert.Equal(t, int64(2), updated.Version)

	// Test GetUserSecretsVersion ДО soft delete
	version, err := storage.SecretRepository().GetUserSecretsVersion(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), version) // После update версия должна быть 2

	// Test SoftDelete
	err = storage.SecretRepository().SoftDelete(ctx, secret.ID, userID)
	require.NoError(t, err)

	// После soft delete секрет не должен быть доступен через обычные методы
	_, err = storage.SecretRepository().GetByID(ctx, secret.ID, userID)
	require.Error(t, err)

	secretsAfterDelete, err := storage.SecretRepository().ListByUser(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, secretsAfterDelete, 0)

	// Но GetChangedSecrets должен вернуть soft deleted секрет
	changedSecrets, err := storage.SecretRepository().GetChangedSecrets(ctx, userID, 0)
	require.NoError(t, err)
	require.Len(t, changedSecrets, 1)
	assert.Equal(t, secret.ID, changedSecrets[0].ID)
	assert.True(t, changedSecrets[0].IsDeleted) // Должен быть помечен как удаленный
}

func TestMemoryStorage_GetChangedSecrets(t *testing.T) {
	storage := memory.NewStorage()
	ctx := context.Background()

	userID := uuid.New().String()

	secret1 := &domain.Secret{
		ID:            uuid.New().String(),
		UserID:        userID,
		Type:          domain.LoginPassword,
		Name:          "Secret 1",
		EncryptedData: []byte("data1"),
		EncryptedMeta: []byte("meta1"),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
	}

	err := storage.SecretRepository().Create(ctx, secret1)
	require.NoError(t, err)

	changedSecrets, err := storage.SecretRepository().GetChangedSecrets(ctx, userID, 0)
	require.NoError(t, err)
	assert.Len(t, changedSecrets, 1)
	assert.Equal(t, secret1.ID, changedSecrets[0].ID)

	secret1.Name = "Updated Secret 1"
	err = storage.SecretRepository().Update(ctx, secret1)
	require.NoError(t, err)

	secret2 := &domain.Secret{
		ID:            uuid.New().String(),
		UserID:        userID,
		Type:          domain.TextData,
		Name:          "Secret 2",
		EncryptedData: []byte("data2"),
		EncryptedMeta: []byte("meta2"),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
	}

	err = storage.SecretRepository().Create(ctx, secret2)
	require.NoError(t, err)

	changedSecretsAfter, err := storage.SecretRepository().GetChangedSecrets(ctx, userID, 1)
	require.NoError(t, err)
	assert.Len(t, changedSecretsAfter, 2)

	foundSecret1 := false
	foundSecret2 := false
	for _, s := range changedSecretsAfter {
		if s.ID == secret1.ID {
			foundSecret1 = true
			assert.Equal(t, int64(2), s.Version)
		}
		if s.ID == secret2.ID {
			foundSecret2 = true
			assert.Equal(t, int64(1), s.Version)
		}
	}
	assert.True(t, foundSecret1, "Secret 1 should be in changed secrets")
	assert.True(t, foundSecret2, "Secret 2 should be in changed secrets")

	err = storage.SecretRepository().SoftDelete(ctx, secret1.ID, userID)
	require.NoError(t, err)

	changedSecretsAfterDelete, err := storage.SecretRepository().GetChangedSecrets(ctx, userID, 2)
	require.NoError(t, err)
	assert.Len(t, changedSecretsAfterDelete, 1)
	assert.Equal(t, secret1.ID, changedSecretsAfterDelete[0].ID)
	assert.True(t, changedSecretsAfterDelete[0].IsDeleted)
	assert.Equal(t, int64(3), changedSecretsAfterDelete[0].Version)
}

func TestMemoryStorage_ListByUserAndType(t *testing.T) {
	storage := memory.NewStorage()
	ctx := context.Background()

	userID := uuid.New().String()

	loginSecret := &domain.Secret{
		ID:            uuid.New().String(),
		UserID:        userID,
		Type:          domain.LoginPassword,
		Name:          "Login Secret",
		EncryptedData: []byte("data"),
		EncryptedMeta: []byte("meta"),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
	}

	textSecret := &domain.Secret{
		ID:            uuid.New().String(),
		UserID:        userID,
		Type:          domain.TextData,
		Name:          "Text Secret",
		EncryptedData: []byte("data"),
		EncryptedMeta: []byte("meta"),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
	}

	err := storage.SecretRepository().Create(ctx, loginSecret)
	require.NoError(t, err)

	err = storage.SecretRepository().Create(ctx, textSecret)
	require.NoError(t, err)

	loginSecrets, err := storage.SecretRepository().ListByUserAndType(ctx, userID, domain.LoginPassword)
	require.NoError(t, err)
	assert.Len(t, loginSecrets, 1)
	assert.Equal(t, loginSecret.ID, loginSecrets[0].ID)

	textSecrets, err := storage.SecretRepository().ListByUserAndType(ctx, userID, domain.TextData)
	require.NoError(t, err)
	assert.Len(t, textSecrets, 1)
	assert.Equal(t, textSecret.ID, textSecrets[0].ID)

	allSecrets, err := storage.SecretRepository().ListByUser(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, allSecrets, 2)
}

func TestMemoryStorage_Transaction(t *testing.T) {
	storage := memory.NewStorage()
	ctx := context.Background()

	tx, err := storage.TransactionManager().BeginTx(ctx)
	require.NoError(t, err)
	require.NotNil(t, tx)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	err = tx.Rollback(ctx)
	require.NoError(t, err)
}

func TestMemoryStorage_SoftDeleteAndVersion(t *testing.T) {
	storage := memory.NewStorage()
	ctx := context.Background()

	userID := uuid.New().String()
	secret := &domain.Secret{
		ID:            uuid.New().String(),
		UserID:        userID,
		Type:          domain.LoginPassword,
		Name:          "Test Secret",
		EncryptedData: []byte("encrypted data"),
		EncryptedMeta: []byte("encrypted meta"),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
	}

	err := storage.SecretRepository().Create(ctx, secret)
	require.NoError(t, err)

	versionBefore, err := storage.SecretRepository().GetUserSecretsVersion(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), versionBefore)

	err = storage.SecretRepository().SoftDelete(ctx, secret.ID, userID)
	require.NoError(t, err)

	_, err = storage.SecretRepository().GetByID(ctx, secret.ID, userID)
	require.Error(t, err)

	secrets, err := storage.SecretRepository().ListByUser(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, secrets, 0)

	changedSecrets, err := storage.SecretRepository().GetChangedSecrets(ctx, userID, 0)
	require.NoError(t, err)
	require.Len(t, changedSecrets, 1)
	assert.Equal(t, secret.ID, changedSecrets[0].ID)
	assert.True(t, changedSecrets[0].IsDeleted)
}

func TestMemoryStorage_GetUserSecretsVersion(t *testing.T) {
	storage := memory.NewStorage()
	ctx := context.Background()

	userID := uuid.New().String()

	secret1 := &domain.Secret{
		ID:            uuid.New().String(),
		UserID:        userID,
		Type:          domain.LoginPassword,
		Name:          "Secret 1",
		EncryptedData: []byte("data1"),
		EncryptedMeta: []byte("meta1"),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
	}

	secret2 := &domain.Secret{
		ID:            uuid.New().String(),
		UserID:        userID,
		Type:          domain.TextData,
		Name:          "Secret 2",
		EncryptedData: []byte("data2"),
		EncryptedMeta: []byte("meta2"),
		Version:       3, // Более высокая версия
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
	}

	err := storage.SecretRepository().Create(ctx, secret1)
	require.NoError(t, err)

	err = storage.SecretRepository().Create(ctx, secret2)
	require.NoError(t, err)

	version, err := storage.SecretRepository().GetUserSecretsVersion(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), version)

	err = storage.SecretRepository().SoftDelete(ctx, secret1.ID, userID)
	require.NoError(t, err)

	err = storage.SecretRepository().SoftDelete(ctx, secret2.ID, userID)
	require.NoError(t, err)

	versionAfterDelete, err := storage.SecretRepository().GetUserSecretsVersion(ctx, userID)
	require.NoError(t, err)

	assert.True(t, versionAfterDelete >= 0)
}
