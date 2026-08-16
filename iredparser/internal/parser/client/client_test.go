package client

import (
	"errors"
	"fmt"
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
		{
			name:           "do error",
			login:          "user@test.com",
			password:       "pass",
			mockResponse:   &http.Response{},
			mockError:      errors.New("post request failed"),
			expectedError:  apperrors.ErrPostRequestFailed,
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

func TestClientGET(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		mockResponse  *http.Response
		mockError     error
		expectedBody  []byte
		expectedError error
	}{
		{
			name: "successful get",
			url:  "http://someurl.com",
			mockResponse: makeResponse(
				http.StatusOK,
				`<html>Welcome</html>`,
				make(http.Header),
			),
			expectedBody:  []byte(`<html>Welcome</html>`),
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "internal server error",
			url:  "http://someurl.com",
			mockResponse: makeResponse(
				http.StatusInternalServerError,
				`<html>Internal serve error</html>`,
				make(http.Header),
			),
			expectedBody:  nil,
			mockError:     nil,
			expectedError: apperrors.ErrInternalServerError,
		},
		{
			name:          "invalid url",
			url:           "://invalid.url",
			mockResponse:  &http.Response{},
			expectedBody:  []byte(nil),
			mockError:     nil,
			expectedError: apperrors.ErrGetRequestCreation,
		},
		{
			name:          "unexpected status code",
			url:           "http://somebadurl.com",
			mockResponse:  makeResponse(http.StatusNotFound, "", make(http.Header)),
			expectedBody:  nil,
			mockError:     nil,
			expectedError: apperrors.ErrUnexpectedStatusCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", req.Method)
				}

				assert.Equal(t, tt.url, req.URL.String())

				return tt.mockResponse, tt.mockError
			}

			client := newTestClient(handler)

			body, err := client.Get(t.Context(), tt.url)
			assert.ErrorIs(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestGetCsrfToken(t *testing.T) {
	tests := []struct {
		name          string
		server        string
		mailbox       string
		mockResponse  *http.Response
		expectedToken string
		mockError     error
		expectedError error
	}{
		{
			name:    "success - token found",
			server:  "test",
			mailbox: "client@test.com",
			mockResponse: makeResponse(
				http.StatusOK,
				`<input name="csrf_token" value="token_value123">`,
				make(http.Header),
			),
			expectedToken: "token_value123",
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:    "error - token not found",
			server:  "test",
			mailbox: "client@test.com",
			mockResponse: makeResponse(
				http.StatusOK,
				`<html>No token here(((</html>`,
				make(http.Header),
			),
			expectedToken: "",
			mockError:     nil,
			expectedError: apperrors.ErrCannoFindCSRFToken,
		},
		{
			name:          "error - get fails",
			server:        "test",
			mailbox:       "client@test.com",
			mockResponse:  nil,
			expectedToken: "",
			mockError:     errors.New("network error"),
			expectedError: apperrors.ErrGetRequestFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Errorf("expected GET, got %v", req.Method)
				}

				return tt.mockResponse, tt.mockError
			}

			client := newTestClient(handler)

			token, err := client.GetCSRFToken(t.Context(), tt.server, tt.mailbox)

			assert.ErrorIs(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestChangePasswordUnit(t *testing.T) {
	tests := []struct {
		name          string
		server        string
		mailbox       string
		newPassword   string
		mockResponses map[string]*mockResponse // URL -> ответ
		expectedError error
	}{
		{
			name:        "successful password change",
			server:      "mail.test.com",
			mailbox:     "test@test.com",
			newPassword: "NewPass123!",
			mockResponses: map[string]*mockResponse{
				"GET /iredadmin/profile/user/password/test@test.com": {
					Response: makeResponse(http.StatusOK, `<input name="csrf_token" value="csrf123">`, make(http.Header)),
					Error:    nil,
				},
				"POST /iredadmin/profile/user/password/test@test.com": {
					Response: makeResponse(http.StatusOK, `<div class="note-success">Password changed</div>`, make(http.Header)),
					Error:    nil,
				},
			},
			expectedError: nil,
		},
		{
			name:        "failed to get CSRF token - Get error",
			server:      "mail.test.com",
			mailbox:     "test@test.com",
			newPassword: "NewPass123!",
			mockResponses: map[string]*mockResponse{
				"GET /iredadmin/profile/user/password/test@test.com": {
					Response: nil,
					Error:    errors.New("network error"),
				},
			},
			expectedError: apperrors.ErrGetRequestFailed,
		},
		{
			name:        "CSRF token not found in response",
			server:      "mail.test.com",
			mailbox:     "test@test.com",
			newPassword: "NewPass123!",
			mockResponses: map[string]*mockResponse{
				"GET /iredadmin/profile/user/password/test@test.com": {
					Response: makeResponse(http.StatusOK, `<html><body>No token here</body></html>`, make(http.Header)),
					Error:    nil,
				},
			},
			expectedError: fmt.Errorf("client: failed to get CSRF token: %w", apperrors.ErrCannoFindCSRFToken),
		},
		{
			name:        "change password - HTTP error 500",
			server:      "mail.test.com",
			mailbox:     "test@test.com",
			newPassword: "NewPass123!",
			mockResponses: map[string]*mockResponse{
				"GET /iredadmin/profile/user/password/test@test.com": {
					Response: makeResponse(http.StatusOK, `<input name="csrf_token" value="csrf123">`, make(http.Header)),
					Error:    nil,
				},
				"POST /iredadmin/profile/user/password/test@test.com": {
					Response: makeResponse(http.StatusInternalServerError, "Internal Server Error", make(http.Header)),
					Error:    nil,
				},
			},
			expectedError: apperrors.ErrInternalServerError,
		},
		{
			name:        "change password - validation error",
			server:      "mail.test.com",
			mailbox:     "test@test.com",
			newPassword: "123",
			mockResponses: map[string]*mockResponse{
				"GET /iredadmin/profile/user/password/test@test.com": {
					Response: makeResponse(http.StatusOK, `<input name="csrf_token" value="csrf123">`, make(http.Header)),
					Error:    nil,
				},
				"POST /iredadmin/profile/user/password/test@test.com": {
					Response: makeResponse(http.StatusOK, `<div class="note-error"><p>Password too weak</p></div>`, make(http.Header)),
					Error:    nil,
				},
			},
			expectedError: apperrors.New(
				apperrors.ErrTypeParsing,
				apperrors.ErrCodeChangePassowrd,
				errors.New("Password too weak"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(req *http.Request) (*http.Response, error) {
				key := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
				mockResp, ok := tt.mockResponses[key]
				if !ok {
					t.Fatalf("unexpected request: %s", key)
				}

				return mockResp.Response, mockResp.Error
			}

			client := newTestClient(handler)

			err := client.ChangePassword(t.Context(), tt.server, tt.mailbox, tt.newPassword)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if strings.Contains(tt.expectedError.Error(), "client: failed to get CSRF token") ||
					strings.Contains(tt.expectedError.Error(), "client: change password request failed") {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				} else {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Вспомогательная структура для мока
type mockResponse struct {
	Response *http.Response
	Error    error
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
