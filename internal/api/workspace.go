package api

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) ListMembers(ctx context.Context) ([]map[string]any, error) {
	var resp dataArrayResponse
	if err := c.do(ctx, http.MethodGet, "/v2/workspace_members", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetMember(ctx context.Context, workspaceMemberID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/workspace_members/%s", pathPart(workspaceMemberID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
