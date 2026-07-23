package controller

import (
	"bytes"
	"encoding/json"
	"iredparser/common"
	"iredparser/internal/database"
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	"log"
	"testing"

	domainparser "iredparser/internal/parser/domain"
	mailboxparser "iredparser/internal/parser/mailbox"
	authservice "iredparser/internal/services/auth_service"
	syncservice "iredparser/internal/sync"
	syncdomain "iredparser/internal/sync/domain"
	syncmailbox "iredparser/internal/sync/mailbox"
	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

func getTestCLIController(buf *bytes.Buffer, config common.ServerConfig) (*CLIController, error) {
	httpClient, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	authService := authservice.NewAuthService()

	db, err := database.Connect(":memory:")
	log.Println(err)
	if err != nil {
		return nil, err
	}

	_, err = db.UpsertServer(&parser.Server{Name: config.Server})
	if err != nil {
		return nil, err
	}

	mailParser := mailboxparser.NewMailboxParser(httpClient, Workers)
	domainParser := domainparser.NewDomainParser(httpClient)

	mailboxSyncService := syncmailbox.NewMailboxSyncService(mailParser, db)
	domainSyncService := syncdomain.NewDomainSyncService(domainParser, db)

	syncService := syncservice.NewSyncService(mailboxSyncService, domainSyncService)

	ctrl := NewCLIController(httpClient, db, authService, syncService, buf)

	return ctrl, nil
}

func TestAuthCheckCLI(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	for _, config := range configs {

		buf := new(bytes.Buffer)
		ctrl, err := getTestCLIController(buf, config)
		assert.NoError(t, err)

		rootCmd := ctrl.InitCommands()
		assert.NotNil(t, rootCmd)

		jsonData, err := json.Marshal(config)
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
		buf := new(bytes.Buffer)
		ctrl, err := getTestCLIController(buf, config)
		assert.NoError(t, err)

		rootCmd := ctrl.InitCommands()
		assert.NotNil(t, rootCmd)

		jsonData, err := json.Marshal(config)
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
