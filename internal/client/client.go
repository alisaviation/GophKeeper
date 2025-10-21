package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
)

// Client gRPC клиент
type Client struct {
	authClient   pb.AuthServiceClient
	secretClient pb.SecretServiceClient
	conn         *grpc.ClientConn
	token        string
}

// NewClient создает нового gRPC клиента
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Client{
		authClient:   pb.NewAuthServiceClient(conn),
		secretClient: pb.NewSecretServiceClient(conn),
		conn:         conn,
	}, nil
}

// SetToken устанавливает токен для аутентификации
func (c *Client) SetToken(token string) {
	c.token = token
}

// createAuthContext создает контекст с токеном аутентификации
func (c *Client) createAuthContext(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+c.token))
}

// Register регистрирует нового пользователя
func (c *Client) Register(ctx context.Context, login, password string) (string, error) {
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
func (c *Client) Login(ctx context.Context, login, password string) (string, string, string, error) {
	resp, err := c.authClient.Login(ctx, &pb.LoginRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", "", "", err
	}
	return resp.GetAccessToken(), resp.GetRefreshToken(), resp.GetUserId(), nil
}

// Sync синхронизирует данные
func (c *Client) Sync(ctx context.Context, userID string, lastSyncVersion int64, secrets []*pb.Secret) (*pb.SyncResponse, error) {
	ctx = c.createAuthContext(ctx)
	return c.secretClient.Sync(ctx, &pb.SyncRequest{
		UserId:          userID,
		LastSyncVersion: lastSyncVersion,
		Secrets:         secrets,
	})
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}
