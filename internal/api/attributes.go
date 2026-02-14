package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) ListAttributes(ctx context.Context, target string, identifier string, showArchived bool, limit int, offset int) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if showArchived {
		query.Set("show_archived", "true")
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := fmt.Sprintf("/v2/%s/%s/attributes", pathPart(target), pathPart(identifier))
	path = withQuery(path, query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateAttribute(ctx context.Context, target string, identifier string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/%s/%s/attributes", pathPart(target), pathPart(identifier))
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetAttribute(ctx context.Context, target string, identifier string, attribute string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/%s/%s/attributes/%s", pathPart(target), pathPart(identifier), pathPart(attribute))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateAttribute(ctx context.Context, target string, identifier string, attribute string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/%s/%s/attributes/%s", pathPart(target), pathPart(identifier), pathPart(attribute))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ListSelectOptions(ctx context.Context, target string, identifier string, attribute string, showArchived bool) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if showArchived {
		query.Set("show_archived", "true")
	}

	path := fmt.Sprintf("/v2/%s/%s/attributes/%s/options", pathPart(target), pathPart(identifier), pathPart(attribute))
	path = withQuery(path, query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateSelectOption(ctx context.Context, target string, identifier string, attribute string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/%s/%s/attributes/%s/options", pathPart(target), pathPart(identifier), pathPart(attribute))
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateSelectOption(ctx context.Context, target string, identifier string, attribute string, option string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/%s/%s/attributes/%s/options/%s", pathPart(target), pathPart(identifier), pathPart(attribute), pathPart(option))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ListStatuses(ctx context.Context, target string, identifier string, attribute string, showArchived bool) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if showArchived {
		query.Set("show_archived", "true")
	}

	path := fmt.Sprintf("/v2/%s/%s/attributes/%s/statuses", pathPart(target), pathPart(identifier), pathPart(attribute))
	path = withQuery(path, query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateStatus(ctx context.Context, target string, identifier string, attribute string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/%s/%s/attributes/%s/statuses", pathPart(target), pathPart(identifier), pathPart(attribute))
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateStatus(ctx context.Context, target string, identifier string, attribute string, status string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/%s/%s/attributes/%s/statuses/%s", pathPart(target), pathPart(identifier), pathPart(attribute), pathPart(status))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
