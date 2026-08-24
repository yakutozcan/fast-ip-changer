# Fast IP Changer

[![CI](https://github.com/yakutozcan/fast-ip-changer/actions/workflows/ci.yml/badge.svg)](https://github.com/yakutozcan/fast-ip-changer/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/yakutozcan/fast-ip-changer?sort=semver)](https://github.com/yakutozcan/fast-ip-changer/releases/latest)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/yakutozcan/fast-ip-changer)](go.mod)

A small desktop utility for switching a machine between static IP configurations and
DHCP, saving named IP profiles, and running basic network diagnostics (ping,
traceroute, connectivity check).

It is a [Wails v2](https://wails.io) application: a Go backend that drives the
operating system's own network tools, and a React 19 + TypeScript + Tailwind CSS v4
frontend. The window is a fixed 480x780 panel (resizing is disabled). **The user
interface is in Turkish**; the code, comments and this document are in English.

## Install

Download the latest build from the
[Releases page](https://github.com/yakutozcan/fast-ip-changer/releases/latest).

**The released binaries are not code-signed.** Signing certificates cost money on both
platforms and this is a free tool, so both operating systems will warn you the first time
you run it. The warning means "we have not seen this publisher before", not "this file is
known to be malicious" — but you should only accept it for a build you downloaded from
this repository's Releases page, and you can always build from source instead (see
[Development](#development)).

**Windows** — run the `.exe` installer.

1. SmartScreen shows "Windows protected your PC": choose **More info → Run anyway**.
2. Accept the UAC prompt. It will say the publisher is unknown, again because the build
   is unsigned. The app asks for elevation at launch because `netsh` cannot change the IP
   configuration without it.

**macOS** — unzip and drag `Fast IP Changer.app` into `/Applications`.

1. The first launch shows "cannot be opened because the developer cannot be verified".
   **Right-click the app → Open**, then confirm. Alternatively clear the quarantine flag:

   ```sh
   xattr -dr com.apple.quarantine "/Applications/Fast IP Changer.app"
   ```

2. The app itself runs unprivileged. Every network change raises the standard macOS
   authorisation dialog and asks for an administrator password — expect one prompt per
   change, not one per command.

Requires macOS 11 or newer, or Windows 10/11.

### Verifying a download

Every release attaches `SHA256SUMS.txt`, listing the SHA-256 of each file in that
release. Because the builds are unsigned, this is the only way to confirm a download
matches what CI produced. Download it next to the file you fetched and check it:

**macOS**

```sh
shasum -a 256 -c SHA256SUMS.txt
```

**Windows** (PowerShell)

```powershell
Get-FileHash .\fast-ip-changer-amd64-installer.exe -Algorithm SHA256
```

Then compare that hash against the matching line in `SHA256SUMS.txt`.

The checksum list is generated in the release workflow, on the same runner that
uploaded the artifacts. It proves the file arrived intact, not that the publisher is
who they claim to be — that is what signing would add, and this project does not sign
(see above). Only trust a `SHA256SUMS.txt` fetched from this repository's Releases
page over HTTPS.

`shasum -c` reports `FAILED` for lines whose file is not present, which is expected
when you downloaded only one of the release's files.

**There is no Linux build**, and one would not be useful: adapter enumeration and every
configuration call return an explicit "unsupported platform" error there. See
[Platform support](#platform-support).

## Platform support

| Platform | Network configuration | Notes |
| --- | --- | --- |
| macOS | `networksetup` | Adapters are the network *services* listed by `networksetup -listallnetworkservices`. |
| Windows | PowerShell / `netsh` | Adapters are read via PowerShell, falling back to `netsh` when it is unavailable; configuration changes go through `netsh`. Child-process console windows are hidden so nothing flashes in the GUI. |
| Linux | not supported | `GetAdapters` and the configuration calls return an "unsupported platform" error. |

## Administrator rights

Changing network configuration is a privileged operation on both supported
platforms. How that is obtained differs:

- **Windows** — the application manifest (`build/windows/wails.exe.manifest`) requests
  `requireAdministrator`, so launching the app triggers a UAC prompt. Windows cannot
  elevate a single child process without re-launching, so the whole app runs
  elevated. If the app is somehow started without elevation (stripped manifest, a
  non-admin account), the network-configuration commands fail with an explicit
  "administrator rights required" error instead of silently doing nothing.
- **macOS** — the app itself runs unprivileged. When a network change is applied, the
  privileged command is re-issued through `osascript`, which shows the standard macOS
  authorisation dialog asking for an administrator password. Arguments are
  shell-quoted before they reach that shell, so values typed into the UI cannot
  escape into it.

**Diagnostics do not need elevation.** Ping, traceroute and the quick connectivity
check run as the current user on both platforms.

Because the app builds privileged commands from values typed into the UI, the quoting
that keeps those values out of the shell is the security boundary that matters most here.
[SECURITY.md](SECURITY.md) documents the threat model and how to report a vulnerability
privately.

## Features

- **Adapter listing** — enumerate network adapters/services with their current IP,
  enabled state and configuration mode (DHCP or static), plus enable/disable of an
  adapter. The mode is left blank rather than guessed at when it cannot be determined.
- **Static IP** — apply an IP address, subnet mask, gateway and DNS server to the
  selected adapter.
- **DHCP** — switch the selected adapter back to obtaining its address automatically.
- **IP profiles** — named sets of IP / subnet / gateway / DNS that can be created,
  updated and deleted, and applied to an adapter in one step.
- **Ping and traceroute** — run against an arbitrary host, with the raw command output
  shown in the app and the ability to cancel a running command.
- **Quick connectivity check** — pings the gateway and an internet address and reports
  both reachability and latency.

## Language

The user interface and the user-facing error messages are **Turkish**. The code, the
comments, this document and the commit history are English.

The app is not internationalised, and that is a deliberate deferral rather than an
oversight: the Turkish strings are not confined to the React layer. Around twenty of them
are `fmt.Errorf` messages inside `pkg/network`, `pkg/profile`, `pkg/diagnostics` and
`pkg/sysexec` that cross the Wails boundary as opaque error strings. Translating the app
therefore means replacing those with typed errors or error codes and resolving the text in
the frontend — a change to the error contract of five packages, not a string-extraction
pass. Contributions are welcome; open an issue first.

## Profile storage

Profiles are stored as JSON in the user's home directory:

```
~/.ip_changer_profiles.json
```

The file is written with mode `0600` (owner read/write only). The app can also reveal
that file in the platform file manager (Finder on macOS, Explorer on Windows).

## Privacy

Everything the app does runs locally against the operating system's own tools, with
one exception: the **optional public-IP lookup** calls the third-party service
`api.ipify.org`. That lookup is **off by default** and has to be explicitly enabled by
the user. No other network requests are made by the app itself, and no telemetry is
collected.

## Project layout

| Path | Contents |
| --- | --- |
| `main.go` | Wails entry point: window options (fixed 480x780, resize disabled), embedded frontend assets, bindings registration. |
| `app.go` | The bindings layer — the `App` struct whose exported methods are what the frontend can call. It only delegates to the `pkg/*` packages. |
| `pkg/network` | Adapter enumeration (including the DHCP/static mode, read from SystemConfiguration's `preferences.plist` on macOS and `Get-NetIPInterface` on Windows), static IP / DHCP switching, adapter enable/disable, and the per-platform parsing of `networksetup` / `netsh` / PowerShell output. |
| `pkg/profile` | Load/save of `~/.ip_changer_profiles.json` and the profile CRUD operations. |
| `pkg/diagnostics` | Ping, traceroute and the quick connectivity check, including latency parsing. |
| `pkg/sysexec` | Shared subprocess helpers used by all of the above: a timeout on every call, hidden console windows on Windows, privileged execution (`RunPrivileged`) and elevation detection (`IsElevated`), and errors that carry the command's stderr instead of a bare exit code. |
| `frontend/` | React + TypeScript UI, its oxlint config (`.oxlintrc.json`) and Vitest suites (`src/**/*.test.ts`). `frontend/wailsjs/` holds the generated Go bindings. |
| `build/` | Wails build assets: the icon, the Windows manifest and NSIS installer template, the macOS `Info.plist` templates. Several files there deviate from the Wails defaults on purpose — see [`build/README.md`](build/README.md). |
| `.github/` | CI (`ci.yml`), the tag-triggered release workflow (`release.yml`), CodeQL and vulnerability scanning (`codeql.yml`, `security.yml`), Dependabot config, issue forms and the PR template. |

## Development

Prerequisites:

- Go 1.25 or newer
- Node.js 22 (the version CI uses; Vite needs `^20.19 || >=22.12`)
- The Wails CLI, pinned to the same version as the `wails/v2` library in `go.mod`:

  ```sh
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
  ```

  Not `@latest`: a CLI newer than the library can generate bindings the code does not
  expect, and CI pins this version too.

Live development:

```sh
wails dev
```

This starts a Vite dev server with hot reload for the frontend. A dev server that also
exposes the bound Go methods runs at <http://localhost:34115> — open that in a browser
and you can call the Go code from devtools.

Production build:

```sh
wails build
```

After changing the bound methods on `App` in `app.go`, regenerate the TypeScript
bindings in `frontend/wailsjs/`:

```sh
wails generate module
```

The project can be configured by editing `wails.json`. More information about the
project settings can be found here: https://wails.io/docs/reference/project-config

### The `frontend/dist` placeholder

`main.go` embeds the frontend build output with `//go:embed all:frontend/dist`, and a Go
embed pattern that matches no files is a compile error. The build output is not
committed, so `frontend/dist/.gitkeep` is tracked (via a negation in `.gitignore`) to
keep `go build`, `go vet` and `go test` working on a fresh clone before any frontend
build has run. Vite empties `dist` on every build, so the same file also lives in
`frontend/public/`, from where Vite copies it back. Leave both in place.

## Testing

Go linting is [golangci-lint](https://golangci-lint.run), configured in
[`.golangci.yml`](.golangci.yml):

```sh
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1 run ./...
```

The set is the default one (errcheck, govet, ineffassign, staticcheck, unused) plus seven
chosen for what this app does: `gosec` because the privileged commands are built from
values typed into the UI, `bodyclose` for the one HTTP client in the tree, `errorlint` and
`nilerr` because the error strings are part of the contract with the frontend,
`copyloopvar`, `misspell` (locale `UK`, matching the project's prose) and `revive` for doc
comments on the exported API that crosses into TypeScript. Formatting is *not* covered by
this command — `golangci-lint run` does not report the enabled formatter's findings, so
`gofmt -l .` is still its own step.

Every exclusion in the config names its reason and each `//nolint` in the source carries
an explanation. The one worth reading is the `gosec` G204 suppression in
[`pkg/sysexec/sysexec.go`](pkg/sysexec/sysexec.go): variable argv is the entire purpose of
that package, the command name is a literal at every call site, and the shell-quoting
helpers rather than the linter are the security boundary.

The version is pinned rather than `@latest`, like the Wails CLI, so a new linter release
cannot turn an unrelated pull request red. CI runs it on all three runners: each OS
compiles a different file in `pkg/sysexec`, and a linter only sees the variant its `GOOS`
selects.

Go packages — unit tests cover the platform output parsers, IP validation and profile
storage:

```sh
go test ./...
```

CI runs that with `-race` and a coverage profile, and writes a per-package coverage table
into the job summary. Coverage is **reported, not gated**: the three runners compile
different subsets of `pkg/sysexec`, so one threshold would either be trivial to meet or
unreachable on the platform that cannot compile the code it covers. The profile is
uploaded as an artifact per runner if you want to open it locally:

```sh
go test ./... -covermode=atomic -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Frontend — lint, type-check (`src` and `vite.config.ts`), unit-test and build:

```sh
cd frontend
npm ci
npm run lint
npm run typecheck
npm run test
npm run build
```

The unit tests run on [Vitest](https://vitest.dev) and live next to the code they cover,
as `src/**/*.test.ts`; `npm run test:watch` re-runs the affected suites as you edit. They
cover the pure logic in `src/lib` — IPv4 and subnet-mask validation, the DNS list parser,
the `localStorage` wrapper and platform detection — not the React components.

Suites run in the `node` environment by default; the ones that need browser globals opt
into a DOM per file with a `@vitest-environment happy-dom` docblock, so no suite pays for
a `document` it does not use. happy-dom rather than jsdom because it provides a real
`Storage`, which lets the storage tests exercise behaviour instead of a mock, and because
it has no peer dependencies to collide with. The test command is deliberately not part of
`npm run build`, which stays a type-check plus bundle.

Linting is [oxlint](https://oxc.rs), configured in
[`frontend/.oxlintrc.json`](frontend/.oxlintrc.json), with `npm run lint:fix` for the
auto-fixable subset. It stands in for ESLint because `typescript-eslint` still caps its
TypeScript peer range below the TypeScript 7 this project uses, and installing it anyway
with `--legacy-peer-deps` would put CI back in the dependency-resolution hole that
`npm ci` exists to prevent. oxlint has no TypeScript peer dependency at all — it parses
TypeScript and JSX itself.

Findings are warnings by default, so both scripts pass `--deny-warnings`; a warning is a
failed build here. `frontend/wailsjs/` and `frontend/dist/` are excluded — both are
generated.

`npm run typecheck` is also what catches stale Wails bindings: if `app.go` changed without
a `wails generate module`, the frontend stops type-checking against the generated
declarations.

The same commands run in CI on every push and pull request, with the Go job repeated
natively on Linux, macOS and Windows. Linux is there for coverage rather than for
shipping: it is the only job that compiles the `!windows && !darwin` files in
`pkg/sysexec` and exercises the unsupported-platform branch.

## Security scanning

Two workflows beyond CI run on every push to `main`, on every pull request, and once a
week on a schedule. [`codeql.yml`](.github/workflows/codeql.yml) builds a CodeQL database
for both languages — Go and the TypeScript frontend — and uploads findings to the
repository's Security tab. Its Go build is a native build plus an explicit cross-compile
of `./pkg/...` for darwin and windows, because a Linux-only compile never sees the
build-tagged halves of `pkg/sysexec` that hand a constructed command to an elevated shell.
[`security.yml`](.github/workflows/security.yml) checks the module against the Go
vulnerability database with `govulncheck`, and on pull requests also runs a dependency
review that fails on a new dependency carrying a high-severity advisory.

The `govulncheck` job tracks the latest stable Go rather than the version in `go.mod`.
`govulncheck` reports standard-library advisories against the toolchain it runs with, and
the `go.mod` directive is a floor, not the toolchain a release is built with — pinning it
there would fail the job every time Go ships a security patch, over findings no change to
this repository can fix. Locally:

```sh
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

A standard-library finding on your machine means your Go toolchain is behind, not that
something here needs changing.

## Contributing

Bug reports, parsing fixtures from locales I cannot test, and pull requests are all
welcome. [CONTRIBUTING.md](CONTRIBUTING.md) covers the setup, the bindings-regeneration
rule, why the OS-output parsers are kept exec-free, and the review expectations for
privileged code. Please read [SECURITY.md](SECURITY.md) before reporting anything
security-related, and note that participation is governed by the
[Code of Conduct](CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE). Third-party material — the Wails Go module and the vendored Wails
JavaScript runtime in `frontend/wailsjs/` — is covered by
[THIRD-PARTY-NOTICES.md](THIRD-PARTY-NOTICES.md). No third-party fonts or images are
bundled; the application icon is original work generated from
[`build/appicon.svg`](build/appicon.svg).
