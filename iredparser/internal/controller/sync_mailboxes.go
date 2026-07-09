package controller

import (
	"github.com/spf13/cobra"
)

func (c *CLIController) NewSyncMailboxesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "syncronize mailboxes in provided sever",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverID, err := c.Storage.GetServerID(c.config.Server)
			if err != nil {
				return err
			}
			amount, err := c.SyncService.Sync(cmd.Context(), serverID)
			if err != nil {
				return err
			}

			c.sendResponse(
				map[string]any{
					"server": c.config.Server,
					"amount": amount,
				},
			)
			return nil
		},
	}

	return cmd
}
