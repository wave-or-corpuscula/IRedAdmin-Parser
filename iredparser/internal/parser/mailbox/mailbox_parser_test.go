package mailboxparser

import (
	"bytes"
	"html/template"
	"io"
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	apperrors "iredparser/pkg/errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func makeResponse(code int, body string, header http.Header) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     header,
	}
}

func getTestClient(handler func(req *http.Request) (*http.Response, error)) *client.Client {
	jar, _ := cookiejar.New(nil)

	httpClient := &http.Client{
		Transport: &mockTransport{
			roundTripFunc: handler,
		},
		Jar: jar,
	}

	return client.NewClientRaw(httpClient)
}

func getTestMailboxParser(handler func(req *http.Request) (*http.Response, error), workers int) *MailboxParser {
	c := getTestClient(handler)

	parser := NewMailboxParser(c, workers)
	return parser
}

type PageData struct {
	Num int
}

const pagesTemplate = `
<span class="pages">
    {{range .}}
        {{if eq . "1"}}
            <a href="#" class="active"><span>{{.}}</span></a>
        {{else}}
            <a href="/iredadmin/users/domain.com/page/{{.}}?order_name=quota"><span>{{.}}</span></a>
        {{end}}
    {{end}}
</span>
`

func TestGetPagesAmount(t *testing.T) {
	workers := 1

	tests := []struct {
		name          string
		pages         []string
		server        string
		domain        parser.Domain
		expectedError error
	}{
		{
			name:          "success - valid row",
			pages:         []string{"1"},
			server:        "mailserver",
			domain:        parser.Domain{Name: "test.com"},
			expectedError: nil,
		},
		{
			name:          "success - valid rows",
			pages:         []string{"1", "2", "3", "4"},
			server:        "mailserver",
			domain:        parser.Domain{Name: "test.com"},
			expectedError: nil,
		},
		{
			name:          "success - invalid in middle",
			pages:         []string{"1", "abc", "invalid", "4"},
			server:        "mailserver",
			domain:        parser.Domain{Name: "test.com"},
			expectedError: nil,
		},
		{
			name:          "error - only invalid",
			pages:         []string{"abc"},
			server:        "mailserver",
			domain:        parser.Domain{Name: "test.com"},
			expectedError: apperrors.ErrInvalidPageValue,
		},
		{
			name:          "error - last invalid",
			pages:         []string{"1", "2", "3", "4.4"},
			server:        "mailserver",
			domain:        parser.Domain{Name: "test.com"},
			expectedError: apperrors.ErrInvalidPageValue,
		},
		{
			name:          "error - negative page",
			pages:         []string{"-1"},
			server:        "mailserver",
			domain:        parser.Domain{Name: "test.com"},
			expectedError: apperrors.ErrInvalidPageValue,
		},
		{
			name:          "error - zero pages (impossible)",
			pages:         []string{"0"},
			server:        "mailserver",
			domain:        parser.Domain{Name: "test.com"},
			expectedError: apperrors.ErrInvalidPageValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatalf("expected GET method, got %v", req.Method)
				}

				t.Log(req.URL)

				tmpl, err := template.New("pagination").Parse(pagesTemplate)
				if err != nil {
					t.Fatalf("cannot compile template: %v", err)
				}

				var buf bytes.Buffer
				err = tmpl.Execute(&buf, tt.pages)
				if err != nil {
					t.Fatalf("cannot execute tamplate: %v", err)
				}

				t.Log(buf.String())
				resp := makeResponse(http.StatusOK, buf.String(), nil)
				return resp, nil
			}

			parser := getTestMailboxParser(handler, workers)
			gotPages, err := parser.getPagesAmount(t.Context(), tt.server, tt.domain)

			assert.ErrorIs(t, err, tt.expectedError)

			if tt.expectedError == nil {
				assert.Equal(t, gotPages, len(tt.pages))
			}
		})
	}
}

