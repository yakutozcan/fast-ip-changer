# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2026-08-24

Tooling and internal quality only. No change to what the app does, with one exception a
user can see: release downloads can now be verified against `SHA256SUMS.txt`.

### Added

- `SHA256SUMS.txt` is attached to every release, so unsigned downloads can be checked
  for integrity. README documents the verification commands.
- Frontend linting with oxlint (`npm run lint`), enforced in CI with `--deny-warnings`.
- Frontend unit tests on Vitest (`npm run test`) covering `src/lib` — IPv4 and mask
  validation, the DNS list parser, the storage wrapper and platform detection — run in CI.
- CI collects a Go coverage profile per runner, prints a per-package table in the job
  summary and uploads the profile as an artifact.
- Go linting with golangci-lint (`.golangci.yml`), pinned to v2.13.1 and run on all three
  CI runners so each platform's `pkg/sysexec` variant is linted.
- CodeQL analysis of the Go and TypeScript sources, `govulncheck` against the Go
  vulnerability database, and dependency review on pull requests.

### Changed

- Doc comments added to the exported API that crosses the Wails boundary, deferred
  `Close`/`Remove` results are now explicitly discarded, and the double close in the
  profile writer says which of the two owns the error.
- `useElevation` uses `async`/`await` like the other hooks instead of a `.then` chain.
- Adapter lists key React children on the adapter name rather than the array index.
- The diagnostics output region is a `role="log"` live region, which is also what makes
  its `tabIndex` correct: a scrollable box has to be focusable to be scrollable by
  keyboard.

## [1.0.0] - 2026-08-21

First public release.

### Added

- Adapter listing with current IPv4 address, administrative state and configuration
  mode (DHCP or static), plus enable/disable of an adapter.
- Static IP configuration — address, subnet mask, gateway and DNS servers — and a
  one-step switch back to DHCP, which also resets DNS to automatic.
- Named IP profiles stored in `~/.ip_changer_profiles.json` (mode `0600`, written
  atomically), with create, update, delete and apply-to-adapter.
- Diagnostics: ping and traceroute against an arbitrary host with cancellation, and a
  quick connectivity check against the gateway and the internet.
- Optional public-IP lookup via `api.ipify.org`, off by default and opt-in only.
- macOS support via `networksetup`, with privileged commands batched behind a single
  system authorisation prompt.
- Windows support via PowerShell with a `netsh` fallback, locale-independent output
  parsing, hidden child-process console windows, and elevation requested by the
  application manifest.

[Unreleased]: https://github.com/yakutozcan/fast-ip-changer/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/yakutozcan/fast-ip-changer/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/yakutozcan/fast-ip-changer/releases/tag/v1.0.0
