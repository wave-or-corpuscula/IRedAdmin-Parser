package controller

import (
	"fmt"
	"iredparser/pkg/utils"
	"log/slog"

	"github.com/spf13/cobra"
)

func (c *CLIController) NewChangePasswordCmd() *cobra.Command {
	var mailbox string
	var password string

	cmd := &cobra.Command{
		Use:  "change-password",
		Long: "change password for provided list of mailbox in provided server",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info(
				"change password command started",
				"server", c.config.Server,
				"mailbox", mailbox,
			)
			if len(password) == 0 {
				pass, err := utils.GeneratePassword(PassLenth)
				if err != nil {
					slog.Error(
						"change password command failed",
						"error while generating password", err,
					)
					return fmt.Errorf("cli: cannot generate password: %w", err)
				}
				password = pass
			}

			err := c.PasswordService.ChangePassword(cmd.Context(), c.config.Server, mailbox, password)
			if err != nil {
				slog.Error(
					"change password command failed",
					"error while changing password", err,
				)
				return err
			}

			c.sendResponse(
				map[string]string{
					"mailbox":  mailbox,
					"password": password,
				},
			)

			slog.Info(
				"change password command finished",
				"server", c.config.Server,
			)

			return nil
		},
	}

	cmd.Flags().StringVarP(&mailbox, "mailbox", "m", "", "mailbox for password change")
	cmd.Flags().StringVarP(&password, "password", "p", "", "change password to this one (default: random valid password (len = 10))")

	cmd.MarkFlagRequired("mailbox")

	return cmd
}
