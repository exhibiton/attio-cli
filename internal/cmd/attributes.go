package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type AttributesCmd struct {
	List     AttributesListCmd     `cmd:"" help:"List attributes"`
	Create   AttributesCreateCmd   `cmd:"" help:"Create attribute"`
	Get      AttributesGetCmd      `cmd:"" help:"Get attribute"`
	Update   AttributesUpdateCmd   `cmd:"" help:"Update attribute"`
	Options  AttributesOptionsCmd  `cmd:"" help:"Manage select options"`
	Statuses AttributesStatusesCmd `cmd:"" help:"Manage statuses"`
}

type AttributesListCmd struct {
	Target       string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier   string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	ShowArchived bool   `name:"show-archived" help:"Include archived attributes"`
	Limit        int    `name:"limit" help:"Page size" default:"0"`
	Offset       int    `name:"offset" help:"Offset" default:"0"`
}

func (c *AttributesListCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	attrs, err := client.ListAttributes(ctx, c.Target, c.Identifier, c.ShowArchived, c.Limit, c.Offset)
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		if err := maybePreviewResults(ctx, len(attrs)); err != nil {
			return err
		}
		return writeOffsetPaginatedJSON(ctx, attrs, c.Limit, c.Offset)
	}
	return writeAttributes(ctx, attrs)
}

type AttributesCreateCmd struct {
	Target     string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Data       string `name:"data" help:"Attribute payload JSON; supports '-' or @file.json" required:""`
}

func (c *AttributesCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "attributes create", map[string]any{"target": c.Target, "identifier": c.Identifier, "data": data}); ok || err != nil {
		return err
	}
	attr, err := client.CreateAttribute(ctx, c.Target, c.Identifier, data)
	if err != nil {
		return err
	}
	return writeSingleAttribute(ctx, attr)
}

type AttributesGetCmd struct {
	Target     string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Attribute  string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
}

func (c *AttributesGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	attr, err := client.GetAttribute(ctx, c.Target, c.Identifier, c.Attribute)
	if err != nil {
		return err
	}
	return writeSingleAttribute(ctx, attr)
}

type AttributesUpdateCmd struct {
	Target     string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Attribute  string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	Data       string `name:"data" help:"Attribute payload JSON; supports '-' or @file.json" required:""`
}

func (c *AttributesUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "attributes update", map[string]any{"target": c.Target, "identifier": c.Identifier, "attribute": c.Attribute, "data": data}); ok || err != nil {
		return err
	}
	attr, err := client.UpdateAttribute(ctx, c.Target, c.Identifier, c.Attribute, data)
	if err != nil {
		return err
	}
	return writeSingleAttribute(ctx, attr)
}

type AttributesOptionsCmd struct {
	List   AttributesOptionsListCmd   `cmd:"" help:"List select options"`
	Create AttributesOptionsCreateCmd `cmd:"" help:"Create select option"`
	Update AttributesOptionsUpdateCmd `cmd:"" help:"Update select option"`
}

type AttributesOptionsListCmd struct {
	Target       string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier   string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Attribute    string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	ShowArchived bool   `name:"show-archived" help:"Include archived options"`
}

func (c *AttributesOptionsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	options, err := client.ListSelectOptions(ctx, c.Target, c.Identifier, c.Attribute, c.ShowArchived)
	if err != nil {
		return err
	}
	return writeSimpleOptions(ctx, options)
}

type AttributesOptionsCreateCmd struct {
	Target     string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Attribute  string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	Data       string `name:"data" help:"Option payload JSON; supports '-' or @file.json" required:""`
}

func (c *AttributesOptionsCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "attributes options create", map[string]any{"target": c.Target, "identifier": c.Identifier, "attribute": c.Attribute, "data": data}); ok || err != nil {
		return err
	}
	option, err := client.CreateSelectOption(ctx, c.Target, c.Identifier, c.Attribute, data)
	if err != nil {
		return err
	}
	return writeSingleSimpleOption(ctx, option)
}

type AttributesOptionsUpdateCmd struct {
	Target     string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Attribute  string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	Option     string `arg:"" name:"option" help:"Option slug or UUID" required:""`
	Data       string `name:"data" help:"Option payload JSON; supports '-' or @file.json" required:""`
}

