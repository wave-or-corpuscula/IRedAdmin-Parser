package syncmailbox

import (
	"testing"

	"iredparser/internal/database"
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	domainparser "iredparser/internal/parser/domain"
	mailboxparser "iredparser/internal/parser/mailbox"
	synccommon "iredparser/internal/sync/common"
	syncdomain "iredparser/internal/sync/domain"
	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

func TestMailboxSyncIntegration(t *testing.T) {
	db, err := synccommon.GetTestDB()
	assert.NoError(t, err)

	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)
	for _, config := range configs {
		server := &parser.Server{Name: config.Server}

		serverModel, err := db.UpsertServer(server)
		assert.NoError(t, err)

		c := client.NewClient(serverModel.Name)
		err = c.Auth(t.Context(), client.AuthConfig(config))
		assert.NoError(t, err)

		parser := domainparser.NewDomainParser(c)

		domainSync := syncdomain.NewDomainSyncService(parser, db)
		domains, err := domainSync.Sync(t.Context(), serverModel.ID)
		assert.NoError(t, err)

		// dbDomains, err := db.GetDomains()
		assert.NoError(t, err)
		// assert.Equal(t, domains, dbDomains)

		mailParser := mailboxparser.NewMailboxParser(c, 20)
		mailSync := NewMailboxSyncService(mailParser, db)

		mailboxModels := []*database.MailboxModel{}

		for _, domain := range domains {
			boxes, err := mailSync.Sync(t.Context(), domain)
			assert.NoError(t, err)

			mailboxModels = append(mailboxModels, boxes...)
		}

		dbBoxes, err := db.GetMailboxes()
		assert.NoError(t, err)

		for _, box := range mailboxModels {
			assert.Contains(t, dbBoxes, box)
		}

	}
}
