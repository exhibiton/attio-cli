package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) ListTasks(ctx context.Context, limit int, offset int, sort string, linkedObject string, linkedRecordID string, assignee string, isCompleted *bool) ([]map[string]any, error) {
	var resp dataArrayResponse

	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}
	if sort != "" {
		query.Set("sort", sort)
	}
	if linkedObject != "" {
		query.Set("linked_object", linkedObject)
	}
	if linkedRecordID != "" {
		query.Set("linked_record_id", linkedRecordID)
	}
	if assignee != "" {
		query.Set("assignee", assignee)
	}
	if isCompleted != nil {
		query.Set("is_completed", strconv.FormatBool(*isCompleted))
	}

	path := withQuery("/v2/tasks", query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateTask(ctx context.Context, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	if err := c.do(ctx, http.MethodPost, "/v2/tasks", body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetTask(ctx context.Context, taskID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/tasks/%s", pathPart(taskID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) UpdateTask(ctx context.Context, taskID string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/tasks/%s", pathPart(taskID))
	if err := c.do(ctx, http.MethodPatch, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) DeleteTask(ctx context.Context, taskID string) error {
	path := fmt.Sprintf("/v2/tasks/%s", pathPart(taskID))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
