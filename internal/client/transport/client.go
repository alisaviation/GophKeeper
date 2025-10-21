package transport

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
)

// GRPCClient gRPC клиент для взаимодействия с сервером
type GRPCClient struct {
	authClient   pb.AuthServiceClient
	secretClient pb.SecretServiceClient
	conn         *grpc.ClientConn
	token        string
}

// NewGRPCClient создает новый gRPC клиент
func NewGRPCClient(serverAddr string) (*GRPCClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		authClient:   pb.NewAuthServiceClient(conn),
		secretClient: pb.NewSecretServiceClient(conn),
		conn:         conn,
	}, nil
}

// SetToken устанавливает токен для аутентификации
func (c *GRPCClient) SetToken(token string) {
	c.token = token
}

// Register регистрирует нового пользователя
func (c *GRPCClient) Register(ctx context.Context, login, password string) (string, error) {
	resp, err := c.authClient.Register(ctx, &pb.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.GetUserId(), nil
}

// Login выполняет вход
func (c *GRPCClient) Login(ctx context.Context, login, password string) (string, string, string, error) {
	resp, err := c.authClient.Login(ctx, &pb.LoginRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", "", "", err
	}
	return resp.GetAccessToken(), resp.GetRefreshToken(), resp.GetUserId(), nil
}

// RefreshToken обновляет токены
func (c *GRPCClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	resp, err := c.authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return "", "", err
	}
	return resp.GetAccessToken(), resp.GetRefreshToken(), nil
}

// Logout выполняет выход
func (c *GRPCClient) Logout(ctx context.Context, refreshToken string) error {
	_, err := c.authClient.Logout(ctx, &pb.LogoutRequest{
		RefreshToken: refreshToken,
	})
	return err
}

// Sync синхронизирует данные
func (c *GRPCClient) Sync(ctx context.Context, userID string, lastSyncVersion int64, secrets []*pb.Secret) (*pb.SyncResponse, error) {
	ctx = c.createAuthContext(ctx)
	return c.secretClient.Sync(ctx, &pb.SyncRequest{
		UserId:          userID,
		LastSyncVersion: lastSyncVersion,
		Secrets:         secrets,
	})
}

// GetSecret получает секрет по ID
func (c *GRPCClient) GetSecret(ctx context.Context, secretID string) (*pb.Secret, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.secretClient.GetSecret(ctx, &pb.GetSecretRequest{
		SecretId: secretID,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetSecret(), nil
}

// ListSecrets получает список секретов
func (c *GRPCClient) ListSecrets(ctx context.Context, userID string, filterType pb.SecretType) ([]*pb.Secret, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.secretClient.ListSecrets(ctx, &pb.ListSecretsRequest{
		UserId:     userID,
		FilterType: filterType,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetSecrets(), nil
}

// UpdateSecret обновляет секрет
func (c *GRPCClient) UpdateSecret(ctx context.Context, secret *pb.Secret) error {
	ctx = c.createAuthContext(ctx)
	_, err := c.secretClient.UpdateSecret(ctx, &pb.UpdateSecretRequest{
		Secret: secret,
	})
	return err
}

// DeleteSecret удаляет секрет
func (c *GRPCClient) DeleteSecret(ctx context.Context, secretID string) error {
	ctx = c.createAuthContext(ctx)
	_, err := c.secretClient.DeleteSecret(ctx, &pb.DeleteSecretRequest{
		SecretId: secretID,
	})
	return err
}

// Close закрывает соединение
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// createAuthContext создает контекст с токеном аутентификации
func (c *GRPCClient) createAuthContext(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+c.token))
}
