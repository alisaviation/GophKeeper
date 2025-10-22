package crypto

import (
	"testing"
)

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid 16 byte key",
			keySize:   16,
			wantError: false,
		},
		{
			name:      "valid 24 byte key",
			keySize:   24,
			wantError: false,
		},
		{
			name:      "valid 32 byte key",
			keySize:   32,
			wantError: false,
		},
		{
			name:      "invalid key size 8",
			keySize:   8,
			wantError: true,
			errorMsg:  "key size must be 16, 24 or 32 bytes",
		},
		{
			name:      "invalid key size 0",
			keySize:   0,
			wantError: true,
			errorMsg:  "key size must be 16, 24 or 32 bytes",
		},
		{
			name:      "invalid key size 64",
			keySize:   64,
			wantError: true,
			errorMsg:  "key size must be 16, 24 or 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateKey(tt.keySize)

			if tt.wantError {
				if err == nil {
					t.Errorf("GenerateKey(%d) expected error, but got nil", tt.keySize)
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("GenerateKey(%d) error = %v, want %v", tt.keySize, err.Error(), tt.errorMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateKey(%d) unexpected error: %v", tt.keySize, err)
				return
			}

			if len(key) != tt.keySize {
				t.Errorf("GenerateKey(%d) key length = %d, want %d", tt.keySize, len(key), tt.keySize)
			}

			allZero := true
			for _, b := range key {
				if b != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				t.Errorf("GenerateKey(%d) generated key is all zeros", tt.keySize)
			}
		})
	}
}

func TestGenerateKey_Uniqueness(t *testing.T) {
	keySize := 16
	key1, err1 := GenerateKey(keySize)
	key2, err2 := GenerateKey(keySize)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to generate keys: %v, %v", err1, err2)
	}

	equal := true
	for i := range key1 {
		if key1[i] != key2[i] {
			equal = false
			break
		}
	}
	if equal {
		t.Errorf("GenerateKey(%d) generated identical keys", keySize)
	}
}
