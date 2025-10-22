package transport_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpc2 "github.com/alisaviation/GophKeeper/internal/generated/grpc"
	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/transport"
)

func TestServer_Integration(t *testing.T) {
	mockUserRepo := app.NewMockUserRepository()
	mockSecretRepo := app.NewMockSecretRepository()
	mockJWTManager := app.NewMockJWTManager()
	mockEncryptor := &app.MockEncryptor{}

	authService := app.NewAuthService(mockUserRepo, mockJWTManager)
	dataService := app.NewDataService(mockSecretRepo, mockEncryptor)

	config := transport.Config{Port: 50052}
	server := transport.NewServer(authService, dataService, config)

	_, err := net.Listen("tcp", ":50051")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil {
			t.Logf("Server stopped: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	client := grpc2.NewAuthServiceClient(conn)

	t.Run("register and login flow", func(t *testing.T) {
		registerResp, err := client.Register(context.Background(), &grpc2.RegisterRequest{
			Login:    "integrationuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Failed to register: %v", err)
		}
		if registerResp.UserId == "" {
			t.Error("Expected non-empty user ID")
		}

		loginResp, err := client.Login(context.Background(), &grpc2.LoginRequest{
			Login:    "integrationuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}
		if loginResp.AccessToken == "" {
			t.Error("Expected non-empty access token")
		}
	})

	server.Stop()
}
