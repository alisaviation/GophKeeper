package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher предоставляет методы для работы с паролями
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher создает новый хешер паролей
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &PasswordHasher{cost: cost}
}

// Hash создает хеш пароля
func (h *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// Check проверяет пароль против хеша
func (h *PasswordHasher) Check(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// DefaultPasswordHasher хешер с настройками по умолчанию
var DefaultPasswordHasher = NewPasswordHasher(bcrypt.DefaultCost)

// HashPassword хеширует пароль с настройками по умолчанию
func HashPassword(password string) (string, error) {
	return DefaultPasswordHasher.Hash(password)
}

// CheckPasswordHash проверяет пароль против хеша
func CheckPasswordHash(password, hash string) bool {
	return DefaultPasswordHasher.Check(password, hash)
}
