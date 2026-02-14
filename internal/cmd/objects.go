package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type ObjectsCmd struct {
	List   ObjectsListCmd   `cmd:"" help:"List objects"`
	Create ObjectsCreateCmd `cmd:"" help:"Create object"`
	Get    ObjectsGetCmd    `cmd:"" help:"Get object"`
	Update ObjectsUpdateCmd `cmd:"" help:"Update object"`
}

type ObjectsListCmd struct{}

func (c *ObjectsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	objects, err := client.ListObjects(ctx)
	if err != nil {
		return err
	}
	if err := maybePreviewResults(ctx, len(objects)); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": objects})
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tAPI_SLUG\tSINGULAR\tPLURAL")
	for _, object := range objects {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			idString(object["id"]),
			mapString(object, "api_slug"),
			mapString(object, "singular_noun"),
			mapString(object, "plural_noun"),
		)
	}
	return nil
}

type ObjectsCreateCmd struct {
	Data string `name:"data" help:"Object payload JSON; supports '-' or @file.json" required:""`
}

func (c *ObjectsCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "objects create", map[string]any{"data": data}); ok || err != nil {
		return err
	}

	object, err := client.CreateObject(ctx, data)
	if err != nil {
		return err
	}
	return writeSingleObject(ctx, object)
}

type ObjectsGetCmd struct {
	Object string `arg:"" name:"object" help:"Object slug or UUID" required:""`
}

func (c *ObjectsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	object, err := client.GetObject(ctx, c.Object)
	if err != nil {
		return err
	}
	return writeSingleObject(ctx, object)
}

type ObjectsUpdateCmd struct {
	Object string `arg:"" name:"object" help:"Object slug or UUID" required:""`
	Data   string `name:"data" help:"Object payload JSON; supports '-' or @file.json" required:""`
}

func (c *ObjectsUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "objects update", map[string]any{"object": c.Object, "data": data}); ok || err != nil {
		return err
	}

	object, err := client.UpdateObject(ctx, c.Object, data)
	if err != nil {
		return err
	}
	return writeSingleObject(ctx, object)
}

func writeSingleObject(ctx context.Context, object map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, object); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": object})
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tAPI_SLUG\tSINGULAR\tPLURAL")
	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		idString(object["id"]),
		mapString(object, "api_slug"),
		mapString(object, "singular_noun"),
		mapString(object, "plural_noun"),
	)
	return nil
}
