package interfaces

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrSecretNotFound         = errors.New("secret not found")
	ErrVersionConflict        = errors.New("version conflict")
	ErrConcurrentModification = errors.New("concurrent modification detected")
	ErrInvalidData            = errors.New("invalid data")
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrInvalidCredentials     = errors.New("invalid credentials")
)
