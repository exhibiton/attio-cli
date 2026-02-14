package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) ListNotes(ctx context.Context, parentObject string, parentRecordID string, limit int, offset int) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if parentObject != "" {
		query.Set("parent_object", parentObject)
	}
	if parentRecordID != "" {
		query.Set("parent_record_id", parentRecordID)
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := withQuery("/v2/notes", query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateNote(ctx context.Context, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	if err := c.do(ctx, http.MethodPost, "/v2/notes", body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetNote(ctx context.Context, noteID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/notes/%s", pathPart(noteID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) DeleteNote(ctx context.Context, noteID string) error {
	path := fmt.Sprintf("/v2/notes/%s", pathPart(noteID))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
