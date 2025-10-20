package memory

import (
	"context"
	"sync"
	"time"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

// memoryStorage реализует Storage в памяти
type memoryStorage struct {
	mu         sync.RWMutex
	users      map[string]*domain.User
	secrets    map[string]*domain.Secret // key: userID_secretID
	userRepo   *memoryUserRepository
	secretRepo *memorySecretRepository
}

// memoryUserRepository реализует UserRepository для памяти
type memoryUserRepository struct {
	storage *memoryStorage
}

// memorySecretRepository реализует SecretRepository для памяти
type memorySecretRepository struct {
	storage *memoryStorage
}

// NewStorage создает новый in-memory Storage
func NewStorage() interfaces.Storage {
	s := &memoryStorage{
		users:   make(map[string]*domain.User),
		secrets: make(map[string]*domain.Secret),
	}

	s.userRepo = &memoryUserRepository{storage: s}
	s.secretRepo = &memorySecretRepository{storage: s}

	return s
}

// UserRepository возвращает in-memory UserRepository
func (s *memoryStorage) UserRepository() interfaces.UserRepository {
	return s.userRepo
}

// SecretRepository возвращает in-memory SecretRepository
func (s *memoryStorage) SecretRepository() interfaces.SecretRepository {
	return s.secretRepo
}

// TransactionManager возвращает менеджер транзакций
func (s *memoryStorage) TransactionManager() interfaces.TransactionManager {
	return s
}

// BeginTx начинает "транзакцию"
func (s *memoryStorage) BeginTx(ctx context.Context) (interfaces.Transaction, error) {
	return s, nil
}

// Commit для in-memory не делает ничего
func (s *memoryStorage) Commit(ctx context.Context) error {
	return nil
}

// Rollback для in-memory не делает ничего
func (s *memoryStorage) Rollback(ctx context.Context) error {
	return nil
}

// Close очищает хранилище
func (s *memoryStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users = make(map[string]*domain.User)
	s.secrets = make(map[string]*domain.Secret)
	return nil
}

func (s *memoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (r *memoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	for _, u := range r.storage.users {
		if u.Login == user.Login {
			return domain.ErrUserAlreadyExists
		}
	}

	r.storage.users[user.ID] = user
	return nil
}

func (r *memoryUserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	for _, user := range r.storage.users {
		if user.Login == login {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *memoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	user, exists := r.storage.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *memoryUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	_, exists := r.storage.users[user.ID]
	if !exists {
		return domain.ErrUserNotFound
	}

	for id, u := range r.storage.users {
		if u.Login == user.Login && id != user.ID {
			return domain.ErrUserAlreadyExists
		}
	}

	r.storage.users[user.ID] = user
	return nil
}

func (r *memoryUserRepository) Delete(ctx context.Context, id string) error {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	_, exists := r.storage.users[id]
	if !exists {
		return domain.ErrUserNotFound
	}

	delete(r.storage.users, id)
	return nil
}

func (r *memorySecretRepository) Create(ctx context.Context, secret *domain.Secret) error {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	key := r.storage.secretKey(secret.UserID, secret.ID)
	r.storage.secrets[key] = secret
	return nil
}

func (r *memorySecretRepository) GetByID(ctx context.Context, id, userID string) (*domain.Secret, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	key := r.storage.secretKey(userID, id)
	secret, exists := r.storage.secrets[key]
	if !exists || secret.IsDeleted {
		return nil, domain.ErrSecretNotFound
	}
	return secret, nil
}

func (r *memorySecretRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Secret, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	var secrets []*domain.Secret
	for key, secret := range r.storage.secrets {
		if r.storage.extractUserID(key) == userID && !secret.IsDeleted {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

func (r *memorySecretRepository) ListByUserAndType(ctx context.Context, userID string, secretType domain.SecretType) ([]*domain.Secret, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	var secrets []*domain.Secret
	for key, secret := range r.storage.secrets {
		if r.storage.extractUserID(key) == userID && secret.Type == secretType && !secret.IsDeleted {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

func (r *memorySecretRepository) Update(ctx context.Context, secret *domain.Secret) error {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	key := r.storage.secretKey(secret.UserID, secret.ID)
	existing, exists := r.storage.secrets[key]
	if !exists {
		return domain.ErrSecretNotFound
	}

	if existing.Version != secret.Version {
		return domain.ErrVersionConflict
	}

	secret.Version = existing.Version + 1
	secret.UpdatedAt = time.Now()
	r.storage.secrets[key] = secret
	return nil
}

func (r *memorySecretRepository) Delete(ctx context.Context, id, userID string) error {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	key := r.storage.secretKey(userID, id)
	_, exists := r.storage.secrets[key]
	if !exists {
		return domain.ErrSecretNotFound
	}

	delete(r.storage.secrets, key)
	return nil
}

func (r *memorySecretRepository) SoftDelete(ctx context.Context, id, userID string) error {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	key := r.storage.secretKey(userID, id)
	secret, exists := r.storage.secrets[key]
	if !exists || secret.IsDeleted {
		return domain.ErrSecretNotFound
	}

	secret.IsDeleted = true
	secret.UpdatedAt = time.Now()
	secret.Version++
	r.storage.secrets[key] = secret
	return nil
}

func (r *memorySecretRepository) GetUserSecretsVersion(ctx context.Context, userID string) (int64, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	var maxVersion int64
	for key, secret := range r.storage.secrets {
		if r.storage.extractUserID(key) == userID && !secret.IsDeleted {
			if secret.Version > maxVersion {
				maxVersion = secret.Version
			}
		}
	}
	return maxVersion, nil
}

func (r *memorySecretRepository) GetChangedSecrets(ctx context.Context, userID string, lastSyncVersion int64) ([]*domain.Secret, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	var secrets []*domain.Secret
	for key, secret := range r.storage.secrets {
		if r.storage.extractUserID(key) == userID && secret.Version >= lastSyncVersion {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

func (s *memoryStorage) secretKey(userID, secretID string) string {
	return userID + "_" + secretID
}

func (s *memoryStorage) extractUserID(key string) string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '_' {
			return key[:i]
		}
	}
	return key
}
