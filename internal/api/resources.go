package api

import (
	"net/url"
	"strings"
)

type dataObjectResponse = Envelope[map[string]any]
type dataArrayResponse = Envelope[[]map[string]any]

func pathPart(value string) string {
	return url.PathEscape(strings.TrimSpace(value))
}

func withQuery(path string, query url.Values) string {
	if query == nil || len(query) == 0 {
		return path
	}
	encoded := query.Encode()
	if encoded == "" {
		return path
	}
	return path + "?" + encoded
}
