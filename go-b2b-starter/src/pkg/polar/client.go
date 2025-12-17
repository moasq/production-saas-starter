package polar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides a low-level HTTP client for Polar API
// This is a generic HTTP wrapper - business logic should be in higher layers
type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
	debug       bool
}

func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Client{
		accessToken: config.AccessToken,
		baseURL:     config.BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		debug: config.Debug,
	}, nil
}

// Get performs a GET request to the Polar API
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.doRequest(ctx, "GET", path, nil)
}

// Patch performs a PATCH request to the Polar API
func (c *Client) Patch(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, "PATCH", path, body)
}

// Post performs a POST request to the Polar API
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, "POST", path, body)
}

// doRequest performs an HTTP request to the Polar API
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.debug {
		fmt.Printf("[Polar Client] %s %s\n", method, url)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Polar API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// DecodeJSON is a helper to decode JSON response
func DecodeJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}
	return nil
}
