# Roadmap

This file tracks intentionally deferred work after `v0.1.0`.

## Completed In v0.1.0

- Core CLI command coverage for all non-SCIM API endpoints.
- Agent UX items from review:
  - structured JSON errors
  - pagination metadata for offset list/query JSON outputs
  - `ATTIO_AUTO_JSON`
  - `--enable-commands`
  - `--id-only`
  - `schema` export command
  - `--timeout`
- Coverage and quality targets:
  - `internal/api` >= 90%
  - `internal/cmd` >= 80%
  - `internal/config` >= 85%
- Golden output tests expanded to:
  - `objects list`
  - `records query`
  - `tasks list`
  - `meetings list`

## Post-v0.1.0 Priorities

## 1) Onboarding / First-Run UX

Status: `Implemented`

- `attio init` is available and supports:
  - interactive prompts for profile, API key, verification, and keyring save preference
  - key verification via `/v2/self`
  - keyring persistence choice in onboarding flow
  - explicit guardrails when `--no-input` is set or no TTY is present
  - guided next-step output

Possible follow-ups:
- richer interactive wizard prompts (profile/base URL confirmation)
- optional smoke-test sequence after init
- installation/path diagnostics (`attio doctor` or `attio init` precheck) for "command not found" and shell PATH issues
- onboarding docs that include shell-specific PATH export guidance and verification steps

## 2) OAuth2 Support

Status: `Planned`

- Implement OAuth2 Authorization Code with PKCE.
- Add token persistence/refresh behavior in config/keyring.
- Keep API key mode supported for headless automation.

Source:
- Deferred in `IMPLEMENTATION_PLAN.md` (API key only in Phase 1).

## 3) Release Distribution Hardening

Status: `In Progress`

- Finalize Goreleaser release pipeline on tags:
  - multi-platform binaries
  - checksums
  - release notes wiring
- Homebrew formula exists in-repo (`Formula/attio.rb`) for one-line install from release binaries.
- Add/maintain dedicated Homebrew tap repo so install is:
  - `brew install exhibiton/tap/attio`
- Automate tap updates from releases (GoReleaser `brews` config + token with write access to `exhibiton/homebrew-tap`).

Source:
- Deferred release tasks in `IMPLEMENTATION_PLAN.md` and review notes.

## 4) Type-Safe API Models (Optional)

Status: `Planned`

- Incrementally add typed request/response structs for high-traffic resources:
  - objects
  - records
  - lists
  - entries
- Keep `map[string]any` where schema is dynamic (attribute values).

Source:
- Phase D in `CODE_REVIEW.md`.

## 5) Batch Operations (Conditional)

Status: `Backlog`

- Add batch create/update/delete commands if/when Attio API provides stable batch endpoints.

Source:
- Future consideration in `REVIEW_v0.1.0.md`.

## 6) Integration Suite Expansion

Status: `In Progress`

- Current integration suite includes read/write coverage for core resources.
- Expand deeper, when feasible and safe for staging environments:
  - lists/entries write paths (optional fixture lifecycle)
  - meetings recordings/transcripts coverage under guarded env toggles
  - stronger cleanup and retry strategy for flaky external dependencies

## Working Principles For Roadmap Items

- Keep API-key auth path first-class for automation and agent workflows.
- Prefer additive, non-breaking CLI evolution.
- Gate risky behavior behind explicit flags/env vars first.
- Add tests and docs in the same change as feature delivery.
