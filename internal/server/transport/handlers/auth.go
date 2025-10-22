package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/generated/grpc"
	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

// AuthHandler обработчик gRPC для аутентификации
type AuthHandler struct {
	grpc.UnimplementedAuthServiceServer
	authService *app.AuthService
}

// NewAuthHandler создает новый обработчик аутентификации
func NewAuthHandler(authService *app.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register регистрирует нового пользователя
func (h *AuthHandler) Register(ctx context.Context, req *grpc.RegisterRequest) (*grpc.RegisterResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := h.authService.Register(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &grpc.RegisterResponse{
		UserId: userID,
	}, nil
}

// Login аутентифицирует пользователя
func (h *AuthHandler) Login(ctx context.Context, req *grpc.LoginRequest) (*grpc.LoginResponse, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accessToken, refreshToken, userID, err := h.authService.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &grpc.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       userID,
	}, nil
}

// RefreshToken обновляет access token
func (h *AuthHandler) RefreshToken(ctx context.Context, req *grpc.RefreshTokenRequest) (*grpc.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	accessToken, refreshToken, err := h.authService.RefreshTokens(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &grpc.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout выполняет выход пользователя
func (h *AuthHandler) Logout(ctx context.Context, req *grpc.LogoutRequest) (*grpc.LogoutResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}
	return &grpc.LogoutResponse{
		Success: true,
	}, nil
}

// validateRegisterRequest валидирует запрос регистрации
func validateRegisterRequest(req *grpc.RegisterRequest) error {
	if req.GetLogin() == "" {
		return domain.ValidationError{Field: "login", Message: "is required"}
	}
	if req.GetPassword() == "" {
		return domain.ValidationError{Field: "password", Message: "is required"}
	}
	if len(req.GetPassword()) < 8 {
		return domain.ValidationError{Field: "password", Message: "must be at least 8 characters"}
	}
	return nil
}

// validateLoginRequest валидирует запрос входа
func validateLoginRequest(req *grpc.LoginRequest) error {
	if req.GetLogin() == "" {
		return domain.ValidationError{Field: "login", Message: "is required"}
	}
	if req.GetPassword() == "" {
		return domain.ValidationError{Field: "password", Message: "is required"}
	}
	return nil
}
