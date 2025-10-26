package domain

import (
	"time"
)

type User struct {
	ID           string
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Secret struct {
	ID            string
	UserID        string
	Type          SecretType
	Name          string
	EncryptedData []byte
	EncryptedMeta []byte
	Version       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	IsDeleted     bool
}

type SecretType string

const (
	SecretTypeUnspecified SecretType = "unspecified"
	LoginPassword         SecretType = "login_password"
	TextData              SecretType = "text_data"
	BinaryData            SecretType = "binary_data"
	BankCard              SecretType = "bank_card"
)
