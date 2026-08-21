## What this changes

<!-- One or two sentences. Link the issue if there is one. -->

## Why

<!-- The code shows what. Explain why, if it isn't obvious. -->

## Checklist

- [ ] `go vet ./...` and `go test ./... -race` pass
- [ ] `gofmt -l .` is empty
- [ ] `npm run typecheck` and `npm run build` pass in `frontend/`
- [ ] Bindings regenerated with `wails generate module` and committed (only if `app.go`
      or a struct crossing the Wails boundary changed)
- [ ] New OS-output parsing is exec-free and covered by a fixture-based test
- [ ] New user-facing strings are Turkish; code, comments and docs are English

## Platforms tested

<!-- Which of macOS / Windows you actually ran this on. Say "none" if it is a
     docs-only or test-only change. If your change touches pkg/network or
     pkg/sysexec and you could only test one platform, say so — the maintainer
     will run the other. -->
