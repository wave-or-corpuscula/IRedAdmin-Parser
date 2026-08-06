package controller

import (
	"context"
	"iredparser/common"
	"iredparser/internal/database"
)

type Storage interface {
	GetServerID(name string) (int64, error)
}

type MailboxesSyncer interface {
	Sync(ctx context.Context)
}

type SyncService interface {
	Sync(ctx context.Context, server *database.ServerModel) (int, error)
}

type AuthChecker interface {
	AuthClient(ctx context.Context, config common.ServerConfig) error
}

type PasswordChanger interface {
	ChangePassword(ctx context.Context, server string, mailbox string, newPassword string) error
}
