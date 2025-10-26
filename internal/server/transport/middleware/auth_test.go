package middleware_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/mocks"
	"github.com/alisaviation/GophKeeper/internal/server/transport/middleware"
)

func TestAuthInterceptor_Unary(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository()
	mockJWTManager := mocks.NewMockJWTManager()
	authService := app.NewAuthService(mockUserRepo, mockJWTManager)
	interceptor := middleware.NewAuthInterceptor(authService)

	user := &domain.User{
		ID:    "test-user-id",
		Login: "testuser",
	}
	mockUserRepo.Users[user.ID] = user
	mockUserRepo.Users[user.Login] = user

	tests := []struct {
		name         string
		fullMethod   string
		token        string
		wantError    bool
		expectedCode codes.Code
		setupContext func(context.Context) context.Context
	}{
		{
			name:       "public method without token",
			fullMethod: "/gophkeeper.v1.AuthService/Register",
			token:      "",
			wantError:  false,
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
		},
		{
			name:         "private method without token",
			fullMethod:   "/gophkeeper.v1.SecretService/Sync",
			token:        "",
			wantError:    true,
			expectedCode: codes.Unauthenticated,
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
		},
		{
			name:       "private method with valid token",
			fullMethod: "/gophkeeper.v1.SecretService/Sync",
			token:      "valid_token",
			wantError:  false,
			setupContext: func(ctx context.Context) context.Context {
				token, _ := mockJWTManager.GenerateAccessToken(user.ID, user.Login)
				md := metadata.Pairs("authorization", "Bearer "+token)
				return metadata.NewIncomingContext(ctx, md)
			},
		},
		{
			name:         "private method with invalid token",
			fullMethod:   "/gophkeeper.v1.SecretService/Sync",
			token:        "invalid_token",
			wantError:    true,
			expectedCode: codes.Unauthenticated,
			setupContext: func(ctx context.Context) context.Context {
				md := metadata.Pairs("authorization", "Bearer invalid_token")
				return metadata.NewIncomingContext(ctx, md)
			},
		},
		{
			name:         "private method with malformed auth header",
			fullMethod:   "/gophkeeper.v1.SecretService/Sync",
			token:        "",
			wantError:    true,
			expectedCode: codes.Unauthenticated,
			setupContext: func(ctx context.Context) context.Context {
				md := metadata.Pairs("authorization", "InvalidFormat")
				return metadata.NewIncomingContext(ctx, md)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.setupContext != nil {
				ctx = tt.setupContext(ctx)
			}

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				if !tt.wantError && tt.fullMethod != "/gophkeeper.v1.AuthService/Register" {
					_, err := middleware.GetUserFromContext(ctx)
					if err != nil {
						t.Errorf("Expected user in context, got error: %v", err)
					}
				}
				return "response", nil
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: tt.fullMethod,
			}

			interceptorFunc := interceptor.Unary()
			resp, err := interceptorFunc(ctx, "request", info, handler)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.expectedCode {
						t.Errorf("Expected code %v, got %v", tt.expectedCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp != "response" {
					t.Errorf("Expected response 'response', got %v", resp)
				}
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	t.Run("user exists in context", func(t *testing.T) {
		user := &domain.User{
			ID:    "test-user-id",
			Login: "testuser",
		}
		ctx := context.WithValue(context.Background(), struct{}{}, user)

		_, err := middleware.GetUserFromContext(ctx)
		if err == nil {
			t.Error("Expected error for wrong context key")
		}
	})

	t.Run("user not found in context", func(t *testing.T) {
		ctx := context.Background()

		_, err := middleware.GetUserFromContext(ctx)
		if err == nil {
			t.Error("Expected error when user not in context")
		}
		if status, ok := status.FromError(err); ok {
			if status.Code() != codes.Unauthenticated {
				t.Errorf("Expected code %v, got %v", codes.Unauthenticated, status.Code())
			}
		}
	})
}
