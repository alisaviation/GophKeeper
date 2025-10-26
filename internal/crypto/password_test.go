package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/crypto"
)

func TestPasswordHasher(t *testing.T) {
	password := "my-secret-password-123"

	hash, err := crypto.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	valid := crypto.CheckPasswordHash(password, hash)
	assert.True(t, valid)

	valid = crypto.CheckPasswordHash("wrong-password", hash)
	assert.False(t, valid)

	hash2, err := crypto.HashPassword(password)
	require.NoError(t, err)
	assert.NotEqual(t, hash, hash2)

	valid = crypto.CheckPasswordHash(password, hash2)
	assert.True(t, valid)
}
