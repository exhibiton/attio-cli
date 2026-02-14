package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/failup-ventures/attio-cli/internal/outfmt"
)

type MembersCmd struct {
	List MembersListCmd `cmd:"" help:"List workspace members"`
	Get  MembersGetCmd  `cmd:"" help:"Get workspace member"`
}

type MembersListCmd struct{}

func (c *MembersListCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	members, err := client.ListMembers(ctx)
	if err != nil {
		return err
	}
	return writeMembers(ctx, members)
}

type MembersGetCmd struct {
	MemberID string `arg:"" name:"member-id" help:"Workspace member UUID" required:""`
}

func (c *MembersGetCmd) Run(ctx context.Context, flags *RootFlags) error {
	client, err := requireClient(flags.Profile)
	if err != nil {
		return err
	}
	member, err := client.GetMember(ctx, c.MemberID)
	if err != nil {
		return err
	}
	return writeSingleMember(ctx, member)
}

func writeSingleMember(ctx context.Context, member map[string]any) error {
	if ok, err := maybeWriteIDOnly(ctx, member); ok || err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": member})
	}
	return writeMembers(ctx, []map[string]any{member})
}

func writeMembers(ctx context.Context, members []map[string]any) error {
	if err := maybePreviewResults(ctx, len(members)); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, os.Stdout, map[string]any{"data": members})
	}
	w, done := tableWriter(ctx)
	defer done()
	_, _ = fmt.Fprintln(w, "ID\tNAME\tEMAIL\tROLE")
	for _, member := range members {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			idString(member["id"]),
			mapString(member, "name"),
			mapString(member, "email_address"),
			mapString(member, "role"),
		)
	}
	return nil
}
