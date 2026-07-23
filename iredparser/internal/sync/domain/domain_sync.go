// Package syncdomain syncronaises domains with db
package syncdomain

import (
	"context"
	"fmt"
	"iredparser/internal/database"
	"iredparser/internal/parser"

	domainparser "iredparser/internal/parser/domain"
)

type DomainStorage interface {
	UpsertDomainMany(domains []*parser.Domain, serverID int64) ([]*database.DomainModel, error)
}

type DomainSyncService struct {
	domainParser *domainparser.DomainParser
	storage      DomainStorage
}

func NewDomainSyncService(parser *domainparser.DomainParser, storage DomainStorage) *DomainSyncService {
	return &DomainSyncService{
		domainParser: parser,
		storage:      storage,
	}
}

func (s *DomainSyncService) Sync(ctx context.Context, server *database.ServerModel) ([]*database.DomainModel, error) {
	domains, err := s.domainParser.Parse(ctx, server.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse domains: %w", err)
	}

	models, err := s.storage.UpsertDomainMany(domains, server.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to sync domains: %w", err)
	}

	return models, nil
}
