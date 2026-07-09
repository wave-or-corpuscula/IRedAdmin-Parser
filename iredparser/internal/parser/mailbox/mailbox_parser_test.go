package mailboxparser

import (
	"context"
	"testing"

	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	apptesting "iredparser/testing"

	"github.com/stretchr/testify/assert"
)

const mailWorkers = 5

func getTestDomain() parser.Domain {
	return parser.Domain{Name: "example.com"}
}

func GetAuthClient(ctx context.Context) (*client.Client, error) {
	configs, err := apptesting.GetAuthConfigs()
	if err != nil {
		return nil, err
	}
	config := client.AuthConfig(configs[0])
	c := client.NewClient(config.Server)
	return c, c.Auth(ctx, client.AuthConfig{Login: config.Login, Password: config.Password})
}

func GetTestMailboxParser(ctx context.Context, workers int) (*MailboxParser, error) {
	c, err := GetAuthClient(ctx)
	if err != nil {
		return nil, err
	}

	parser := NewMailboxParser(c, workers)
	return parser, nil
}

func TestGetPages(t *testing.T) {
	p, err := GetTestMailboxParser(t.Context(), mailWorkers)
	assert.NoError(t, err)

	domain := getTestDomain()

	pages, err := p.getPagesAmount(t.Context(), domain)
	assert.NoError(t, err)
	assert.True(t, pages > 20)
}

func TestParseMailboxes(t *testing.T) {
	p, err := GetTestMailboxParser(t.Context(), mailWorkers)
	assert.NoError(t, err)

	domain := getTestDomain()

	boxes, err := p.Parse(t.Context(), domain)
	assert.NoError(t, err)

	assert.True(t, len(boxes) >= 1007)
}
