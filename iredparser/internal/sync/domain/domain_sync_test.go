package syncdomain

import (
	"context"
	"errors"
	"iredparser/internal/database"
	"iredparser/internal/parser"
	apperrors "iredparser/pkg/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockDomainParser struct {
	result *parser.ParseDomainResult
	err    error
}

func (m *mockDomainParser) Parse(
	ctx context.Context,
	server string,
) (*parser.ParseDomainResult, error) {
	return m.result, m.err
}

type mockDomainStorage struct {
	called   bool
	domains  []*parser.Domain
	serverID int64
}

func (m *mockDomainStorage) UpsertDomainMany(
	domains []*parser.Domain,
	serverID int64,
) ([]*database.DomainModel, error) {
	m.called = true
	m.domains = domains
	m.serverID = serverID

	return nil, nil
}

func TestDomainSyncService_Sync_ParseErrors(t *testing.T) {
	err1 := errors.New("failed to parse domain 1")
	err2 := errors.New("failed to parse domain 2")
	err3 := errors.New("failed to parse domain 3")

	domains := []*parser.Domain{
		{Name: "valid1.com"},
		{Name: "valid2.com"},
	}

	tests := []struct {
		name        string
		parseResult *parser.ParseDomainResult
		wantErrors  []error
	}{
		{
			name: "one parsing error",
			parseResult: &parser.ParseDomainResult{
				Domains: domains,
				Errors:  []error{err1},
			},
			wantErrors: []error{err1},
		},
		{
			name: "multiple parsing errors",
			parseResult: &parser.ParseDomainResult{
				Domains: domains,
				Errors:  []error{err1, err2, err3},
			},
			wantErrors: []error{err1, err2, err3},
		},
		{
			name: "multiple parsing errors with valid domains",
			parseResult: &parser.ParseDomainResult{
				Domains: domains,
				Errors:  []error{err1, err2},
			},
			wantErrors: []error{err1, err2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domainParser := &mockDomainParser{
				result: tt.parseResult,
			}

			storage := &mockDomainStorage{}

			service := NewDomainSyncService(domainParser, storage)

			server := &database.ServerModel{
				ID: 42,
				Server: parser.Server{
					Name: "mail.example.com",
				},
			}

			result, err := service.Sync(context.Background(), server)

			assert.Error(t, err)
			assert.Nil(t, result)

			var multiErr *apperrors.IRedMultiError
			assert.ErrorAs(t, err, &multiErr)

			assert.Equal(t, tt.wantErrors, multiErr.Errors)

			assert.False(t, storage.called)
		})
	}
}

func TestDomainSyncService_Sync_ParseError(t *testing.T) {
	parseErr := errors.New("parser is unavailable")

	domainParser := &mockDomainParser{
		err: parseErr,
	}

	storage := &mockDomainStorage{}

	service := NewDomainSyncService(domainParser, storage)

	server := &database.ServerModel{
		ID: 42,
		Server: parser.Server{
			Name: "mail.example.com",
		},
	}

	result, err := service.Sync(context.Background(), server)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, parseErr)

	assert.False(t, storage.called)
}
