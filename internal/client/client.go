package client

//
//import (
//	"context"
//
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/metadata"
//
//	pb "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
//)
//
//// Client gRPC клиент
//type Client struct {
//	authClient   pb.AuthServiceClient
//	secretClient pb.SecretServiceClient
//	conn         *grpc.ClientConn
//	token        string
//}
//
//// NewClient создает нового gRPC клиента
//func NewClient(serverAddr string) (*Client, error) {
//	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
//	if err != nil {
//		return nil, err
//	}
//
//	return &Client{
//		authClient:   pb.NewAuthServiceClient(conn),
//		secretClient: pb.NewSecretServiceClient(conn),
//		conn:         conn,
//	}, nil
//}
//
//// SetToken устанавливает токен для аутентификации
//func (c *Client) SetToken(token string) {
//	c.token = token
//}
//
//// createAuthContext создает контекст с токеном аутентификации
//func (c *Client) createAuthContext(ctx context.Context) context.Context {
//	if c.token == "" {
//		return ctx
//	}
//	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+c.token))
//}
//
//// Register регистрирует нового пользователя
//func (c *Client) Register(ctx context.Context, login, password string) (string, error) {
//	resp, err := c.authClient.Register(ctx, &pb.RegisterRequest{
//		Login:    login,
//		Password: password,
//	})
//	if err != nil {
//		return "", err
//	}
//	return resp.GetUserId(), nil
//}
//
//// Login выполняет вход
//func (c *Client) Login(ctx context.Context, login, password string) (string, string, string, error) {
//	resp, err := c.authClient.Login(ctx, &pb.LoginRequest{
//		Login:    login,
//		Password: password,
//	})
//	if err != nil {
//		return "", "", "", err
//	}
//	return resp.GetAccessToken(), resp.GetRefreshToken(), resp.GetUserId(), nil
//}
//
//// Sync синхронизирует данные
//func (c *Client) Sync(ctx context.Context, userID string, lastSyncVersion int64, secrets []*pb.Secret) (*pb.SyncResponse, error) {
//	ctx = c.createAuthContext(ctx)
//	return c.secretClient.Sync(ctx, &pb.SyncRequest{
//		UserId:          userID,
//		LastSyncVersion: lastSyncVersion,
//		Secrets:         secrets,
//	})
//}
//
//// Close закрывает соединение
//func (c *Client) Close() error {
//	return c.conn.Close()
//}
//
//// RefreshToken обновляет токены
//func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
//	resp, err := c.authClient.RefreshToken(ctx, &pb.RefreshTokenRequest{
//		RefreshToken: refreshToken,
//	})
//	if err != nil {
//		return "", "", err
//	}
//	return resp.GetAccessToken(), resp.GetRefreshToken(), nil
//}
//
//// Logout выполняет выход
//func (c *Client) Logout(ctx context.Context, refreshToken string) error {
//	_, err := c.authClient.Logout(ctx, &pb.LogoutRequest{
//		RefreshToken: refreshToken,
//	})
//	return err
//}
//
//// GetSecret получает секрет по ID
//func (c *Client) GetSecret(ctx context.Context, secretID string) (*pb.Secret, error) {
//	ctx = c.createAuthContext(ctx)
//	resp, err := c.secretClient.GetSecret(ctx, &pb.GetSecretRequest{
//		SecretId: secretID,
//	})
//	if err != nil {
//		return nil, err
//	}
//	return resp.GetSecret(), nil
//}
//
//// ListSecrets получает список секретов
//func (c *Client) ListSecrets(ctx context.Context, userID string, filterType pb.SecretType) ([]*pb.Secret, error) {
//	ctx = c.createAuthContext(ctx)
//	resp, err := c.secretClient.ListSecrets(ctx, &pb.ListSecretsRequest{
//		UserId:     userID,
//		FilterType: filterType,
//	})
//	if err != nil {
//		return nil, err
//	}
//	return resp.GetSecrets(), nil
//}
//
//// UpdateSecret обновляет секрет
//func (c *Client) UpdateSecret(ctx context.Context, secret *pb.Secret) error {
//	ctx = c.createAuthContext(ctx)
//	_, err := c.secretClient.UpdateSecret(ctx, &pb.UpdateSecretRequest{
//		Secret: secret,
//	})
//	return err
//}
//
//// DeleteSecret удаляет секрет
//func (c *Client) DeleteSecret(ctx context.Context, secretID string) error {
//	ctx = c.createAuthContext(ctx)
//	_, err := c.secretClient.DeleteSecret(ctx, &pb.DeleteSecretRequest{
//		SecretId: secretID,
//	})
//	return err
//}
