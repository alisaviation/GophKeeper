package crypto

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

type JWTManagerInterface interface {
	GenerateAccessToken(userID, login string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	IsAccessToken(claims *TokenClaims) bool
	IsRefreshToken(claims *TokenClaims) bool
}

// JWTManager управляет JWT токенами
type JWTManager struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// JWTConfig конфигурация для JWT менеджера
type JWTConfig struct {
	Secret        string        `yaml:"secret" env:"JWT_SECRET"`
	AccessExpiry  time.Duration `yaml:"access_expiry" env:"JWT_ACCESS_EXPIRY" default:"15m"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry" env:"JWT_REFRESH_EXPIRY" default:"168h"` // 7 дней
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *JWTConfig {
	return &JWTConfig{
		Secret:        "default-jwt-secret-key-change-in-production",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
}

// NewJWTManager создает новый JWT менеджер
func NewJWTManager(config *JWTConfig) *JWTManager {
	if config == nil {
		config = DefaultConfig()
	}

	return &JWTManager{
		secret:        []byte(config.Secret),
		accessExpiry:  config.AccessExpiry,
		refreshExpiry: config.RefreshExpiry,
	}
}

// TokenClaims кастомные claims для JWT токенов
type TokenClaims struct {
	UserID string `json:"user_id"`
	Login  string `json:"login,omitempty"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

// GenerateAccessToken генерирует access token
func (m *JWTManager) GenerateAccessToken(userID, login string) (string, error) {
	claims := &TokenClaims{
		UserID: userID,
		Login:  login,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gophkeeper",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, nil
}

// GenerateRefreshToken генерирует refresh token
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := &TokenClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gophkeeper",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken проверяет и парсит JWT токен
func (m *JWTManager) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

// IsAccessToken проверяет, является ли токен access token'ом
func (m *JWTManager) IsAccessToken(claims *TokenClaims) bool {
	return claims.Type == "access"
}

// IsRefreshToken проверяет, является ли токен refresh token'ом
func (m *JWTManager) IsRefreshToken(claims *TokenClaims) bool {
	return claims.Type == "refresh"
}
