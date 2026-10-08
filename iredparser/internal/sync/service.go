// Package syncservice provides service for syncing domains and mailboxes
package syncservice

import (
	"context"
	"errors"
	"iredparser/internal/database"
	"log/slog"
	"sync"
)

type MailSyncServiceType interface {
	Sync(ctx context.Context, server *database.ServerModel, domain *database.DomainModel) ([]*database.MailboxModel, error)
}

type DomainSyncServiceType interface {
	Sync(ctx context.Context, server *database.ServerModel) ([]*database.DomainModel, error)
}

type SyncService struct {
	mailSync   MailSyncServiceType
	domainSync DomainSyncServiceType
}

func NewSyncService(mailSync MailSyncServiceType, domainSync DomainSyncServiceType) *SyncService {
	return &SyncService{
		mailSync:   mailSync,
		domainSync: domainSync,
	}
}

func (s *SyncService) Sync(ctx context.Context, server *database.ServerModel) (int, error) {
	slog.Info("sync started", "server", server.Name)

	domains, err := s.domainSync.Sync(ctx, server)
	if err != nil {
		slog.Info(
			"domain sync failed",
			"server", server.Name,
			"error", err,
		)
		return -1, err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(domains))
	amountCh := make(chan int, len(domains))

	for _, domain := range domains {
		wg.Go(func() {
			mailboxes, err := s.mailSync.Sync(ctx, server, domain)
			if err != nil {
				errCh <- err
				return
			}
			amountCh <- len(mailboxes)
		})
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(amountCh)
	}()

	var syncWg sync.WaitGroup
	syncErrors := []error{}

	syncWg.Add(1)
	go func() {
		defer syncWg.Done()
		for err := range errCh {
			syncErrors = append(syncErrors, err)
		}
	}()

	total := 0
	for amount := range amountCh {
		total += amount
	}

	syncWg.Wait()


	if len(syncErrors) > 0 {
		return -1, errors.Join(syncErrors...)
	}

	slog.Info(
		"sync completed",
		"server", server.Name,
		"domains", len(domains),
		"mailboxes", total,
	)
	return total, nil
}
