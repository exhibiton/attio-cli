package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/api"
	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type EntriesCmd struct {
	Create  EntriesCreateCmd  `cmd:"" help:"Create an entry"`
	Assert  EntriesAssertCmd  `cmd:"" aliases:"upsert" help:"Assert (upsert) a list entry by parent"`
	Query   EntriesQueryCmd   `cmd:"" help:"Query entries"`
	Get     EntriesGetCmd     `cmd:"" help:"Get entry"`
	Update  EntriesUpdateCmd  `cmd:"" help:"Update entry (PATCH append multiselect)"`
	Replace EntriesReplaceCmd `cmd:"" help:"Replace entry (PUT overwrite multiselect)"`
	Delete  EntriesDeleteCmd  `cmd:"" help:"Delete entry"`
	Values  EntriesValuesCmd  `cmd:"" help:"Entry attribute values"`
}

type EntriesCreateCmd struct {
	List string `arg:"" name:"list" help:"List slug or UUID" required:""`
	Data string `name:"data" help:"Entry data JSON; supports '-' or @file.json" required:""`
}

func (c *EntriesCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "entries create", map[string]any{"list": c.List, "data": data}); ok || err != nil {
		return err
	}
	entry, err := client.CreateEntry(ctx, c.List, data)
	if err != nil {
		return err
	}
	return writeSingleEntry(ctx, entry)
}

type EntriesAssertCmd struct {
	List string `arg:"" name:"list" help:"List slug or UUID" required:""`
	Data string `name:"data" help:"Entry data JSON; supports '-' or @file.json" required:""`
}

func (c *EntriesAssertCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "entries assert", map[string]any{"list": c.List, "data": data}); ok || err != nil {
		return err
	}
	entry, err := client.AssertEntry(ctx, c.List, data)
	if err != nil {
		return err
	}
	return writeSingleEntry(ctx, entry)
}

type EntriesQueryCmd struct {
	List     string `arg:"" name:"list" help:"List slug or UUID" required:""`
	Filter   string `name:"filter" help:"Filter JSON object"`
	Sorts    string `name:"sorts" help:"Sorts JSON array"`
	Limit    int    `name:"limit" help:"Page size" default:"500"`
	Offset   int    `name:"offset" help:"Offset for first page" default:"0"`
	All      bool   `name:"all" help:"Fetch all pages"`
	MaxPages int    `name:"max-pages" help:"Maximum pages to fetch when --all is set" default:"100"`
}

func (c *EntriesQueryCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	var filter any
	if c.Filter != "" {
		filter, err = readJSONValueInput(c.Filter)
		if err != nil {
			return err
		}
	}

	var sorts any
	if c.Sorts != "" {
		sorts, err = readJSONValueInput(c.Sorts)
		if err != nil {
			return err
		}
	}

	limit := c.Limit
	if limit <= 0 {
		limit = 500
	}

	var entries []map[string]any
	if c.All {
		start := c.Offset
		entries, err = api.FetchAllOffset(ctx, limit, c.MaxPages, func(offset int) ([]map[string]any, error) {
			return client.QueryEntries(ctx, c.List, filter, sorts, limit, start+offset)
		})
		if err != nil {
			return err
		}
	} else {
		entries, err = client.QueryEntries(ctx, c.List, filter, sorts, limit, c.Offset)
		if err != nil {
			return err
		}
	}

	if outfmt.IsJSON(ctx) {
		if err := maybePreviewResults(ctx, len(entries)); err != nil {
			return err
		}
		offset := c.Offset
		if c.All {
			// When --all is used, pagination has been fully consumed.
			return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
				"data": entries,
				"pagination": map[string]any{
					"limit":    limit,
					"offset":   offset,
					"has_more": false,
				},
			})
		}
		return writeOffsetPaginatedJSON(ctx, entries, limit, offset)
	}

	return writeEntryRows(ctx, entries)
}

type EntriesGetCmd struct {
	List    string `arg:"" name:"list" help:"List slug or UUID" required:""`
	EntryID string `arg:"" name:"entry-id" help:"Entry UUID" required:""`
}

func (c *EntriesGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	entry, err := client.GetEntry(ctx, c.List, c.EntryID)
	if err != nil {
		return err
	}
	return writeSingleEntry(ctx, entry)
}

type EntriesUpdateCmd struct {
	List    string `arg:"" name:"list" help:"List slug or UUID" required:""`
	EntryID string `arg:"" name:"entry-id" help:"Entry UUID" required:""`
	Data    string `name:"data" help:"Entry data JSON; supports '-' or @file.json" required:""`
}

func (c *EntriesUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "entries update", map[string]any{"list": c.List, "entry_id": c.EntryID, "data": data}); ok || err != nil {
		return err
	}
	entry, err := client.UpdateEntry(ctx, c.List, c.EntryID, data)
	if err != nil {
		return err
	}
	return writeSingleEntry(ctx, entry)
}

type EntriesReplaceCmd struct {
	List    string `arg:"" name:"list" help:"List slug or UUID" required:""`
	EntryID string `arg:"" name:"entry-id" help:"Entry UUID" required:""`
	Data    string `name:"data" help:"Entry data JSON; supports '-' or @file.json" required:""`
}

func (c *EntriesReplaceCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "entries replace", map[string]any{"list": c.List, "entry_id": c.EntryID, "data": data}); ok || err != nil {
		return err
	}
	entry, err := client.ReplaceEntry(ctx, c.List, c.EntryID, data)
	if err != nil {
		return err
	}
	return writeSingleEntry(ctx, entry)
}

type EntriesDeleteCmd struct {
	List    string `arg:"" name:"list" help:"List slug or UUID" required:""`
	EntryID string `arg:"" name:"entry-id" help:"Entry UUID" required:""`
}

func (c *EntriesDeleteCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "entries delete", map[string]any{"list": c.List, "entry_id": c.EntryID}); ok || err != nil {
		return err
	}
	if err := client.DeleteEntry(ctx, c.List, c.EntryID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"deleted":  true,
			"list":     c.List,
			"entry_id": c.EntryID,
		})
	}
	_, _ = os.Stdout.WriteString("Deleted entry " + c.EntryID + "\n")
	return nil
}

func writeSingleEntry(ctx context.Context, entry map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, entry); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": entry})
	}
	return writeEntryRows(ctx, []map[string]any{entry})
}

func writeEntryRows(ctx context.Context, entries []map[string]any) error {
	if err := maybePreviewResults(ctx, len(entries)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": entries})
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tCREATED_AT\tWEB_URL")
	for _, entry := range entries {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			idString(entry["id"]),
			mapString(entry, "created_at"),
			mapString(entry, "web_url"),
		)
	}
	return nil
}
