# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/yakutozcan/fast-ip-changer/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yakutozcan/fast-ip-changer/releases/tag/v1.0.0
