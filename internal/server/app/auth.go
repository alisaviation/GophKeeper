package app

import (
	"context"
	"fmt"
	"time"

	"github.com/alisaviation/GophKeeper/internal/crypto"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

// AuthService предоставляет методы для аутентификации и авторизации
type AuthService struct {
	users      interfaces.UserRepository
	jwtManager crypto.JWTManagerInterface
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(users interfaces.UserRepository, jwtManager crypto.JWTManagerInterface) *AuthService {
	return &AuthService{
		users:      users,
		jwtManager: jwtManager,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	err := validateCredentials(login, password)
	if err != nil {
		return "", fmt.Errorf("invalid credentials: %w", err)
	}

	existing, err := s.users.GetByLogin(ctx, login)
	if err == nil && existing != nil {
		return "", domain.ErrUserAlreadyExists
	}
	if err != nil && err != domain.ErrUserNotFound {
		return "", fmt.Errorf("failed to check user existence: %w", err)
	}

	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:           generateUserID(),
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return user.ID, nil
}

// Login аутентифицирует пользователя и возвращает JWT токен
func (s *AuthService) Login(ctx context.Context, login, password string) (string, string, string, error) {
	err := validateCredentials(login, password)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid credentials: %w", err)
	}

	user, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return "", "", "", domain.ErrInvalidCredentials
		}
		return "", "", "", fmt.Errorf("failed to get user: %w", err)
	}

	if !crypto.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", "", domain.ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Login)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, user.ID, nil
}

// ValidateToken проверяет валидность JWT токена
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !s.jwtManager.IsAccessToken(claims) {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

// RefreshTokens обновляет пару токенов
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	if !s.jwtManager.IsRefreshToken(claims) {
		return "", "", domain.ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", "", domain.ErrUserNotFound
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Login)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}
