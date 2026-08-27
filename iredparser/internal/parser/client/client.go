// Package client provides http client for local requests
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"iredparser/common"
	"iredparser/internal/parser"
	"iredparser/pkg/utils"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	apperrors "iredparser/pkg/errors"
)

const (
	RequestTimeout = 30
	TestMailbox    = "test@rmzu.by"
)

type Client struct {
	httpClient   *http.Client
	cookieString string
}

func createHTTPClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	customTransport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Jar:       jar,
		Transport: customTransport,
		Timeout:   RequestTimeout * time.Second,
	}
	return client, nil
}

func NewClientRaw(client *http.Client) *Client {
	return &Client{
		httpClient: client,
	}
}

func NewClient() (*Client, error) {
	client, err := createHTTPClient()
	if err != nil {
		return nil, err
	}

	return &Client{httpClient: client}, nil
}

func (c *Client) ConfigureClient(config common.ServerConfig) {
}

func (c *Client) AuthServer(ctx context.Context, server string, login string, password string) error {
	baseURL := parser.CreateBaseURL(server)
	loginURL := baseURL + parser.LoginPath

	return c.AuthServerURL(ctx, loginURL, login, password)
}

func (c *Client) AuthServerURL(ctx context.Context, loginURL string, login string, password string) error {
	data := url.Values{}
	data.Set("username", login)
	data.Set("password", password)
	data.Set("form_login", "Login")
	data.Set("lang", "en_EN")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return apperrors.ErrPostRequestCreation
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
	req.Header.Set("Referer", loginURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apperrors.ErrPostRequestFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return apperrors.ErrInternalServerError
	}

	if resp.StatusCode != http.StatusOK {
		return apperrors.ErrStatusCodeNotOK.Wrapf("received status: %d", resp.StatusCode)
	}

	err = utils.ExtractLoginError(resp.Body)
	if err != nil {
		return err
	}

	cookies := c.httpClient.Jar.Cookies(req.URL)
	for _, cookie := range cookies {
		if cookie.Name == "iRedAdmin-LDAP" {
			c.cookieString = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
			break
		}
	}

	return nil
}

func (c *Client) Auth(ctx context.Context, config common.ServerConfig) error {
	return c.AuthServer(ctx, config.Server, config.Login, config.Password)
}

func (c *Client) ChangePassword(ctx context.Context, server string, mailbox string, newPassword string) error {
	csrfToken, err := c.GetCSRFToken(ctx, server, mailbox)
	if err != nil {
		return fmt.Errorf("client: failed to get CSRF token: %w", err)
	}

	changeURL := parser.CreateChangePasswordPath(server, mailbox)

	data := url.Values{}
	data.Set("csrf_token", csrfToken)
	data.Set("newpw", newPassword)
	data.Set("confirmpw", newPassword)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, changeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return apperrors.ErrPostRequestCreation
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
	req.Header.Set("Referer", changeURL)
	req.Header.Set("Origin", server)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: change password request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return apperrors.ErrInternalServerError
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("change password failed with status %d: %s", resp.StatusCode, string(body))
	}

	err = utils.ExtractChangePasswordErrors(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetCSRFToken(ctx context.Context, server string, mailbox string) (string, error) {
	url := parser.CreateChangePasswordPath(server, mailbox)
	body, err := c.Get(ctx, url)
	if err != nil {
		return "", err
	}

	token, err := utils.ExtractCSRFToken(body)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, apperrors.ErrGetRequestCreation
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.ErrGetRequestFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return nil, apperrors.ErrInternalServerError
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.ErrUnexpectedStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) GetFromServer(ctx context.Context, server string, path string) ([]byte, error) {
	baseURL := parser.CreateBaseURL(server)
	url := baseURL + path
	return c.Get(ctx, url)
}

func (c *Client) GetCookieString() string {
	return c.cookieString
}
