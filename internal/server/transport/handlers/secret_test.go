package handlers_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
	pb "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
	"github.com/alisaviation/GophKeeper/internal/server/transport/handlers"
	"github.com/alisaviation/GophKeeper/internal/server/transport/middleware"
)

// testContextWithUser создает контекст с тестовым пользователем
func testContextWithUser(user *domain.User) context.Context {
	return context.WithValue(context.Background(), middleware.UserContextKey{}, user)
}

func TestSecretHandler_Sync(t *testing.T) {
	mockSecretRepo := app.NewMockSecretRepository()
	mockEncryptor := &app.MockEncryptor{}
	dataService := app.NewDataService(mockSecretRepo, mockEncryptor)
	handler := handlers.NewSecretHandler(dataService)

	user := &domain.User{
		ID:    "test-user-id",
		Login: "testuser",
	}

	tests := []struct {
		name          string
		request       *pb.SyncRequest
		wantError     bool
		errorCode     codes.Code
		wantConflicts bool
		setupMock     func()
		description   string
	}{
		{
			name: "successful sync with no changes",
			request: &pb.SyncRequest{
				LastSyncVersion: 0,
				Secrets:         []*pb.Secret{},
			},
			wantError:     false,
			wantConflicts: false,
			description:   "sync with empty secrets should succeed",
		},
		{
			name: "successful sync with new secrets",
			request: &pb.SyncRequest{
				LastSyncVersion: 0,
				Secrets: []*pb.Secret{
					{
						Id:            "secret-1",
						UserId:        user.ID,
						Type:          pb.SecretType_LOGIN_PASSWORD,
						Name:          "Test Secret",
						EncryptedData: []byte("encrypted data"),
						Version:       1,
					},
				},
			},
			wantError:     false,
			wantConflicts: false,
			description:   "sync with own secrets should succeed",
		},
		{
			name: "sync with foreign user secret",
			request: &pb.SyncRequest{
				LastSyncVersion: 0,
				Secrets: []*pb.Secret{
					{
						Id:            "foreign-secret",
						UserId:        "other-user",
						Type:          pb.SecretType_LOGIN_PASSWORD,
						Name:          "Foreign Secret",
						EncryptedData: []byte("encrypted data"),
						Version:       1,
					},
				},
			},
			wantError:     true,
			errorCode:     codes.PermissionDenied,
			wantConflicts: false,
			description:   "sync with foreign user secret should return PermissionDenied",
		},
		{
			name: "sync with version conflict",
			request: &pb.SyncRequest{
				LastSyncVersion: 0,
				Secrets: []*pb.Secret{
					{
						Id:            "conflict-secret",
						UserId:        user.ID,
						Type:          pb.SecretType_LOGIN_PASSWORD,
						Name:          "Updated Secret",
						EncryptedData: []byte("updated data"),
						Version:       1, // Устаревшая версия
					},
				},
			},
			wantError:     false,
			wantConflicts: true,
			setupMock: func() {
				existingSecret := &domain.Secret{
					ID:            "conflict-secret",
					UserID:        user.ID,
					Type:          domain.LoginPassword,
					Name:          "Existing Secret",
					EncryptedData: []byte("existing data"),
					Version:       2,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				mockSecretRepo.Secrets[existingSecret.ID] = existingSecret
				mockSecretRepo.UserSecret[user.ID] = []string{existingSecret.ID}
				mockSecretRepo.Versions[existingSecret.ID] = existingSecret.Version
			},
			description: "sync with version conflict should return conflicts list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSecretRepo.Secrets = make(map[string]*domain.Secret)
			mockSecretRepo.UserSecret = make(map[string][]string)
			mockSecretRepo.Versions = make(map[string]int64)
			mockSecretRepo.Deleted = make(map[string]bool)

			if tt.setupMock != nil {
				tt.setupMock()
			}

			ctx := testContextWithUser(user)
			resp, err := handler.Sync(ctx, tt.request)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil. Test: %s", tt.description)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error, got: %v. Test: %s", err, tt.description)
					return
				}
				if st.Code() != tt.errorCode {
					t.Errorf("Expected code %v, got %v. Message: %s. Test: %s",
						tt.errorCode, st.Code(), st.Message(), tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Test: %s", err, tt.description)
					return
				}
				if resp == nil {
					t.Error("Expected non-nil response")
					return
				}

				if tt.wantConflicts {
					if len(resp.Conflicts) == 0 {
						t.Errorf("Expected conflicts, but got none. Test: %s", tt.description)
					} else if !contains(resp.Conflicts, "conflict-secret") {
						t.Errorf("Expected conflict for secret 'conflict-secret', got conflicts: %v. Test: %s",
							resp.Conflicts, tt.description)
					} else {
						t.Logf("✓ Correctly detected version conflict for secret: %v", resp.Conflicts)
					}
				} else {
					if len(resp.Conflicts) > 0 {
						t.Errorf("Unexpected conflicts: %v. Test: %s", resp.Conflicts, tt.description)
					}
				}
			}
		})
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestSecretHandler_GetSecret(t *testing.T) {
	mockSecretRepo := app.NewMockSecretRepository()
	mockEncryptor := &app.MockEncryptor{}
	dataService := app.NewDataService(mockSecretRepo, mockEncryptor)
	handler := handlers.NewSecretHandler(dataService)

	user := &domain.User{
		ID:    "test-user-id",
		Login: "testuser",
	}

	// Setup test secret
	testSecret := &domain.Secret{
		ID:            "test-secret-id",
		UserID:        user.ID,
		Type:          domain.LoginPassword,
		Name:          "Test Secret",
		EncryptedData: []byte("encrypted data"),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	tests := []struct {
		name      string
		request   *pb.GetSecretRequest
		wantError bool
		errorCode codes.Code
		setupMock func()
	}{
		{
			name: "successful get secret",
			request: &pb.GetSecretRequest{
				SecretId: "test-secret-id",
			},
			wantError: false,
			setupMock: func() {
				mockSecretRepo.Secrets[testSecret.ID] = testSecret
				mockSecretRepo.UserSecret[user.ID] = []string{testSecret.ID}
			},
		},
		{
			name: "secret not found",
			request: &pb.GetSecretRequest{
				SecretId: "nonexistent-secret",
			},
			wantError: true,
			errorCode: codes.NotFound,
			setupMock: func() {
				// Не добавляем секрет
			},
		},
		{
			name: "empty secret id",
			request: &pb.GetSecretRequest{
				SecretId: "",
			},
			wantError: true,
			errorCode: codes.InvalidArgument,
			setupMock: func() {
				// Ничего не настраиваем
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockSecretRepo.Secrets = make(map[string]*domain.Secret)
			mockSecretRepo.UserSecret = make(map[string][]string)

			if tt.setupMock != nil {
				tt.setupMock()
			}

			ctx := testContextWithUser(user)
			resp, err := handler.GetSecret(ctx, tt.request)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errorCode {
						t.Errorf("Expected code %v, got %v", tt.errorCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if resp.Secret == nil {
					t.Error("Expected non-nil secret")
				}
				if resp.Secret.Id != testSecret.ID {
					t.Errorf("Expected secret ID %s, got %s", testSecret.ID, resp.Secret.Id)
				}
			}
		})
	}
}

func TestSecretHandler_ListSecrets(t *testing.T) {
	mockSecretRepo := app.NewMockSecretRepository()
	mockEncryptor := &app.MockEncryptor{}
	dataService := app.NewDataService(mockSecretRepo, mockEncryptor)
	handler := handlers.NewSecretHandler(dataService)

	user := &domain.User{
		ID:    "test-user-id",
		Login: "testuser",
	}

	// Setup test secrets
	secrets := []*domain.Secret{
		{
			ID:            "secret-1",
			UserID:        user.ID,
			Type:          domain.LoginPassword,
			Name:          "Secret 1",
			EncryptedData: []byte("data1"),
			Version:       1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "secret-2",
			UserID:        user.ID,
			Type:          domain.TextData,
			Name:          "Secret 2",
			EncryptedData: []byte("data2"),
			Version:       1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	tests := []struct {
		name      string
		request   *pb.ListSecretsRequest
		wantCount int
		wantError bool
		setupMock func()
	}{
		{
			name:      "list all secrets",
			request:   &pb.ListSecretsRequest{},
			wantCount: 2,
			wantError: false,
			setupMock: func() {
				for _, secret := range secrets {
					mockSecretRepo.Secrets[secret.ID] = secret
					mockSecretRepo.UserSecret[user.ID] = append(mockSecretRepo.UserSecret[user.ID], secret.ID)
				}
			},
		},
		{
			name: "list filtered by type",
			request: &pb.ListSecretsRequest{
				FilterType: pb.SecretType_LOGIN_PASSWORD,
			},
			wantCount: 1,
			wantError: false,
			setupMock: func() {
				for _, secret := range secrets {
					mockSecretRepo.Secrets[secret.ID] = secret
					mockSecretRepo.UserSecret[user.ID] = append(mockSecretRepo.UserSecret[user.ID], secret.ID)
				}
			},
		},
		{
			name:      "list no secrets",
			request:   &pb.ListSecretsRequest{},
			wantCount: 0,
			wantError: false,
			setupMock: func() {
				// Не добавляем секреты
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockSecretRepo.Secrets = make(map[string]*domain.Secret)
			mockSecretRepo.UserSecret = make(map[string][]string)

			if tt.setupMock != nil {
				tt.setupMock()
			}

			ctx := testContextWithUser(user)
			resp, err := handler.ListSecrets(ctx, tt.request)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if len(resp.Secrets) != tt.wantCount {
					t.Errorf("Expected %d secrets, got %d", tt.wantCount, len(resp.Secrets))
				}
			}
		})
	}
}
