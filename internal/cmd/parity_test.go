package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

type openAPISpec struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

func TestOpenAPICommandParity(t *testing.T) {
	path := findOpenAPIPath(t)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read openapi.json: %v", err)
	}

	var spec openAPISpec
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatalf("parse openapi.json: %v", err)
	}

	parser, _, err := newParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	commands := map[string]struct{}{}
	collectCommandPaths(parser.Model.Node, nil, commands)

	var unmapped []string
	var missing []string
	for p, methods := range spec.Paths {
		if strings.HasPrefix(p, "/scim/") {
			continue
		}
		for m := range methods {
			method := strings.ToUpper(strings.TrimSpace(m))
			cmd, ok := endpointToCommand(method, p)
			key := method + " " + p
			if !ok {
				unmapped = append(unmapped, key)
				continue
			}
			if _, exists := commands[cmd]; !exists {
				missing = append(missing, key+" -> "+cmd)
			}
		}
	}

	sort.Strings(unmapped)
	sort.Strings(missing)

	if len(unmapped) > 0 {
		t.Fatalf("unmapped OpenAPI endpoints:\n%s", strings.Join(unmapped, "\n"))
	}
	if len(missing) > 0 {
		t.Fatalf("mapped endpoints missing CLI commands:\n%s", strings.Join(missing, "\n"))
	}
}

func collectCommandPaths(node *kong.Node, prefix []string, out map[string]struct{}) {
	for _, child := range node.Children {
		if child.Hidden || strings.TrimSpace(child.Name) == "" {
			continue
		}
		next := append(prefix, child.Name)
		if len(child.Children) == 0 {
			out[strings.Join(next, " ")] = struct{}{}
		}
		collectCommandPaths(child, next, out)
	}
}

