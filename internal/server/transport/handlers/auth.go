package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
	pb "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
)

// AuthHandler обработчик gRPC для аутентификации
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService *app.AuthService
}

// NewAuthHandler создает новый обработчик аутентификации
func NewAuthHandler(authService *app.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register регистрирует нового пользователя
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := h.authService.Register(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &pb.RegisterResponse{
		UserId: userID,
	}, nil
}

// Login аутентифицирует пользователя
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accessToken, refreshToken, userID, err := h.authService.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       userID,
	}, nil
}

// RefreshToken обновляет access token
func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	accessToken, refreshToken, err := h.authService.RefreshTokens(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout выполняет выход пользователя
func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}
	return &pb.LogoutResponse{
		Success: true,
	}, nil
}

// validateRegisterRequest валидирует запрос регистрации
func validateRegisterRequest(req *pb.RegisterRequest) error {
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
func validateLoginRequest(req *pb.LoginRequest) error {
	if req.GetLogin() == "" {
		return domain.ValidationError{Field: "login", Message: "is required"}
	}
	if req.GetPassword() == "" {
		return domain.ValidationError{Field: "password", Message: "is required"}
	}
	return nil
}
