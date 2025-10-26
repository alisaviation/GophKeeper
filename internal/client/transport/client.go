package transport

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	grpc2 "github.com/alisaviation/GophKeeper/internal/generated/grpc"
)

// GRPCClient gRPC клиент для взаимодействия с сервером
type GRPCClient struct {
	authClient   grpc2.AuthServiceClient
	secretClient grpc2.SecretServiceClient
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
		authClient:   grpc2.NewAuthServiceClient(conn),
		secretClient: grpc2.NewSecretServiceClient(conn),
		conn:         conn,
	}, nil
}

// SetToken устанавливает токен для аутентификации
func (c *GRPCClient) SetToken(token string) {
	c.token = token
}

// Register регистрирует нового пользователя
func (c *GRPCClient) Register(ctx context.Context, login, password string) (string, error) {
	resp, err := c.authClient.Register(ctx, &grpc2.RegisterRequest{
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
	resp, err := c.authClient.Login(ctx, &grpc2.LoginRequest{
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
	resp, err := c.authClient.RefreshToken(ctx, &grpc2.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return "", "", err
	}
	return resp.GetAccessToken(), resp.GetRefreshToken(), nil
}

// Logout выполняет выход
func (c *GRPCClient) Logout(ctx context.Context, refreshToken string) error {
	_, err := c.authClient.Logout(ctx, &grpc2.LogoutRequest{
		RefreshToken: refreshToken,
	})
	return err
}

// Sync синхронизирует данные
func (c *GRPCClient) Sync(ctx context.Context, userID string, lastSyncVersion int64, secrets []*grpc2.Secret) (*grpc2.SyncResponse, error) {
	ctx = c.createAuthContext(ctx)
	return c.secretClient.Sync(ctx, &grpc2.SyncRequest{
		UserId:          userID,
		LastSyncVersion: lastSyncVersion,
		Secrets:         secrets,
	})
}

// GetSecret получает секрет по ID
func (c *GRPCClient) GetSecret(ctx context.Context, secretID string) (*grpc2.Secret, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.secretClient.GetSecret(ctx, &grpc2.GetSecretRequest{
		SecretId: secretID,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetSecret(), nil
}

// ListSecrets получает список секретов
func (c *GRPCClient) ListSecrets(ctx context.Context, userID string, filterType grpc2.SecretType) ([]*grpc2.Secret, error) {
	ctx = c.createAuthContext(ctx)
	resp, err := c.secretClient.ListSecrets(ctx, &grpc2.ListSecretsRequest{
		UserId:     userID,
		FilterType: filterType,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetSecrets(), nil
}

// UpdateSecret обновляет секрет
func (c *GRPCClient) UpdateSecret(ctx context.Context, secret *grpc2.Secret) error {
	ctx = c.createAuthContext(ctx)
	_, err := c.secretClient.UpdateSecret(ctx, &grpc2.UpdateSecretRequest{
		Secret: secret,
	})
	return err
}

// DeleteSecret удаляет секрет
func (c *GRPCClient) DeleteSecret(ctx context.Context, secretID string) error {
	ctx = c.createAuthContext(ctx)
	_, err := c.secretClient.DeleteSecret(ctx, &grpc2.DeleteSecretRequest{
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
