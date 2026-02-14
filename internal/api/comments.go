package api

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) CreateComment(ctx context.Context, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	if err := c.do(ctx, http.MethodPost, "/v2/comments", body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetComment(ctx context.Context, commentID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/comments/%s", pathPart(commentID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) DeleteComment(ctx context.Context, commentID string) error {
	path := fmt.Sprintf("/v2/comments/%s", pathPart(commentID))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
