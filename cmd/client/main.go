package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alisaviation/GophKeeper/internal/client/app"
	"github.com/alisaviation/GophKeeper/internal/client/commands"
	"github.com/alisaviation/GophKeeper/internal/client/storage"
	"github.com/alisaviation/GophKeeper/internal/client/transport"
	"github.com/alisaviation/GophKeeper/internal/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cfg := config.SetClientConfig()

	clientApp, err := initApp(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:     "gophkeeper",
		Short:   "GophKeeper - Secure password manager",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	rootCmd.AddCommand(
		commands.NewAuthCommand(clientApp),
		commands.NewSyncCommand(clientApp),
		commands.NewSecretsCommand(clientApp),
		commands.NewVersionCommand(version, commit, date),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func initApp(cfg config.ClientConfig) (*app.Client, error) {
	localStorage, err := storage.NewFileStorage(cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to init local storage: %w", err)
	}

	grpcClient, err := transport.NewGRPCClient(cfg.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to init gRPC client: %w", err)
	}

	clientApp := app.NewClient(localStorage, grpcClient)

	return clientApp, nil
}
