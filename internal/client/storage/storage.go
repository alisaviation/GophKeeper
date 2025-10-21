package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alisaviation/GophKeeper/internal/client/domain"
)

// FileStorage файловое хранилище
type FileStorage struct {
	basePath string
}

// NewFileStorage создает новое файловое хранилище
func NewFileStorage(basePath string) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStorage{
		basePath: basePath,
	}, nil
}

// SaveSession сохраняет сессию
func (s *FileStorage) SaveSession(session *domain.Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.sessionPath(), data, 0600)
}

// GetSession получает сессию
func (s *FileStorage) GetSession() (*domain.Session, error) {
	data, err := os.ReadFile(s.sessionPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// DeleteSession удаляет сессию
func (s *FileStorage) DeleteSession() error {
	return os.Remove(s.sessionPath())
}

// SaveSecret сохраняет секрет
func (s *FileStorage) SaveSecret(secret *domain.SecretData) error {
	secrets, err := s.GetSecrets()
	if err != nil {
		return err
	}

	// Обновляем или добавляем секрет
	found := false
	for i, existingSecret := range secrets {
		if existingSecret.ID == secret.ID {
			secrets[i] = secret
			found = true
			break
		}
	}
	if !found {
		secrets = append(secrets, secret)
	}

	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.secretsPath(), data, 0600)
}

// GetSecret получает секрет по ID
func (s *FileStorage) GetSecret(id string) (*domain.SecretData, error) {
	secrets, err := s.GetSecrets()
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets {
		if secret.ID == id {
			return secret, nil
		}
	}

	return nil, fmt.Errorf("secret not found: %s", id)
}

// GetSecrets получает все секреты
func (s *FileStorage) GetSecrets() ([]*domain.SecretData, error) {
	data, err := os.ReadFile(s.secretsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []*domain.SecretData{}, nil
		}
		return nil, err
	}

	var secrets []*domain.SecretData
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

// DeleteSecret удаляет секрет
func (s *FileStorage) DeleteSecret(id string) error {
	secrets, err := s.GetSecrets()
	if err != nil {
		return err
	}

	var filtered []*domain.SecretData
	for _, secret := range secrets {
		if secret.ID != id {
			filtered = append(filtered, secret)
		}
	}

	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.secretsPath(), data, 0600)
}

// Вспомогательные методы путей
func (s *FileStorage) sessionPath() string {
	return filepath.Join(s.basePath, "session.json")
}

func (s *FileStorage) secretsPath() string {
	return filepath.Join(s.basePath, "secrets.json")
}

func (s *FileStorage) configPath() string {
	return filepath.Join(s.basePath, "config.json")
}
