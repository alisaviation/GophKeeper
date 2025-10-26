package handlers

import (
	"github.com/alisaviation/GophKeeper/internal/generated/grpc"
	"github.com/alisaviation/GophKeeper/internal/server/domain"
)

// validateRegisterRequest валидирует запрос регистрации
func validateRegisterRequest(req *grpc.RegisterRequest) error {
	if req.GetLogin() == "" {
		return domain.ValidationError{Field: "login", Message: "is required"}
	}
	if req.GetPassword() == "" {
		return domain.ValidationError{Field: "password", Message: "is required"}
	}
	if len(req.GetPassword()) < 8 {
		return domain.ValidationError{Field: "password", Message: "must be at least 8 characters"}
	}
	return nil
}

// validateLoginRequest валидирует запрос входа
func validateLoginRequest(req *grpc.LoginRequest) error {
	if req.GetLogin() == "" {
		return domain.ValidationError{Field: "login", Message: "is required"}
	}
	if req.GetPassword() == "" {
		return domain.ValidationError{Field: "password", Message: "is required"}
	}
	return nil
}
