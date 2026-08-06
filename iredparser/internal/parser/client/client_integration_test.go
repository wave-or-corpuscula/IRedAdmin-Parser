package client

import (
	"bytes"
	"context"
	"fmt"
	"iredparser/common"
	"iredparser/pkg/errors"
	"iredparser/pkg/utils"
	"testing"

	apptesting "iredparser/testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

const testMailbox = "test@rmzu.by"

func GetTestAuthClient(ctx context.Context) (*Client, error) {
	configs, err := apptesting.GetAuthConfigs()
	if err != nil {
		return nil, err
	}
	config := configs[0]
	c, err := NewClient()
	if err != nil {
		return nil, err
	}

	return c, c.Auth(ctx, config)
}

func TestClientAuth(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	tests := []struct {
		name   string
		config common.ServerConfig
		err    error
	}{
		{
			name:   "correct credentials",
			config: configs[0],
			err:    nil,
		},
		{
			name:   "invalid email",
			config: common.ServerConfig{Server: configs[0].Server},
			err:    errors.ErrInvalidUsername,
		},
		{
			name:   "incorrect credentials",
			config: common.ServerConfig{Server: configs[0].Server, Login: "incorrect@mail.com", Password: "incorrect"},
			err:    errors.ErrIncorrectCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient()
			assert.NoError(t, err)

			err = c.Auth(t.Context(), tt.config)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientGet(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	client, err := GetTestAuthClient(t.Context())
	assert.NoError(t, err)

	body, err := client.Get(
		t.Context(),
		fmt.Sprintf("https://%s/iredadmin/dashboard", configs[0].Server),
	)
	assert.NoError(t, err)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	assert.NoError(t, err)

	assert.True(t, len(body) != 0)

	title := doc.Find(".title").Text()
	assert.NotContains(t, title, "Login")
}

func TestGetCSRFToken(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	config := configs[0]

	c, err := NewClient()
	assert.NoError(t, err)

	err = c.Auth(t.Context(), config)
	assert.NoError(t, err)

	mailbox := config.Login

	token, err := c.GetCSRFToken(t.Context(), config.Server, mailbox)
	assert.NoError(t, err)
	assert.True(t, len(token) > 0)
}

func TestChangePassword(t *testing.T) {
	configs, err := apptesting.GetAuthConfigs()
	assert.NoError(t, err)

	config := configs[0]

	c, err := NewClient()
	assert.NoError(t, err)

	err = c.Auth(t.Context(), config)
	assert.NoError(t, err)

	newPass, err := utils.GeneratePassword(10)
	assert.NoError(t, err)

	err = c.ChangePassword(t.Context(), config.Server, testMailbox, newPass)
	assert.NoError(t, err)
}
