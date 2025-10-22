package main

import (
	"log"
	"runtime"

	"github.com/alisaviation/GophKeeper/internal/config"
	"github.com/alisaviation/GophKeeper/internal/crypto"
	"github.com/alisaviation/GophKeeper/internal/server/app"
	"github.com/alisaviation/GophKeeper/internal/server/storage"
	"github.com/alisaviation/GophKeeper/internal/server/transport"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cfg := config.SetServerConfig()
	log.Printf("Starting GophKeeper Server %s (commit: %s, built: %s, go: %s)", version, commit, date, runtime.Version())

	if err := storage.RunMigrations(cfg.Database); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	storage, err := storage.NewStorage(storage.Config{
		Type:     storage.TypePostgreSQL,
		Database: cfg.Database,
	})
	if err != nil {
		log.Fatal("Failed to create storage:", err)
	}
	defer storage.Close()

	jwtManager := crypto.NewJWTManager(&crypto.JWTConfig{
		Secret:        cfg.JWT.Secret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	})

	var encryptor crypto.Encryptor
	if cfg.Encryption.Key != "" {
		encryptor, err = crypto.NewAESGCMEncryptor([]byte(cfg.Encryption.Key))
		if err != nil {
			log.Fatal("Failed to create encryptor:", err)
		}
	} else {
		encryptor = &crypto.NoopEncryptor{}
	}

	authService := app.NewAuthService(storage.UserRepository(), jwtManager)
	dataService := app.NewDataService(storage.SecretRepository(), encryptor)

	grpcConfig := transport.Config{
		Port: cfg.GRPCPort,
	}

	grpcServer := transport.NewServer(authService, dataService, grpcConfig)

	log.Printf("Starting gRPC server on port %d", grpcConfig.Port)
	if err = grpcServer.Start(); err != nil {
		log.Fatal("Failed to start gRPC server:", err)
	}
}
