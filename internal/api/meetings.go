package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type cursorArrayResponse struct {
	Data       []map[string]any `json:"data"`
	Pagination *struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"pagination"`
}

func (c *Client) ListMeetings(ctx context.Context, limit int, cursor string, sort string, participants string, linkedObject string, linkedRecordID string, endsFrom string, startsBefore string, timezone string) ([]map[string]any, string, error) {
	var resp cursorArrayResponse

	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	if sort != "" {
		query.Set("sort", sort)
	}
	if participants != "" {
		query.Set("participants", participants)
	}
	if linkedObject != "" {
		query.Set("linked_object", linkedObject)
	}
	if linkedRecordID != "" {
		query.Set("linked_record_id", linkedRecordID)
	}
	if endsFrom != "" {
		query.Set("ends_from", endsFrom)
	}
	if startsBefore != "" {
		query.Set("starts_before", startsBefore)
	}
	if timezone != "" {
		query.Set("timezone", timezone)
	}

	path := withQuery("/v2/meetings", query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, "", err
	}

	next := ""
	if resp.Pagination != nil && resp.Pagination.NextCursor != nil {
		next = *resp.Pagination.NextCursor
	}
	return resp.Data, next, nil
}

func (c *Client) FindOrCreateMeeting(ctx context.Context, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	if err := c.do(ctx, http.MethodPost, "/v2/meetings", body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetMeeting(ctx context.Context, meetingID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/meetings/%s", pathPart(meetingID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ListCallRecordings(ctx context.Context, meetingID string, limit int, cursor string) ([]map[string]any, string, error) {
	var resp cursorArrayResponse

	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		query.Set("cursor", cursor)
	}

	path := fmt.Sprintf("/v2/meetings/%s/call_recordings", pathPart(meetingID))
	path = withQuery(path, query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, "", err
	}

	next := ""
	if resp.Pagination != nil && resp.Pagination.NextCursor != nil {
		next = *resp.Pagination.NextCursor
	}
	return resp.Data, next, nil
}

func (c *Client) CreateCallRecording(ctx context.Context, meetingID string, data map[string]any) (map[string]any, error) {
	var resp dataObjectResponse
	body := map[string]any{"data": data}
	path := fmt.Sprintf("/v2/meetings/%s/call_recordings", pathPart(meetingID))
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetCallRecording(ctx context.Context, meetingID string, callRecordingID string) (map[string]any, error) {
	var resp dataObjectResponse
	path := fmt.Sprintf("/v2/meetings/%s/call_recordings/%s", pathPart(meetingID), pathPart(callRecordingID))
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) DeleteCallRecording(ctx context.Context, meetingID string, callRecordingID string) error {
	path := fmt.Sprintf("/v2/meetings/%s/call_recordings/%s", pathPart(meetingID), pathPart(callRecordingID))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) GetTranscript(ctx context.Context, meetingID string, callRecordingID string, cursor string) ([]map[string]any, string, error) {
	var resp cursorArrayResponse

	query := url.Values{}
	if cursor != "" {
		query.Set("cursor", cursor)
	}

	path := fmt.Sprintf("/v2/meetings/%s/call_recordings/%s/transcript", pathPart(meetingID), pathPart(callRecordingID))
	path = withQuery(path, query)
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, "", err
	}

	next := ""
	if resp.Pagination != nil && resp.Pagination.NextCursor != nil {
		next = *resp.Pagination.NextCursor
	}
	return resp.Data, next, nil
}
