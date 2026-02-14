package api

// Envelope wraps most Attio API responses.
type Envelope[T any] struct {
	Data T `json:"data"`
}

// Self represents GET /v2/self response.
type Self struct {
	Active                        bool    `json:"active"`
	Scope                         string  `json:"scope,omitempty"`
	ClientID                      string  `json:"client_id,omitempty"`
	TokenType                     string  `json:"token_type,omitempty"`
	Exp                           *int64  `json:"exp,omitempty"`
	IAT                           *int64  `json:"iat,omitempty"`
	Sub                           string  `json:"sub,omitempty"`
	Aud                           string  `json:"aud,omitempty"`
	Iss                           string  `json:"iss,omitempty"`
	AuthorizedByWorkspaceMemberID string  `json:"authorized_by_workspace_member_id,omitempty"`
	WorkspaceID                   string  `json:"workspace_id,omitempty"`
	WorkspaceName                 string  `json:"workspace_name,omitempty"`
	WorkspaceSlug                 string  `json:"workspace_slug,omitempty"`
	WorkspaceLogoURL              *string `json:"workspace_logo_url,omitempty"`
}
