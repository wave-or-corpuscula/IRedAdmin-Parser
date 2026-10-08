package syncmailbox

import (
	"context"
	"errors"
	"iredparser/internal/database"
	"iredparser/internal/parser"
	apperrors "iredparser/pkg/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMailboxParser struct {
	result *parser.ParseMailboxesResult
	err    error
}

func (m *mockMailboxParser) Parse(
	ctx context.Context,
	server string,
	domain parser.Domain,
) (*parser.ParseMailboxesResult, error) {
	return m.result, m.err
}

type mockMailboxStorage struct {
	called    bool
	mailboxes []*parser.Mailbox
	domainID  int64
	result    []*database.MailboxModel
	err       error
}

func (m *mockMailboxStorage) UpsertMailboxMany(
	mailboxes []*parser.Mailbox,
	domainID int64,
) ([]*database.MailboxModel, error) {
	m.called = true
	m.mailboxes = mailboxes
	m.domainID = domainID

	return m.result, m.err
}

func TestMailboxSyncService_Sync_ParseErrors(t *testing.T) {
	err1 := errors.New("failed to parse mailbox 1")
	err2 := errors.New("failed to parse mailbox 2")
	err3 := errors.New("failed to parse mailbox 3")

	mailboxes := []*parser.Mailbox{
		{
			Address: "user1@example.com",
		},
		{
			Address: "user2@example.com",
		},
	}

	tests := []struct {
		name        string
		parseResult *parser.ParseMailboxesResult
		wantErrors  []error
	}{
		{
			name: "one parsing error",
			parseResult: &parser.ParseMailboxesResult{
				Mailboxes: mailboxes,
				Errors:    []error{err1},
			},
			wantErrors: []error{err1},
		},
		{
			name: "multiple parsing errors",
			parseResult: &parser.ParseMailboxesResult{
				Mailboxes: mailboxes,
				Errors:    []error{err1, err2, err3},
			},
			wantErrors: []error{err1, err2, err3},
		},
		{
			name: "valid mailboxes with multiple parsing errors",
			parseResult: &parser.ParseMailboxesResult{
				Mailboxes: mailboxes,
				Errors:    []error{err1, err2},
			},
			wantErrors: []error{err1, err2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailboxParser := &mockMailboxParser{
				result: tt.parseResult,
			}

			storage := &mockMailboxStorage{}

			service := NewMailboxSyncService(mailboxParser, storage)

			server := &database.ServerModel{
				ID: 1,
				Server: parser.Server{
					Name: "mail.example.com",
				},
			}

			domain := &database.DomainModel{
				ID: 42,
				Domain: parser.Domain{
					Name: "example.com",
				},
			}

			result, err := service.Sync(
				context.Background(),
				server,
				domain,
			)

			assert.Error(t, err)
			assert.Nil(t, result)

			var multiErr *apperrors.IRedMultiError
			assert.ErrorAs(t, err, &multiErr)

			assert.Equal(t, tt.wantErrors, multiErr.Errors)

			assert.False(t, storage.called)
		})
	}
}

func TestMailboxSyncService_Sync_ParseError(t *testing.T) {
	parseErr := errors.New("parser unavailable")

	mailboxParser := &mockMailboxParser{
		err: parseErr,
	}

	storage := &mockMailboxStorage{}

	service := NewMailboxSyncService(mailboxParser, storage)

	server := &database.ServerModel{
		ID: 1,
		Server: parser.Server{
			Name: "mail.example.com",
		},
	}

	domain := &database.DomainModel{
		ID: 42,
		Domain: parser.Domain{
			Name: "example.com",
		},
	}

	result, err := service.Sync(
		context.Background(),
		server,
		domain,
	)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, parseErr)
	assert.Equal(
		t,
		`failed to parse mailboxes for domain "example.com": parser unavailable`,
		err.Error(),
	)

	assert.False(t, storage.called)
}

func TestMailboxSyncService_Sync_StorageError(t *testing.T) {
	storageErr := errors.New("database unavailable")

	mailboxes := []*parser.Mailbox{
		{},
		{},
	}

	mailboxParser := &mockMailboxParser{
		result: &parser.ParseMailboxesResult{
			Mailboxes: mailboxes,
		},
	}

	storage := &mockMailboxStorage{
		err: storageErr,
	}

	service := NewMailboxSyncService(mailboxParser, storage)

	server := &database.ServerModel{
		ID: 1,
		Server: parser.Server{
			Name: "mail.example.com",
		},
	}

	domain := &database.DomainModel{
		ID: 42,
		Domain: parser.Domain{
			Name: "example.com",
		},
	}

	result, err := service.Sync(
		context.Background(),
		server,
		domain,
	)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, storageErr)

	assert.True(t, storage.called)
	assert.Equal(t, mailboxes, storage.mailboxes)
	assert.Equal(t, int64(42), storage.domainID)
}

func TestMailboxSyncService_Sync_Success(t *testing.T) {
	mailboxes := []*parser.Mailbox{
		{},
		{},
	}

	models := []*database.MailboxModel{
		{},
		{},
	}

	mailboxParser := &mockMailboxParser{
		result: &parser.ParseMailboxesResult{
			Mailboxes: mailboxes,
		},
	}

	storage := &mockMailboxStorage{
		result: models,
	}

	service := NewMailboxSyncService(mailboxParser, storage)

	server := &database.ServerModel{
		ID: 1,
		Server: parser.Server{
			Name: "mail.example.com",
		},
	}

	domain := &database.DomainModel{
		ID: 42,
		Domain: parser.Domain{
			Name: "example.com",
		},
	}

	result, err := service.Sync(
		context.Background(),
		server,
		domain,
	)

	assert.NoError(t, err)
	assert.Equal(t, models, result)

	assert.True(t, storage.called)
	assert.Equal(t, mailboxes, storage.mailboxes)
	assert.Equal(t, int64(42), storage.domainID)
}
