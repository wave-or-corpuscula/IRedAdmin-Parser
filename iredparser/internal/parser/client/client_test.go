package client

import (
	"iredparser/common"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
