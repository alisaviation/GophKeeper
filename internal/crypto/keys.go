package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// GenerateKey генерирует случайный ключ для AES шифрования
func GenerateKey(keySize int) ([]byte, error) {
	if keySize != 16 && keySize != 24 && keySize != 32 {
		return nil, errors.New("key size must be 16, 24 or 32 bytes")
	}

	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	return key, nil
}

// GenerateKeyBase64 генерирует случайный ключ и возвращает в base64
func GenerateKeyBase64(keySize int) (string, error) {
	key, err := GenerateKey(keySize)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// KeyFromBase64 декодирует ключ из base64
func KeyFromBase64(keyBase64 string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}
	return key, nil
}

// GenerateEncryptionKey генерирует ключ для шифрования данных по умолчанию (AES-256)
func GenerateEncryptionKey() ([]byte, error) {
	return GenerateKey(32)
}

// GenerateEncryptionKeyBase64 генерирует ключ в base64
func GenerateEncryptionKeyBase64() (string, error) {
	return GenerateKeyBase64(32)
}
