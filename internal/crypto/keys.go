package crypto

import (
	"crypto/rand"
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