func findOpenAPIPath(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"openapi.json",
		filepath.Join("..", "..", "openapi.json"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	t.Fatalf("openapi.json not found")
	return ""
}

func endpointToCommand(method string, path string) (string, bool) {
	switch {
	case method == "GET" && path == "/v2/self":
		return "self", true

	case method == "GET" && path == "/v2/objects":
		return "objects list", true
	case method == "POST" && path == "/v2/objects":
		return "objects create", true
	case method == "GET" && path == "/v2/objects/{object}":
		return "objects get", true
	case method == "PATCH" && path == "/v2/objects/{object}":
		return "objects update", true

	case method == "POST" && path == "/v2/objects/{object}/records":
		return "records create", true
	case method == "PUT" && path == "/v2/objects/{object}/records":
		return "records assert", true
	case method == "POST" && path == "/v2/objects/{object}/records/query":
		return "records query", true
	case method == "POST" && path == "/v2/objects/records/search":
		return "records search", true
	case method == "GET" && path == "/v2/objects/{object}/records/{record_id}":
		return "records get", true
	case method == "PATCH" && path == "/v2/objects/{object}/records/{record_id}":
		return "records update", true
	case method == "PUT" && path == "/v2/objects/{object}/records/{record_id}":
		return "records replace", true
	case method == "DELETE" && path == "/v2/objects/{object}/records/{record_id}":
		return "records delete", true
	case method == "GET" && path == "/v2/objects/{object}/records/{record_id}/attributes/{attribute}/values":
		return "records values list", true
	case method == "GET" && path == "/v2/objects/{object}/records/{record_id}/entries":
		return "records entries list", true

	case method == "GET" && path == "/v2/lists":
		return "lists list", true
	case method == "POST" && path == "/v2/lists":
		return "lists create", true
	case method == "GET" && path == "/v2/lists/{list}":
		return "lists get", true
	case method == "PATCH" && path == "/v2/lists/{list}":
		return "lists update", true

	case method == "POST" && path == "/v2/lists/{list}/entries":
		return "entries create", true
	case method == "PUT" && path == "/v2/lists/{list}/entries":
		return "entries assert", true
	case method == "POST" && path == "/v2/lists/{list}/entries/query":
		return "entries query", true
	case method == "GET" && path == "/v2/lists/{list}/entries/{entry_id}":
		return "entries get", true
	case method == "PATCH" && path == "/v2/lists/{list}/entries/{entry_id}":
		return "entries update", true
	case method == "PUT" && path == "/v2/lists/{list}/entries/{entry_id}":
		return "entries replace", true
	case method == "DELETE" && path == "/v2/lists/{list}/entries/{entry_id}":
		return "entries delete", true
	case method == "GET" && path == "/v2/lists/{list}/entries/{entry_id}/attributes/{attribute}/values":
		return "entries values list", true

	case method == "GET" && path == "/v2/notes":
		return "notes list", true
	case method == "POST" && path == "/v2/notes":
		return "notes create", true
	case method == "GET" && path == "/v2/notes/{note_id}":
		return "notes get", true
	case method == "DELETE" && path == "/v2/notes/{note_id}":
		return "notes delete", true

	case method == "GET" && path == "/v2/tasks":
		return "tasks list", true
	case method == "POST" && path == "/v2/tasks":
		return "tasks create", true
	case method == "GET" && path == "/v2/tasks/{task_id}":
		return "tasks get", true
	case method == "PATCH" && path == "/v2/tasks/{task_id}":
		return "tasks update", true
	case method == "DELETE" && path == "/v2/tasks/{task_id}":
		return "tasks delete", true

	case method == "POST" && path == "/v2/comments":
		return "comments create", true
	case method == "GET" && path == "/v2/comments/{comment_id}":
		return "comments get", true
	case method == "DELETE" && path == "/v2/comments/{comment_id}":
		return "comments delete", true

	case method == "GET" && path == "/v2/threads":
		return "threads list", true
	case method == "GET" && path == "/v2/threads/{thread_id}":
		return "threads get", true

	case method == "GET" && path == "/v2/meetings":
		return "meetings list", true
	case method == "POST" && path == "/v2/meetings":
		return "meetings create", true
	case method == "GET" && path == "/v2/meetings/{meeting_id}":
		return "meetings get", true
	case method == "GET" && path == "/v2/meetings/{meeting_id}/call_recordings":
		return "meetings recordings list", true
	case method == "POST" && path == "/v2/meetings/{meeting_id}/call_recordings":
		return "meetings recordings create", true
	case method == "GET" && path == "/v2/meetings/{meeting_id}/call_recordings/{call_recording_id}":
		return "meetings recordings get", true
	case method == "DELETE" && path == "/v2/meetings/{meeting_id}/call_recordings/{call_recording_id}":
		return "meetings recordings delete", true
	case method == "GET" && path == "/v2/meetings/{meeting_id}/call_recordings/{call_recording_id}/transcript":
		return "meetings transcript", true

	case method == "GET" && path == "/v2/webhooks":
		return "webhooks list", true
	case method == "POST" && path == "/v2/webhooks":
		return "webhooks create", true
	case method == "GET" && path == "/v2/webhooks/{webhook_id}":
		return "webhooks get", true
	case method == "PATCH" && path == "/v2/webhooks/{webhook_id}":
		return "webhooks update", true
	case method == "DELETE" && path == "/v2/webhooks/{webhook_id}":
		return "webhooks delete", true

	case method == "GET" && path == "/v2/workspace_members":
		return "members list", true
	case method == "GET" && path == "/v2/workspace_members/{workspace_member_id}":
		return "members get", true

	case method == "GET" && path == "/v2/{target}/{identifier}/attributes":
		return "attributes list", true
	case method == "POST" && path == "/v2/{target}/{identifier}/attributes":
		return "attributes create", true
	case method == "GET" && path == "/v2/{target}/{identifier}/attributes/{attribute}":
		return "attributes get", true
	case method == "PATCH" && path == "/v2/{target}/{identifier}/attributes/{attribute}":
		return "attributes update", true
	case method == "GET" && path == "/v2/{target}/{identifier}/attributes/{attribute}/options":
		return "attributes options list", true
	case method == "POST" && path == "/v2/{target}/{identifier}/attributes/{attribute}/options":
		return "attributes options create", true
	case method == "PATCH" && path == "/v2/{target}/{identifier}/attributes/{attribute}/options/{option}":
		return "attributes options update", true
	case method == "GET" && path == "/v2/{target}/{identifier}/attributes/{attribute}/statuses":
		return "attributes statuses list", true
	case method == "POST" && path == "/v2/{target}/{identifier}/attributes/{attribute}/statuses":
		return "attributes statuses create", true
	case method == "PATCH" && path == "/v2/{target}/{identifier}/attributes/{attribute}/statuses/{status}":
		return "attributes statuses update", true
	}

	return "", false
}
