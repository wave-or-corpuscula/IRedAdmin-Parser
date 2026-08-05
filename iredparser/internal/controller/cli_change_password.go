package controller

import "github.com/spf13/cobra"

func (c *CLIController) NewChangePasswordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "change-password",
		Long: "change passwords for provided list of mailboxes in provided server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
