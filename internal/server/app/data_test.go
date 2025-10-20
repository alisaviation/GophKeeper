package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/crypto"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/memory"
)

func TestDataService_CreateAndGetSecret(t *testing.T) {
	storage := memory.NewStorage()
	crypto := &crypto.NoopEncryptor{}
	dataService := app.NewDataService(storage.SecretRepository(), crypto)
	ctx := context.Background()

	userID := domain.GenerateID()
	secret := &domain.Secret{
		Type:          domain.LoginPassword,
		Name:          "Test Secret",
		EncryptedData: []byte("encrypted data"),
		EncryptedMeta: []byte("encrypted meta"),
	}

	err := dataService.CreateSecret(ctx, userID, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, secret.ID)

	retrieved, err := dataService.GetSecret(ctx, userID, secret.ID)
	require.NoError(t, err)
	assert.Equal(t, secret.ID, retrieved.ID)
	assert.Equal(t, secret.Name, retrieved.Name)
}

func TestDataService_Sync(t *testing.T) {
	storage := memory.NewStorage()
	crypto := &crypto.NoopEncryptor{}
	dataService := app.NewDataService(storage.SecretRepository(), crypto)
	ctx := context.Background()

	userID := domain.GenerateID()

	serverSecret := &domain.Secret{
		Type:          domain.LoginPassword,
		Name:          "Server Secret",
		EncryptedData: []byte("server data"),
		EncryptedMeta: []byte("server meta"),
	}
	err := dataService.CreateSecret(ctx, userID, serverSecret)
	require.NoError(t, err)

	result, err := dataService.Sync(ctx, userID, []*domain.Secret{}, 0)
	require.NoError(t, err)
	assert.Len(t, result.ServerSecrets, 1)
	assert.Equal(t, serverSecret.ID, result.ServerSecrets[0].ID)
}

func TestDataService_UpdateSecret(t *testing.T) {
	storage := memory.NewStorage()
	crypto := &crypto.NoopEncryptor{}
	dataService := app.NewDataService(storage.SecretRepository(), crypto)
	ctx := context.Background()

	userID := domain.GenerateID()

	secret := &domain.Secret{
		Type:          domain.LoginPassword,
		Name:          "Test Secret",
		EncryptedData: []byte("encrypted data"),
		EncryptedMeta: []byte("encrypted meta"),
	}

	err := dataService.CreateSecret(ctx, userID, secret)
	require.NoError(t, err)
	assert.Equal(t, int64(1), secret.Version)

	current, err := dataService.GetSecret(ctx, userID, secret.ID)
	require.NoError(t, err)

	current.Name = "Updated Secret"

	err = dataService.UpdateSecret(ctx, userID, current)
	require.NoError(t, err)

	updated, err := dataService.GetSecret(ctx, userID, secret.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Secret", updated.Name)
	assert.Equal(t, int64(2), updated.Version)
}

func TestDataService_UpdateSecret_VersionConflict(t *testing.T) {
	storage := memory.NewStorage()
	crypto := &crypto.NoopEncryptor{}
	dataService := app.NewDataService(storage.SecretRepository(), crypto)
	ctx := context.Background()

	userID := domain.GenerateID()

	secret := &domain.Secret{
		Type:          domain.LoginPassword,
		Name:          "Test Secret",
		EncryptedData: []byte("encrypted data"),
		EncryptedMeta: []byte("encrypted meta"),
	}

	err := dataService.CreateSecret(ctx, userID, secret)
	require.NoError(t, err)

	current, err := dataService.GetSecret(ctx, userID, secret.ID)
	require.NoError(t, err)

	current.Name = "First Update"
	err = dataService.UpdateSecret(ctx, userID, current)
	require.NoError(t, err)

	outdatedSecret := &domain.Secret{
		ID:            secret.ID,
		UserID:        userID,
		Type:          secret.Type,
		Name:          "Outdated Update",
		EncryptedData: secret.EncryptedData,
		EncryptedMeta: secret.EncryptedMeta,
		Version:       1,
		CreatedAt:     secret.CreatedAt,
		UpdatedAt:     secret.UpdatedAt,
		IsDeleted:     secret.IsDeleted,
	}

	err = dataService.UpdateSecret(ctx, userID, outdatedSecret)
	require.Error(t, err)
	assert.Equal(t, domain.ErrVersionConflict, err)

	final, err := dataService.GetSecret(ctx, userID, secret.ID)
	require.NoError(t, err)
	assert.Equal(t, "First Update", final.Name)
	assert.Equal(t, int64(2), final.Version)
}

func TestDataService_DeleteSecret(t *testing.T) {
	storage := memory.NewStorage()
	encryptor := &crypto.NoopEncryptor{}
	dataService := app.NewDataService(storage.SecretRepository(), encryptor)
	ctx := context.Background()

	userID := domain.GenerateID()
	secret := &domain.Secret{
		Type:          domain.LoginPassword,
		Name:          "Test Secret",
		EncryptedData: []byte("encrypted data"),
		EncryptedMeta: []byte("encrypted meta"),
	}

	err := dataService.CreateSecret(ctx, userID, secret)
	require.NoError(t, err)

	err = dataService.DeleteSecret(ctx, userID, secret.ID)
	require.NoError(t, err)

	_, err = dataService.GetSecret(ctx, userID, secret.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrSecretNotFound)
}

func TestDataService_ListSecrets(t *testing.T) {
	storage := memory.NewStorage()
	encryptor := &crypto.NoopEncryptor{}
	dataService := app.NewDataService(storage.SecretRepository(), encryptor)
	ctx := context.Background()

	userID := domain.GenerateID()

	secret1 := &domain.Secret{
		Type:          domain.LoginPassword,
		Name:          "Secret 1",
		EncryptedData: []byte("data1"),
		EncryptedMeta: []byte("meta1"),
	}

	secret2 := &domain.Secret{
		Type:          domain.TextData,
		Name:          "Secret 2",
		EncryptedData: []byte("data2"),
		EncryptedMeta: []byte("meta2"),
	}

	err := dataService.CreateSecret(ctx, userID, secret1)
	require.NoError(t, err)

	err = dataService.CreateSecret(ctx, userID, secret2)
	require.NoError(t, err)

	secrets, err := dataService.ListSecrets(ctx, userID, nil)
	require.NoError(t, err)
	assert.Len(t, secrets, 2)

	loginPasswordType := domain.LoginPassword
	loginSecrets, err := dataService.ListSecrets(ctx, userID, &loginPasswordType)
	require.NoError(t, err)
	assert.Len(t, loginSecrets, 1)
	assert.Equal(t, secret1.ID, loginSecrets[0].ID)

	textDataType := domain.TextData
	textSecrets, err := dataService.ListSecrets(ctx, userID, &textDataType)
	require.NoError(t, err)
	assert.Len(t, textSecrets, 1)
	assert.Equal(t, secret2.ID, textSecrets[0].ID)
}
