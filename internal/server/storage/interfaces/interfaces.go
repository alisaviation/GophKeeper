package interfaces

import (
	"context"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

// UserRepository определяет контракт для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

// SecretRepository определяет контракт для работы с секретами
type SecretRepository interface {
	Create(ctx context.Context, secret *domain.Secret) error
	GetByID(ctx context.Context, id, userID string) (*domain.Secret, error)
	ListByUser(ctx context.Context, userID string) ([]*domain.Secret, error)
	ListByUserAndType(ctx context.Context, userID string, secretType domain.SecretType) ([]*domain.Secret, error)
	Update(ctx context.Context, secret *domain.Secret) error
	Delete(ctx context.Context, id, userID string) error
	SoftDelete(ctx context.Context, id, userID string) error
	GetUserSecretsVersion(ctx context.Context, userID string) (int64, error)
	GetChangedSecrets(ctx context.Context, userID string, lastSyncVersion int64) ([]*domain.Secret, error)
}

// TransactionManager определяет контракт для управления транзакциями
type TransactionManager interface {
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction представляет транзакцию базы данных
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	UserRepository() UserRepository
	SecretRepository() SecretRepository
}

// Storage объединяет все репозитории
type Storage interface {
	UserRepository() UserRepository
	SecretRepository() SecretRepository
	TransactionManager() TransactionManager
	Close() error
	Ping(ctx context.Context) error
}
