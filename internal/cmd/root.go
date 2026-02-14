package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/failup-ventures/attio-cli/internal/errfmt"
	"github.com/failup-ventures/attio-cli/internal/outfmt"
	"github.com/failup-ventures/attio-cli/internal/ui"
)

const (
	colorAuto  = "auto"
	colorNever = "never"
)

type CLI struct {
	RootFlags `embed:""`

	Version kong.VersionFlag `help:"Print version and exit"`

	Search RecordsSearchCmd `cmd:"" name:"search" help:"Alias for 'records search' (Beta)"`
	Query  RecordsQueryCmd  `cmd:"" name:"query" help:"Alias for 'records query'"`

	Auth       AuthCmd               `cmd:"" help:"Manage authentication"`
	Self       SelfCmd               `cmd:"" name:"self" aliases:"whoami,me" help:"Show current token info"`
	Objects    ObjectsCmd            `cmd:"" help:"Manage objects"`
	Records    RecordsCmd            `cmd:"" help:"Manage records"`
	Lists      ListsCmd              `cmd:"" help:"Manage lists"`
	Entries    EntriesCmd            `cmd:"" help:"Manage list entries"`
	Notes      NotesCmd              `cmd:"" help:"Manage notes"`
	Tasks      TasksCmd              `cmd:"" help:"Manage tasks"`
	Comments   CommentsCmd           `cmd:"" help:"Manage comments"`
	Threads    ThreadsCmd            `cmd:"" help:"Manage threads"`
	Webhooks   WebhooksCmd           `cmd:"" help:"Manage webhooks"`
	Meetings   MeetingsCmd           `cmd:"" help:"Manage meetings"`
	Attributes AttributesCmd         `cmd:"" aliases:"attrs" help:"Manage attributes"`
	Members    MembersCmd            `cmd:"" help:"Manage workspace members"`
	VersionCmd VersionCmd            `cmd:"" name:"version" help:"Print version"`
	Completion CompletionCmd         `cmd:"" help:"Generate shell completion scripts"`
	Complete   CompletionInternalCmd `cmd:"" name:"__complete" hidden:"" help:"Internal completion helper"`
	Schema     SchemaCmd             `cmd:"" help:"Print command schema"`
}

func Execute(args []string) error {
	parser, cli, err := newParser()
	if err != nil {
		return err
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		err = newUsageError(err)
		emitCLIError(err, jsonRequestedFromArgs(args), nil)
		return err
	}

	if shouldAutoJSON(cli.RootFlags) {
		cli.JSON = true
	}

	mode, err := outfmt.FromFlags(cli.JSON, cli.Plain)
	if err != nil {
		err = newUsageError(err)
		emitCLIError(err, cli.JSON, nil)
		return err
	}

	timeout, err := parseTimeout(cli.Timeout)
	if err != nil {
		err = newUsageError(err)
		emitCLIError(err, mode.JSON, nil)
		return err
	}
	setClientRuntimeOptions("attio-cli/"+Version, timeout)

	logLevel := slog.LevelWarn
	if cli.Verbose {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, mode)
	ctx = outfmt.WithJSONTransform(ctx, outfmt.JSONTransform{
		ResultsOnly: cli.ResultsOnly,
		Select:      splitCommaList(cli.Select),
	})
	ctx = withRuntimeOptions(ctx, RuntimeOptions{
		DryRun:    cli.DryRun,
		FailEmpty: cli.FailEmpty,
		IDOnly:    cli.IDOnly,
	})

	uiColor := cli.Color
	if outfmt.IsJSON(ctx) || outfmt.IsPlain(ctx) {
		uiColor = colorNever
	}

	u, err := ui.New(ui.Options{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Color:  uiColor,
	})
	if err != nil {
		return err
	}
	ctx = ui.WithUI(ctx, u)

	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.Bind(&cli.RootFlags)
	kctx.Bind(parser)

	if err := enforceCommandAllowlist(kctx.Command(), cli.EnableCommands); err != nil {
		emitCLIError(err, mode.JSON, u)
		return err
	}

	err = kctx.Run()
	if err == nil {
		return nil
	}

	err = stableExitCode(err)
	emitCLIError(err, mode.JSON, u)
	return err
}

func newParser() (*kong.Kong, *CLI, error) {
	cli := &CLI{}
	parser, err := kong.New(
		cli,
		kong.Name("attio"),
		kong.Description("Command-line interface for Attio API"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"color":   colorAuto,
			"version": buildVersionString(),
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return parser, cli, nil
}

func emitCLIError(err error, jsonOutput bool, u *ui.UI) {
	if err == nil {
		return
	}
	if jsonOutput {
		writeErrorJSON(os.Stderr, err)
		return
	}
	msg := strings.TrimSpace(errfmt.Format(err))
	if msg == "" {
		return
	}
	if u != nil {
		u.Err().Error(msg)
		return
	}
	_, _ = fmt.Fprintln(os.Stderr, msg)
}
