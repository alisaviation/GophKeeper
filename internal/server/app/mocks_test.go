package app_test

import (
	"context"
	"fmt"

	"github.com/alisaviation/GophKeeper/internal/crypto"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

type mockUserRepository struct {
	users map[string]*domain.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.users[user.Login]; exists {
		return domain.ErrUserAlreadyExists
	}
	m.users[user.Login] = user
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	user, exists := m.users[login]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return domain.ErrUserNotFound
	}
	m.users[user.ID] = user
	m.users[user.Login] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	user, exists := m.users[id]
	if !exists {
		return domain.ErrUserNotFound
	}
	delete(m.users, user.Login)
	delete(m.users, user.ID)
	return nil
}

type mockSecretRepository struct {
	secrets    map[string]*domain.Secret
	versions   map[string]int64
	deleted    map[string]bool
	userSecret map[string][]string
}

func newMockSecretRepository() *mockSecretRepository {
	return &mockSecretRepository{
		secrets:    make(map[string]*domain.Secret),
		versions:   make(map[string]int64),
		deleted:    make(map[string]bool),
		userSecret: make(map[string][]string),
	}
}

func (m *mockSecretRepository) Create(ctx context.Context, secret *domain.Secret) error {
	if _, exists := m.secrets[secret.ID]; exists {
		return domain.ErrSecretAlreadyExists
	}
	m.secrets[secret.ID] = secret
	m.versions[secret.ID] = secret.Version
	m.deleted[secret.ID] = false

	m.userSecret[secret.UserID] = append(m.userSecret[secret.UserID], secret.ID)
	return nil
}

func (m *mockSecretRepository) GetByID(ctx context.Context, id, userID string) (*domain.Secret, error) {
	secret, exists := m.secrets[id]
	if !exists || m.deleted[id] {
		return nil, domain.ErrSecretNotFound
	}
	if secret.UserID != userID {
		return nil, domain.ErrSecretNotFound
	}
	return secret, nil
}

func (m *mockSecretRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Secret, error) {
	var result []*domain.Secret
	for _, secretID := range m.userSecret[userID] {
		if secret, exists := m.secrets[secretID]; exists && !m.deleted[secretID] {
			result = append(result, secret)
		}
	}
	return result, nil
}

func (m *mockSecretRepository) ListByUserAndType(ctx context.Context, userID string, secretType domain.SecretType) ([]*domain.Secret, error) {
	var result []*domain.Secret
	for _, secretID := range m.userSecret[userID] {
		if secret, exists := m.secrets[secretID]; exists && !m.deleted[secretID] && secret.Type == secretType {
			result = append(result, secret)
		}
	}
	return result, nil
}

func (m *mockSecretRepository) Update(ctx context.Context, secret *domain.Secret) error {
	existing, exists := m.secrets[secret.ID]
	if !exists || m.deleted[secret.ID] {
		return domain.ErrSecretNotFound
	}
	if existing.Version != secret.Version {
		return domain.ErrVersionConflict
	}

	secret.Version = existing.Version + 1
	m.secrets[secret.ID] = secret
	m.versions[secret.ID] = secret.Version
	return nil
}

func (m *mockSecretRepository) Delete(ctx context.Context, id, userID string) error {
	secret, exists := m.secrets[id]
	if !exists || m.deleted[id] {
		return domain.ErrSecretNotFound
	}
	if secret.UserID != userID {
		return domain.ErrSecretNotFound
	}
	m.deleted[id] = true
	return nil
}

func (m *mockSecretRepository) SoftDelete(ctx context.Context, id, userID string) error {
	return m.Delete(ctx, id, userID)
}

func (m *mockSecretRepository) GetUserSecretsVersion(ctx context.Context, userID string) (int64, error) {
	return int64(len(m.userSecret[userID])), nil
}

func (m *mockSecretRepository) GetChangedSecrets(ctx context.Context, userID string, lastSyncVersion int64) ([]*domain.Secret, error) {
	var result []*domain.Secret
	for _, secretID := range m.userSecret[userID] {
		if secret, exists := m.secrets[secretID]; exists && !m.deleted[secretID] {
			result = append(result, secret)
		}
	}
	return result, nil
}

type mockJWTManager struct {
	tokens       map[string]*crypto.TokenClaims
	tokenCounter int
}

func newMockJWTManager() *mockJWTManager {
	return &mockJWTManager{
		tokens:       make(map[string]*crypto.TokenClaims),
		tokenCounter: 0,
	}
}

func (m *mockJWTManager) GenerateAccessToken(userID, login string) (string, error) {
	m.tokenCounter++
	token := fmt.Sprintf("access_%s_%s_%d", userID, login, m.tokenCounter)
	m.tokens[token] = &crypto.TokenClaims{
		UserID: userID,
		Login:  login,
		Type:   "access",
	}
	return token, nil
}

func (m *mockJWTManager) GenerateRefreshToken(userID string) (string, error) {
	m.tokenCounter++
	token := fmt.Sprintf("refresh_%s_%d", userID, m.tokenCounter)
	m.tokens[token] = &crypto.TokenClaims{
		UserID: userID,
		Type:   "refresh",
	}
	return token, nil
}

func (m *mockJWTManager) ValidateToken(tokenString string) (*crypto.TokenClaims, error) {
	claims, exists := m.tokens[tokenString]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (m *mockJWTManager) IsAccessToken(claims *crypto.TokenClaims) bool {
	return claims.Type == "access"
}

func (m *mockJWTManager) IsRefreshToken(claims *crypto.TokenClaims) bool {
	return claims.Type == "refresh"
}

// mockEncryptor мок шифровальщика
type mockEncryptor struct{}

func (m *mockEncryptor) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (m *mockEncryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	return encryptedData, nil
}
