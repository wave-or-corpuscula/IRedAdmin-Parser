package controller

import (
	"github.com/spf13/cobra"
)

func (c *CLIController) NewAuthCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth-check",
		Short: "check authentication credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			c.sendResponse(
				map[string]any{
					"authenticated": true,
					"server":        c.config.Server,
					"cookie_string": c.Client.GetCookieString(),
				},
			)
			return nil
		},
	}

	return cmd
}
