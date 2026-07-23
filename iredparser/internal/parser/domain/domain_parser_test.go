package domainparser

import (
	"iredparser/internal/parser/client"
	"testing"

	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDomains(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	for _, config := range configs {

		c, err := client.NewClient()
		assert.NoError(t, err)
		err = c.Auth(t.Context(), config)
		assert.NoError(t, err)

		parser := NewDomainParser(c)

		domains, err := parser.Parse(t.Context(), config.Server)
		assert.NoError(t, err)

		t.Logf("got %d domains from %s\n", len(domains), config.Server)
	}
}
