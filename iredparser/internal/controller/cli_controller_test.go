package controller

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	"iredparser/internal/database"
	"iredparser/internal/parser/client"
	domainparser "iredparser/internal/parser/domain"
	mailboxparser "iredparser/internal/parser/mailbox"
	authservice "iredparser/internal/services/auth_service"
	syncservice "iredparser/internal/sync"
	syncdomain "iredparser/internal/sync/domain"
	syncmailbox "iredparser/internal/sync/mailbox"
	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

func getTestCLIController(buf *bytes.Buffer) *CLIController {
	httpClient := client.NewClient("")
	authService := authservice.NewAuthService()

	db, err := database.Connect(DSN)
	if err != nil {
		log.Fatalln(err)
	}

	mailParser := mailboxparser.NewMailboxParser(httpClient, Workers)
	domainParser := domainparser.NewDomainParser(httpClient)

	mailboxSyncService := syncmailbox.NewMailboxSyncService(mailParser, db)
	domainSyncService := syncdomain.NewDomainSyncService(domainParser, db)

	syncService := syncservice.NewSyncService(mailboxSyncService, domainSyncService)

	ctrl := NewCLIController(httpClient, db, authService, syncService, buf)

	return ctrl
}

func TestAuthCheckCLI(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	for _, config := range configs {

		buf := new(bytes.Buffer)
		ctrl := getTestCLIController(buf)
		rootCmd := ctrl.InitCommands()
		assert.NotNil(t, rootCmd)

		cfg := client.AuthConfig(config)

		jsonData, err := json.Marshal(cfg)
		assert.NoError(t, err)

		rootCmd.SetArgs([]string{
			"auth-check",
			"--config", string(jsonData),
		})

		err = rootCmd.Execute()
		assert.NoError(t, err)

		var resp Response
		err = json.Unmarshal(buf.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	}
}

func TestSyncCLI(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	for _, config := range configs {
		t.Log("start syncing for", config.Server)

		buf := new(bytes.Buffer)
		ctrl := getTestCLIController(buf)
		rootCmd := ctrl.InitCommands()
		assert.NotNil(t, rootCmd)

		cfg := client.AuthConfig(config)

		jsonData, err := json.Marshal(cfg)
		assert.NoError(t, err)

		rootCmd.SetArgs([]string{
			"sync",
			"--config", string(jsonData),
		})

		err = rootCmd.Execute()
		assert.NoError(t, err)

		var resp Response
		err = json.Unmarshal(buf.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		ctrl.Storage.Close()
	}
}
