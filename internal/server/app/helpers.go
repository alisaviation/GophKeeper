package app

import (
	"context"
	"fmt"

	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

func validateCredentials(login, password string) error {
	if len(login) < 3 || len(login) > 50 {
		return &domain.ValidationError{
			Field:   "login",
			Message: "must be between 3 and 50 characters",
		}
	}

	if len(password) < 8 {
		return &domain.ValidationError{
			Field:   "password",
			Message: "must be at least 8 characters",
		}
	}

	for _, char := range login {
		if !isValidLoginChar(char) {
			return &domain.ValidationError{
				Field:   "login",
				Message: "can only contain letters, numbers and underscores",
			}
		}
	}

	return nil
}

func isValidLoginChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_'
}

func generateUserID() string {
	return domain.GenerateID()
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
