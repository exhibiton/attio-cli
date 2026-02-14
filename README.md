# attio-cli

`attio` is a Go CLI for the Attio API.

## Requirements

- Go 1.23+
- Attio API key (workspace-scoped)

## Install / Build

```bash
make build
./bin/attio --help
```

or run without building:

```bash
go run ./cmd/attio --help
```

## Authentication

Priority order:

1. `ATTIO_API_KEY`
2. Keyring value for selected profile (`attio auth login`)
3. Config file profile value (`~/.config/attio-cli/config.json`)

First-time setup (recommended):

```bash
attio init
```

This starts an onboarding flow that asks for profile, API key, verification preference, and keyring save preference.

Store a key in keyring:

```bash
attio auth login --api-key <YOUR_KEY>
```

Check auth resolution:

```bash
attio auth status --json
```

Use profile-specific credentials:

```bash
attio auth login --profile staging --api-key <STAGING_KEY>
attio --profile staging auth status --json
```

## Global Flags

- `--json`: JSON output
- `--plain`: stable TSV-like output
- `--results-only`: unwrap JSON envelope
- `--select`: project JSON fields
- `--dry-run`: print intended write operation and skip API mutation
- `--fail-empty`: exit code `3` when list/query/search returns no results
- `--id-only`: print only the resource ID for create/get/update-style responses
- `--timeout`: request timeout (for example `15s`, `1m`)
- `--enable-commands`: comma-separated allowlist for command sandboxing
- `--profile`: profile name (default: `default`)

Environment options:
- `ATTIO_AUTO_JSON=1`: auto-switch to JSON when stdout is piped
- `ATTIO_TIMEOUT=30s`: default timeout override
- `ATTIO_ENABLE_COMMANDS=records,objects`: default command allowlist

## Desire-Path Aliases

- `attio search ...` -> `attio records search ...`
- `attio query ...` -> `attio records query ...`

## Query / Filter Examples

Query records with filter/sort JSON:

```bash
attio --json records query people \
  --filter '{"name":{"$contains":"Ada"}}' \
  --sorts '[{"attribute":"created_at","direction":"desc"}]' \
  --limit 10 --offset 0
```

Search across specific objects:

```bash
attio records search "ada" --objects people,companies --limit 10
```

Use JSON projection for script-friendly output:

```bash
attio --json --select data.0.id.record_id records get people rec_123
```

## Pagination Guide

Offset endpoints (`records query`, `entries query`, `tasks list`, etc.):

```bash
# page 1
attio --json records query people --limit 50 --offset 0

# page 2
attio --json records query people --limit 50 --offset 50
```

Auto-fetch all pages where supported:

```bash
attio --json records query people --all --limit 100 --max-pages 20
attio --json meetings list --all --limit 50 --max-pages 20
```

## Data Input Patterns

Inline JSON:

```bash
attio records create people --data '{"values":{"name":[{"full_name":"Ada"}]}}'
```

JSON from file:

```bash
attio records create people --data @payloads/new_record.json
```

JSON from stdin:

```bash
cat payloads/new_record.json | attio records create people --data -
```

## Shell Completion

Generate completion scripts:

```bash
attio completion bash
attio completion zsh
attio completion fish
attio completion powershell
```

## Key Commands

- `attio init`
- `attio self`
- `attio objects ...`
- `attio records ...`
- `attio lists ...`
- `attio entries ...`
- `attio notes ...`
- `attio tasks ...`
- `attio comments ...`
- `attio threads ...`
- `attio meetings ...`
- `attio webhooks ...`
- `attio attributes ...`
- `attio members ...`

## Fixed-Schema Create/Update Flags

These command groups support named flags (with optional `--data` JSON overlays):

```bash
attio notes create --parent-object people --parent-record <record-id> --title "Intro" --content "Summary"
attio tasks create --content "Follow up with prospect" --assignees alice@company.com
attio comments create --author <member-id> --content "Looks good" --record-object people --record-id <record-id>
attio webhooks create --target-url "https://example.com/webhook" --subscriptions '[{"event_type":"task.created","filter":null}]'
```

`comments create --body` is kept as a compatibility alias for `--content`.

## Agent-Oriented Helpers

Command schema export:

```bash
attio --json schema
```

ID-only output:

```bash
attio --id-only records create people --data @payloads/new_record.json
```

Command sandboxing:

```bash
attio --enable-commands records,objects self   # blocked (not allowlisted)
attio --enable-commands records,objects records query people --limit 5
```

## Profile Management

Use multiple profiles for different workspaces/environments:

```bash
attio --profile default self
attio --profile staging self
```

Profile selection order:
- explicit `--profile`
- `default_profile` from config
- fallback `default`

## Config File

See `docs/config.md`.

## Integration Tests

Run live integration tests (opt-in):

```bash
ATTIO_API_KEY=<key> go test -tags=integration ./internal/integration -v
```

Skip expensive resources when needed:

```bash
ATTIO_IT_SKIP=meetings,webhooks go test -tags=integration ./internal/integration -v
```

## Coverage

```bash
make test-cover          # writes coverage.out
make cover-report        # prints per-function + total coverage
make cover-check         # enforces minimum threshold (default 70.0%)
```
