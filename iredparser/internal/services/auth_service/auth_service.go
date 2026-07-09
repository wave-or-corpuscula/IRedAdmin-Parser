package authservice

import (
	"context"

	"iredparser/internal/parser/client"
)

type AuthService struct {
	client *client.Client
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (a *AuthService) AuthClient(ctx context.Context, c *client.Client, config client.AuthConfig) error {
	return c.Auth(ctx, config)
}
