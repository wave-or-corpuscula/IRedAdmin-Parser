package authservice

import (
	"context"
	"iredparser/common"
	"iredparser/internal/parser/client"
)

type AuthService struct {
	client *client.Client
}

func NewAuthService(client *client.Client) *AuthService {
	return &AuthService{
		client: client,
	}
}

func (a *AuthService) AuthClient(ctx context.Context, config common.ServerConfig) error {
	return a.client.Auth(ctx, config)
}
