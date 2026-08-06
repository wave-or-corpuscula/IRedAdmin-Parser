package controller

import (
	"fmt"
	"iredparser/pkg/utils"

	"github.com/spf13/cobra"
)

func (c *CLIController) NewChangePasswordCmd() *cobra.Command {
	var mailbox string
	var password string

	cmd := &cobra.Command{
		Use:  "change-password",
		Long: "change password for provided list of mailbox in provided server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(password) == 0 {
				pass, err := utils.GeneratePassword(PassLenth)
				if err != nil {
					return fmt.Errorf("cli: cannot generate password: %w", err)
				}
				password = pass
			}

			err := c.PasswordService.ChangePassword(cmd.Context(), c.config.Server, mailbox, password)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&mailbox, "mailbox", "m", "", "mailbox for password change")
	cmd.Flags().StringVarP(&password, "password", "p", "", "change password to this one (default: random valid password (len = 10))")

	cmd.MarkFlagRequired("mailbox")

	return cmd
}
