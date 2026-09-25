package polar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrDisabled = errors.New("billing is disabled")
var ErrNotFound = errors.New("billing resource not found")
var ErrInvalidResponse = errors.New("invalid billing provider response")

// Match Polar's current stable contract, independently of SDK/package releases.
const APIVersion = "2026-04"
const maxResponseBytes = 2 << 20

type Client struct {
	config     Config
	httpClient *http.Client
}

func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Client{config: *config, httpClient: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (c *Client) ProductID() string { return c.config.ProductID }

func (c *Client) Enabled() bool { return c.config.Enabled }

// GetJSON never includes provider bodies or credentials in user-visible errors.
func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	if !c.Enabled() {
		return ErrDisabled
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.BaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Polar-Version", APIVersion)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("billing provider request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("billing provider returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return ErrInvalidResponse
	}
	if resp.StatusCode == http.StatusNotFound {
		// An invalid/retired API version also returns 404. Only a documented
		// missing resource may be treated as a customer with no subscription.
		var problem struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &problem) == nil && problem.Error == "ResourceNotFound" {
			return ErrNotFound
		}
		return ErrInvalidResponse
	}
	body = bytes.TrimSpace(body)
	// JSON null decodes successfully into a struct without setting its fields.
	// Require a single object so malformed data cannot masquerade as free status.
	if len(body) == 0 || body[0] != '{' || json.Unmarshal(body, out) != nil {
		return ErrInvalidResponse
	}
	return nil
}
