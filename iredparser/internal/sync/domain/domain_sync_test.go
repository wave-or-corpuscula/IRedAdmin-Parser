package syncdomain

import (
	"testing"

	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	domainparser "iredparser/internal/parser/domain"
	synccommon "iredparser/internal/sync/common"
	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncDomains(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	config := client.AuthConfig(configs[0])

	db, err := synccommon.GetTestDB()
	assert.NoError(t, err)

	server := &parser.Server{Name: config.Server}

	serverModel, err := db.UpsertServer(server)
	assert.NoError(t, err)

	c := client.NewClient(serverModel.Name)
	c.Auth(t.Context(), config)

	parser := domainparser.NewDomainParser(c)

	domainSync := NewDomainSyncService(parser, db)
	domains, err := domainSync.Sync(t.Context(), serverModel.ID)
	assert.NoError(t, err)

	dbDomains, err := db.GetDomains()
	assert.NoError(t, err)

	assert.Equal(t, domains, dbDomains)
}
