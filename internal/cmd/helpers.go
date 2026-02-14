package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/failup-ventures/attio-cli/internal/api"
	"github.com/failup-ventures/attio-cli/internal/config"
	"github.com/failup-ventures/attio-cli/internal/outfmt"
	"github.com/failup-ventures/attio-cli/internal/ui"
)

var (
	clientRuntimeMu sync.RWMutex
	clientUserAgent = "attio-cli/dev"
	clientTimeout   = 30 * time.Second
)

func setClientRuntimeOptions(userAgent string, timeout time.Duration) {
	clientRuntimeMu.Lock()
	defer clientRuntimeMu.Unlock()

	if strings.TrimSpace(userAgent) != "" {
		clientUserAgent = strings.TrimSpace(userAgent)
	}
	if timeout > 0 {
		clientTimeout = timeout
	}
}

func getClientRuntimeOptions() (string, time.Duration) {
	clientRuntimeMu.RLock()
	defer clientRuntimeMu.RUnlock()
	return clientUserAgent, clientTimeout
}

func requireClient(profile string) (*api.Client, error) {
	apiKey, err := config.ResolveAPIKey(profile)
	if err != nil {
		return nil, err
	}
	baseURL := config.ResolveBaseURL(profile)
	client := api.NewClient(apiKey, baseURL)
	userAgent, timeout := getClientRuntimeOptions()
	client.SetUserAgent(userAgent)
	client.SetTimeout(timeout)
	return client, nil
}

func tableWriter(ctx context.Context) (*tabwriter.Writer, func()) {
	var w io.Writer = os.Stdout
	if u := ui.FromContext(ctx); u != nil {
		w = u.OutWriter()
	}
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	flush := func() {
		_ = tw.Flush()
	}
	return tw, flush
}

func splitCommaList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return fmt.Sprintf("%s%s", strings.Repeat("*", len(value)-4), value[len(value)-4:])
}

func maybeWriteIDOnly(ctx context.Context, resource map[string]any) (bool, error) {
	if !isIDOnly(ctx) {
		return false, nil
	}
	id := idString(resource["id"])
	if id == "" {
		return false, newUsageError(errors.New("--id-only requires a response with an id field"))
	}
	_, err := os.Stdout.WriteString(id + "\n")
	return true, err
}

func writeOffsetPaginatedJSON(ctx context.Context, items []map[string]any, limit int, offset int) error {
	return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
		"data": items,
		"pagination": map[string]any{
			"limit":    limit,
			"offset":   offset,
			"has_more": limit > 0 && len(items) >= limit,
		},
	})
}

func readJSONObjectInput(input string) (map[string]any, error) {
	value, err := readJSONValueInput(input)
	if err != nil {
		return nil, err
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, newUsageError(errors.New("expected JSON object"))
	}
	return obj, nil
}

func readJSONValueInput(input string) (any, error) {
	raw, err := readRawInput(input)
	if err != nil {
		return nil, err
	}

	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, newUsageError(fmt.Errorf("invalid JSON: %w", err))
	}
	return out, nil
}

func readRawInput(input string) ([]byte, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, newUsageError(errors.New("missing --data value"))
	}
	switch {
	case input == "-":
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		return b, nil
	case strings.HasPrefix(input, "@"):
		path := strings.TrimSpace(strings.TrimPrefix(input, "@"))
		if path == "" {
			return nil, newUsageError(errors.New("missing file path after @"))
		}
		expanded, err := expandPath(path)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(expanded) //nolint:gosec // user-specified local file path
	default:
		return []byte(input), nil
	}
}

func hasMapKey(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}

func parseOptionalBoolFlag(value string, flagName string) (*bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	v, err := strconv.ParseBool(value)
	if err != nil {
		return nil, newUsageError(fmt.Errorf("%s must be true or false", flagName))
	}
	return &v, nil
}

