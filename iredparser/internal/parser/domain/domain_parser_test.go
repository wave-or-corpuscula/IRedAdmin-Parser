package domainparser

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	apperrors "iredparser/pkg/errors"
	"iredparser/pkg/utils"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

type DomainData struct {
	Disabled    bool
	Name        string
	DisplayName string
	UsedMemory  string
	Quota       string
	UsersAmount int
}

func makeDomain(disabled bool, name string, displayName string, usedMemory string, quota string, usersAmount int) DomainData {
	return DomainData{
		Disabled:    disabled,
		Name:        name,
		DisplayName: displayName,
		UsedMemory:  usedMemory,
		Quota:       quota,
		UsersAmount: usersAmount,
	}
}

func (d DomainData) Compare(domain parser.Domain) (bool, error) {
	if d.Disabled != domain.Disabled {
		return false, nil
	}

	if d.Name != domain.Name {
		return false, nil
	}

	if d.DisplayName != domain.DisplayName {
		return false, nil
	}

	quota, err := utils.GetMemoryBytes(d.Quota)
	if err != nil {
		return false, err
	}

	if quota != domain.QuotaBytes {
		return false, nil
	}

	used, err := utils.GetMemoryBytes(d.UsedMemory)
	if err != nil {
		return false, err
	}

	if used != domain.UsedMemoryBytes {
		return false, nil
	}

	return true, nil
}

const domainTemplate = `
<table>
	<tbody>
		{{ range . }}
		<tr class="{{ if .Disabled }}disabled{{ else }}active{{ end }}">
			<td class="checkbox vcenter">
					<input type="checkbox" class="checkbox" name="domainName" value="{{ .Name }}">
			</td>
			<td class="vcenter">
					<a href="/iredadmin/profile/domain/general/{{ .Name }}">{{ .Name }}</a>
					<a href="/iredadmin/profile/domain/general/{{ .Name }}">
							<i class="fa fa-cog fa-lg fr-space"></i>
					</a>
			</td>
			<td class="vcenter">{{ .DisplayName }}</td>
			<td class="vcenter" data-sort-value="">
					<span class="color-grey">{{ .UsedMemory }}</span> / {{ .Quota }}
			</td>
			<td class="vcenter" data-sort-value="0">
					<a href="/iredadmin/users/{{ .Name }}" style="text-decoration: none; display: block; padding: 0 10px 0 10px;">{{ .UsersAmount }}</a>
			</td>
		</tr>
	 {{ end }}
	</tbody>
</table>
`

type MockClient struct{}

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func newTestClient(handler func(r *http.Request) (*http.Response, error)) *client.Client {
	jar, _ := cookiejar.New(nil)

	httpClient := &http.Client{
		Transport: &mockTransport{
			roundTripFunc: handler,
		},
		Jar: jar,
	}

	return client.NewClientRaw(httpClient)
}

func newTestDomainParser(handler func(r *http.Request) (*http.Response, error)) *DomainParser {
	testClient := newTestClient(handler)

	return NewDomainParser(testClient)
}

func makeResponse(statusCode int, body string, header http.Header) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     header,
	}
}

