package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/crypto"
)

func TestAESGCMEncryptor(t *testing.T) {
	key, err := crypto.GenerateKey(32)
	require.NoError(t, err)

	encryptor, err := crypto.NewAESGCMEncryptor(key)
	require.NoError(t, err)

	plaintext := []byte("Hello, World! This is a secret message.")

	ciphertext, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := encryptor.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	ciphertext2, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, ciphertext, ciphertext2)

	decrypted2, err := encryptor.Decrypt(ciphertext2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestAESGCMEncryptor_InvalidKey(t *testing.T) {
	_, err := crypto.NewAESGCMEncryptor([]byte("short"))
	require.Error(t, err)
}

func TestAESGCMEncryptor_InvalidCiphertext(t *testing.T) {
	key, err := crypto.GenerateKey(32)
	require.NoError(t, err)

	encryptor, err := crypto.NewAESGCMEncryptor(key)
	require.NoError(t, err)

	_, err = encryptor.Decrypt([]byte("short"))
	require.Error(t, err)

	_, err = encryptor.Decrypt([]byte("this-is-not-valid-ciphertext-data"))
	require.Error(t, err)
}
