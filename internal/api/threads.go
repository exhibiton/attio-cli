package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) ListThreads(ctx context.Context, object string, recordID string, list string, entryID string, limit int, offset int) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if object != "" {
		query.Set("object", object)
	}
	if recordID != "" {
		query.Set("record_id", recordID)
	}
	if list != "" {
		query.Set("list", list)
	}
	if entryID != "" {
		query.Set("entry_id", entryID)
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := withQuery("/v2/threads", query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetThread(ctx context.Context, threadID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/threads/%s", pathPart(threadID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
