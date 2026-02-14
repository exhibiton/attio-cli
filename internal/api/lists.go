package api

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) ListLists(ctx context.Context) ([]map[string]any, error) {
	var resp dataArrayResponse
	if err := c.do(ctx, http.MethodGet, "/v2/lists", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateList(ctx context.Context, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	if err := c.do(ctx, http.MethodPost, "/v2/lists", body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetList(ctx context.Context, list string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/lists/%s", pathPart(list))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateList(ctx context.Context, list string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/lists/%s", pathPart(list))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
