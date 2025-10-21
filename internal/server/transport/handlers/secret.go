package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
	pb "github.com/alisaviation/GophKeeper/internal/server/transport/grpc"
	"github.com/alisaviation/GophKeeper/internal/server/transport/middleware"
)

// SecretHandler обработчик gRPC для работы с секретами
type SecretHandler struct {
	pb.UnimplementedSecretServiceServer
	dataService *app.DataService
}

// NewSecretHandler создает новый обработчик секретов
func NewSecretHandler(dataService *app.DataService) *SecretHandler {
	return &SecretHandler{
		dataService: dataService,
	}
}

// Sync синхронизирует данные между клиентом и сервером
func (h *SecretHandler) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	clientSecrets := make([]*domain.Secret, 0, len(req.GetSecrets()))
	for _, pbSecret := range req.GetSecrets() {
		clientSecrets = append(clientSecrets, domain.SecretFromProto(pbSecret))
	}

	result, err := h.dataService.Sync(ctx, user.ID, clientSecrets, req.GetLastSyncVersion())
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	serverSecrets := make([]*pb.Secret, 0, len(result.ServerSecrets))
	for _, secret := range result.ServerSecrets {
		serverSecrets = append(serverSecrets, secret.ToProto())
	}

	return &pb.SyncResponse{
		CurrentVersion: result.CurrentVersion,
		Secrets:        serverSecrets,
		Conflicts:      result.Conflicts,
	}, nil
}

// GetSecret возвращает секрет по ID
func (h *SecretHandler) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id is required")
	}

	secret, err := h.dataService.GetSecret(ctx, user.ID, req.GetSecretId())
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &pb.GetSecretResponse{
		Secret: secret.ToProto(),
	}, nil
}

// ListSecrets возвращает список секретов пользователя
func (h *SecretHandler) ListSecrets(ctx context.Context, req *pb.ListSecretsRequest) (*pb.ListSecretsResponse, error) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var secrets []*domain.Secret
	var filterType *domain.SecretType

	if req.GetFilterType() != pb.SecretType_SECRET_TYPE_UNSPECIFIED {
		st := domain.SecretTypeFromProto(req.GetFilterType())
		filterType = &st
	}

	secrets, err = h.dataService.ListSecrets(ctx, user.ID, filterType)
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	pbSecrets := make([]*pb.Secret, 0, len(secrets))
	for _, secret := range secrets {
		pbSecrets = append(pbSecrets, secret.ToProto())
	}

	return &pb.ListSecretsResponse{
		Secrets: pbSecrets,
	}, nil
}

// UpdateSecret обновляет существующий секрет
func (h *SecretHandler) UpdateSecret(ctx context.Context, req *pb.UpdateSecretRequest) (*pb.UpdateSecretResponse, error) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetSecret() == nil {
		return nil, status.Error(codes.InvalidArgument, "secret is required")
	}

	secret := domain.SecretFromProto(req.GetSecret())
	if err := h.dataService.UpdateSecret(ctx, user.ID, secret); err != nil {
		return nil, MapErrorToStatus(err)
	}

	updatedSecret, err := h.dataService.GetSecret(ctx, user.ID, secret.ID)
	if err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &pb.UpdateSecretResponse{
		Secret: updatedSecret.ToProto(),
	}, nil
}

// DeleteSecret удаляет секрет
func (h *SecretHandler) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id is required")
	}

	if err := h.dataService.DeleteSecret(ctx, user.ID, req.GetSecretId()); err != nil {
		return nil, MapErrorToStatus(err)
	}

	return &pb.DeleteSecretResponse{
		Success: true,
	}, nil
}
