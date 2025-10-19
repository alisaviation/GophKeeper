package domain

import (
	"errors"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSecretNotFound     = errors.New("secret not found")
	ErrVersionConflict    = errors.New("version conflict")
	ErrInvalidSecretType  = errors.New("invalid secret type")
	ErrAccessDenied       = errors.New("access denied")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
