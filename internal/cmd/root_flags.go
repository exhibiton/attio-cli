package cmd

type RootFlags struct {
	Profile        string `help:"Named profile to use" default:"default"`
	JSON           bool   `help:"Output JSON to stdout" short:"j"`
	Plain          bool   `help:"Output stable TSV for piping" short:"p"`
	IDOnly         bool   `name:"id-only" help:"Output only the primary resource ID when available"`
	ResultsOnly    bool   `name:"results-only" help:"In JSON mode, emit only the data array"`
	Select         string `name:"select" help:"In JSON mode, project comma-separated fields"`
	DryRun         bool   `name:"dry-run" aliases:"noop,preview,dryrun" help:"Do not make changes; print intended action and exit successfully"`
	FailEmpty      bool   `name:"fail-empty" help:"Exit code 3 when a list/query command returns no results"`
	EnableCommands string `name:"enable-commands" env:"ATTIO_ENABLE_COMMANDS" help:"Comma-separated command allowlist (supports top-level or full command path)"`
	Timeout        string `name:"timeout" env:"ATTIO_TIMEOUT" default:"30s" help:"HTTP request timeout (for example: 30s, 2m)"`
	Color          string `help:"Color output: auto|always|never" default:"${color}"`
	Verbose        bool   `help:"Enable debug logging" short:"v"`
	NoInput        bool   `name:"no-input" help:"Never prompt; fail instead"`
}
