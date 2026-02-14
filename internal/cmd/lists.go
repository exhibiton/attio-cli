package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type ListsCmd struct {
	List   ListsListCmd   `cmd:"" help:"List lists"`
	Create ListsCreateCmd `cmd:"" help:"Create list"`
	Get    ListsGetCmd    `cmd:"" help:"Get list"`
	Update ListsUpdateCmd `cmd:"" help:"Update list"`
}

type ListsListCmd struct{}

func (c *ListsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	lists, err := client.ListLists(ctx)
	if err != nil {
		return err
	}
	if err := maybePreviewResults(ctx, len(lists)); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": lists})
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tAPI_SLUG\tNAME\tPARENT_OBJECT")
	for _, list := range lists {
		parentObject := mapString(mapMap(list, "parent_object"), "api_slug")
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			idString(list["id"]),
			mapString(list, "api_slug"),
			mapString(list, "name"),
			parentObject,
		)
	}
	return nil
}

type ListsCreateCmd struct {
	Data string `name:"data" help:"List payload JSON; supports '-' or @file.json" required:""`
}

func (c *ListsCreateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "lists create", map[string]any{"data": data}); ok || err != nil {
		return err
	}

	list, err := client.CreateList(ctx, data)
	if err != nil {
		return err
	}
	return writeSingleList(ctx, list)
}

type ListsGetCmd struct {
	List string `arg:"" name:"list" help:"List slug or UUID" required:""`
}

func (c *ListsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	list, err := client.GetList(ctx, c.List)
	if err != nil {
		return err
	}
	return writeSingleList(ctx, list)
}

type ListsUpdateCmd struct {
	List string `arg:"" name:"list" help:"List slug or UUID" required:""`
	Data string `name:"data" help:"List payload JSON; supports '-' or @file.json" required:""`
}

func (c *ListsUpdateCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	data, err := readJSONObjectInput(c.Data)
	if err != nil {
		return err
	}
	if ok, err := maybeDryRun(ctx, "lists update", map[string]any{"list": c.List, "data": data}); ok || err != nil {
		return err
	}

	list, err := client.UpdateList(ctx, c.List, data)
	if err != nil {
		return err
	}
	return writeSingleList(ctx, list)
}

func writeSingleList(ctx context.Context, list map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, list); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": list})
	}

	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tAPI_SLUG\tNAME\tPARENT_OBJECT")
	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		idString(list["id"]),
		mapString(list, "api_slug"),
		mapString(list, "name"),
		mapString(mapMap(list, "parent_object"), "api_slug"),
	)
	return nil
}
