package api

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) ListObjects(ctx context.Context) ([]map[string]any, error) {
	var resp dataArrayResponse
	if err := c.do(ctx, http.MethodGet, "/v2/objects", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateObject(ctx context.Context, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	if err := c.do(ctx, http.MethodPost, "/v2/objects", body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetObject(ctx context.Context, object string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/objects/%s", pathPart(object))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateObject(ctx context.Context, object string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/objects/%s", pathPart(object))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
