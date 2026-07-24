// Package client provides http client for local requests
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"iredparser/common"
	"iredparser/internal/parser"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	apperrors "iredparser/pkg/errors"

	"github.com/PuerkitoBio/goquery"
)

const RequestTimeout = 30

type Client struct {
	httpClient *http.Client
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

func NewClient() (*Client, error) {
	client, err := createHTTPClient()
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

func (c *Client) ConfigureClient(config common.ServerConfig) {
}

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
		return fmt.Errorf("client: post request failed: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("client: failed to parse html: %w", err)
	}

	title := doc.Find(".title").Text()
	if strings.Contains(title, "Login") {
		return apperrors.ErrInvalidCredentials
	}

	return nil
}

func (c *Client) Auth(ctx context.Context, config common.ServerConfig) error {
	return c.AuthServer(ctx, config.Server, config.Login, config.Password)
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

func (c *Client) GetFromServer(ctx context.Context, server string, path string) ([]byte, error) {
	baseURL := parser.CreateBaseURL(server)
	url := baseURL + path
	return c.Get(ctx, url)
}
