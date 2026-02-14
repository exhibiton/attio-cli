package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) ListWebhooks(ctx context.Context, limit int, offset int) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := withQuery("/v2/webhooks", query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateWebhook(ctx context.Context, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	if err := c.do(ctx, http.MethodPost, "/v2/webhooks", body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetWebhook(ctx context.Context, webhookID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/webhooks/%s", pathPart(webhookID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateWebhook(ctx context.Context, webhookID string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/webhooks/%s", pathPart(webhookID))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, webhookID string) error {
	path := fmt.Sprintf("/v2/webhooks/%s", pathPart(webhookID))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
