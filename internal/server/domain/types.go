package domain

import (
	"time"

	gophkeeper_v1 "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
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
	IsDeleted     bool // Soft delete
}

type SecretType string

const (
	SecretTypeUnspecified SecretType = "unspecified"
	LoginPassword         SecretType = "login_password"
	TextData              SecretType = "text_data"
	BinaryData            SecretType = "binary_data"
	BankCard              SecretType = "bank_card"
)

type SyncState struct {
	UserID         string
	CurrentVersion int64
	LastSyncAt     time.Time
}

type SecretVersion struct {
	SecretID  string
	Version   int64
	Data      []byte
	Meta      []byte
	CreatedAt time.Time
}

func (s *Secret) ToProto() *gophkeeper_v1.Secret {
	return &gophkeeper_v1.Secret{
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

func SecretFromProto(pb *gophkeeper_v1.Secret) *Secret {
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

func SecretTypeToProto(t SecretType) gophkeeper_v1.SecretType {
	switch t {
	case LoginPassword:
		return gophkeeper_v1.SecretType_LOGIN_PASSWORD
	case TextData:
		return gophkeeper_v1.SecretType_TEXT_DATA
	case BinaryData:
		return gophkeeper_v1.SecretType_BINARY_DATA
	case BankCard:
		return gophkeeper_v1.SecretType_BANK_CARD
	default:
		return gophkeeper_v1.SecretType_SECRET_TYPE_UNSPECIFIED
	}
}

func SecretTypeFromProto(pb gophkeeper_v1.SecretType) SecretType {
	switch pb {
	case gophkeeper_v1.SecretType_LOGIN_PASSWORD:
		return LoginPassword
	case gophkeeper_v1.SecretType_TEXT_DATA:
		return TextData
	case gophkeeper_v1.SecretType_BINARY_DATA:
		return BinaryData
	case gophkeeper_v1.SecretType_BANK_CARD:
		return BankCard
	default:
		return SecretTypeUnspecified
	}
}
