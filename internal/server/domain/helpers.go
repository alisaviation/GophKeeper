package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/alisaviation/GophKeeper/internal/generated/grpc"
)

func (s *Secret) ToProto() *grpc.Secret {
	return &grpc.Secret{
		Id:            s.ID,
		UserId:        s.UserID,
		Type:          SecretTypeToProto(s.Type),
		Name:          s.Name,
		EncryptedData: s.EncryptedData,
		EncryptedMeta: s.EncryptedMeta,
		Version:       s.Version,
		CreatedAt:     s.CreatedAt.Unix(),
		UpdatedAt:     s.UpdatedAt.Unix(),
		IsDeleted:     s.IsDeleted,
	}
}

func SecretFromProto(pb *grpc.Secret) *Secret {
	return &Secret{
		ID:            pb.GetId(),
		UserID:        pb.GetUserId(),
		Type:          SecretTypeFromProto(pb.GetType()),
		Name:          pb.GetName(),
		EncryptedData: pb.GetEncryptedData(),
		EncryptedMeta: pb.GetEncryptedMeta(),
		Version:       pb.GetVersion(),
		CreatedAt:     time.Unix(pb.GetCreatedAt(), 0),
		UpdatedAt:     time.Unix(pb.GetUpdatedAt(), 0),
		IsDeleted:     pb.GetIsDeleted(),
	}
}

func SecretTypeToProto(t SecretType) grpc.SecretType {
	switch t {
	case LoginPassword:
		return grpc.SecretType_LOGIN_PASSWORD
	case TextData:
		return grpc.SecretType_TEXT_DATA
	case BinaryData:
		return grpc.SecretType_BINARY_DATA
	case BankCard:
		return grpc.SecretType_BANK_CARD
	default:
		return grpc.SecretType_SECRET_TYPE_UNSPECIFIED
	}
}

func SecretTypeFromProto(pb grpc.SecretType) SecretType {
	switch pb {
	case grpc.SecretType_LOGIN_PASSWORD:
		return LoginPassword
	case grpc.SecretType_TEXT_DATA:
		return TextData
	case grpc.SecretType_BINARY_DATA:
		return BinaryData
	case grpc.SecretType_BANK_CARD:
		return BankCard
	default:
		return SecretTypeUnspecified
	}
}

// GenerateID генерирует уникальный ID
func GenerateID() string {
	return uuid.New().String()
}

// Now возвращает текущее время
func Now() time.Time {
	return time.Now()
}
