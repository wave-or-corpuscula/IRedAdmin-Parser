package syncdomain

import (
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	"testing"

	domainparser "iredparser/internal/parser/domain"
	synccommon "iredparser/internal/sync/common"
	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncDomains(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	db, err := synccommon.GetTestDB()
	assert.NoError(t, err)

	for _, config := range configs {

		server := &parser.Server{Name: config.Server}

		serverModel, err := db.UpsertServer(server)
		assert.NoError(t, err)

		c, err := client.NewClient()
		assert.NoError(t, err)

		err = c.Auth(t.Context(), config)
		assert.NoError(t, err)

		parser := domainparser.NewDomainParser(c)

		domainSync := NewDomainSyncService(parser, db)
		domains, err := domainSync.Sync(t.Context(), serverModel)
		assert.NoError(t, err)

		dbDomains, err := db.GetDomains()
		assert.NoError(t, err)

		for _, domain := range domains {
			assert.Contains(t, dbDomains, domain)
		}
	}
}
