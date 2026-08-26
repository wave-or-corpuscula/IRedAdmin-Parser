package mailboxparser

import (
	"context"
	"iredparser/common"
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	"testing"

	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

const mailWorkers = 5

func getTestDomain() parser.Domain {
	return parser.Domain{Name: "example.com"}
}

func GetAuthClient(ctx context.Context, config common.ServerConfig) (*client.Client, error) {
	c, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	return c, c.Auth(ctx, config)
}

func GetTestMailboxParser(ctx context.Context, config common.ServerConfig) (*MailboxParser, error) {
	c, err := GetAuthClient(ctx, config)
	if err != nil {
		return nil, err
	}

	parser := NewMailboxParser(c, mailWorkers)
	return parser, nil
}

func TestGetPages(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	for _, config := range configs {
		p, err := GetTestMailboxParser(t.Context(), config)
		assert.NoError(t, err)

		domain := getTestDomain()

		pages, err := p.getPagesAmount(t.Context(), config.Server, domain)
		assert.NoError(t, err)
		assert.True(t, pages > 0)

		t.Logf("got %d pages from %s\n", pages, config.Server)
	}
}

func TestParseMailboxes(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	for _, config := range configs {
		p, err := GetTestMailboxParser(t.Context(), config)
		assert.NoError(t, err)

		domain := getTestDomain()

		boxes, err := p.Parse(t.Context(), config.Server, domain)
		assert.NoError(t, err)

		t.Logf("got %d mailboxes from %s\n", len(boxes), config.Server)
	}
}
