package syncservice

import (
	"context"
	"errors"
	"iredparser/internal/database"
	"iredparser/internal/parser"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMailSync struct {
	mu      sync.Mutex
	calls   int
	results map[int64][]*database.MailboxModel
	errors  map[int64]error
}

func (m *mockMailSync) Sync(
	ctx context.Context,
	server *database.ServerModel,
	domain *database.DomainModel,
) ([]*database.MailboxModel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls++

	return m.results[domain.ID], m.errors[domain.ID]
}

func (m *mockMailSync) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.calls
}

type mockDomainSync struct {
	domains []*database.DomainModel
	err     error
	called  bool
}

func (m *mockDomainSync) Sync(
	ctx context.Context,
	server *database.ServerModel,
) ([]*database.DomainModel, error) {
	m.called = true

	return m.domains, m.err
}

func TestSyncService_Sync(t *testing.T) {
	domainErr := errors.New("domain sync error")
	mailErr1 := errors.New("mailbox sync error 1")
	mailErr2 := errors.New("mailbox sync error 2")

	domain1 := &database.DomainModel{ID: 1}
	domain2 := &database.DomainModel{ID: 2}
	domain3 := &database.DomainModel{ID: 3}

	server := &database.ServerModel{
		ID: 1,
		Server: parser.Server{
			Name: "test-server",
		},
	}

	tests := []struct {
		name          string
		domains       []*database.DomainModel
		domainErr     error
		mailResults   map[int64][]*database.MailboxModel
		mailErrors    map[int64]error
		wantTotal     int
		wantErr       error
		wantMailCalls int
	}{
		{
			name:      "domain sync error",
			domainErr: domainErr,
			wantTotal: -1,
			wantErr:   domainErr,
		},
		{
			name:          "no domains",
			domains:       nil,
			wantTotal:     0,
			wantMailCalls: 0,
		},
		{
			name: "all domains synced successfully",
			domains: []*database.DomainModel{
				domain1,
				domain2,
			},
			mailResults: map[int64][]*database.MailboxModel{
				1: {
					{ID: 1},
					{ID: 2},
				},
				2: {
					{ID: 3},
					{ID: 4},
					{ID: 5},
				},
			},
			wantTotal:     5,
			wantMailCalls: 2,
		},
		{
			name: "one mailbox sync error",
			domains: []*database.DomainModel{
				domain1,
				domain2,
			},
			mailResults: map[int64][]*database.MailboxModel{
				1: {
					{ID: 1},
					{ID: 2},
				},
			},
			mailErrors: map[int64]error{
				2: mailErr1,
			},
			wantTotal:     -1,
			wantErr:       mailErr1,
			wantMailCalls: 2,
		},
		{
			name: "multiple mailbox sync errors",
			domains: []*database.DomainModel{
				domain1,
				domain2,
				domain3,
			},
			mailErrors: map[int64]error{
				1: mailErr1,
				2: mailErr2,
			},
			mailResults: map[int64][]*database.MailboxModel{
				3: {
					{ID: 1},
					{ID: 2},
				},
			},
			wantTotal:     -1,
			wantErr:       mailErr1,
			wantMailCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domainSync := &mockDomainSync{
				domains: tt.domains,
				err:     tt.domainErr,
			}

			mailSync := &mockMailSync{
				results: tt.mailResults,
				errors:  tt.mailErrors,
			}

			service := NewSyncService(mailSync, domainSync)

			total, err := service.Sync(context.Background(), server)

			assert.Equal(t, tt.wantTotal, total)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantMailCalls, mailSync.Calls())
		})
	}
}
