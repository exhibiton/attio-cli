package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type NotesCmd struct {
	List   NotesListCmd   `cmd:"" help:"List notes"`
	Create NotesCreateCmd `cmd:"" help:"Create note"`
	Get    NotesGetCmd    `cmd:"" help:"Get note"`
	Delete NotesDeleteCmd `cmd:"" help:"Delete note"`
}

type NotesListCmd struct {
	ParentObject   string `name:"parent-object" help:"Parent object slug or UUID" required:""`
	ParentRecordID string `name:"parent-record" help:"Parent record UUID" required:""`
	Limit          int    `name:"limit" help:"Page size" default:"10"`
	Offset         int    `name:"offset" help:"Offset" default:"0"`
}

func (c *NotesListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	notes, err := client.ListNotes(ctx, c.ParentObject, c.ParentRecordID, c.Limit, c.Offset)
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		if err := maybePreviewResults(ctx, len(notes)); err != nil {
			return err
		}
		return writeOffsetPaginatedJSON(ctx, notes, c.Limit, c.Offset)
	}
	return writeNotes(ctx, notes)
}

type NotesCreateCmd struct {
	Data           string `name:"data" help:"Optional note payload JSON; supports '-' or @file.json"`
	ParentObject   string `name:"parent-object" help:"Parent object slug or UUID"`
	ParentRecordID string `name:"parent-record" help:"Parent record UUID"`
	Title          string `name:"title" help:"Note title"`
	Content        string `name:"content" help:"Note content"`
	Format         string `name:"format" help:"Content format (plaintext|markdown, defaults to plaintext)"`
	CreatedAt      string `name:"created-at" help:"Optional created timestamp (ISO 8601)"`
	MeetingID      string `name:"meeting-id" help:"Optional meeting UUID or 'null'"`
}

func (c *NotesCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := c.payload()
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "notes create", map[string]any{"data": data}); ok || err != nil {
		return err
	}
	note, err := client.CreateNote(ctx, data)
	if err != nil {
		return err
	}
	return writeSingleNote(ctx, note)
}

func (c *NotesCreateCmd) payload() (map[string]any, error) {
	data := map[string]any{}
	if strings.TrimSpace(c.Data) != "" {
		parsed, err := readJSONObjectInput(c.Data)
		if err != nil {
			return nil, err
		}
		data = parsed
	}

	if c.ParentObject != "" {
		data["parent_object"] = c.ParentObject
	}
	if c.ParentRecordID != "" {
		data["parent_record_id"] = c.ParentRecordID
	}
	if c.Title != "" {
		data["title"] = c.Title
	}
	if c.Content != "" {
		data["content"] = c.Content
	}
	if c.Format != "" {
		if c.Format != "plaintext" && c.Format != "markdown" {
			return nil, newUsageError(errors.New("--format must be one of: plaintext, markdown"))
		}
		data["format"] = c.Format
	} else if !hasMapKey(data, "format") {
		data["format"] = "plaintext"
	}
	if c.CreatedAt != "" {
		data["created_at"] = c.CreatedAt
	}
	if c.MeetingID != "" {
		if strings.EqualFold(c.MeetingID, "null") {
			data["meeting_id"] = nil
		} else {
			data["meeting_id"] = c.MeetingID
		}
	}

	missing := make([]string, 0, 5)
	for _, required := range []string{"parent_object", "parent_record_id", "title", "format", "content"} {
		if !hasMapKey(data, required) {
			missing = append(missing, required)
		}
	}
	if len(missing) > 0 {
		return nil, newUsageError(errors.New("notes create requires: " + strings.Join(missing, ", ")))
	}
	return data, nil
}

type NotesGetCmd struct {
	NoteID string `arg:"" name:"note-id" help:"Note UUID" required:""`
}

func (c *NotesGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	note, err := client.GetNote(ctx, c.NoteID)
	if err != nil {
		return err
	}
	return writeSingleNote(ctx, note)
}

type NotesDeleteCmd struct {
	NoteID string `arg:"" name:"note-id" help:"Note UUID" required:""`
}

func (c *NotesDeleteCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "notes delete", map[string]any{"note_id": c.NoteID}); ok || err != nil {
		return err
	}
	if err := client.DeleteNote(ctx, c.NoteID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"deleted": true, "note_id": c.NoteID})
	}
	_, _ = os.Stdout.WriteString("Deleted note " + c.NoteID + "\n")
	return nil
}

func writeSingleNote(ctx context.Context, note map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, note); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": note})
	}
	return writeNotes(ctx, []map[string]any{note})
}

func writeNotes(ctx context.Context, notes []map[string]any) error {
	if err := maybePreviewResults(ctx, len(notes)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": notes})
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tTITLE\tPARENT_RECORD_ID\tCREATED_AT")
	for _, note := range notes {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			idString(note["id"]),
			mapString(note, "title"),
			mapString(mapMap(note, "parent_record"), "record_id"),
			mapString(note, "created_at"),
		)
	}
	return nil
}
