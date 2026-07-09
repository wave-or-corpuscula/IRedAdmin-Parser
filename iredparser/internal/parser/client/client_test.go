package client

import (
	"bytes"
	"context"
	"testing"

	apptesting "iredparser/testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func GetTestAuthClient(ctx context.Context) (*Client, error) {
	configs, err := apptesting.GetAuthConfigs()
	if err != nil {
		return nil, err
	}
	config := AuthConfig(configs[0])
	c := NewClient(config.Server)
	return c, c.Auth(ctx, config)
}

func TestClientAuth(t *testing.T) {
	_, err := GetTestAuthClient(t.Context())
	assert.NoError(t, err)
}

func TestClientGet(t *testing.T) {
	client, err := GetTestAuthClient(t.Context())
	assert.NoError(t, err)

	body, err := client.Get(t.Context(), "https://mail01/iredadmin/dashboard")
	assert.NoError(t, err)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	assert.NoError(t, err)

	assert.True(t, len(body) != 0)

	title := doc.Find(".title").Text()
	assert.NotContains(t, title, "Login")
}
