package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alisaviation/GophKeeper/internal/client/app"
)

// newSyncCommand создает команды для синхронизации
func NewSyncCommand(clientApp *app.Client) *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data with server",
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			resolveStrategy, _ := cmd.Flags().GetString("resolve")

			ctx := context.Background()
			result, err := clientApp.Sync(ctx, force, resolveStrategy)
			if err != nil {
				fmt.Printf("Sync failed: %v\n", err)
				return
			}

			fmt.Printf("Sync completed successfully:\n")
			fmt.Printf("  Uploaded: %d secrets\n", result.Uploaded)
			fmt.Printf("  Downloaded: %d secrets\n", result.Downloaded)

			if len(result.Conflicts) > 0 {
				fmt.Printf("  Conflicts: %d\n", len(result.Conflicts))
				for _, conflict := range result.Conflicts {
					fmt.Printf("    - %s\n", conflict)
				}
			}
		},
	}

	syncCmd.Flags().BoolP("force", "f", false, "Force sync all secrets")
	syncCmd.Flags().StringP("resolve", "r", "server", "Conflict resolution strategy (server|local)")

	return syncCmd
}
