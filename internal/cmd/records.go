package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/api"
	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type RecordsCmd struct {
	Create  RecordsCreateCmd  `cmd:"" help:"Create a record"`
	Assert  RecordsAssertCmd  `cmd:"" aliases:"upsert" help:"Assert (upsert) a record"`
	Query   RecordsQueryCmd   `cmd:"" help:"Query records"`
	Search  RecordsSearchCmd  `cmd:"" help:"Search records (Beta)"`
	Get     RecordsGetCmd     `cmd:"" help:"Get record"`
	Update  RecordsUpdateCmd  `cmd:"" help:"Update record (PATCH append multiselect)"`
	Replace RecordsReplaceCmd `cmd:"" help:"Replace record (PUT overwrite multiselect)"`
	Delete  RecordsDeleteCmd  `cmd:"" help:"Delete record"`
	Values  RecordsValuesCmd  `cmd:"" help:"Record attribute values"`
	Entries RecordsEntriesCmd `cmd:"" help:"List entries for a record"`
}

type RecordsCreateCmd struct {
	Object string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	Data   string `name:"data" help:"Record data JSON; supports '-' or @file.json" required:""`
}

func (c *RecordsCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "records create", map[string]any{"object": c.Object, "data": data}); ok || err != nil {
		return err
	}
	record, err := client.CreateRecord(ctx, c.Object, data)
	if err != nil {
		return err
	}
	return writeSingleRecord(ctx, record)
}

type RecordsAssertCmd struct {
	Object            string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	MatchingAttribute string `name:"matching-attribute" help:"Matching attribute slug or UUID" required:""`
	Data              string `name:"data" help:"Record data JSON; supports '-' or @file.json" required:""`
}

func (c *RecordsAssertCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "records assert", map[string]any{"object": c.Object, "matching_attribute": c.MatchingAttribute, "data": data}); ok || err != nil {
		return err
	}
	record, err := client.AssertRecord(ctx, c.Object, c.MatchingAttribute, data)
	if err != nil {
		return err
	}
	return writeSingleRecord(ctx, record)
}

type RecordsQueryCmd struct {
	Object   string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	Filter   string `name:"filter" help:"Filter JSON object"`
	Sorts    string `name:"sorts" help:"Sorts JSON array"`
	Limit    int    `name:"limit" help:"Page size" default:"500"`
	Offset   int    `name:"offset" help:"Offset for first page" default:"0"`
	All      bool   `name:"all" help:"Fetch all pages"`
	MaxPages int    `name:"max-pages" help:"Maximum pages to fetch when --all is set" default:"100"`
}

func (c *RecordsQueryCmd) Run(ctx context.Context, flags *RootFlags) error {
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

	var records []map[string]any
	if c.All {
		start := c.Offset
		records, err = api.FetchAllOffset(ctx, limit, c.MaxPages, func(offset int) ([]map[string]any, error) {
			return client.QueryRecords(ctx, c.Object, filter, sorts, limit, start+offset)
		})
		if err != nil {
			return err
		}
	} else {
		records, err = client.QueryRecords(ctx, c.Object, filter, sorts, limit, c.Offset)
		if err != nil {
			return err
		}
	}

	if outfmt.IsJSON(ctx) {
		if err := maybePreviewResults(ctx, len(records)); err != nil {
			return err
		}
		offset := c.Offset
		if c.All {
			return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
				"data": records,
				"pagination": map[string]any{
					"limit":    limit,
					"offset":   offset,
					"has_more": false,
				},
			})
		}
		return writeOffsetPaginatedJSON(ctx, records, limit, offset)
	}

	return writeRecordRows(ctx, records)
}

type RecordsSearchCmd struct {
	Query     string `arg:"" name:"query" help:"Search query" required:""`
	Limit     int    `name:"limit" help:"Max results (1-25)" default:"25"`
	Objects   string `name:"objects" help:"Comma-separated object slugs/UUIDs" required:""`
	RequestAs string `name:"request-as" help:"Request context JSON (defaults to {\"type\":\"workspace\"})"`
}

func (c *RecordsSearchCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	objects := splitCommaList(c.Objects)
	if len(objects) == 0 {
		return newUsageError(fmt.Errorf("--objects is required"))
	}

	var requestAs any
	if c.RequestAs != "" {
		requestAs, err = readJSONValueInput(c.RequestAs)
		if err != nil {
			return err
		}
	}

	records, err := client.SearchRecords(ctx, c.Query, c.Limit, objects, requestAs)
	if err != nil {
		return err
	}
	return writeRecordRows(ctx, records)
}

type RecordsGetCmd struct {
	Object   string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	RecordID string `arg:"" name:"record-id" help:"Record UUID" required:""`
}

func (c *RecordsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	record, err := client.GetRecord(ctx, c.Object, c.RecordID)
	if err != nil {
		return err
	}
	return writeSingleRecord(ctx, record)
}

type RecordsUpdateCmd struct {
	Object   string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	RecordID string `arg:"" name:"record-id" help:"Record UUID" required:""`
	Data     string `name:"data" help:"Record data JSON; supports '-' or @file.json" required:""`
}

func (c *RecordsUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "records update", map[string]any{"object": c.Object, "record_id": c.RecordID, "data": data}); ok || err != nil {
		return err
	}
	record, err := client.UpdateRecord(ctx, c.Object, c.RecordID, data)
	if err != nil {
		return err
	}
	return writeSingleRecord(ctx, record)
}

type RecordsReplaceCmd struct {
	Object   string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	RecordID string `arg:"" name:"record-id" help:"Record UUID" required:""`
	Data     string `name:"data" help:"Record data JSON; supports '-' or @file.json" required:""`
}

func (c *RecordsReplaceCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "records replace", map[string]any{"object": c.Object, "record_id": c.RecordID, "data": data}); ok || err != nil {
		return err
	}
	record, err := client.ReplaceRecord(ctx, c.Object, c.RecordID, data)
	if err != nil {
		return err
	}
	return writeSingleRecord(ctx, record)
}

type RecordsDeleteCmd struct {
	Object   string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	RecordID string `arg:"" name:"record-id" help:"Record UUID" required:""`
}

func (c *RecordsDeleteCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "records delete", map[string]any{"object": c.Object, "record_id": c.RecordID}); ok || err != nil {
		return err
	}
	if err := client.DeleteRecord(ctx, c.Object, c.RecordID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{
			"deleted":   true,
			"object":    c.Object,
			"record_id": c.RecordID,
		})
	}
	_, _ = os.Stdout.WriteString("Deleted record " + c.RecordID + "\n")
	return nil
}

func writeSingleRecord(ctx context.Context, record map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, record); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": record})
	}
	return writeRecordRows(ctx, []map[string]any{record})
}

func writeRecordRows(ctx context.Context, records []map[string]any) error {
	if err := maybePreviewResults(ctx, len(records)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": records})
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tNAME\tEMAIL\tCREATED_AT\tWEB_URL")
	for _, record := range records {
		name := recordValueSummary(record, "name", "full_name", "company_name")
		email := recordValueSummary(record, "email_addresses", "email_address")
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			idString(record["id"]),
			name,
			email,
			mapString(record, "created_at"),
			mapString(record, "web_url"),
		)
	}
	return nil
}
