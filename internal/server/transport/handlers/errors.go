package handlers

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

// MapErrorToStatus преобразует доменные ошибки в gRPC статусы
func MapErrorToStatus(err error) error {
	switch err {
	case domain.ErrUserAlreadyExists:
		return status.Error(codes.AlreadyExists, "user already exists")
	case domain.ErrUserNotFound:
		return status.Error(codes.NotFound, "user not found")
	case domain.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case domain.ErrSecretNotFound:
		return status.Error(codes.NotFound, "secret not found")
	case domain.ErrVersionConflict:
		return status.Error(codes.FailedPrecondition, "version conflict")
	case domain.ErrInvalidSecretType:
		return status.Error(codes.InvalidArgument, "invalid secret type")
	case domain.ErrAccessDenied:
		return status.Error(codes.PermissionDenied, "access denied")
	case domain.ErrInvalidToken:
		return status.Error(codes.Unauthenticated, "invalid token")
	case domain.ErrTokenExpired:
		return status.Error(codes.Unauthenticated, "token expired")
	case domain.ErrSecretAlreadyExists:
		return status.Error(codes.AlreadyExists, "secret already exists")
	case domain.ErrInvalidSecret:
		return status.Error(codes.InvalidArgument, "invalid secret")
	}
	if ve, ok := err.(domain.ValidationError); ok {
		return status.Error(codes.InvalidArgument, ve.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
