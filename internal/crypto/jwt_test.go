package crypto_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/crypto"
)

func TestJWTManager(t *testing.T) {
	config := &crypto.JWTConfig{
		Secret:        "test-secret-key",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	jwtManager := crypto.NewJWTManager(config)

	userID := "user-123"
	login := "testuser"

	// Генерируем access token
	accessToken, err := jwtManager.GenerateAccessToken(userID, login)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	// Генерируем refresh token
	refreshToken, err := jwtManager.GenerateRefreshToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Проверяем access token
	accessClaims, err := jwtManager.ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, login, accessClaims.Login)
	assert.Equal(t, "access", accessClaims.Type)
	assert.True(t, jwtManager.IsAccessToken(accessClaims))
	assert.False(t, jwtManager.IsRefreshToken(accessClaims))

	// Проверяем refresh token
	refreshClaims, err := jwtManager.ValidateToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, "refresh", refreshClaims.Type)
	assert.True(t, jwtManager.IsRefreshToken(refreshClaims))
	assert.False(t, jwtManager.IsAccessToken(refreshClaims))

	// Проверяем невалидный токен
	_, err = jwtManager.ValidateToken("invalid-token")
	require.Error(t, err)
}
