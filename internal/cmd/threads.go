package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type ThreadsCmd struct {
	List ThreadsListCmd `cmd:"" help:"List threads"`
	Get  ThreadsGetCmd  `cmd:"" help:"Get thread"`
}

type ThreadsListCmd struct {
	RecordID string `name:"record-id" help:"Record UUID" required:""`
	Object   string `name:"object" help:"Object slug or UUID" required:""`
	EntryID  string `name:"entry-id" help:"Entry UUID"`
	List     string `name:"list" help:"List slug or UUID"`
	Limit    int    `name:"limit" help:"Page size" default:"20"`
	Offset   int    `name:"offset" help:"Offset" default:"0"`
}

func (c *ThreadsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}

	threads, err := client.ListThreads(ctx, c.Object, c.RecordID, c.List, c.EntryID, c.Limit, c.Offset)
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		if err := maybePreviewResults(ctx, len(threads)); err != nil {
			return err
		}
		return writeOffsetPaginatedJSON(ctx, threads, c.Limit, c.Offset)
	}
	return writeThreads(ctx, threads)
}

type ThreadsGetCmd struct {
	ThreadID string `arg:"" name:"thread-id" help:"Thread UUID" required:""`
}

func (c *ThreadsGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	thread, err := client.GetThread(ctx, c.ThreadID)
	if err != nil {
		return err
	}
	return writeSingleThread(ctx, thread)
}

func writeSingleThread(ctx context.Context, thread map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, thread); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": thread})
	}
	return writeThreads(ctx, []map[string]any{thread})
}

func writeThreads(ctx context.Context, threads []map[string]any) error {
	if err := maybePreviewResults(ctx, len(threads)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": threads})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tIS_RESOLVED\tCREATED_AT")
	for _, thread := range threads {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			idString(thread["id"]),
			mapString(thread, "is_resolved"),
			mapString(thread, "created_at"),
		)
	}
	return nil
}
