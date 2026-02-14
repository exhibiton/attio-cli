package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) CreateRecord(ctx context.Context, object string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/objects/%s/records", pathPart(object))
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) AssertRecord(ctx context.Context, object string, matchingAttribute string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}

	query := url.Values{}
	if matchingAttribute != "" {
		query.Set("matching_attribute", matchingAttribute)
	}
	path := withQuery(fmt.Sprintf("/v2/objects/%s/records", pathPart(object)), query)
	if err := c.do(ctx, http.MethodPut, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) QueryRecords(ctx context.Context, object string, filter any, sorts any, limit int, offset int) ([]map[string]any, error) {
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

	path := fmt.Sprintf("/v2/objects/%s/records/query", pathPart(object))
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) SearchRecords(ctx context.Context, query string, limit int, objects []string, requestAs any) ([]map[string]any, error) {
	var resp dataArrayResponse

	body := map[string]any{
		"query":   query,
		"objects": objects,
	}
	if limit > 0 {
		body["limit"] = limit
	}
	if requestAs == nil {
		body["request_as"] = map[string]any{"type": "workspace"}
	} else {
		body["request_as"] = requestAs
	}

	if err := c.do(ctx, http.MethodPost, "/v2/objects/records/search", body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetRecord(ctx context.Context, object string, recordID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/objects/%s/records/%s", pathPart(object), pathPart(recordID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateRecord(ctx context.Context, object string, recordID string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/objects/%s/records/%s", pathPart(object), pathPart(recordID))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ReplaceRecord(ctx context.Context, object string, recordID string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/objects/%s/records/%s", pathPart(object), pathPart(recordID))
	if err := c.do(ctx, http.MethodPut, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) DeleteRecord(ctx context.Context, object string, recordID string) error {
	path := fmt.Sprintf("/v2/objects/%s/records/%s", pathPart(object), pathPart(recordID))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) ListRecordAttributeValues(ctx context.Context, object string, recordID string, attribute string, showHistoric bool, limit int, offset int) ([]map[string]any, error) {
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

	path := fmt.Sprintf("/v2/objects/%s/records/%s/attributes/%s/values", pathPart(object), pathPart(recordID), pathPart(attribute))
	path = withQuery(path, query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ListRecordEntries(ctx context.Context, object string, recordID string, limit int, offset int) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := fmt.Sprintf("/v2/objects/%s/records/%s/entries", pathPart(object), pathPart(recordID))
	path = withQuery(path, query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
