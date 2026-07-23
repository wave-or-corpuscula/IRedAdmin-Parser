// Package controller controls cli side of the application
package controller

import (
	"context"
	"encoding/json"
	"errors"
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
	Workers = 30
	DSN     = "data/ireddata.db"
)

type Response struct {
	Success bool           `json:"success"`
	Error   *ErrorResponse `json:"error"`
	Data    any            `json:"data"`
}

type CLIError struct {
	Code string
	Err  error
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %v", e.Code, e.Err)
	}
	return fmt.Sprintf("[%s] unknown error", e.Code)
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

// Error codes
const (
	ErrCliInvalidConfig       = "INVALID_CONFIG"
	ErrCliInvalidCredentials  = "INVALID_CREDENTIALS"
	ErrCliAuthenticationError = "AUTHENTICATION_ERROR"
	ErrCli                    = "UNKNOWN_CLI_ERROR"
)

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
	AuthClient(ctx context.Context, c *client.Client, config common.ServerConfig) error
}

type SyncService interface {
	Sync(ctx context.Context, server *database.ServerModel) (int, error)
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
			return &CLIError{Code: ErrCli, Err: fmt.Errorf("cli controller: failed to parse config string: %w", err)}
		}

		err = json.Unmarshal([]byte(configString), &cfg)
		if err != nil {
			return &CLIError{Code: ErrCliInvalidConfig, Err: fmt.Errorf("cli controller: unable to unmarshal config: %w", err)}
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(10)*time.Second)
		defer cancel()

		err = c.AuthService.AuthClient(ctx, c.Client, cfg)
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			return &CLIError{
				Code: ErrCliInvalidCredentials,
				Err:  err,
			}
		} else if err != nil {
			return &CLIError{
				Code: ErrCliAuthenticationError,
				Err:  err,
			}
		}

		c.config = cfg

		return nil
	}

	rootCmd.AddCommand(c.NewAuthCheckCmd())
	rootCmd.AddCommand(c.NewSyncMailboxesCmd())

	return rootCmd
}

func Execute(rootCmd *cobra.Command) error {
	return rootCmd.Execute()
}