func TestParseDomainsTable(t *testing.T) {
	tests := []struct {
		name          string
		rowData       []DomainData
		expectedCount int
		mockResonse   *http.Response
		mockError     error
		expectedError error
	}{
		{
			name:          "success - valid row",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "11 GB", "Unlimited", 10)},
			expectedCount: 1,
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "success - valid rows",
			rowData: []DomainData{
				makeDomain(false, "test_domain1", "Test Domain 1", "11 GB", "Unlimited", 24),
				makeDomain(false, "test_domain2", "Test Domain 2", "11 GB", "Unlimited", 24),
				makeDomain(true, "test_domain3", "Test Domain 3", "11 GB", "Unlimited", 24),
				makeDomain(false, "test_domain4", "Test Domain 4", "11 GB", "Unlimited", 24),
				makeDomain(true, "test_domain5", "Test Domain 5", "11 GB", "Unlimited", 24),
			},
			expectedCount: 5,
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "error - empty domain",
			rowData: []DomainData{
				makeDomain(false, "", "Test Domain 1", "11 GB", "Unlimited", 23),
				makeDomain(false, "test_domain4", "Test Domain 4", "11 GG", "Unlimited", 23),
				makeDomain(false, "test_domain5", "Test Domain 5", "abc GB", "Unlimited", 23),
				makeDomain(false, "test_domain6", "Test Domain 6", "Unlimited", "Unlimited", 23),
			},
			expectedCount: 0,
			mockError:     nil,
			expectedError: apperrors.ErrEmptyDomain,
		},
		{
			name: "error - invalid memory suffix",
			rowData: []DomainData{
				makeDomain(false, "", "Test Domain 1", "11 GB", "Unlimited", 22),
				makeDomain(true, "test_domain4", "Test Domain 4", "11 GG", "Unlimited", 22),
				makeDomain(false, "test_domain5", "Test Domain 5", "abc GB", "Unlimited", 22),
				makeDomain(true, "test_domain6", "Test Domain 6", "Unlimited", "Unlimited", 22),
			},
			expectedCount: 0,
			mockError:     nil,
			expectedError: apperrors.ErrInvalidMemorySuffix,
		},
		{
			name: "error - invalid memory value",
			rowData: []DomainData{
				makeDomain(true, "", "Test Domain 1", "11 GB", "Unlimited", 27),
				makeDomain(true, "test_domain4", "Test Domain 4", "11 GG", "Unlimited", 27),
				makeDomain(false, "test_domain5", "Test Domain 5", "abc GB", "Unlimited", 27),
				makeDomain(false, "test_domain6", "Test Domain 6", "Unlimited", "Unlimited", 27),
			},
			expectedCount: 0,
			mockError:     nil,
			expectedError: apperrors.ErrInvalidMemoryValue,
		},
		{
			name:          "error - get request",
			rowData:       nil,
			expectedCount: 0,
			mockError:     errors.New("some http error"),
			expectedError: apperrors.ErrGetRequestFailed,
		},
		{
			name:          "error - empty table",
			rowData:       []DomainData{},
			expectedCount: 0,
			mockError:     nil,
			expectedError: nil,
		},
	}

	tmpl, err := template.New("domain").Parse(domainTemplate)
	if err != nil {
		t.Fatalf("Cannot compile template: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, r.Method)

				var result bytes.Buffer
				err = tmpl.Execute(&result, tt.rowData)
				assert.NoError(t, err)

				resp := makeResponse(http.StatusOK, result.String(), nil)
				return resp, tt.mockError
			}

			domainParser := newTestDomainParser(handler)
			domains, err := domainParser.Parse(t.Context(), "test_server")

			assert.ErrorIs(t, err, tt.expectedError)

			if tt.expectedError == nil {
				assert.Len(t, domains, tt.expectedCount)

				for i := range len(domains) {
					ok, err := tt.rowData[i].Compare(*domains[i])
					assert.NoError(t, err)
					assert.True(t, ok)
				}
			}
		})
	}
}

func TestParseRow(t *testing.T) {
	tests := []struct {
		name          string
		rowData       []DomainData
		expectedError error
	}{
		{
			name:          "success - valid row",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "11 GB", "Unlimited", 10)},
			expectedError: nil,
		},
		{
			name:          "success - valid row, zero quota/used",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "0", "0", 11)},
			expectedError: nil,
		},
		{
			name:          "success - valid row, empty display name",
			rowData:       []DomainData{makeDomain(false, "", "", "11 GB", "Unlimited", 24)},
			expectedError: apperrors.ErrEmptyDomain,
		},
		{
			name:          "error - invalid memory suffix",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "11 GG", "Unlimited", 12)},
			expectedError: apperrors.ErrInvalidMemorySuffix,
		},
		{
			name:          "error - invalid memory value (letters)",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "abc GB", "Unlimited", 20)},
			expectedError: apperrors.ErrInvalidMemoryValue,
		},
		{
			name:          "error - invalid memory value (invalid float)",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "10,2 GB", "Unlimited", 23)},
			expectedError: apperrors.ErrInvalidMemoryValue,
		},
		{
			name:          "error - invalid memory value (unlimited quota)",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "Unlimited", "Unlimited", 23)},
			expectedError: apperrors.ErrInvalidMemoryValue,
		},
		{
			name:          "error - invalid memory value (negative memory)",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "-1 GB", "Unlimited", 23)},
			expectedError: apperrors.ErrInvalidMemoryValue,
		},
		{
			name:          "error - invalid unlimited value",
			rowData:       []DomainData{makeDomain(false, "test_domain", "Test Domain", "1 GB", "Unlimited invalid", 23)},
			expectedError: apperrors.ErrInvalidMemorySuffix,
		},
	}

	tmpl, err := template.New("domain-row").Parse(domainTemplate)
	if err != nil {
		t.Fatalf("cannot compile template: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bytes.Buffer
			err = tmpl.Execute(&result, tt.rowData)
			assert.NoError(t, err)

			row, err := goquery.NewDocumentFromReader(&result)
			assert.NoError(t, err)

			domain, err := parseRow(row.Find("tbody tr"))
			assert.ErrorIs(t, err, tt.expectedError)

			if tt.expectedError == nil {
				ok, err := tt.rowData[0].Compare(*domain)
				assert.NoError(t, err)
				assert.True(t, ok)
			}
		})
	}
}
