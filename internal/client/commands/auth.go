package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/alisaviation/GophKeeper/internal/client/app"
)

// newAuthCommand создает команды для аутентификации
func NewAuthCommand(clientApp *app.Client) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	authCmd.AddCommand(
		&cobra.Command{
			Use:   "register [login]",
			Short: "Register a new user",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				login := args[0]
				password := readPassword("Enter password: ")
				confirmPassword := readPassword("Confirm password: ")

				if password != confirmPassword {
					fmt.Println("Error: Passwords do not match")
					return
				}

				ctx := context.Background()
				userID, err := clientApp.Register(ctx, login, password)
				if err != nil {
					fmt.Printf("Registration failed: %v\n", err)
					return
				}

				fmt.Printf("Successfully registered user %s with ID: %s\n", login, userID)
			},
		},
		&cobra.Command{
			Use:   "login [login]",
			Short: "Login to the service",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				login := args[0]
				password := readPassword("Enter password: ")

				ctx := context.Background()
				err := clientApp.Login(ctx, login, password)
				if err != nil {
					fmt.Printf("Login failed: %v\n", err)
					return
				}

				fmt.Printf("Successfully logged in as %s\n", login)
			},
		},
		&cobra.Command{
			Use:   "logout",
			Short: "Logout from the service",
			Run: func(cmd *cobra.Command, args []string) {
				err := clientApp.Logout()
				if err != nil {
					fmt.Printf("Logout failed: %v\n", err)
					return
				}

				fmt.Println("Successfully logged out")
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show current authentication status",
			Run: func(cmd *cobra.Command, args []string) {
				session, err := clientApp.GetSession()
				if err != nil || session == nil || session.AccessToken == "" {
					fmt.Println("Status: Not authenticated")
					return
				}

				fmt.Printf("Status: Authenticated as %s (UserID: %s)\n", session.Login, session.UserID)
				fmt.Printf("Last sync: %s\n", time.Unix(session.LastSync, 0).Format(time.RFC3339))
			},
		},
	)

	return authCmd
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		password, _ := reader.ReadString('\n')
		return strings.TrimSpace(password)
	}
	return string(password)
}
