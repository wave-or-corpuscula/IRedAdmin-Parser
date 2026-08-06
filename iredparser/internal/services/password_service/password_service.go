package passwordservice

import (
	"context"
	"iredparser/internal/parser/client"
)

type PasswordService struct {
	client *client.Client
}

func NewPasswordService(client *client.Client) *PasswordService {
	return &PasswordService{
		client: client,
	}
}

func (cps *PasswordService) ChangePassword(ctx context.Context, server string, mailbox string, newPassword string) error {
	return cps.client.ChangePassword(ctx, server, mailbox, newPassword)
}
