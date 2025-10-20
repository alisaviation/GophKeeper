package app

import (
	"context"
	"fmt"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

// DataService предоставляет методы для управления секретами
type DataService struct {
	secrets interfaces.SecretRepository
	crypto  Encryptor
}

// NewDataService создает новый сервис управления данными
func NewDataService(secrets interfaces.SecretRepository, crypto Encryptor) *DataService {
	return &DataService{
		secrets: secrets,
		crypto:  crypto,
	}
}

// Sync синхронизирует данные между клиентом и сервером
func (s *DataService) Sync(ctx context.Context, userID string, clientSecrets []*domain.Secret, lastSyncVersion int64) (*SyncResult, error) {

	serverVersion, err := s.secrets.GetUserSecretsVersion(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	serverChanges, err := s.secrets.GetChangedSecrets(ctx, userID, lastSyncVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get server changes: %w", err)
	}

	conflicts, err := s.processClientChanges(ctx, userID, clientSecrets)
	if err != nil {
		return nil, fmt.Errorf("failed to process client changes: %w", err)
	}

	result := &SyncResult{
		CurrentVersion: serverVersion,
		ServerSecrets:  serverChanges,
		Conflicts:      conflicts,
	}

	return result, nil
}

// SyncResult представляет результат синхронизации
type SyncResult struct {
	CurrentVersion int64
	ServerSecrets  []*domain.Secret
	Conflicts      []string // ID конфликтных секретов
}

// GetSecret возвращает секрет по ID
func (s *DataService) GetSecret(ctx context.Context, userID, secretID string) (*domain.Secret, error) {
	secret, err := s.secrets.GetByID(ctx, secretID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	return secret, nil
}

// ListSecrets возвращает все секреты пользователя
func (s *DataService) ListSecrets(ctx context.Context, userID string, filterType *domain.SecretType) ([]*domain.Secret, error) {
	var secrets []*domain.Secret
	var err error

	if filterType != nil {
		secrets, err = s.secrets.ListByUserAndType(ctx, userID, *filterType)
	} else {
		secrets, err = s.secrets.ListByUser(ctx, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secrets, nil
}

// CreateSecret создает новый секрет
func (s *DataService) CreateSecret(ctx context.Context, userID string, secret *domain.Secret) error {
	secret.ID = domain.GenerateID()
	secret.UserID = userID
	secret.Version = 1
	secret.CreatedAt = domain.Now()
	secret.UpdatedAt = domain.Now()
	secret.IsDeleted = false

	if err := s.validateSecret(secret); err != nil {
		return fmt.Errorf("invalid secret: %w", err)
	}

	if err := s.secrets.Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// UpdateSecret обновляет существующий секрет
func (s *DataService) UpdateSecret(ctx context.Context, userID string, secret *domain.Secret) error {

	existing, err := s.secrets.GetByID(ctx, secret.ID, userID)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	if existing.Version != secret.Version {
		return interfaces.ErrVersionConflict
	}

	secret.UserID = userID
	secret.UpdatedAt = domain.Now()

	if err := s.validateSecret(secret); err != nil {
		return fmt.Errorf("invalid secret: %w", err)
	}

	if err := s.secrets.Update(ctx, secret); err != nil {
		if err == interfaces.ErrVersionConflict {
			return interfaces.ErrVersionConflict
		}
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// DeleteSecret удаляет секрет
func (s *DataService) DeleteSecret(ctx context.Context, userID, secretID string) error {
	if err := s.secrets.SoftDelete(ctx, secretID, userID); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

func (s *DataService) processClientChanges(ctx context.Context, userID string, clientSecrets []*domain.Secret) ([]string, error) {
	var conflicts []string

	for _, clientSecret := range clientSecrets {
		if clientSecret.UserID != userID {
			return nil, fmt.Errorf("access denied for secret %s", clientSecret.ID)
		}

		if clientSecret.IsDeleted {
			if err := s.secrets.SoftDelete(ctx, clientSecret.ID, userID); err != nil && err != interfaces.ErrSecretNotFound {
				if err != interfaces.ErrSecretNotFound {
					conflicts = append(conflicts, clientSecret.ID)
				}
			}
		} else {
			existing, err := s.secrets.GetByID(ctx, clientSecret.ID, userID)
			if err != nil && err != interfaces.ErrSecretNotFound {
				conflicts = append(conflicts, clientSecret.ID)
				continue
			}

			if existing == nil {
				clientSecret.Version = 1
				clientSecret.CreatedAt = domain.Now()
				clientSecret.UpdatedAt = domain.Now()
				if err := s.secrets.Create(ctx, clientSecret); err != nil {
					conflicts = append(conflicts, clientSecret.ID)
				}
			} else {
				if existing.Version != clientSecret.Version {
					conflicts = append(conflicts, clientSecret.ID)
				} else {
					clientSecret.Version = existing.Version
					clientSecret.UpdatedAt = domain.Now()
					if err := s.secrets.Update(ctx, clientSecret); err != nil {
						conflicts = append(conflicts, clientSecret.ID)
					}
				}
			}
		}
	}

	return conflicts, nil
}

func (s *DataService) validateSecret(secret *domain.Secret) error {
	if secret.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if secret.Name == "" {
		return fmt.Errorf("secret name is required")
	}

	if len(secret.EncryptedData) == 0 {
		return fmt.Errorf("encrypted data is required")
	}

	if len(secret.EncryptedData) > 10*1024*1024 { // 10MB
		return fmt.Errorf("secret data too large")
	}

	switch secret.Type {
	case domain.LoginPassword, domain.TextData, domain.BinaryData, domain.BankCard:
	default:
		return fmt.Errorf("invalid secret type: %s", secret.Type)
	}

	return nil
}