const mailboxTemplate = `
<table>
	<tbody>
		{{ range . }}
		<tr class="{{ if .Disabled }}disabled{{ else }}active{{ end }}">
			<td class="checkbox">
					<input type="checkbox" name="mail" class="checkbox" value="{{ .Addres }}" alt="active">
			</td>
			<td class="vcenter">
					<a href="/iredadmin/profile/user/general/{{ .Addres }}">
							<i class="fa fa-cog fa-lg fr-space"></i>
					</a>
					<a href="/iredadmin/profile/user/general/{{ .Addres }}">{{ .DisplayName }}</a>
			</td>
			<td class="vcenter">
					<span><strong>{{ .DisplayName }}</strong></span>
					<span class="color-grey"><em>@{{ .Domain }}</em></span>
			</td>
			<td class="vcenter"></td>
			<td class="vcenter"></td>
			<td class="vcenter" data-sort-value="0">{{ .Quota }}</td>
		</tr>
	 {{ end }}
	</tbody>
</table>
`

type MailboxData struct {
	Disabled    bool
	IsAdmin     bool
	DisplayName string
	Addres      string
	Domain      string
	Quota       string
}

func makeMailbox(disabled bool, isAdmin bool, addres string, displayName string, quota string) *MailboxData {
	if !strings.Contains(addres, "@") {
		log.Fatalln("no '@' in mailbox addres")
	}

	domain := strings.Split(addres, "@")[1]

	return &MailboxData{
		Disabled:    disabled,
		IsAdmin:     isAdmin,
		DisplayName: displayName,
		Addres:      addres,
		Domain:      domain,
		Quota:       quota,
	}
}

func makeMailboxesTable(pages int, mailboxes ...*MailboxData) string {
	mailTmpl, err := template.New("mailboxes").Parse(mailboxTemplate)
	if err != nil {
		log.Fatalf("cannot compile mailbox template: %v", err)
	}

	var mailboxesTable bytes.Buffer
	err = mailTmpl.Execute(&mailboxesTable, mailboxes)
	if err != nil {
		log.Fatalf("cannot execute template: %v", err)
	}

	pagesTmpl, err := template.New("pages").Parse(pagesTemplate)
	if err != nil {
		log.Fatalf("cannot compile pages template: %v", err)
	}

	pageValues := make([]string, pages)
	for i := range pages {
		pageValues[i] = strconv.Itoa(i + 1)
	}

	var pagesFooter bytes.Buffer
	err = pagesTmpl.Execute(&pagesFooter, pageValues)
	if err != nil {
		log.Fatalf("cannot execute pages template: %v", err)
	}

	mailboxPage := mailboxesTable.String() + pagesFooter.String()

	return mailboxPage
}

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		pages         int
		mailboxes     []*MailboxData
		workers       int
		expectedError error
	}{
		{
			name:  "success - valid mailboxes",
			pages: 1,
			mailboxes: []*MailboxData{
				makeMailbox(false, false, "test1@domain.com", "test1", "0 Bytes / Unlimited"),
				makeMailbox(false, false, "test2@domain.com", "test2", "2 GB / 4 GB"),
				makeMailbox(false, false, "test3@domain.com", "test2", "345 KB / 2 GB"),
				makeMailbox(false, false, "test4@domain.com", "test4", "1 MB / Unlimited"),
			},
			workers:       1,
			expectedError: nil,
		}, // TODO: Добавить кейсов
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatalf("expected GET method, got %v", req.Method)
				}

				table := makeMailboxesTable(tt.pages, tt.mailboxes...)
				resp := makeResponse(http.StatusOK, table, nil)

				return resp, nil
			}

			mailboxParser := getTestMailboxParser(handler, tt.workers)
			result, err := mailboxParser.Parse(t.Context(), "testmailserver", parser.Domain{Name: "test.com"})
			assert.ErrorIs(t, err, tt.expectedError)

			if tt.expectedError == nil {
				assert.Len(t, result.Mailboxes, len(tt.mailboxes))
			}
		})
	}
}
