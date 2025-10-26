package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alisaviation/GophKeeper/internal/client/domain"
	"github.com/alisaviation/GophKeeper/internal/crypto"
	pb "github.com/alisaviation/GophKeeper/internal/generated/grpc"
)

func (c *Client) ensureAuthenticated(ctx context.Context) (*domain.Session, error) {
	session, err := c.storage.GetSession()
	if err != nil {
		return nil, fmt.Errorf("not authenticated: %w", err)
	}

	if session.AccessToken == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	c.transport.SetToken(session.AccessToken)
	return session, nil
}

func (c *Client) encryptSecret(secret *domain.SecretData) (*pb.Secret, error) {
	session, err := c.storage.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	encryptor, err := crypto.NewAESGCMEncryptor(session.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	dataBytes, err := json.Marshal(secret.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	encryptedData, err := encryptor.Encrypt(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	pbSecret := &pb.Secret{
		Id:            secret.ID,
		UserId:        secret.UserID,
		Type:          mapSecretTypeToProto(secret.Type),
		Name:          secret.Name,
		EncryptedData: encryptedData,
		Version:       secret.Version,
		CreatedAt:     secret.CreatedAt.Unix(),
		UpdatedAt:     secret.UpdatedAt.Unix(),
		IsDeleted:     secret.IsDeleted,
	}

	return pbSecret, nil
}

func (c *Client) decryptSecret(pbSecret *pb.Secret) (*domain.SecretData, error) {
	session, err := c.storage.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	encryptor, err := crypto.NewAESGCMEncryptor(session.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	decryptedData, err := encryptor.Decrypt(pbSecret.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	var data interface{}
	switch mapSecretTypeFromProto(pbSecret.Type) {
	case domain.SecretTypeLoginPassword:
		var loginPassData domain.LoginPasswordData
		if err := json.Unmarshal(decryptedData, &loginPassData); err != nil {
			return nil, fmt.Errorf("failed to deserialize login password data: %w", err)
		}
		data = loginPassData
	case domain.SecretTypeText:
		var textData domain.TextData
		if err := json.Unmarshal(decryptedData, &textData); err != nil {
			return nil, fmt.Errorf("failed to deserialize text data: %w", err)
		}
		data = textData
	case domain.SecretTypeBankCard:
		var cardData domain.BankCardData
		if err := json.Unmarshal(decryptedData, &cardData); err != nil {
			return nil, fmt.Errorf("failed to deserialize bank card data: %w", err)
		}
		data = cardData
	default:
		return nil, fmt.Errorf("unknown secret type: %v", pbSecret.Type)
	}

	secret := &domain.SecretData{
		ID:        pbSecret.Id,
		UserID:    pbSecret.UserId,
		Type:      mapSecretTypeFromProto(pbSecret.Type),
		Name:      pbSecret.Name,
		Data:      data,
		Version:   pbSecret.Version,
		CreatedAt: time.Unix(pbSecret.CreatedAt, 0),
		UpdatedAt: time.Unix(pbSecret.UpdatedAt, 0),
		IsDirty:   false,
		IsDeleted: pbSecret.IsDeleted,
	}

	return secret, nil
}

func generateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}

func mapSecretTypeToProto(secretType domain.SecretType) pb.SecretType {
	switch secretType {
	case domain.SecretTypeLoginPassword:
		return pb.SecretType_LOGIN_PASSWORD
	case domain.SecretTypeText:
		return pb.SecretType_TEXT_DATA
	case domain.SecretTypeBankCard:
		return pb.SecretType_BANK_CARD
	default:
		return pb.SecretType_SECRET_TYPE_UNSPECIFIED
	}
}

func mapSecretTypeFromProto(pbType pb.SecretType) domain.SecretType {
	switch pbType {
	case pb.SecretType_LOGIN_PASSWORD:
		return domain.SecretTypeLoginPassword
	case pb.SecretType_TEXT_DATA:
		return domain.SecretTypeText
	case pb.SecretType_BANK_CARD:
		return domain.SecretTypeBankCard
	default:
		return domain.SecretTypeLoginPassword
	}
}