func (c *AttributesOptionsUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "attributes options update", map[string]any{"target": c.Target, "identifier": c.Identifier, "attribute": c.Attribute, "option": c.Option, "data": data}); ok || err != nil {
		return err
	}
	option, err := client.UpdateSelectOption(ctx, c.Target, c.Identifier, c.Attribute, c.Option, data)
	if err != nil {
		return err
	}
	return writeSingleSimpleOption(ctx, option)
}

type AttributesStatusesCmd struct {
	List   AttributesStatusesListCmd   `cmd:"" help:"List statuses"`
	Create AttributesStatusesCreateCmd `cmd:"" help:"Create status"`
	Update AttributesStatusesUpdateCmd `cmd:"" help:"Update status"`
}

type AttributesStatusesListCmd struct {
	Target       string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier   string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Attribute    string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	ShowArchived bool   `name:"show-archived" help:"Include archived statuses"`
}

func (c *AttributesStatusesListCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	statuses, err := client.ListStatuses(ctx, c.Target, c.Identifier, c.Attribute, c.ShowArchived)
	if err != nil {
		return err
	}
	return writeSimpleOptions(ctx, statuses)
}

type AttributesStatusesCreateCmd struct {
	Target     string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Attribute  string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	Data       string `name:"data" help:"Status payload JSON; supports '-' or @file.json" required:""`
}

func (c *AttributesStatusesCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "attributes statuses create", map[string]any{"target": c.Target, "identifier": c.Identifier, "attribute": c.Attribute, "data": data}); ok || err != nil {
		return err
	}
	status, err := client.CreateStatus(ctx, c.Target, c.Identifier, c.Attribute, data)
	if err != nil {
		return err
	}
	return writeSingleSimpleOption(ctx, status)
}

type AttributesStatusesUpdateCmd struct {
	Target     string `arg:"" name:"target" help:"Target resource (objects|lists)" required:""`
	Identifier string `arg:"" name:"identifier" help:"Object/list slug or UUID" required:""`
	Attribute  string `arg:"" name:"attribute" help:"Attribute slug or UUID" required:""`
	Status     string `arg:"" name:"status" help:"Status slug or UUID" required:""`
	Data       string `name:"data" help:"Status payload JSON; supports '-' or @file.json" required:""`
}

func (c *AttributesStatusesUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	if err := validateAttributesTarget(c.Target); err != nil {
		return err
	}
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "attributes statuses update", map[string]any{"target": c.Target, "identifier": c.Identifier, "attribute": c.Attribute, "status": c.Status, "data": data}); ok || err != nil {
		return err
	}
	status, err := client.UpdateStatus(ctx, c.Target, c.Identifier, c.Attribute, c.Status, data)
	if err != nil {
		return err
	}
	return writeSingleSimpleOption(ctx, status)
}

func validateAttributesTarget(target string) error {
	if target == "objects" || target == "lists" {
		return nil
	}
	return newUsageError(fmt.Errorf("target must be 'objects' or 'lists'"))
}

func writeSingleAttribute(ctx context.Context, attr map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, attr); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": attr})
	}
	return writeAttributes(ctx, []map[string]any{attr})
}

func writeAttributes(ctx context.Context, attrs []map[string]any) error {
	if err := maybePreviewResults(ctx, len(attrs)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": attrs})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tTITLE\tTYPE\tIS_ARCHIVED")
	for _, attr := range attrs {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			idString(attr["id"]),
			mapString(attr, "title"),
			mapString(attr, "api_type"),
			mapString(attr, "is_archived"),
		)
	}
	return nil
}

func writeSingleSimpleOption(ctx context.Context, item map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, item); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": item})
	}
	return writeSimpleOptions(ctx, []map[string]any{item})
}

func writeSimpleOptions(ctx context.Context, items []map[string]any) error {
	if err := maybePreviewResults(ctx, len(items)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": items})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tTITLE\tIS_ARCHIVED")
	for _, item := range items {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			idString(item["id"]),
			mapString(item, "title"),
			mapString(item, "is_archived"),
		)
	}
	return nil
}
