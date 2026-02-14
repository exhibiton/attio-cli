package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) CreateEntry(ctx context.Context, list string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/lists/%s/entries", pathPart(list))
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) AssertEntry(ctx context.Context, list string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/lists/%s/entries", pathPart(list))
	if err := c.do(ctx, http.MethodPut, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) QueryEntries(ctx context.Context, list string, filter any, sorts any, limit int, offset int) ([]map[string]any, error) {
	var resp dataArrayResponse

	body := map[string]any{}
	if filter != nil {
		body["filter"] = filter
	}
	if sorts != nil {
		body["sorts"] = sorts
	}
	if limit > 0 {
		body["limit"] = limit
	}
	if offset > 0 {
		body["offset"] = offset
	}

	path := fmt.Sprintf("/v2/lists/%s/entries/query", pathPart(list))
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetEntry(ctx context.Context, list string, entryID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/lists/%s/entries/%s", pathPart(list), pathPart(entryID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateEntry(ctx context.Context, list string, entryID string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/lists/%s/entries/%s", pathPart(list), pathPart(entryID))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ReplaceEntry(ctx context.Context, list string, entryID string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/lists/%s/entries/%s", pathPart(list), pathPart(entryID))
	if err := c.do(ctx, http.MethodPut, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) DeleteEntry(ctx context.Context, list string, entryID string) error {
	path := fmt.Sprintf("/v2/lists/%s/entries/%s", pathPart(list), pathPart(entryID))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) ListEntryAttributeValues(ctx context.Context, list string, entryID string, attribute string, showHistoric bool, limit int, offset int) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if showHistoric {
		query.Set("show_historic", "true")
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := fmt.Sprintf("/v2/lists/%s/entries/%s/attributes/%s/values", pathPart(list), pathPart(entryID), pathPart(attribute))
	path = withQuery(path, query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
