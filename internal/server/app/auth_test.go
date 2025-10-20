package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

func TestAuthService_Register(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := newMockJWTManager()
	authService := app.NewAuthService(userRepo, jwtManager)
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		userID, err := authService.Register(ctx, "testuser", "password123")
		require.NoError(t, err)
		assert.NotEmpty(t, userID)

		user, err := userRepo.GetByLogin(ctx, "testuser")
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Login)
	})

	t.Run("duplicate registration", func(t *testing.T) {
		_, err := authService.Register(ctx, "testuser", "password456")
		require.Error(t, err)
		assert.Equal(t, domain.ErrUserAlreadyExists, err)
	})

	t.Run("invalid login too short", func(t *testing.T) {
		_, err := authService.Register(ctx, "ab", "password123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login: must be between 3 and 50 characters")
	})

	t.Run("invalid login too long", func(t *testing.T) {
		longLogin := make([]rune, 51)
		for i := range longLogin {
			longLogin[i] = 'a'
		}
		_, err := authService.Register(ctx, string(longLogin), "password123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login: must be between 3 and 50 characters")
	})

	t.Run("invalid password", func(t *testing.T) {
		_, err := authService.Register(ctx, "newuser", "short")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password: must be at least 8 characters")
	})

	t.Run("invalid login characters", func(t *testing.T) {
		_, err := authService.Register(ctx, "user@name", "password123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login: can only contain letters, numbers and underscores")
	})
}

func TestAuthService_Login(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := newMockJWTManager()
	authService := app.NewAuthService(userRepo, jwtManager)
	ctx := context.Background()

	userID, err := authService.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	t.Run("successful login", func(t *testing.T) {
		accessToken, refreshToken, returnedUserID, err := authService.Login(ctx, "testuser", "password123")
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, userID, returnedUserID)

		claims, err := jwtManager.ValidateToken(accessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, "testuser", claims.Login)
		assert.True(t, jwtManager.IsAccessToken(claims))

		refreshClaims, err := jwtManager.ValidateToken(refreshToken)
		require.NoError(t, err)
		assert.Equal(t, userID, refreshClaims.UserID)
		assert.True(t, jwtManager.IsRefreshToken(refreshClaims))
	})

	t.Run("wrong password", func(t *testing.T) {
		_, _, _, err = authService.Login(ctx, "testuser", "wrongpassword")
		require.Error(t, err)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
	})

	t.Run("nonexistent user", func(t *testing.T) {
		_, _, _, err = authService.Login(ctx, "nonexistent", "password123")
		require.Error(t, err)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
	})

	t.Run("invalid credentials format", func(t *testing.T) {
		_, _, _, err = authService.Login(ctx, "ab", "password123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := newMockJWTManager()
	authService := app.NewAuthService(userRepo, jwtManager)
	ctx := context.Background()

	_, err := authService.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	accessToken, _, _, err := authService.Login(ctx, "testuser", "password123")
	require.NoError(t, err)

	t.Run("successful token validation", func(t *testing.T) {
		user, err := authService.ValidateToken(ctx, accessToken)
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Login)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err = authService.ValidateToken(ctx, "invalid-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate token")
	})

	t.Run("user not found during validation", func(t *testing.T) {
		accessToken, err := jwtManager.GenerateAccessToken("non-existent-user", "testuser")
		require.NoError(t, err)

		_, err = authService.ValidateToken(ctx, accessToken)
		require.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestAuthService_RefreshTokens(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := newMockJWTManager()
	authService := app.NewAuthService(userRepo, jwtManager)
	ctx := context.Background()

	_, err := authService.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	_, refreshToken, _, err := authService.Login(ctx, "testuser", "password123")
	require.NoError(t, err)

	t.Run("successful token refresh", func(t *testing.T) {
		newAccessToken, newRefreshToken, err := authService.RefreshTokens(ctx, refreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, refreshToken, newRefreshToken)

		user, err := authService.ValidateToken(ctx, newAccessToken)
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Login)

		claims, err := jwtManager.ValidateToken(newRefreshToken)
		require.NoError(t, err)
		assert.True(t, jwtManager.IsRefreshToken(claims))
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		_, _, err = authService.RefreshTokens(ctx, "invalid-refresh-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid refresh token")
	})

	t.Run("user not found during refresh", func(t *testing.T) {
		refreshToken, err := jwtManager.GenerateRefreshToken("non-existent-user")
		require.NoError(t, err)

		_, _, err = authService.RefreshTokens(ctx, refreshToken)
		require.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})
}

func TestAuthService_ValidateToken_WrongTokenType(t *testing.T) {
	userRepo := newMockUserRepository()
	jwtManager := newMockJWTManager()
	authService := app.NewAuthService(userRepo, jwtManager)
	ctx := context.Background()

	_, err := authService.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	_, refreshToken, _, err := authService.Login(ctx, "testuser", "password123")
	require.NoError(t, err)

	t.Run("refresh token as access token", func(t *testing.T) {
		_, err = authService.ValidateToken(ctx, refreshToken)
		require.Error(t, err)
		assert.Equal(t, domain.ErrInvalidToken, err)
	})

	t.Run("access token as refresh token", func(t *testing.T) {
		accessToken, _, _, err := authService.Login(ctx, "testuser", "password123")
		require.NoError(t, err)

		_, _, err = authService.RefreshTokens(ctx, accessToken)
		require.Error(t, err)
		assert.Equal(t, domain.ErrInvalidToken, err)
	})
}
