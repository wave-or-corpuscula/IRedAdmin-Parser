// Package parsetesting contains comminly used tool in parsing
package parsetesting

import (
	"context"

	"iredparser/internal/parser/client"
	apptesting "iredparser/testing"
)



func GetAuthClient(ctx context.Context) (*client.Client, error) {
	configs, err := apptesting.GetAuthConfigs()
	if err != nil {
		return nil, err
	}
	config := client.AuthConfig(configs[0])
	c := client.NewClient(config.Server)
	return c, c.Auth(ctx, client.AuthConfig{Login: config.Login, Password: config.Password})
}
