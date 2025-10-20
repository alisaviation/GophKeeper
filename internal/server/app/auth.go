package app

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

// AuthService предоставляет методы для аутентификации и авторизации
type AuthService struct {
	users     interfaces.UserRepository
	jwtSecret string
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(users interfaces.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	if err := validateCredentials(login, password); err != nil {
		return "", fmt.Errorf("invalid credentials: %w", err)
	}

	existing, err := s.users.GetByLogin(ctx, login)
	if err == nil && existing != nil {
		return "", interfaces.ErrUserAlreadyExists
	}
	if err != nil && err != interfaces.ErrUserNotFound {
		return "", fmt.Errorf("failed to check user existence: %w", err)
	}

	passwordHash, err := hashPassword(password)
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

	var err error
	if err == interfaces.ErrUserNotFound {
		return "", "", "", interfaces.ErrInvalidCredentials
	}

	user, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		if err == interfaces.ErrUserNotFound {
			return "", "", "", interfaces.ErrInvalidCredentials
		}
		return "", "", "", fmt.Errorf("failed to get user: %w", err)
	}

	if !checkPasswordHash(password, user.PasswordHash) {
		return "", "", "", interfaces.ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Login)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, user.ID, nil
}

// ValidateToken проверяет валидность JWT токена
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}

		user, err := s.users.GetByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshTokens обновляет пару токенов
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	// todo проверка refresh token в базе данных и логика отзыва старых токенов?

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("invalid user id in token")
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Login)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

func (s *AuthService) generateAccessToken(userID, login string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"login":   login,
		"exp":     time.Now().Add(15 * time.Minute).Unix(), // 15 минут
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) generateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 дней
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func validateCredentials(login, password string) error {
	if len(login) < 3 || len(login) > 50 {
		return fmt.Errorf("login must be between 3 and 50 characters")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	for _, char := range login {
		if !isValidLoginChar(char) {
			return fmt.Errorf("login can only contain letters, numbers and underscores")
		}
	}

	return nil
}

func isValidLoginChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_'
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateUserID() string {
	return domain.GenerateID() // Будем использовать UUID
}
