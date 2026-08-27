package syncmailbox

import (
	"context"
	"fmt"
	"iredparser/internal/database"
	"iredparser/internal/parser"

	mailboxparser "iredparser/internal/parser/mailbox"
	apperrors "iredparser/pkg/errors"
)

type MailboxStorage interface {
	UpsertMailboxMany(mailboxes []*parser.Mailbox, DomainID int64) ([]*database.MailboxModel, error)
}

type MailboxSyncService struct {
	mailboxParser *mailboxparser.MailboxParser
	storage       MailboxStorage
}

func NewMailboxSyncService(parser *mailboxparser.MailboxParser, storage MailboxStorage) *MailboxSyncService {
	return &MailboxSyncService{
		mailboxParser: parser,
		storage:       storage,
	}
}

func (s *MailboxSyncService) Sync(ctx context.Context, server *database.ServerModel, domain *database.DomainModel) ([]*database.MailboxModel, error) {
	results, err := s.mailboxParser.Parse(ctx, server.Name, domain.Domain)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse mailboxes for domain %q: %w",
			domain.Name,
			err,
		)
	}

	if len(results.Errors) > 0 {
		return nil, apperrors.NewMultiError(results.Errors)
	}

	mailboxModels, err := s.storage.UpsertMailboxMany(results.Mailboxes, domain.ID)
	if err != nil {
		return nil, fmt.Errorf("error while mailbox syncing: %w", err)
	}

	return mailboxModels, nil
}
