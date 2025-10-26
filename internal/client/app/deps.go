package app

import (
	"context"

	"github.com/alisaviation/GophKeeper/internal/client/domain"
	pb "github.com/alisaviation/GophKeeper/internal/generated/grpc"
)

// Storage интерфейс для хранилища
type Storage interface {
	SaveSession(session *domain.Session) error
	GetSession() (*domain.Session, error)
	DeleteSession() error
	SaveSecret(secret *domain.SecretData) error
	GetSecret(id string) (*domain.SecretData, error)
	GetSecrets() ([]*domain.SecretData, error)
}

// Transport интерфейс для транспорта
type Transport interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	Sync(ctx context.Context, userID string, lastSyncVersion int64, secrets []*pb.Secret) (*pb.SyncResponse, error)
	SetToken(token string)
}
