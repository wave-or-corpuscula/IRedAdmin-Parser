// Package controller controls cli side of the application
package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"iredparser/common"
	"iredparser/internal/database"
	"iredparser/internal/parser/client"

	apperrors "iredparser/pkg/errors"

	"github.com/spf13/cobra"
)

const (
	ErrCliInvalidConfig       = "INVALID_CONFIG"
	ErrCliInvalidCredentials  = "INVALID_CREDENTIALS"
	ErrCliAuthenticationError = "AUTHENTICATION_ERROR"
	ErrCli                    = "ERROR_IN_PROCESS"
)

const (
	Workers = 30
	DSN     = "data/ireddata.db"
)

type Response struct {
	Success bool           `json:"success"`
	Error   *ErrorResponse `json:"error"`
	Data    any            `json:"data"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *CLIController) sendResponse(data any) {
	resp := Response{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(c.outWriter).Encode(resp)
}

func (c *CLIController) SendError(code string, err error) {
	resp := Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    code,
			Message: err.Error(),
		},
	}

	json.NewEncoder(c.outWriter).Encode(resp)
}

type AuthChecker interface {
	AuthClient(ctx context.Context, c *client.Client, config client.AuthConfig) error
}

type SyncService interface {
	Sync(ctx context.Context, serverID int64) (int, error)
}

type Storage interface {
	GetServerID(name string) (int64, error)
}

type MailboxesSyncer interface {
	Sync(ctx context.Context)
}

type CLIController struct {
	Client      *client.Client
	Storage     *database.Database
	AuthService AuthChecker
	SyncService SyncService
	outWriter   io.Writer
	config      common.ServerConfig
}

func NewCLIController(client *client.Client, storage *database.Database, authcService AuthChecker, syncService SyncService, out io.Writer) *CLIController {
	return &CLIController{
		Client:      client,
		Storage:     storage,
		SyncService: syncService,
		AuthService: authcService,
		outWriter:   out,
	}
}

func (c *CLIController) InitCommands() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "parser-cli",
		Short: "Parser CLi Utility from human for human",
	}

	rootCmd.PersistentFlags().StringP("config", "c", "{}", "json config for server")
	_ = rootCmd.MarkPersistentFlagRequired("config")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		var cfg common.ServerConfig

		configString, err := cmd.Flags().GetString("config")
		if err != nil {
			return
		}

		err = json.Unmarshal([]byte(configString), &cfg)
		if err != nil {
			c.SendError(
				ErrCliInvalidConfig,
				err,
			)
			return
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(10)*time.Second)
		defer cancel()

		err = c.AuthService.AuthClient(ctx, c.Client, client.AuthConfig(cfg))
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			c.SendError(
				ErrCliInvalidCredentials,
				err,
			)
			return
		} else if err != nil {
			c.SendError(
				ErrCliAuthenticationError,
				err,
			)
			return
		}

		c.config = cfg
	}

	rootCmd.AddCommand(c.NewAuthCheckCmd())
	rootCmd.AddCommand(c.NewSyncMailboxesCmd())

	return rootCmd
}

func Execute(rootCmd *cobra.Command) error {
	return rootCmd.Execute()
}
