package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSecretConversion(t *testing.T) {
	userID := uuid.New().String()
	secretID := uuid.New().String()
	now := time.Now()

	domSecret := &Secret{
		ID:            secretID,
		UserID:        userID,
		Type:          LoginPassword,
		Name:          "test-secret",
		EncryptedData: []byte("encrypted-data"),
		EncryptedMeta: []byte("encrypted-meta"),
		Version:       1,
		CreatedAt:     now,
		UpdatedAt:     now,
		IsDeleted:     false,
	}

	pbSecret := domSecret.ToProto()

	if pbSecret.GetId() != secretID {
		t.Errorf("Expected ID %s, got %s", secretID, pbSecret.GetId())
	}

	if pbSecret.GetUserId() != userID {
		t.Errorf("Expected UserID %s, got %s", userID, pbSecret.GetUserId())
	}

	if pbSecret.GetVersion() != 1 {
		t.Errorf("Expected Version 1, got %d", pbSecret.GetVersion())
	}

	domSecret2 := SecretFromProto(pbSecret)

	if domSecret2.ID != secretID {
		t.Errorf("Expected ID %s after conversion, got %s", secretID, domSecret2.ID)
	}
}

func TestSecretTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		domType  SecretType
		protoVal int32
	}{
		{"LoginPassword", LoginPassword, 1},
		{"TextData", TextData, 2},
		{"BinaryData", BinaryData, 3},
		{"BankCard", BankCard, 4},
		{"Unspecified", SecretTypeUnspecified, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protoType := SecretTypeToProto(tt.domType)
			if int32(protoType) != tt.protoVal {
				t.Errorf("Expected proto value %d, got %d", tt.protoVal, protoType)
			}

			domType := SecretTypeFromProto(protoType)
			if domType != tt.domType {
				t.Errorf("Expected domain type %v, got %v", tt.domType, domType)
			}
		})
	}
}
