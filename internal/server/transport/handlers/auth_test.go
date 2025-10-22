package handlers_test

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/crypto"
	pb "github.com/alisaviation/GophKeeper/internal/generated/grpc"
	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/transport/handlers"
)

func TestAuthHandler_Register(t *testing.T) {
	mockUserRepo := app.NewMockUserRepository()
	mockJWTManager := app.NewMockJWTManager()
	authService := app.NewAuthService(mockUserRepo, mockJWTManager)
	handler := handlers.NewAuthHandler(authService)

	tests := []struct {
		name      string
		request   *pb.RegisterRequest
		wantError bool
		errorCode codes.Code
		setupMock func()
	}{
		{
			name: "successful registration",
			request: &pb.RegisterRequest{
				Login:    "newuser",
				Password: "password123",
			},
			wantError: false,
		},
		{
			name: "duplicate user",
			request: &pb.RegisterRequest{
				Login:    "existinguser",
				Password: "password123",
			},
			wantError: true,
			errorCode: codes.AlreadyExists,
			setupMock: func() {
				// Создаем существующего пользователя
				user := &domain.User{
					ID:    "existing-id",
					Login: "existinguser",
				}
				mockUserRepo.Users[user.Login] = user
			},
		},
		{
			name: "empty login",
			request: &pb.RegisterRequest{
				Login:    "",
				Password: "password123",
			},
			wantError: true,
			errorCode: codes.InvalidArgument,
		},
		{
			name: "empty password",
			request: &pb.RegisterRequest{
				Login:    "user",
				Password: "",
			},
			wantError: true,
			errorCode: codes.InvalidArgument,
		},
		{
			name: "short password",
			request: &pb.RegisterRequest{
				Login:    "user",
				Password: "123",
			},
			wantError: true,
			errorCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockUserRepo.Users = make(map[string]*domain.User)

			if tt.setupMock != nil {
				tt.setupMock()
			}

			ctx := context.Background()
			resp, err := handler.Register(ctx, tt.request)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errorCode {
						t.Errorf("Expected code %v, got %v", tt.errorCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if resp.UserId == "" {
					t.Error("Expected non-empty user ID")
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	mockUserRepo := app.NewMockUserRepository()
	mockJWTManager := app.NewMockJWTManager()
	authService := app.NewAuthService(mockUserRepo, mockJWTManager)
	handler := handlers.NewAuthHandler(authService)

	setupTestUser := func() {
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := &domain.User{
			ID:           "test-user-id",
			Login:        "testuser",
			PasswordHash: string(hashedPassword),
		}
		mockUserRepo.Users[user.ID] = user
		mockUserRepo.Users[user.Login] = user
	}

	tests := []struct {
		name      string
		request   *pb.LoginRequest
		wantError bool
		errorCode codes.Code
	}{
		{
			name: "successful login",
			request: &pb.LoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			wantError: false,
		},
		{
			name: "user not found",
			request: &pb.LoginRequest{
				Login:    "nonexistent",
				Password: "password123",
			},
			wantError: true,
			errorCode: codes.Unauthenticated,
		},
		{
			name: "wrong password",
			request: &pb.LoginRequest{
				Login:    "testuser",
				Password: "wrongpassword",
			},
			wantError: true,
			errorCode: codes.Unauthenticated,
		},
		{
			name: "empty login",
			request: &pb.LoginRequest{
				Login:    "",
				Password: "password123",
			},
			wantError: true,
			errorCode: codes.InvalidArgument,
		},
		{
			name: "empty password",
			request: &pb.LoginRequest{
				Login:    "testuser",
				Password: "",
			},
			wantError: true,
			errorCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo.Users = make(map[string]*domain.User)
			setupTestUser()

			ctx := context.Background()
			resp, err := handler.Login(ctx, tt.request)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errorCode {
						t.Errorf("Expected code %v, got %v", tt.errorCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if resp.AccessToken == "" {
					t.Error("Expected non-empty access token")
				}
				if resp.RefreshToken == "" {
					t.Error("Expected non-empty refresh token")
				}
				if resp.UserId != "test-user-id" {
					t.Errorf("Expected user ID %s, got %s", "test-user-id", resp.UserId)
				}
			}
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	mockUserRepo := app.NewMockUserRepository()
	mockJWTManager := app.NewMockJWTManager()
	authService := app.NewAuthService(mockUserRepo, mockJWTManager)
	handler := handlers.NewAuthHandler(authService)

	user := &domain.User{
		ID:    "test-user-id",
		Login: "testuser",
	}

	tests := []struct {
		name        string
		request     *pb.RefreshTokenRequest
		wantError   bool
		errorCode   codes.Code
		setupMock   func() *pb.RefreshTokenRequest
		description string
	}{
		{
			name:      "successful token refresh",
			wantError: false,
			setupMock: func() *pb.RefreshTokenRequest {
				mockUserRepo.Users[user.ID] = user
				token, _ := mockJWTManager.GenerateRefreshToken(user.ID)
				return &pb.RefreshTokenRequest{RefreshToken: token}
			},
			description: "valid refresh token with existing user",
		},
		{
			name:      "empty refresh token",
			wantError: true,
			errorCode: codes.InvalidArgument,
			setupMock: func() *pb.RefreshTokenRequest {
				mockUserRepo.Users[user.ID] = user
				return &pb.RefreshTokenRequest{RefreshToken: ""}
			},
			description: "empty token should return invalid argument",
		},
		{
			name:      "invalid refresh token",
			wantError: true,
			errorCode: codes.Internal,
			setupMock: func() *pb.RefreshTokenRequest {
				mockUserRepo.Users[user.ID] = user
				return &pb.RefreshTokenRequest{RefreshToken: "invalid_token"}
			},
			description: "invalid token returns internal error",
		},
		{
			name:      "user not found for valid token",
			wantError: true,
			errorCode: codes.NotFound,
			setupMock: func() *pb.RefreshTokenRequest {
				token, _ := mockJWTManager.GenerateRefreshToken("non-existent-user")
				mockJWTManager.Tokens[token] = &crypto.TokenClaims{
					UserID: "non-existent-user",
					Type:   "refresh",
				}
				return &pb.RefreshTokenRequest{RefreshToken: token}
			},
			description: "valid token but user doesn't exist returns not found",
		},
		{
			name:      "expired refresh token",
			wantError: true,
			errorCode: codes.Internal,
			setupMock: func() *pb.RefreshTokenRequest {
				mockUserRepo.Users[user.ID] = user
				return &pb.RefreshTokenRequest{RefreshToken: "expired_token"}
			},
			description: "expired token returns internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo.Users = make(map[string]*domain.User)
			mockJWTManager.Tokens = make(map[string]*crypto.TokenClaims)
			mockJWTManager.TokenCounter = 0

			var req *pb.RefreshTokenRequest
			if tt.setupMock != nil {
				req = tt.setupMock()
			}

			ctx := context.Background()
			resp, err := handler.RefreshToken(ctx, req)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil. Test: %s", tt.description)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error, got: %v. Test: %s", err, tt.description)
					return
				}
				if st.Code() != tt.errorCode {
					t.Errorf("Expected code %v, got %v. Message: %s. Test: %s",
						tt.errorCode, st.Code(), st.Message(), tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Test: %s", err, tt.description)
					return
				}
				if resp.AccessToken == "" {
					t.Error("Expected non-empty access token")
				}
				if resp.RefreshToken == "" {
					t.Error("Expected non-empty refresh token")
				}
				if req != nil && resp.RefreshToken == req.RefreshToken {
					t.Error("Expected new refresh token to be different from old one")
				}
			}
		})
	}
}
