package controller

import (
	"context"
	"iredparser/common"
	"iredparser/internal/database"
	"iredparser/internal/parser/client"
)

type AuthChecker interface {
	AuthClient(ctx context.Context, c *client.Client, config common.ServerConfig) error
}

type SyncService interface {
	Sync(ctx context.Context, server *database.ServerModel) (int, error)
}

type Storage interface {
	GetServerID(name string) (int64, error)
}

type MailboxesSyncer interface {
	Sync(ctx context.Context)
}
