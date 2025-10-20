package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
	"github.com/alisaviation/GophKeeper/internal/server/storage/memory"
)

func TestAuthService_Register(t *testing.T) {
	storage := memory.NewStorage()
	authService := app.NewAuthService(storage.UserRepository(), "test-secret")
	ctx := context.Background()

	userID, err := authService.Register(ctx, "testuser", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	_, err = authService.Register(ctx, "testuser", "password456")
	require.Error(t, err)
	assert.Equal(t, interfaces.ErrUserAlreadyExists, err)

	_, err = authService.Register(ctx, "ab", "password123")
	require.Error(t, err)

	_, err = authService.Register(ctx, "newuser", "short")
	require.Error(t, err)
}

func TestAuthService_Login(t *testing.T) {
	storage := memory.NewStorage()
	authService := app.NewAuthService(storage.UserRepository(), "test-secret")
	ctx := context.Background()

	userID, err := authService.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	accessToken, refreshToken, returnedUserID, err := authService.Login(ctx, "testuser", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, userID, returnedUserID)

	_, _, _, err = authService.Login(ctx, "testuser", "wrongpassword")
	require.Error(t, err)
	assert.Equal(t, interfaces.ErrInvalidCredentials, err)

	_, _, _, err = authService.Login(ctx, "nonexistent", "password123")
	require.Error(t, err)
	assert.Equal(t, interfaces.ErrInvalidCredentials, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	storage := memory.NewStorage()
	authService := app.NewAuthService(storage.UserRepository(), "test-secret")
	ctx := context.Background()

	_, err := authService.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	accessToken, _, _, err := authService.Login(ctx, "testuser", "password123")
	require.NoError(t, err)

	user, err := authService.ValidateToken(ctx, accessToken)
	require.NoError(t, err)
	assert.Equal(t, "testuser", user.Login)

	_, err = authService.ValidateToken(ctx, "invalid-token")
	require.Error(t, err)
}
