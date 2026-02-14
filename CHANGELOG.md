# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

### Added
- Global CLI flags: `--id-only`, `--enable-commands`, and `--timeout`.
- Agent-oriented behavior:
  - structured JSON error envelopes when `--json` is active
  - `ATTIO_AUTO_JSON=1` auto-switch for piped stdout
  - `schema` command for machine-readable command tree export
- Offset pagination metadata in JSON list/query outputs (`limit`, `offset`, `has_more`).
- Client runtime configuration hooks for user-agent and timeout.
- Expanded command/API/config test matrices and coverage-focused test suites.
- Golden output tests for `records query`, `tasks list`, and `meetings list` (table/JSON/plain).
- Additional integration read-path tests for lists/entries and meetings (opt-in `integration` tag).

### Changed
- `comments create` now uses `--content` as the primary flag (`--body` kept as alias).
- API client now sends `User-Agent: attio-cli/<version>` by default from CLI runtime.
- Pagination helpers now honor context cancellation between pages.
- Meetings cursor pagination response handling now guards nullable pagination blocks.

### Fixed
- Improved diagnostics for non-JSON API error bodies via debug logging in error parser.
- Removed unused response envelope duplication by reusing generic envelope aliases.

### Documentation
- README expanded with filtering, pagination, `--data @file`, and agent-focused examples.
- AGENTS guide expanded with test patterns, golden-file workflow, and coverage workflow.
