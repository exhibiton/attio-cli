package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.attio.com"
const defaultUserAgent = "attio-cli/dev"

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string
}

func NewClient(apiKey string, baseURL string) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	transport := &RetryTransport{
		Base:       http.DefaultTransport,
		MaxRetries: 3,
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		baseURL:   baseURL,
		apiKey:    apiKey,
		userAgent: defaultUserAgent,
	}
}

func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = strings.TrimSpace(userAgent)
}

func (c *Client) SetTimeout(timeout time.Duration) {
	if timeout <= 0 {
		return
	}
	c.httpClient.Timeout = timeout
}

func (c *Client) GetSelf(ctx context.Context) (*Self, error) {
	var out Self
	if err := c.do(ctx, http.MethodGet, "/v2/self", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) do(ctx context.Context, method string, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}

	if result == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
