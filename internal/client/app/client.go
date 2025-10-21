// internal/client/app/client.go
package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alisaviation/GophKeeper/internal/client/domain"
	"github.com/alisaviation/GophKeeper/internal/client/storage"
	"github.com/alisaviation/GophKeeper/internal/client/transport"
	"github.com/alisaviation/GophKeeper/internal/crypto"
	pb "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
)

// Client основное клиентское приложение
type Client struct {
	storage   *storage.FileStorage
	transport *transport.GRPCClient
}

// NewClient создает новый клиент
func NewClient(storage *storage.FileStorage, transport *transport.GRPCClient) *Client {
	return &Client{
		storage:   storage,
		transport: transport,
	}
}

// SyncResult результат синхронизации
type SyncResult struct {
	Uploaded   int
	Downloaded int
	Conflicts  []string
}

// SecretDisplay отображаемый секрет
type SecretDisplay struct {
	ID        string
	Name      string
	Type      string
	Data      interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Register регистрирует нового пользователя
func (c *Client) Register(ctx context.Context, login, password string) (string, error) {
	userID, err := c.transport.Register(ctx, login, password)
	if err != nil {
		return "", fmt.Errorf("registration failed: %w", err)
	}

	// Генерируем ключ шифрования
	encryptionKey, err := generateEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Сохраняем начальную сессию
	session := &domain.Session{
		UserID:        userID,
		Login:         login,
		LastSync:      time.Now().Unix(),
		EncryptionKey: encryptionKey,
	}

	if err := c.storage.SaveSession(session); err != nil {
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	return userID, nil
}

// Login выполняет аутентификацию
func (c *Client) Login(ctx context.Context, login, password string) error {
	accessToken, refreshToken, userID, err := c.transport.Login(ctx, login, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Получаем или создаем сессию
	session, err := c.storage.GetSession()
	if err != nil || session == nil {
		// Создаем новую сессию
		encryptionKey, err := generateEncryptionKey()
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}

		session = &domain.Session{
			UserID:        userID,
			Login:         login,
			EncryptionKey: encryptionKey,
		}
	}

	// Обновляем токены
	session.AccessToken = accessToken
	session.RefreshToken = refreshToken
	session.LastSync = time.Now().Unix()

	if err := c.storage.SaveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Устанавливаем токен в транспорт
	c.transport.SetToken(accessToken)

	return nil
}

// Logout выходит из системы
func (c *Client) Logout() error {
	session, err := c.storage.GetSession()
	if err != nil {
		return fmt.Errorf("no active session: %w", err)
	}

	// Вызываем logout на сервере
	if err := c.transport.Logout(context.Background(), session.RefreshToken); err != nil {
		fmt.Printf("Warning: logout call failed: %v\n", err)
	}

	// Удаляем локальную сессию
	if err := c.storage.DeleteSession(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	c.transport.SetToken("")
	return nil
}

// GetSession возвращает текущую сессию
func (c *Client) GetSession() (*domain.Session, error) {
	return c.storage.GetSession()
}

// CreateSecret создает новый секрет
func (c *Client) CreateSecret(ctx context.Context, secretData *domain.SecretData) (string, error) {
	// Проверяем аутентификацию
	session, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return "", err
	}

	// Устанавливаем метаданные
	secretData.ID = domain.GenerateID()
	secretData.UserID = session.UserID
	secretData.CreatedAt = time.Now()
	secretData.UpdatedAt = time.Now()
	secretData.IsDirty = true

	// Сохраняем локально
	if err := c.storage.SaveSecret(secretData); err != nil {
		return "", fmt.Errorf("failed to save secret locally: %w", err)
	}

	return secretData.ID, nil
}

// GetSecret получает секрет по ID
func (c *Client) GetSecret(ctx context.Context, id string) (*SecretDisplay, error) {
	secret, err := c.storage.GetSecret(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return &SecretDisplay{
		ID:        secret.ID,
		Name:      secret.Name,
		Type:      string(secret.Type),
		Data:      secret.Data,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}, nil
}

// ListSecrets возвращает список секретов
func (c *Client) ListSecrets(ctx context.Context, filterType string) ([]*SecretDisplay, error) {
	secrets, err := c.storage.GetSecrets()
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	var result []*SecretDisplay
	for _, secret := range secrets {
		if filterType == "" || string(secret.Type) == filterType {
			result = append(result, &SecretDisplay{
				ID:        secret.ID,
				Name:      secret.Name,
				Type:      string(secret.Type),
				CreatedAt: secret.CreatedAt,
				UpdatedAt: secret.UpdatedAt,
			})
		}
	}

	return result, nil
}

// DeleteSecret удаляет секрет
func (c *Client) DeleteSecret(ctx context.Context, id string) error {
	// Помечаем секрет как удаленный локально
	secret, err := c.storage.GetSecret(id)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	secret.IsDeleted = true
	secret.IsDirty = true
	secret.UpdatedAt = time.Now()

	if err := c.storage.SaveSecret(secret); err != nil {
		return fmt.Errorf("failed to mark secret as deleted: %w", err)
	}

	return nil
}

// Sync синхронизирует данные с сервером
func (c *Client) Sync(ctx context.Context, force bool, resolveStrategy string) (*SyncResult, error) {
	session, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	// Получаем локальные секреты
	localSecrets, err := c.storage.GetSecrets()
	if err != nil {
		return nil, fmt.Errorf("failed to get local secrets: %w", err)
	}

	// Подготавливаем секреты для отправки
	var secretsToSync []*pb.Secret
	for _, localSecret := range localSecrets {
		if localSecret.IsDirty || force {
			encryptedSecret, err := c.encryptSecret(localSecret)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt secret %s: %w", localSecret.ID, err)
			}
			secretsToSync = append(secretsToSync, encryptedSecret)
		}
	}

	// Выполняем синхронизацию
	syncResponse, err := c.transport.Sync(ctx, session.UserID, session.LastSyncVersion, secretsToSync)
	if err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}

	// Обрабатываем полученные секреты
	downloaded := 0
	var conflicts []string

	for _, serverSecret := range syncResponse.Secrets {
		decryptedSecret, err := c.decryptSecret(serverSecret)
		if err != nil {
			fmt.Printf("Warning: failed to decrypt secret %s: %v\n", serverSecret.Id, err)
			continue
		}

		// Проверяем конфликты
		localSecret, err := c.storage.GetSecret(decryptedSecret.ID)
		if err == nil && localSecret.Version != decryptedSecret.Version {
			conflicts = append(conflicts, fmt.Sprintf("%s: version conflict (local=%d, server=%d)",
				decryptedSecret.ID, localSecret.Version, decryptedSecret.Version))

			if resolveStrategy == "server" {
				// Используем версию с сервера
				if err := c.storage.SaveSecret(decryptedSecret); err != nil {
					fmt.Printf("Warning: failed to save secret %s: %v\n", decryptedSecret.ID, err)
				} else {
					downloaded++
				}
			}
		} else {
			// Сохраняем секрет
			if err := c.storage.SaveSecret(decryptedSecret); err != nil {
				fmt.Printf("Warning: failed to save secret %s: %v\n", decryptedSecret.ID, err)
			} else {
				downloaded++
			}
		}
	}

	// Обновляем информацию о сессии
	session.LastSync = time.Now().Unix()
	session.LastSyncVersion = syncResponse.CurrentVersion
	if err := c.storage.SaveSession(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &SyncResult{
		Uploaded:   len(secretsToSync),
		Downloaded: downloaded,
		Conflicts:  conflicts,
	}, nil
}

// ensureAuthenticated проверяет аутентификацию
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

// encryptSecret шифрует секрет для отправки на сервер
func (c *Client) encryptSecret(secret *domain.SecretData) (*pb.Secret, error) {
	session, err := c.storage.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Создаем шифровальщик с ключом из сессии
	encryptor, err := crypto.NewAESGCMEncryptor(session.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	// Сериализуем данные
	dataBytes, err := json.Marshal(secret.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	// Шифруем данные
	encryptedData, err := encryptor.Encrypt(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Создаем protobuf секрет
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

// decryptSecret расшифровывает секрет с сервера
func (c *Client) decryptSecret(pbSecret *pb.Secret) (*domain.SecretData, error) {
	session, err := c.storage.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Создаем шифровальщик с ключом из сессии
	encryptor, err := crypto.NewAESGCMEncryptor(session.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	// Расшифровываем данные
	decryptedData, err := encryptor.Decrypt(pbSecret.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Определяем тип данных и десериализуем
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

// Вспомогательные функции
func generateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
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
		return domain.SecretTypeLoginPassword // fallback
	}
}
