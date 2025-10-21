package app

import (
	"context"
	"fmt"

	"github.com/alisaviation/GophKeeper/internal/crypto"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

type MockUserRepository struct {
	Users map[string]*domain.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.Users[user.Login]; exists {
		return domain.ErrUserAlreadyExists
	}
	m.Users[user.Login] = user
	m.Users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, exists := m.Users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	user, exists := m.Users[login]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if _, exists := m.Users[user.ID]; !exists {
		return domain.ErrUserNotFound
	}
	m.Users[user.ID] = user
	m.Users[user.Login] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	user, exists := m.Users[id]
	if !exists {
		return domain.ErrUserNotFound
	}
	delete(m.Users, user.Login)
	delete(m.Users, user.ID)
	return nil
}

type MockSecretRepository struct {
	Secrets    map[string]*domain.Secret
	Versions   map[string]int64
	Deleted    map[string]bool
	UserSecret map[string][]string
}

func NewMockSecretRepository() *MockSecretRepository {
	return &MockSecretRepository{
		Secrets:    make(map[string]*domain.Secret),
		Versions:   make(map[string]int64),
		Deleted:    make(map[string]bool),
		UserSecret: make(map[string][]string),
	}
}

func (m *MockSecretRepository) Create(ctx context.Context, secret *domain.Secret) error {
	if _, exists := m.Secrets[secret.ID]; exists {
		return domain.ErrSecretAlreadyExists
	}
	m.Secrets[secret.ID] = secret
	m.Versions[secret.ID] = secret.Version
	m.Deleted[secret.ID] = false

	m.UserSecret[secret.UserID] = append(m.UserSecret[secret.UserID], secret.ID)
	return nil
}

func (m *MockSecretRepository) GetByID(ctx context.Context, id, userID string) (*domain.Secret, error) {
	secret, exists := m.Secrets[id]
	if !exists || m.Deleted[id] {
		return nil, domain.ErrSecretNotFound
	}
	if secret.UserID != userID {
		return nil, domain.ErrSecretNotFound
	}
	return secret, nil
}

func (m *MockSecretRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Secret, error) {
	var result []*domain.Secret
	for _, secretID := range m.UserSecret[userID] {
		if secret, exists := m.Secrets[secretID]; exists && !m.Deleted[secretID] {
			result = append(result, secret)
		}
	}
	return result, nil
}

func (m *MockSecretRepository) ListByUserAndType(ctx context.Context, userID string, secretType domain.SecretType) ([]*domain.Secret, error) {
	var result []*domain.Secret
	for _, secretID := range m.UserSecret[userID] {
		if secret, exists := m.Secrets[secretID]; exists && !m.Deleted[secretID] && secret.Type == secretType {
			result = append(result, secret)
		}
	}
	return result, nil
}

func (m *MockSecretRepository) Update(ctx context.Context, secret *domain.Secret) error {
	existing, exists := m.Secrets[secret.ID]
	if !exists || m.Deleted[secret.ID] {
		return domain.ErrSecretNotFound
	}
	if existing.Version != secret.Version {
		return domain.ErrVersionConflict
	}

	secret.Version = existing.Version + 1
	m.Secrets[secret.ID] = secret
	m.Versions[secret.ID] = secret.Version
	return nil
}

func (m *MockSecretRepository) Delete(ctx context.Context, id, userID string) error {
	secret, exists := m.Secrets[id]
	if !exists || m.Deleted[id] {
		return domain.ErrSecretNotFound
	}
	if secret.UserID != userID {
		return domain.ErrSecretNotFound
	}
	m.Deleted[id] = true
	return nil
}

func (m *MockSecretRepository) SoftDelete(ctx context.Context, id, userID string) error {
	return m.Delete(ctx, id, userID)
}

func (m *MockSecretRepository) GetUserSecretsVersion(ctx context.Context, userID string) (int64, error) {
	return int64(len(m.UserSecret[userID])), nil
}

func (m *MockSecretRepository) GetChangedSecrets(ctx context.Context, userID string, lastSyncVersion int64) ([]*domain.Secret, error) {
	var result []*domain.Secret
	for _, secretID := range m.UserSecret[userID] {
		if secret, exists := m.Secrets[secretID]; exists && !m.Deleted[secretID] {
			result = append(result, secret)
		}
	}
	return result, nil
}

type MockJWTManager struct {
	Tokens       map[string]*crypto.TokenClaims
	TokenCounter int
}

func NewMockJWTManager() *MockJWTManager {
	return &MockJWTManager{
		Tokens:       make(map[string]*crypto.TokenClaims),
		TokenCounter: 0,
	}
}

func (m *MockJWTManager) GenerateAccessToken(userID, login string) (string, error) {
	m.TokenCounter++
	token := fmt.Sprintf("access_%s_%s_%d", userID, login, m.TokenCounter)
	m.Tokens[token] = &crypto.TokenClaims{
		UserID: userID,
		Login:  login,
		Type:   "access",
	}
	return token, nil
}

func (m *MockJWTManager) GenerateRefreshToken(userID string) (string, error) {
	m.TokenCounter++
	token := fmt.Sprintf("refresh_%s_%d", userID, m.TokenCounter)
	m.Tokens[token] = &crypto.TokenClaims{
		UserID: userID,
		Type:   "refresh",
	}
	return token, nil
}

func (m *MockJWTManager) ValidateToken(tokenString string) (*crypto.TokenClaims, error) {
	claims, exists := m.Tokens[tokenString]
	if !exists {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

func (m *MockJWTManager) IsAccessToken(claims *crypto.TokenClaims) bool {
	return claims.Type == "access"
}

func (m *MockJWTManager) IsRefreshToken(claims *crypto.TokenClaims) bool {
	return claims.Type == "refresh"
}

// MockEncryptor мок шифровальщика
type MockEncryptor struct{}

func (m *MockEncryptor) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (m *MockEncryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	return encryptedData, nil
}
