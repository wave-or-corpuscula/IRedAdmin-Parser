package domainparser

import (
	"testing"

	"iredparser/internal/parser/client"
	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDomains(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	config := client.AuthConfig(configs[0])

	c := client.NewClient(config.Server)
	err = c.Auth(t.Context(), client.AuthConfig{Login: config.Login, Password: config.Password})
	assert.NoError(t, err)

	parser := NewDomainParser(c)

	domains, err := parser.Parse(t.Context())
	assert.NoError(t, err)

	assert.True(t, len(domains) > 0)
}
