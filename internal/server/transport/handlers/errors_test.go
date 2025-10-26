package handlers_test

import (
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/transport/handlers"
)

func TestMapErrorToStatus(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{
			name:         "user already exists",
			err:          domain.ErrUserAlreadyExists,
			expectedCode: codes.AlreadyExists,
		},
		{
			name:         "user not found",
			err:          domain.ErrUserNotFound,
			expectedCode: codes.NotFound,
		},
		{
			name:         "invalid credentials",
			err:          domain.ErrInvalidCredentials,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "secret not found",
			err:          domain.ErrSecretNotFound,
			expectedCode: codes.NotFound,
		},
		{
			name:         "version conflict",
			err:          domain.ErrVersionConflict,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "invalid secret type",
			err:          domain.ErrInvalidSecretType,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "access denied",
			err:          domain.ErrAccessDenied,
			expectedCode: codes.PermissionDenied,
		},
		{
			name:         "invalid token",
			err:          domain.ErrInvalidToken,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "token expired",
			err:          domain.ErrTokenExpired,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "secret already exists",
			err:          domain.ErrSecretAlreadyExists,
			expectedCode: codes.AlreadyExists,
		},
		{
			name:         "invalid secret",
			err:          domain.ErrInvalidSecret,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "validation error",
			err:          domain.ValidationError{Field: "login", Message: "invalid format"},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "unknown error",
			err:          fmt.Errorf("some unknown error"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mappedErr := handlers.MapErrorToStatus(tt.err)

			st, ok := status.FromError(mappedErr)
			if !ok {
				t.Fatalf("Expected gRPC status error, got: %v", mappedErr)
			}

			if st.Code() != tt.expectedCode {
				t.Errorf("MapErrorToStatus(%v): expected code %v, got %v. Message: %s",
					tt.err, tt.expectedCode, st.Code(), st.Message())
			}

			if st.Message() == "" {
				t.Error("Expected non-empty error message")
			}
		})
	}
}
