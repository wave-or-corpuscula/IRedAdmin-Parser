package controller

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func (c *CLIController) NewSyncMailboxesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "syncronize mailboxes in provided sever",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info(
				"mailbox sync command started",
				"server", c.config.Server,
			)

			server, err := c.Storage.GetServer(c.config.Server)
			if err != nil {
				slog.Error(
					"mailbox sync command failed",
					"server", c.config.Server,
					"cannot get server from db", err,
				)
				return err
			}
			amount, err := c.SyncService.Sync(cmd.Context(), server)
			if err != nil {
				slog.Error(
					"mailbox sync command failed",
					"server", c.config.Server,
					"cannot sync mailboxes", err,
				)
				return err
			}

			c.sendResponse(
				map[string]any{
					"server": c.config.Server,
					"amount": amount,
				},
			)

			slog.Info(
				"mailbox sync command completed",
				"server", c.config.Server,
				"mailboxes", amount,
			)

			return nil
		},
	}

	return cmd
}
