package cmd

import (
	"context"
	"fmt"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type RecordsEntriesCmd struct {
	List RecordsEntriesListCmd `cmd:"" help:"List entries for a record"`
}

type RecordsEntriesListCmd struct {
	Object   string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	RecordID string `arg:"" name:"record-id" help:"Record UUID" required:""`
	Limit    int    `name:"limit" help:"Page size" default:"0"`
	Offset   int    `name:"offset" help:"Offset" default:"0"`
}

func (c *RecordsEntriesListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	entries, err := client.ListRecordEntries(ctx, c.Object, c.RecordID, c.Limit, c.Offset)
	if err != nil {
		return err
	}
	if err := maybePreviewResults(ctx, len(entries)); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeOffsetPaginatedJSON(ctx, entries, c.Limit, c.Offset)
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ENTRY_ID\tCREATED_AT\tWEB_URL")
	for _, entry := range entries {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			idString(entry["id"]),
			mapString(entry, "created_at"),
			mapString(entry, "web_url"),
		)
	}
	return nil
}
