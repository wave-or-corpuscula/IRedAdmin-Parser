// Package controller controls cli side of the application
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iredparser/common"
	"iredparser/internal/database"
	"iredparser/internal/parser/client"
	"time"

	apperrors "iredparser/pkg/errors"

	"github.com/spf13/cobra"
)

const (
	Workers   = 30
	PassLenth = 10
	DSN       = "data/ireddata.db"
)

type CLIController struct {
	Client          *client.Client
	Storage         *database.Database
	AuthService     AuthChecker
	SyncService     SyncService
	PasswordService PasswordChanger
	outWriter       io.Writer
	config          common.ServerConfig
}

func NewCLIController(
	client *client.Client,
	storage *database.Database,
	authcService AuthChecker,
	syncService SyncService,
	passwordService PasswordChanger,
	out io.Writer,
) *CLIController {
	return &CLIController{
		Client:          client,
		Storage:         storage,
		SyncService:     syncService,
		AuthService:     authcService,
		PasswordService: passwordService,
		outWriter:       out,
	}
}

func (c *CLIController) InitCommands() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "parser-cli",
		Short:         "Parser CLi Utility from human for human",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringP("config", "c", "{}", "json config for server")
	_ = rootCmd.MarkPersistentFlagRequired("config")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var cfg common.ServerConfig

		configString, err := cmd.Flags().GetString("config")
		if err != nil {
			return apperrors.New(
				apperrors.ErrTypeCLI,
				apperrors.ErrCodeCliInvalidConfig,
				fmt.Errorf("cli controller: failed to parse config string: %w", err),
			)
		}

		err = json.Unmarshal([]byte(configString), &cfg)
		if err != nil {
			return apperrors.New(
				apperrors.ErrTypeCLI,
				apperrors.ErrCodeCliInvalidConfig,
				fmt.Errorf("cli controller: unable to unmarshal config: %w", err),
			)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(10)*time.Second)
		defer cancel()

		err = c.AuthService.AuthClient(ctx, cfg)
		if apperrors.IsType(err, apperrors.ErrTypeAuthentication) {
			return err
		} else if err != nil {
			return apperrors.New(
				apperrors.ErrTypeCLI,
				apperrors.ErrCodeCLIUnknown,
				err,
			)
		}

		c.config = cfg

		return nil
	}

	rootCmd.AddCommand(c.NewAuthCheckCmd())
	rootCmd.AddCommand(c.NewSyncMailboxesCmd())
	rootCmd.AddCommand(c.NewChangePasswordCmd())

	return rootCmd
}

func Execute(rootCmd *cobra.Command) error {
	return rootCmd.Execute()
}
