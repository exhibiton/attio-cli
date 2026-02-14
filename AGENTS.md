# Repository Guidelines

## Project Structure

- `cmd/attio/`: CLI entrypoint.
- `internal/api/`: Attio HTTP client and endpoint methods.
- `internal/cmd/`: Kong command tree and command handlers.
- `internal/config/`: config file + keyring auth resolution.
- `internal/outfmt/`, `internal/errfmt/`, `internal/ui/`: output/error/terminal behavior.
- `openapi.json`: authoritative endpoint reference used for parity checks.

## Build, Test, and Dev Commands

- `make build`: build `bin/attio`.
- `make fmt`: run `gofmt` on source.
- `make test`: run unit tests.
- `make ci`: local quality gate (`fmt` + `test`).
- `go run ./cmd/attio --help`: inspect command tree.

## Coding Standards

- Keep dependencies minimal; prefer stdlib unless there is clear value.
- Add tests for any behavior change (API, command routing, output/error behavior).
- Keep stdout script-friendly; use `--json`/`--plain` for machine output.
- Keep user-facing errors actionable; prefer explicit remediation hints.

## Testing Guidance

- Unit tests live next to code as `*_test.go`.
- Prefer table-driven tests and `httptest.Server` for API/client/command tests.
- Use deterministic tests (inject clocks/sleep funcs where retry behavior exists).
- Integration tests should be opt-in and gated with build tags/env vars.
- For CLI handler tests, use `setupCLIEnv(t)` + `captureExecute(t, args)` from `internal/cmd/root_test.go`.
- Prefer non-retry HTTP statuses (for example `400`/`404`) in test fixtures to keep runtime fast.

## Golden Files

- Golden outputs are under `internal/cmd/testdata/`.
- Validate with normal test runs:
  - `go test ./internal/cmd -run TestGoldenObjectsListOutputs`
- Update goldens intentionally (review diffs carefully):
  - `go test ./internal/cmd -run TestGoldenObjectsListOutputs -args -update`
- Keep golden updates focused to formatting/output changes only.

## Coverage Workflow

- Full package coverage:
  - `go test ./... -cover`
- Function-level coverage report:
  - `go test ./... -coverprofile=coverage.out`
  - `go tool cover -func=coverage.out`
- CI gate and local parity:
  - `make ci`

## Security and Secrets

- Never commit API keys, credentials, or keyring exports.
- Use `ATTIO_API_KEY` for CI/headless runs.
- Use `attio auth login` for local keyring-backed auth.
