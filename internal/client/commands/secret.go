package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alisaviation/GophKeeper/internal/client/app"
	"github.com/alisaviation/GophKeeper/internal/client/domain"
)

// newSecretsCommand создает команды для управления секретами
func NewSecretsCommand(clientApp *app.Client) *cobra.Command {
	secretsCmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage secrets",
	}

	secretsCmd.AddCommand(
		&cobra.Command{
			Use:   "list [type]",
			Short: "List all secrets",
			Args:  cobra.MaximumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				var filterType string
				if len(args) > 0 {
					filterType = args[0]
				}

				ctx := context.Background()
				secrets, err := clientApp.ListSecrets(ctx, filterType)
				if err != nil {
					fmt.Printf("Failed to list secrets: %v\n", err)
					return
				}

				if len(secrets) == 0 {
					fmt.Println("No secrets found")
					return
				}

				fmt.Printf("Found %d secrets:\n", len(secrets))
				for _, secret := range secrets {
					fmt.Printf("  %s [%s] - %s (created: %s)\n",
						secret.ID[:8], secret.Type, secret.Name,
						secret.CreatedAt.Format("2006-01-02 15:04"))
				}
			},
		},
		&cobra.Command{
			Use:   "get [id]",
			Short: "Get secret by ID",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				id := args[0]

				ctx := context.Background()
				secret, err := clientApp.GetSecret(ctx, id)
				if err != nil {
					fmt.Printf("Failed to get secret: %v\n", err)
					return
				}

				fmt.Printf("Secret: %s [%s]\n", secret.Name, secret.Type)
				fmt.Printf("ID: %s\n", secret.ID)
				fmt.Printf("Created: %s\n", secret.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("Updated: %s\n", secret.UpdatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("Data: %+v\n", secret.Data)
			},
		},
		&cobra.Command{
			Use:   "create-login [name] [login] [password]",
			Short: "Create login/password secret",
			Args:  cobra.ExactArgs(3),
			Run: func(cmd *cobra.Command, args []string) {
				name, login, password := args[0], args[1], args[2]

				website, _ := cmd.Flags().GetString("website")
				notes, _ := cmd.Flags().GetString("notes")

				secretData := &domain.SecretData{
					Type: domain.SecretTypeLoginPassword,
					Name: name,
					Data: domain.LoginPasswordData{
						Login:    login,
						Password: password,
						Website:  website,
						Notes:    notes,
					},
				}

				ctx := context.Background()
				id, err := clientApp.CreateSecret(ctx, secretData)
				if err != nil {
					fmt.Printf("Failed to create secret: %v\n", err)
					return
				}

				fmt.Printf("Successfully created secret with ID: %s\n", id)
			},
		},
		&cobra.Command{
			Use:   "create-text [name] [content]",
			Short: "Create text secret",
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				name, content := args[0], args[1]

				secretData := &domain.SecretData{
					Type: domain.SecretTypeText,
					Name: name,
					Data: domain.TextData{
						Content: content,
					},
				}

				ctx := context.Background()
				id, err := clientApp.CreateSecret(ctx, secretData)
				if err != nil {
					fmt.Printf("Failed to create secret: %v\n", err)
					return
				}

				fmt.Printf("Successfully created secret with ID: %s\n", id)
			},
		},
		&cobra.Command{
			Use:   "create-card [name] [cardholder] [number] [expiry] [cvv]",
			Short: "Create bank card secret",
			Args:  cobra.ExactArgs(5),
			Run: func(cmd *cobra.Command, args []string) {
				name, cardholder, number, expiry, cvv := args[0], args[1], args[2], args[3], args[4]

				bankName, _ := cmd.Flags().GetString("bank")

				secretData := &domain.SecretData{
					Type: domain.SecretTypeBankCard,
					Name: name,
					Data: domain.BankCardData{
						CardHolder: cardholder,
						CardNumber: number,
						ExpiryDate: expiry,
						CVV:        cvv,
						BankName:   bankName,
					},
				}

				ctx := context.Background()
				id, err := clientApp.CreateSecret(ctx, secretData)
				if err != nil {
					fmt.Printf("Failed to create secret: %v\n", err)
					return
				}

				fmt.Printf("Successfully created secret with ID: %s\n", id)
			},
		},
		&cobra.Command{
			Use:   "delete [id]",
			Short: "Delete secret by ID",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				id := args[0]

				ctx := context.Background()
				err := clientApp.DeleteSecret(ctx, id)
				if err != nil {
					fmt.Printf("Failed to delete secret: %v\n", err)
					return
				}

				fmt.Printf("Successfully marked secret %s for deletion\n", id)
			},
		},
		&cobra.Command{
			Use:   "create-binary [name] [file-path]",
			Short: "Create binary secret from file",
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				name, filePath := args[0], args[1]

				data, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Printf("Failed to read file: %v\n", err)
					return
				}

				description, _ := cmd.Flags().GetString("description")

				secretData := &domain.SecretData{
					Type: domain.SecretTypeBinary,
					Name: name,
					Data: domain.BinaryData{
						Data:        data,
						Description: description,
						FileName:    filepath.Base(filePath),
					},
				}

				ctx := context.Background()
				id, err := clientApp.CreateSecret(ctx, secretData)
				if err != nil {
					fmt.Printf("Failed to create secret: %v\n", err)
					return
				}

				fmt.Printf("Successfully created binary secret with ID: %s\n", id)
			},
		},
	)

	createLoginCmd := secretsCmd.Commands()[2]
	createLoginCmd.Flags().String("website", "", "Website URL")
	createLoginCmd.Flags().String("notes", "", "Additional notes")

	createCardCmd := secretsCmd.Commands()[4]
	createCardCmd.Flags().String("bank", "", "Bank name")

	return secretsCmd
}
