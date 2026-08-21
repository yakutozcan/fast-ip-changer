# Contributing

Thanks for considering a contribution. This is a small project, so the process is short —
but a few of the rules below are specific to what this app does, and they will otherwise be
learned the hard way in review.

## Getting set up

Prerequisites:

- Go 1.25 or newer
- Node.js 22 (the version CI uses; Vite needs `^20.19 || >=22.12`)
- The Wails CLI, **pinned to the version in `go.mod`**:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
```

Do not use `@latest`. A CLI newer than the `wails/v2` library can generate different
bindings than the code expects, and CI pins the same version.

Then:

```bash
wails dev
```

Before opening a pull request:

```bash
go vet ./... && go test ./... -race && gofmt -l .
cd frontend && npm ci && npm run typecheck && npm run build
```

CI runs exactly this, on Linux, macOS and Windows.

## Project rules

### 1. Bindings are generated, and the generated files are committed

`frontend/wailsjs/` holds the Go↔TypeScript bindings. If you change an exported method on
`App` in [`app.go`](app.go), or a struct that crosses the boundary, run:

```bash
wails generate module
```

and commit the resulting diff. The tree is currently an exact 1:1 match with the bound
methods; drift is a bug, not a formality. CI's `npm run typecheck` is the backstop.

### 2. Platform code needs platform testing

Changes to `pkg/network` or `pkg/sysexec` need to be exercised on **both** macOS and
Windows. A contributor on one platform structurally cannot validate the other — say which
platform you tested on in the PR, and ask; the maintainer will run the other one.

Linux is deliberately unsupported: `GetAdapters` and every configuration call return an
unsupported-platform error. Please do not send Linux network implementations without
opening an issue to discuss it first.

### 3. Parsers stay pure, and ship with fixtures

Operating-system command output is locale-dependent — Turkish `netsh` prints
`Devre Dışı` where English prints `Disabled`, and that difference has already broken this
app once. That is exactly why [`pkg/network/parse.go`](pkg/network/parse.go) contains no
`exec` calls at all: every function takes command output as a string and returns parsed
data.

New parsing follows that shape and comes with a table-driven unit test carrying **real
captured output** as a fixture, in the style of
[`pkg/network/network_test.go`](pkg/network/network_test.go). Do not add an integration
test that shells out to `netsh` or `networksetup` to cover parsing.

### 4. Privileged code gets extra scrutiny

Every command that changes the network configuration goes through
`sysexec.RunPrivileged` / `RunPrivilegedBatch`. On macOS these run through `osascript`,
which means a shell — so every argument must pass through `shellQuote`. A new privileged
call site should:

- build an `[]string` argv, never a pre-joined command string,
- validate its inputs before the command is constructed (see `pkg/network/validate.go`),
- and prefer adding to an existing batch over issuing a second privileged call, so the
  user is not asked to authorise twice.

See [SECURITY.md](SECURITY.md) for the threat model.

### 5. Language: Turkish UI, English everything else

- **User-facing strings are Turkish** — React copy and the `fmt.Errorf` messages in
  `pkg/**` that surface in the UI.
- **Code, comments, documentation, commit messages and PR descriptions are English.**

A PR that adds an English UI string will be asked to change it, and vice versa. If you
want to work on internationalising the app, that is welcome — but note it is not a
string-extraction job: the Turkish messages inside `pkg/**` cross the Wails boundary as
opaque error strings, so real i18n means replacing them with typed errors or error codes
and translating in the frontend. Open an issue first.

## Commits and pull requests

- One logical change per PR. Keep unrelated reformatting out of it.
- Present-tense, imperative commit subjects (`fix netsh parsing on Turkish consoles`).
- Explain *why* in the body when the change is not obvious. The code says what.
- Fill in the PR checklist — it exists so review can start with the interesting part.

## Reporting bugs

Use the [issue forms](https://github.com/yakutozcan/fast-ip-changer/issues/new/choose).
For a parsing bug, the single most useful thing you can attach is the **raw output** of the
command the app ran (`netsh interface show interface`, `networksetup -listnetworkserviceorder`,
`ping`), copied verbatim from your own console. Security issues go through
[SECURITY.md](SECURITY.md) instead — not a public issue.
