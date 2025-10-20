package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
	"github.com/alisaviation/GophKeeper/internal/server/storage/memory"
)

func TestDataService_CreateAndGetSecret(t *testing.T) {
	storage := memory.NewStorage()
	crypto := &app.NoopEncryptor{}
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
	crypto := &app.NoopEncryptor{}
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
	crypto := &app.NoopEncryptor{}
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
	assert.Equal(t, int64(1), secret.Version) // После создания версия 1

	current, err := dataService.GetSecret(ctx, userID, secret.ID)
	require.NoError(t, err)

	current.Name = "Updated Secret"

	err = dataService.UpdateSecret(ctx, userID, current)
	require.NoError(t, err)

	updated, err := dataService.GetSecret(ctx, userID, secret.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Secret", updated.Name)
	assert.Equal(t, int64(2), updated.Version) // Версия должна увеличиться до 2
}

func TestDataService_UpdateSecret_VersionConflict(t *testing.T) {
	storage := memory.NewStorage()
	crypto := &app.NoopEncryptor{}
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
		Version:       1, // Устаревшая версия (должна быть 2)
		CreatedAt:     secret.CreatedAt,
		UpdatedAt:     secret.UpdatedAt,
		IsDeleted:     secret.IsDeleted,
	}

	err = dataService.UpdateSecret(ctx, userID, outdatedSecret)
	require.Error(t, err)
	assert.Equal(t, interfaces.ErrVersionConflict, err)

	final, err := dataService.GetSecret(ctx, userID, secret.ID)
	require.NoError(t, err)
	assert.Equal(t, "First Update", final.Name) // Осталось первое обновление
	assert.Equal(t, int64(2), final.Version)    // Версия осталась 2
}
