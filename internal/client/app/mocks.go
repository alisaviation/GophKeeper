package app

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/alisaviation/GophKeeper/internal/client/domain"
	pb "github.com/alisaviation/GophKeeper/internal/generated/grpc"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveSession(session *domain.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockStorage) GetSession() (*domain.Session, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockStorage) DeleteSession() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorage) SaveSecret(secret *domain.SecretData) error {
	args := m.Called(secret)
	return args.Error(0)
}

func (m *MockStorage) GetSecret(id string) (*domain.SecretData, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecretData), args.Error(1)
}

func (m *MockStorage) GetSecrets() ([]*domain.SecretData, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SecretData), args.Error(1)
}

type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) Register(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

func (m *MockTransport) Login(ctx context.Context, login, password string) (string, string, string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockTransport) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockTransport) Sync(ctx context.Context, userID string, lastSyncVersion int64, secrets []*pb.Secret) (*pb.SyncResponse, error) {
	args := m.Called(ctx, userID, lastSyncVersion, secrets)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SyncResponse), args.Error(1)
}

func (m *MockTransport) SetToken(token string) {
	m.Called(token)
}
