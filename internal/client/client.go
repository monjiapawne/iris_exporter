package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Client holds configuration required for the http request
type Client struct {
	Host       string
	APIKey     string
	SSLVerify  bool
	httpClient *http.Client
	logger     *slog.Logger
}

// newHTTPClient wrapper to create an production ready http client
func newHTTPClient(sslVerify bool) *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: !sslVerify},
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

// NewClient creates a new httpClient along with the parameters to connect to the target for scraping
func NewClient(apiKey, schema, host, port string, sslVerify bool, logger *slog.Logger) *Client {
	return &Client{
		Host:       schema + "://" + net.JoinHostPort(host, port),
		APIKey:     apiKey,
		SSLVerify:  sslVerify,
		httpClient: newHTTPClient(sslVerify),
		logger:     logger,
	}
}

// APICall is a generic get request to the api configured in c *Client
// it decodes into to the generic type provided by the caller
func APICall[R any](c *Client, path string) (R, error) {
	// Limit the response from the target's API
	const maxResponseBytes = 10 * 1024 * 1024 // 10 MB
	var dest R
	endpoint := c.Host + path

	// Setup http request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		return dest, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Send http request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("http request failed", "endpoint", endpoint, "err", err)
		return dest, err
	}
	defer resp.Body.Close()
	c.logger.Debug("http request", "endpoint", endpoint, "status code", resp.StatusCode, "content length", resp.ContentLength)

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Error("http request failed", "endpoint", endpoint, "status code", resp.StatusCode)
		return dest, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	// Unmarshal response into provided type
	err = json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&dest)
	if err != nil {
		c.logger.Error("failed to decode response", "endpoint", endpoint, "err", err)
		return dest, err
	}

	return dest, nil
}
