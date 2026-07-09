// Package syncservice provides service for syncing domains and mailboxes
package syncservice

import (
	"context"
	"errors"
	"sync"

	"iredparser/internal/database"
)

type MailSyncServiceType interface {
	Sync(ctx context.Context, domain *database.DomainModel) ([]*database.MailboxModel, error)
}

type DomainSyncServiceType interface {
	Sync(ctx context.Context, serverID int64) ([]*database.DomainModel, error)
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

func (s *SyncService) Sync(ctx context.Context, serverID int64) (int, error) {
	domains, err := s.domainSync.Sync(ctx, serverID)
	if err != nil {
		return -1, err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(domains))
	amountCh := make(chan int, len(domains))

	for _, domain := range domains {
		wg.Go(func() {
			mailboxes, err := s.mailSync.Sync(ctx, domain)
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

	syncErrors := []error{}
	for err := range errCh {
		syncErrors = append(syncErrors, err)
	}
	if len(syncErrors) > 0 {
		return -1, errors.Join(syncErrors...)
	}

	total := 0
	for amount := range amountCh {
		total += amount
	}

	return total, nil
}
