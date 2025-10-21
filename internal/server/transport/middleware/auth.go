package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

// AuthInterceptor перехватчик для аутентификации
type AuthInterceptor struct {
	authService *app.AuthService
}

// NewAuthInterceptor создает новый перехватчик аутентификации
func NewAuthInterceptor(authService *app.AuthService) *AuthInterceptor {
	return &AuthInterceptor{
		authService: authService,
	}
}

// Unary возвращает unary interceptor для аутентификации
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		user, err := i.authenticate(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}

		ctx = context.WithValue(ctx, UserContextKey{}, user)
		return handler(ctx, req)
	}
}

// authenticate извлекает и проверяет JWT токен
func (i *AuthInterceptor) authenticate(ctx context.Context) (*domain.User, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	authHeaders := md["authorization"]
	if len(authHeaders) == 0 {
		return nil, domain.ErrInvalidToken
	}

	token := strings.TrimPrefix(authHeaders[0], "Bearer ")
	if token == "" {
		return nil, domain.ErrInvalidToken
	}

	user, err := i.authService.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// isPublicMethod проверяет, является ли метод публичным
func isPublicMethod(fullMethod string) bool {
	publicMethods := map[string]bool{
		"/gophkeeper.v1.AuthService/Register": true,
		"/gophkeeper.v1.AuthService/Login":    true,
	}

	return publicMethods[fullMethod]
}

// UserContextKey ключ для хранения пользователя в контексте
type UserContextKey struct{}

// GetUserFromContext извлекает пользователя из контекста
func GetUserFromContext(ctx context.Context) (*domain.User, error) {
	user, ok := ctx.Value(UserContextKey{}).(*domain.User)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}
	return user, nil
}
