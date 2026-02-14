package cmd

import (
	"context"
	"fmt"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type EntriesValuesCmd struct {
	List EntriesValuesListCmd `cmd:"" help:"List entry attribute values"`
}

type EntriesValuesListCmd struct {
	List         string `arg:"" name:"list" help:"List slug or UUID" required:""`
	EntryID      string `arg:"" name:"entry-id" help:"Entry UUID" required:""`
	Attribute    string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	ShowHistoric bool   `name:"show-historic" help:"Include historic values"`
	Limit        int    `name:"limit" help:"Page size" default:"0"`
	Offset       int    `name:"offset" help:"Offset" default:"0"`
}

func (c *EntriesValuesListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	values, err := client.ListEntryAttributeValues(ctx, c.List, c.EntryID, c.Attribute, c.ShowHistoric, c.Limit, c.Offset)
	if err != nil {
		return err
	}
	if err := maybePreviewResults(ctx, len(values)); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeOffsetPaginatedJSON(ctx, values, c.Limit, c.Offset)
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ACTIVE_FROM\tACTIVE_UNTIL\tVALUE")
	for _, value := range values {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			mapString(value, "active_from"),
			mapString(value, "active_until"),
			anyString(value),
		)
	}
	return nil
}
