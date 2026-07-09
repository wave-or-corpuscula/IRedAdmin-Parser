// Package client provides http client for local requests
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"iredparser/internal/parser"
	apperrors "iredparser/pkg/errors"

	"github.com/PuerkitoBio/goquery"
)

type AuthConfig struct {
	Server   string `json:"server"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Client struct {
	httpClient *http.Client
	server     string
}

func NewClient(serverName string) *Client {
	jar, _ := cookiejar.New(nil)

	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Jar:       jar,
		Transport: customTransport,
		Timeout:   time.Duration(parser.HTTPTimeoutSeconds) * time.Second,
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	return http.ErrUseLastResponse
		// },
	}

	return &Client{client, serverName}
}

func (c *Client) ConfigureClient(config AuthConfig) {
}

func (c *Client) GetServer() string {
	return c.server
}

func (c *Client) GetBaseURL() string {
	return parser.CreateBaseURL(c.server)
}

// func AuthClient(ctx context.Context, c *http.Client, config AuthConfig) (*Client, error) {
// 	err := c.Auth(ctx, config)
// 	if err != nil {
// 		return nil, fmt.Errorf("client: %w", err)
// 	}
// 	return httpClient, nil
// }

func (c *Client) AuthServer(ctx context.Context, server string, login string, password string) error {
	baseURL := parser.CreateBaseURL(server)
	loginURL := baseURL + parser.LoginPath

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
	req.Header.Set("Origin", baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// return apperrors.ErrPostRequestFailed
		return fmt.Errorf("post request failed: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse html: %w", err)
	}

	title := doc.Find(".title").Text()
	if strings.Contains(title, "Login") {
		return apperrors.ErrInvalidCredentials
	}

	return nil
}

func (c *Client) Auth(ctx context.Context, config AuthConfig) error {
	var targetServer string
	if config.Server != "" {
		targetServer = config.Server
	} else if c.server != "" {
		targetServer = c.server
	} else {
		return fmt.Errorf("client: server address is required")
	}
	c.server = targetServer
	return c.AuthServer(ctx, targetServer, config.Login, config.Password)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) GetFromBase(ctx context.Context, path string) ([]byte, error) {
	baseURL := parser.CreateBaseURL(c.server)
	return c.Get(ctx, baseURL+path)
}
