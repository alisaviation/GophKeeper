package domain

import (
	"time"

	"github.com/google/uuid"
)

// SecretType тип секрета
type SecretType string

const (
	SecretTypeLoginPassword SecretType = "login_password"
	SecretTypeText          SecretType = "text"
	SecretTypeBankCard      SecretType = "bank_card"
	SecretTypeBinary        SecretType = "binary_data"
)

// SecretData локальное представление секрета
type SecretData struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Type      SecretType  `json:"type"`
	Name      string      `json:"name"`
	Data      interface{} `json:"data"`
	Version   int64       `json:"version"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	IsDirty   bool        `json:"is_dirty"` // Флаг изменений для синхронизации
	IsDeleted bool        `json:"is_deleted"`
}

// LoginPasswordData данные логина/пароля
type LoginPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Website  string `json:"website,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

// TextData текстовые данные
type TextData struct {
	Content     string `json:"content"`
	Description string `json:"description,omitempty"`
}

// BankCardData данные банковской карты
type BankCardData struct {
	CardHolder string `json:"card_holder"`
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
	BankName   string `json:"bank_name,omitempty"`
}

type BinaryData struct {
	Data        []byte `json:"data"`
	Description string `json:"description,omitempty"`
	FileName    string `json:"file_name"`
}

// Session сессия пользователя
type Session struct {
	UserID          string `json:"user_id"`
	Login           string `json:"login"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	LastSync        int64  `json:"last_sync"`
	LastSyncVersion int64  `json:"last_sync_version"`
	EncryptionKey   []byte `json:"encryption_key"`
}

// GenerateID генерирует уникальный ID
func GenerateID() string {
	return uuid.New().String()
}
