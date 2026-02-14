package cmd

import (
	"context"
	"fmt"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type RecordsValuesCmd struct {
	List RecordsValuesListCmd `cmd:"" help:"List record attribute values"`
}

type RecordsValuesListCmd struct {
	Object       string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	RecordID     string `arg:"" name:"record-id" help:"Record UUID" required:""`
	Attribute    string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	ShowHistoric bool   `name:"show-historic" help:"Include historic values"`
	Limit        int    `name:"limit" help:"Page size" default:"0"`
	Offset       int    `name:"offset" help:"Offset" default:"0"`
}

func (c *RecordsValuesListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	values, err := client.ListRecordAttributeValues(ctx, c.Object, c.RecordID, c.Attribute, c.ShowHistoric, c.Limit, c.Offset)
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