func parseJSONArrayFlag(value string, flagName string) ([]any, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := readJSONValueInput(value)
	if err != nil {
		return nil, err
	}
	items, ok := parsed.([]any)
	if !ok {
		return nil, newUsageError(fmt.Errorf("%s must be a JSON array", flagName))
	}
	return items, nil
}

func parseTaskAssigneesFlag(value string) ([]any, error) {
	parts := splitCommaList(value)
	if len(parts) == 0 {
		return nil, nil
	}
	assignees := make([]any, 0, len(parts))
	for _, part := range parts {
		if strings.Contains(part, "@") {
			assignees = append(assignees, map[string]any{
				"workspace_member_email_address": part,
			})
			continue
		}
		assignees = append(assignees, map[string]any{
			"referenced_actor_type": "workspace-member",
			"referenced_actor_id":   part,
		})
	}
	return assignees, nil
}

func mapFromJSONField(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	switch v := m[key].(type) {
	case map[string]any:
		return v
	case []any:
		if len(v) == 0 {
			return nil
		}
		first, _ := v[0].(map[string]any)
		return first
	default:
		return nil
	}
}

func stringOrSummaryFromValue(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case map[string]any:
		for _, key := range []string{
			"full_name",
			"email_address",
			"workspace_member_email_address",
			"name",
			"title",
			"value",
			"domain",
			"original_phone_number",
			"target_record_id",
			"record_id",
		} {
			if s := anyString(x[key]); s != "" {
				return s
			}
		}
		return anyString(x)
	case []any:
		for _, item := range x {
			if s := stringOrSummaryFromValue(item); s != "" {
				return s
			}
		}
		return ""
	default:
		return anyString(x)
	}
}

func recordValueSummary(record map[string]any, keys ...string) string {
	values, _ := record["values"].(map[string]any)
	for _, key := range keys {
		if s := stringOrSummaryFromValue(values[key]); s != "" {
			return s
		}
	}
	return ""
}

func taskStatusSummary(task map[string]any) string {
	v, ok := task["is_completed"].(bool)
	if !ok {
		return anyString(task["is_completed"])
	}
	if v {
		return "completed"
	}
	return "open"
}

func taskAssigneeSummary(task map[string]any) string {
	assignees, _ := task["assignees"].([]any)
	out := make([]string, 0, len(assignees))
	for _, item := range assignees {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		s := anyString(m["workspace_member_email_address"])
		if s == "" {
			s = anyString(m["referenced_actor_id"])
		}
		if s == "" {
			actor := mapFromJSONField(m, "actor")
			if actor != nil {
				s = anyString(actor["id"])
			}
		}
		if s != "" {
			out = append(out, s)
		}
	}
	return strings.Join(out, ",")
}

func meetingParticipantsSummary(meeting map[string]any) string {
	if count, ok := intFromAnyValue(meeting["participant_count"]); ok && count >= 0 {
		return strconv.Itoa(count)
	}
	participants, _ := meeting["participants"].([]any)
	if participants == nil {
		return ""
	}
	return strconv.Itoa(len(participants))
}

func intFromAnyValue(v any) (int, bool) {
	switch x := v.(type) {
	case nil:
		return 0, false
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	case json.Number:
		i, err := x.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(x))
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func expandPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return path, nil
}

func anyString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	case json.Number:
		return x.String()
	case bool:
		return strconv.FormatBool(x)
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprintf("%v", x)
		}
		return string(b)
	}
}

func mapString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	return anyString(m[key])
}

func mapMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	v, _ := m[key].(map[string]any)
	return v
}

func idString(v any) string {
	if m, ok := v.(map[string]any); ok {
		for _, key := range []string{
			"record_id",
			"entry_id",
			"object_id",
			"list_id",
			"note_id",
			"task_id",
			"comment_id",
			"thread_id",
			"webhook_id",
			"workspace_member_id",
			"meeting_id",
			"call_recording_id",
			"attribute_id",
			"option_id",
			"status_id",
		} {
			if id := mapString(m, key); id != "" {
				return id
			}
		}
	}
	return anyString(v)
}
