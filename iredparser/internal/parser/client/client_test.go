package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iredparser/common"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func newMockClient(roundTripFunc func(req *http.Request) (*http.Response, error)) *http.Client {
	jar, _ := cookiejar.New(nil)

	return &http.Client{
		Transport: &mockRoundTripper{roundTripFunc: roundTripFunc},
		Timeout:   30 * time.Second,
		Jar:       jar,
	}
}

func mockResponse(statusCode int, body string, cookies []*http.Cookie) *http.Response {
	resp := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}

	if len(cookies) > 0 {
		for _, c := range cookies {
			resp.Header.Add("Set-Cookie", c.String())
		}
	}
	return resp
}

func TestNewClient(t *testing.T) {
	t.Run("creates client with default config", func(t *testing.T) {
		c, err := NewClient()
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.NotNil(t, c.httpClient)
		assert.Equal(t, 30*time.Second, c.httpClient.Timeout)
		assert.NotNil(t, c.httpClient.Jar)
		assert.NotNil(t, c.httpClient.Transport)
	})
}

func TestClient_AuthServerURL(t *testing.T) {
	tests := []struct {
		name           string
		loginURL       string
		login          string
		password       string
		mockResponse   *http.Response
		mockError      error
		expectedError  error
		expectedCookie string
	}{
		{
			name:     "successful authentication",
			loginURL: "https://test.com/login",
			login:    "admin",
			password: "pass123",
			mockResponse: mockResponse(http.StatusOK, `<html>Login successful</html>`, []*http.Cookie{
				{Name: "iRedAdmin-LDAP", Value: "session123"},
			}),
			expectedError:  nil,
			expectedCookie: "iRedAdmin-LDAP=session123",
		},
		{
			name:         "authentication failed - wrong credentials",
			loginURL:     "https://test.com/login",
			login:        "wrong",
			password:     "wrong",
			mockResponse: mockResponse(http.StatusOK, `<html><div class="title">Login</div></html>`, nil),
			// expectedError:  apperrors.ErrLoginFailed,
			expectedError:  errors.New("login failed"),
			expectedCookie: "",
		},
		{
			name:           "authentication failed - request error",
			loginURL:       "https://test.com/login",
			login:          "admin",
			password:       "pass",
			mockResponse:   nil,
			mockError:      errors.New("connection refused"),
			expectedError:  errors.New("client: post request failed: connection refused"),
			expectedCookie: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
				if req.Method != "POST" {
					return nil, fmt.Errorf("expected POST, got %s", req.Method)
				}
				if req.URL.String() != tt.loginURL {
					return nil, fmt.Errorf("expected URL %s, got %s", tt.loginURL, req.URL.String())
				}
				if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
					return nil, fmt.Errorf("wrong Content-Type")
				}

				body, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				bodyStr := string(body)
				if !strings.Contains(bodyStr, "username="+tt.login) {
					return nil, fmt.Errorf("username not found in body")
				}
				if !strings.Contains(bodyStr, "password="+tt.password) {
					return nil, fmt.Errorf("password not found in body")
				}

				if tt.mockError != nil {
					return nil, tt.mockError
				}

				if tt.mockResponse == nil {
					return nil, fmt.Errorf("mockResponse is nil")
				}

				return tt.mockResponse, nil
			})

			c := &Client{httpClient: mockClient}

			err := c.AuthServerURL(context.Background(), tt.loginURL, tt.login, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCookie, c.cookieString)
			}
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

func TestAuthTestServer(t *testing.T) {
	server := getTestServer()
	defer server.Close()

	config := common.ServerConfig{}

	c, err := NewClient()
	assert.NoError(t, err)

	err = c.AuthServerURL(t.Context(), server.URL, config.Login, config.Password)
	assert.NoError(t, err)

	t.Log(c.GetCookieString())
}
