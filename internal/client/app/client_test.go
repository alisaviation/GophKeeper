package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/alisaviation/GophKeeper/internal/client/domain"
	pb "github.com/alisaviation/GophKeeper/internal/generated/grpc"
)

func TestClient_Register(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		password    string
		setupMocks  func(*MockStorage, *MockTransport)
		expectError bool
	}{
		{
			name:     "successful registration",
			login:    "testuser",
			password: "testpass",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				mt.On("Register", mock.Anything, "testuser", "testpass").
					Return("user123", nil)
				ms.On("SaveSession", mock.AnythingOfType("*domain.Session")).
					Run(func(args mock.Arguments) {
						session := args.Get(0).(*domain.Session)
						assert.Equal(t, "user123", session.UserID)
						assert.Equal(t, "testuser", session.Login)
						assert.NotEmpty(t, session.EncryptionKey)
					}).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:     "registration failed on transport",
			login:    "testuser",
			password: "testpass",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				mt.On("Register", mock.Anything, "testuser", "testpass").
					Return("", errors.New("transport error"))
			},
			expectError: true,
		},
		{
			name:     "failed to save session",
			login:    "testuser",
			password: "testpass",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				mt.On("Register", mock.Anything, "testuser", "testpass").
					Return("user123", nil)
				ms.On("SaveSession", mock.AnythingOfType("*domain.Session")).
					Return(errors.New("storage error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockTransport := &MockTransport{}
			tt.setupMocks(mockStorage, mockTransport)

			client := NewClient(mockStorage, mockTransport)
			userID, err := client.Register(context.Background(), tt.login, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "user123", userID)
			}

			mockStorage.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestClient_Login(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		password    string
		setupMocks  func(*MockStorage, *MockTransport)
		expectError bool
	}{
		{
			name:     "successful login with new session",
			login:    "testuser",
			password: "testpass",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				ms.On("GetSession").Return(nil, errors.New("no session"))
				mt.On("Login", mock.Anything, "testuser", "testpass").
					Return("access123", "refresh123", "user123", nil)
				ms.On("SaveSession", mock.AnythingOfType("*domain.Session")).
					Run(func(args mock.Arguments) {
						session := args.Get(0).(*domain.Session)
						assert.Equal(t, "user123", session.UserID)
						assert.Equal(t, "testuser", session.Login)
						assert.Equal(t, "access123", session.AccessToken)
						assert.Equal(t, "refresh123", session.RefreshToken)
						assert.NotEmpty(t, session.EncryptionKey)
					}).
					Return(nil)
				mt.On("SetToken", "access123").Once()
			},
			expectError: false,
		},
		{
			name:     "successful login with existing session",
			login:    "testuser",
			password: "testpass",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				existingSession := &domain.Session{
					UserID:        "user123",
					Login:         "testuser",
					EncryptionKey: []byte("existingkey"),
				}
				ms.On("GetSession").Return(existingSession, nil)
				mt.On("Login", mock.Anything, "testuser", "testpass").
					Return("access123", "refresh123", "user123", nil)
				ms.On("SaveSession", mock.AnythingOfType("*domain.Session")).
					Run(func(args mock.Arguments) {
						session := args.Get(0).(*domain.Session)
						assert.Equal(t, "user123", session.UserID)
						assert.Equal(t, "testuser", session.Login)
						assert.Equal(t, "access123", session.AccessToken)
						assert.Equal(t, "refresh123", session.RefreshToken)
						assert.Equal(t, []byte("existingkey"), session.EncryptionKey)
					}).
					Return(nil)
				mt.On("SetToken", "access123").Once()
			},
			expectError: false,
		},
		{
			name:     "login failed on transport",
			login:    "testuser",
			password: "testpass",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				ms.On("GetSession").Return(nil, errors.New("no session")).Maybe()
				mt.On("Login", mock.Anything, "testuser", "testpass").
					Return("", "", "", errors.New("auth failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockTransport := &MockTransport{}
			tt.setupMocks(mockStorage, mockTransport)

			client := NewClient(mockStorage, mockTransport)
			err := client.Login(context.Background(), tt.login, tt.password)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStorage.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestClient_Logout(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockStorage, *MockTransport)
		expectError bool
	}{
		{
			name: "successful logout",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				session := &domain.Session{
					RefreshToken: "refresh123",
				}
				ms.On("GetSession").Return(session, nil)
				mt.On("Logout", mock.Anything, "refresh123").Return(nil)
				ms.On("DeleteSession").Return(nil)
				mt.On("SetToken", "").Once()
			},
			expectError: false,
		},
		{
			name: "logout with transport error",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				session := &domain.Session{
					RefreshToken: "refresh123",
				}
				ms.On("GetSession").Return(session, nil)
				mt.On("Logout", mock.Anything, "refresh123").Return(errors.New("network error"))
				ms.On("DeleteSession").Return(nil)
				mt.On("SetToken", "").Once()
			},
			expectError: false,
		},
		{
			name: "no active session",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				ms.On("GetSession").Return(nil, errors.New("no session"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockTransport := &MockTransport{}
			tt.setupMocks(mockStorage, mockTransport)

			client := NewClient(mockStorage, mockTransport)
			err := client.Logout()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStorage.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestClient_CreateSecret(t *testing.T) {
	tests := []struct {
		name        string
		secretData  *domain.SecretData
		setupMocks  func(*MockStorage, *MockTransport)
		expectError bool
	}{
		{
			name: "successful secret creation",
			secretData: &domain.SecretData{
				Name: "test secret",
				Type: domain.SecretTypeLoginPassword,
				Data: domain.LoginPasswordData{
					Login:    "user",
					Password: "pass",
				},
			},
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				session := &domain.Session{
					UserID:        "user123",
					AccessToken:   "access123",
					EncryptionKey: []byte("testkey123456789012345678901234"),
				}
				ms.On("GetSession").Return(session, nil)
				mt.On("SetToken", "access123").Once()
				ms.On("SaveSecret", mock.AnythingOfType("*domain.SecretData")).
					Run(func(args mock.Arguments) {
						secret := args.Get(0).(*domain.SecretData)
						assert.NotEmpty(t, secret.ID)
						assert.Equal(t, "user123", secret.UserID)
						assert.Equal(t, "test secret", secret.Name)
						assert.True(t, secret.IsDirty)
						assert.False(t, secret.IsDeleted)
					}).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "not authenticated",
			secretData: &domain.SecretData{
				Name: "test secret",
				Type: domain.SecretTypeText,
				Data: domain.TextData{Content: "test text"},
			},
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				ms.On("GetSession").Return(nil, errors.New("not authenticated"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockTransport := &MockTransport{}
			tt.setupMocks(mockStorage, mockTransport)

			client := NewClient(mockStorage, mockTransport)
			secretID, err := client.CreateSecret(context.Background(), tt.secretData)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, secretID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, secretID)
			}

			mockStorage.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestClient_GetSecret(t *testing.T) {
	tests := []struct {
		name        string
		secretID    string
		setupMocks  func(*MockStorage, *MockTransport)
		expectError bool
	}{
		{
			name:     "successful get secret",
			secretID: "secret123",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				secret := &domain.SecretData{
					ID:        "secret123",
					Name:      "test secret",
					Type:      domain.SecretTypeLoginPassword,
					Data:      domain.LoginPasswordData{Login: "user", Password: "pass"},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				ms.On("GetSecret", "secret123").Return(secret, nil)
			},
			expectError: false,
		},
		{
			name:     "secret not found",
			secretID: "nonexistent",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				ms.On("GetSecret", "nonexistent").Return(nil, errors.New("not found"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockTransport := &MockTransport{}
			tt.setupMocks(mockStorage, mockTransport)

			client := NewClient(mockStorage, mockTransport)
			result, err := client.GetSecret(context.Background(), tt.secretID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "secret123", result.ID)
				assert.Equal(t, "test secret", result.Name)
			}

			mockStorage.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestClient_ListSecrets(t *testing.T) {
	tests := []struct {
		name        string
		filterType  string
		setupMocks  func(*MockStorage, *MockTransport)
		expectCount int
		expectError bool
	}{
		{
			name:       "list all secrets",
			filterType: "",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				secrets := []*domain.SecretData{
					{
						ID:        "1",
						Name:      "secret1",
						Type:      domain.SecretTypeLoginPassword,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					{
						ID:        "2",
						Name:      "secret2",
						Type:      domain.SecretTypeText,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				ms.On("GetSecrets").Return(secrets, nil)
			},
			expectCount: 2,
			expectError: false,
		},
		{
			name:       "filter by type",
			filterType: "login_password",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				secrets := []*domain.SecretData{
					{
						ID:        "1",
						Name:      "secret1",
						Type:      domain.SecretTypeLoginPassword,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					{
						ID:        "2",
						Name:      "secret2",
						Type:      domain.SecretTypeText,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				ms.On("GetSecrets").Return(secrets, nil)
			},
			expectCount: 1,
			expectError: false,
		},
		{
			name:       "storage error",
			filterType: "",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				ms.On("GetSecrets").Return(nil, errors.New("storage error"))
			},
			expectCount: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockTransport := &MockTransport{}
			tt.setupMocks(mockStorage, mockTransport)

			client := NewClient(mockStorage, mockTransport)
			result, err := client.ListSecrets(context.Background(), tt.filterType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectCount)
			}

			mockStorage.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestClient_DeleteSecret(t *testing.T) {
	tests := []struct {
		name        string
		secretID    string
		setupMocks  func(*MockStorage, *MockTransport)
		expectError bool
	}{
		{
			name:     "successful delete",
			secretID: "secret123",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				secret := &domain.SecretData{
					ID:   "secret123",
					Name: "test secret",
				}
				ms.On("GetSecret", "secret123").Return(secret, nil)
				ms.On("SaveSecret", mock.AnythingOfType("*domain.SecretData")).
					Run(func(args mock.Arguments) {
						secret := args.Get(0).(*domain.SecretData)
						assert.True(t, secret.IsDeleted)
						assert.True(t, secret.IsDirty)
					}).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:     "secret not found",
			secretID: "nonexistent",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				ms.On("GetSecret", "nonexistent").Return(nil, errors.New("not found"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockTransport := &MockTransport{}
			tt.setupMocks(mockStorage, mockTransport)

			client := NewClient(mockStorage, mockTransport)
			err := client.DeleteSecret(context.Background(), tt.secretID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStorage.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestClient_Sync(t *testing.T) {
	tests := []struct {
		name            string
		force           bool
		resolveStrategy string
		setupMocks      func(*MockStorage, *MockTransport)
		expectError     bool
	}{
		{
			name:            "successful sync with no conflicts",
			force:           false,
			resolveStrategy: "server",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				encryptionKey := make([]byte, 32)
				copy(encryptionKey, "testkey12345678901234567890123456")

				session := &domain.Session{
					UserID:          "user123",
					AccessToken:     "access123",
					LastSyncVersion: 1,
					EncryptionKey:   encryptionKey,
				}
				ms.On("GetSession").Return(session, nil).Maybe()
				mt.On("SetToken", "access123").Maybe()

				localSecrets := []*domain.SecretData{
					{
						ID:        "secret1",
						UserID:    "user123",
						Name:      "test1",
						Type:      domain.SecretTypeLoginPassword,
						Data:      domain.LoginPasswordData{Login: "user1", Password: "pass1"},
						Version:   1,
						IsDirty:   true,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				ms.On("GetSecrets").Return(localSecrets, nil).Once()

				syncResponse := &pb.SyncResponse{
					CurrentVersion: 2,
					Secrets:        []*pb.Secret{},
				}

				mt.On("Sync",
					mock.Anything,
					"user123",
					int64(1),
					mock.MatchedBy(func(secrets []*pb.Secret) bool {
						return len(secrets) > 0 && secrets[0].Id == "secret1"
					})).
					Return(syncResponse, nil).Once()

				ms.On("SaveSession", mock.AnythingOfType("*domain.Session")).
					Run(func(args mock.Arguments) {
						session := args.Get(0).(*domain.Session)
						assert.Equal(t, int64(2), session.LastSyncVersion)
					}).
					Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:            "not authenticated",
			force:           false,
			resolveStrategy: "server",
			setupMocks: func(ms *MockStorage, mt *MockTransport) {
				ms.On("GetSession").Return(nil, errors.New("not authenticated"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockTransport := &MockTransport{}
			tt.setupMocks(mockStorage, mockTransport)

			client := NewClient(mockStorage, mockTransport)
			result, err := client.Sync(context.Background(), tt.force, tt.resolveStrategy)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockStorage.AssertExpectations(t)
			mockTransport.AssertExpectations(t)
		})
	}
}

func TestClient_GetSession(t *testing.T) {
	mockStorage := &MockStorage{}
	mockTransport := &MockTransport{}

	expectedSession := &domain.Session{
		UserID: "user123",
		Login:  "testuser",
	}

	mockStorage.On("GetSession").Return(expectedSession, nil)

	client := NewClient(mockStorage, mockTransport)
	session, err := client.GetSession()

	assert.NoError(t, err)
	assert.Equal(t, expectedSession, session)
	mockStorage.AssertExpectations(t)
}

func TestMapSecretTypeFunctions(t *testing.T) {
	tests := []struct {
		name       string
		domainType domain.SecretType
		protoType  pb.SecretType
	}{
		{
			name:       "login password",
			domainType: domain.SecretTypeLoginPassword,
			protoType:  pb.SecretType_LOGIN_PASSWORD,
		},
		{
			name:       "text data",
			domainType: domain.SecretTypeText,
			protoType:  pb.SecretType_TEXT_DATA,
		},
		{
			name:       "bank card",
			domainType: domain.SecretTypeBankCard,
			protoType:  pb.SecretType_BANK_CARD,
		},
		{
			name:       "unspecified",
			domainType: domain.SecretType("unknown"),
			protoType:  pb.SecretType_SECRET_TYPE_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protoResult := mapSecretTypeToProto(tt.domainType)
			assert.Equal(t, tt.protoType, protoResult)

			domainResult := mapSecretTypeFromProto(tt.protoType)
			if tt.protoType == pb.SecretType_SECRET_TYPE_UNSPECIFIED {
				assert.Equal(t, domain.SecretTypeLoginPassword, domainResult)
			} else {
				assert.Equal(t, tt.domainType, domainResult)
			}
		})
	}
}

func TestGenerateEncryptionKey(t *testing.T) {
	key, err := generateEncryptionKey()

	assert.NoError(t, err)
	assert.Len(t, key, 32)
	assert.NotEmpty(t, key)
}

func TestEncryptDecryptSecret(t *testing.T) {
	mockStorage := &MockStorage{}
	mockTransport := &MockTransport{}

	encryptionKey, err := generateEncryptionKey()
	require.NoError(t, err)

	session := &domain.Session{
		UserID:        "user123",
		EncryptionKey: encryptionKey,
	}

	secret := &domain.SecretData{
		ID:        "secret123",
		UserID:    "user123",
		Name:      "test secret",
		Type:      domain.SecretTypeLoginPassword,
		Data:      domain.LoginPasswordData{Login: "testuser", Password: "testpass"},
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockStorage.On("GetSession").Return(session, nil)

	client := NewClient(mockStorage, mockTransport)

	encrypted, err := client.encryptSecret(secret)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)

	decrypted, err := client.decryptSecret(encrypted)
	assert.NoError(t, err)
	assert.NotNil(t, decrypted)

	assert.Equal(t, secret.ID, decrypted.ID)
	assert.Equal(t, secret.Name, decrypted.Name)
	assert.Equal(t, secret.Type, decrypted.Type)

	decryptedData, ok := decrypted.Data.(domain.LoginPasswordData)
	assert.True(t, ok)
	originalData, ok := secret.Data.(domain.LoginPasswordData)
	assert.True(t, ok)
	assert.Equal(t, originalData.Login, decryptedData.Login)
	assert.Equal(t, originalData.Password, decryptedData.Password)

	mockStorage.AssertExpectations(t)
}
