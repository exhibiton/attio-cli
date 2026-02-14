package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type CommentsCmd struct {
	Create CommentsCreateCmd `cmd:"" help:"Create comment"`
	Get    CommentsGetCmd    `cmd:"" help:"Get comment"`
	Delete CommentsDeleteCmd `cmd:"" help:"Delete comment"`
}

type CommentsCreateCmd struct {
	Data string `name:"data" help:"Optional comment payload JSON; supports '-' or @file.json"`

	Author    string `name:"author" help:"Author workspace member UUID"`
	Content   string `name:"content" aliases:"body" help:"Comment content text"`
	Format    string `name:"format" help:"Body format (plaintext|markdown, defaults to plaintext)"`
	CreatedAt string `name:"created-at" help:"Optional created timestamp (ISO 8601)"`

	ThreadID     string `name:"thread" help:"Existing thread UUID to reply to"`
	RecordObject string `name:"record-object" help:"Record object slug/UUID for new top-level comment"`
	RecordID     string `name:"record-id" help:"Record UUID for new top-level comment"`
	EntryList    string `name:"entry-list" help:"List slug/UUID for new top-level comment"`
	EntryID      string `name:"entry-id" help:"Entry UUID for new top-level comment"`
}

func (c *CommentsCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := c.payload()
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "comments create", map[string]any{"data": data}); ok || err != nil {
		return err
	}
	comment, err := client.CreateComment(ctx, data)
	if err != nil {
		return err
	}
	return writeSingleComment(ctx, comment)
}

func (c *CommentsCreateCmd) payload() (map[string]any, error) {
	data := map[string]any{}
	if strings.TrimSpace(c.Data) != "" {
		parsed, err := readJSONObjectInput(c.Data)
		if err != nil {
			return nil, err
		}
		data = parsed
	}

	if c.Author != "" {
		data["author"] = map[string]any{
			"type": "workspace-member",
			"id":   c.Author,
		}
	}
	if c.Content != "" {
		data["content"] = c.Content
	}
	if c.Format != "" {
		data["format"] = c.Format
	} else if !hasMapKey(data, "format") {
		data["format"] = "plaintext"
	}
	if c.CreatedAt != "" {
		data["created_at"] = c.CreatedAt
	}

	flagTargets := 0
	if c.ThreadID != "" {
		flagTargets++
	}
	if c.RecordObject != "" || c.RecordID != "" {
		if c.RecordObject == "" || c.RecordID == "" {
			return nil, newUsageError(errors.New("--record-object and --record-id must be provided together"))
		}
		flagTargets++
	}
	if c.EntryList != "" || c.EntryID != "" {
		if c.EntryList == "" || c.EntryID == "" {
			return nil, newUsageError(errors.New("--entry-list and --entry-id must be provided together"))
		}
		flagTargets++
	}
	if flagTargets > 1 {
		return nil, newUsageError(errors.New("only one target mode can be set: --thread, --record-object/--record-id, or --entry-list/--entry-id"))
	}

	if c.ThreadID != "" {
		data["thread_id"] = c.ThreadID
		delete(data, "record")
		delete(data, "entry")
	}
	if c.RecordObject != "" && c.RecordID != "" {
		data["record"] = map[string]any{
			"object":    c.RecordObject,
			"record_id": c.RecordID,
		}
		delete(data, "thread_id")
		delete(data, "entry")
	}
	if c.EntryList != "" && c.EntryID != "" {
		data["entry"] = map[string]any{
			"list":     c.EntryList,
			"entry_id": c.EntryID,
		}
		delete(data, "thread_id")
		delete(data, "record")
	}

	required := []string{"author", "content", "format"}
	missing := make([]string, 0, len(required))
	for _, key := range required {
		if !hasMapKey(data, key) {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return nil, newUsageError(errors.New("comments create requires: " + strings.Join(missing, ", ")))
	}

	targets := 0
	if hasMapKey(data, "thread_id") {
		targets++
	}
	if hasMapKey(data, "record") {
		targets++
	}
	if hasMapKey(data, "entry") {
		targets++
	}
	if targets != 1 {
		return nil, newUsageError(errors.New("comments create requires one of: --thread, --record-object/--record-id, or --entry-list/--entry-id"))
	}

	return data, nil
}

type CommentsGetCmd struct {
	CommentID string `arg:"" name:"comment-id" help:"Comment UUID" required:""`
}

func (c *CommentsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	comment, err := client.GetComment(ctx, c.CommentID)
	if err != nil {
		return err
	}
	return writeSingleComment(ctx, comment)
}

type CommentsDeleteCmd struct {
	CommentID string `arg:"" name:"comment-id" help:"Comment UUID" required:""`
}

func (c *CommentsDeleteCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "comments delete", map[string]any{"comment_id": c.CommentID}); ok || err != nil {
		return err
	}
	if err := client.DeleteComment(ctx, c.CommentID); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"deleted": true, "comment_id": c.CommentID})
	}
	_, _ = os.Stdout.WriteString("Deleted comment " + c.CommentID + "\n")
	return nil
}

func writeSingleComment(ctx context.Context, comment map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, comment); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": comment})
	}
	return writeComments(ctx, []map[string]any{comment})
}

func writeComments(ctx context.Context, comments []map[string]any) error {
	if err := maybePreviewResults(ctx, len(comments)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": comments})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tTHREAD_ID\tCREATED_AT")
	for _, comment := range comments {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			idString(comment["id"]),
			mapString(mapMap(comment, "thread_id"), "thread_id"),
			mapString(comment, "created_at"),
		)
	}
	return nil
}
