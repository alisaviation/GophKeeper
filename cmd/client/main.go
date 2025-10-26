package main

import (
	"fmt"
	"os"
	"runtime"

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
	builtBy = "unknown"
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper - Secure password manager",
		Version: fmt.Sprintf("%s (commit: %s, built: %s, by: %s, go: %s)",
			version, commit, date, builtBy, runtime.Version()),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	clientApp, err := initApp()
	if err != nil {
		fmt.Printf("Failed to initialize app: %v\n", err)
		os.Exit(1)
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

func initApp() (*app.Client, error) {
	cfg := config.SetClientConfig()
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
