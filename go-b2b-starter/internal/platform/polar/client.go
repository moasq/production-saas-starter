package polar

import (
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
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("billing provider request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("billing provider returned HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(out)
}
