package app

import (
	"context"
	"fmt"

	"github.com/alisaviation/GophKeeper/internal/crypto"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
	"github.com/alisaviation/GophKeeper/internal/server/storage/interfaces"
)

// DataService предоставляет методы для управления секретами
type DataService struct {
	secrets interfaces.SecretRepository
	crypto  crypto.Encryptor
}

// NewDataService создает новый сервис управления данными
func NewDataService(secrets interfaces.SecretRepository, crypto crypto.Encryptor) *DataService {
	return &DataService{
		secrets: secrets,
		crypto:  crypto,
	}
}

// Sync синхронизирует данные между клиентом и сервером
func (s *DataService) Sync(ctx context.Context, userID string, clientSecrets []*domain.Secret, lastSyncVersion int64) (*SyncResult, error) {
	for _, secret := range clientSecrets {
		if secret.UserID != userID {
			return nil, domain.ErrAccessDenied
		}
	}
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
	Conflicts      []string
}

// GetSecret возвращает секрет по ID
func (s *DataService) GetSecret(ctx context.Context, userID, secretID string) (*domain.Secret, error) {
	secret, err := s.secrets.GetByID(ctx, secretID, userID)
	if err != nil {
		return nil, domain.ErrSecretNotFound
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

	err := s.validateSecret(secret)
	if err != nil {
		return domain.ErrInvalidSecret
	}

	err = s.secrets.Create(ctx, secret)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// UpdateSecret обновляет существующий секрет
func (s *DataService) UpdateSecret(ctx context.Context, userID string, secret *domain.Secret) error {

	existing, err := s.secrets.GetByID(ctx, secret.ID, userID)
	if err != nil {
		return domain.ErrSecretNotFound
	}

	if existing.Version != secret.Version {
		return domain.ErrVersionConflict
	}

	secret.UserID = userID
	secret.UpdatedAt = domain.Now()

	if err := s.validateSecret(secret); err != nil {
		return domain.ErrInvalidSecret
	}

	if err := s.secrets.Update(ctx, secret); err != nil {
		return domain.ErrVersionConflict
	}

	return nil
}

// DeleteSecret удаляет секрет
func (s *DataService) DeleteSecret(ctx context.Context, userID, secretID string) error {
	err := s.secrets.SoftDelete(ctx, secretID, userID)
	if err != nil {
		return domain.ErrSecretNotFound
	}
	return nil
}

func (s *DataService) processClientChanges(ctx context.Context, userID string, clientSecrets []*domain.Secret) ([]string, error) {
	var conflicts []string

	for _, clientSecret := range clientSecrets {
		if clientSecret.UserID != userID {
			return nil, fmt.Errorf("access denied for secret %s: %w", clientSecret.ID, domain.ErrAccessDenied)
		}

		if clientSecret.IsDeleted {
			err := s.secrets.SoftDelete(ctx, clientSecret.ID, userID)
			if err != nil && err != domain.ErrSecretNotFound {
				conflicts = append(conflicts, clientSecret.ID)
			}
		} else {
			existing, err := s.secrets.GetByID(ctx, clientSecret.ID, userID)
			if err != nil && err != domain.ErrSecretNotFound {
				conflicts = append(conflicts, clientSecret.ID)
				continue
			}

			if existing == nil {
				clientSecret.Version = 1
				clientSecret.CreatedAt = domain.Now()
				clientSecret.UpdatedAt = domain.Now()
				err = s.secrets.Create(ctx, clientSecret)
				if err != nil {
					conflicts = append(conflicts, clientSecret.ID)
				}
			} else {
				if existing.Version != clientSecret.Version {
					conflicts = append(conflicts, clientSecret.ID)
				} else {
					clientSecret.Version = existing.Version
					clientSecret.UpdatedAt = domain.Now()
					err = s.secrets.Update(ctx, clientSecret)
					if err != nil {
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
		return &domain.ValidationError{
			Field:   "user_id",
			Message: "is required",
		}
	}

	if secret.Name == "" {
		return &domain.ValidationError{
			Field:   "name",
			Message: "is required",
		}
	}

	if len(secret.EncryptedData) == 0 {
		return &domain.ValidationError{
			Field:   "encrypted_data",
			Message: "is required",
		}
	}

	if len(secret.EncryptedData) > 10*1024*1024 {
		return &domain.ValidationError{
			Field:   "encrypted_data",
			Message: "too large",
		}
	}

	switch secret.Type {
	case domain.LoginPassword, domain.TextData, domain.BinaryData, domain.BankCard:
	default:
		return fmt.Errorf("invalid secret type: %w", domain.ErrInvalidSecretType)
	}

	return nil
}
