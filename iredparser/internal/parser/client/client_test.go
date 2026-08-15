package client

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "iredparser/pkg/errors"
)

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func newTestClient(handler func(req *http.Request) (*http.Response, error)) *Client {
	jar, _ := cookiejar.New(nil)

	client := &Client{
		httpClient: &http.Client{
			Transport: &mockTransport{
				roundTripFunc: handler,
			},
			Jar: jar,
		},
	}

	return client
}

func makeResponse(statusCode int, body string, header http.Header) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     header,
	}
}

func TestAuthServerURL(t *testing.T) {
	tests := []struct {
		name           string
		login          string
		password       string
		mockResponse   *http.Response
		mockError      error
		expectedError  error
		expectedCookie string
	}{
		{
			name:     "successful login",
			login:    "user@test.com",
			password: "correct123",
			mockResponse: makeResponse(
				http.StatusOK,
				`<html>Welcome</html>`,
				http.Header{"Set-Cookie": []string{"iRedAdmin-LDAP=abc123; Path=/"}},
			),
			mockError:      nil,
			expectedError:  nil,
			expectedCookie: "iRedAdmin-LDAP=abc123",
		},
		{
			name:     "login failed - incorrect credentials",
			login:    "user@test.com",
			password: "wrong",
			mockResponse: makeResponse(
				http.StatusOK,
				`<div class="note-error"><p><strong>Error:</strong> username or password is incorrect</p></div>`,
				make(http.Header),
			),
			mockError:      nil,
			expectedError:  apperrors.ErrIncorrectCredentials,
			expectedCookie: "",
		},
		{
			name:     "login failed - invalid username format",
			login:    "invalid",
			password: "pass",
			mockResponse: makeResponse(
				http.StatusOK,
				`<div class="note-error"><p><strong>Error:</strong> must be an valid email</p></div>`,
				make(http.Header),
			),
			mockError:      nil,
			expectedError:  apperrors.ErrInvalidUsername,
			expectedCookie: "",
		},
		{
			name:     "login failed - missing required fields",
			login:    "",
			password: "",
			mockResponse: makeResponse(
				http.StatusOK,
				`<div class="note-error"><p><strong>Error:</strong> username required</p></div>`,
				make(http.Header),
			),
			mockError:      nil,
			expectedError:  apperrors.ErrLoginRequired,
			expectedCookie: "",
		},
		{
			name:     "http server error",
			login:    "user@test.com",
			password: "pass",
			mockResponse: makeResponse(
				http.StatusInternalServerError,
				`Internal Server Error`,
				make(http.Header),
			),
			mockError:      nil,
			expectedError:  apperrors.ErrInternalServerError,
			expectedCookie: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Errorf("expexted POST, got %s", req.Method)
				}

				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}
				defer req.Body.Close()

				values, err := url.ParseQuery(string(body))
				if err != nil {
					t.Fatalf("failed to parse form: %s", err)
				}

				assert.Equal(t, values.Get("username"), tt.login)
				assert.Equal(t, values.Get("password"), tt.password)
				assert.Equal(t, values.Get("form_login"), "Login")
				assert.Equal(t, values.Get("lang"), "en_EN")

				return tt.mockResponse, tt.mockError
			}

			client := newTestClient(handler)

			err := client.AuthServerURL(t.Context(), "http://test.com/login", tt.login, tt.password)

			assert.ErrorIs(t, err, tt.expectedError)
			assert.Equal(t, client.cookieString, tt.expectedCookie)
		})
	}
}

func getTestServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/iredadmin/login":
			http.SetCookie(w, &http.Cookie{
				Name:  "iRedAdmin-LDAP",
				Value: "test-session",
				Path:  "/",
			})
			w.Header().Set("Location", "/iredadmin/dashboard")
			w.WriteHeader(http.StatusFound)

		case "/iredadmin/dashboard":
			if _, err := r.Cookie("iRedAdmin-LDAP"); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`<html><div class="title">Login</div></html>`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><div class="title">Dashboard</div></html>`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return server
}
