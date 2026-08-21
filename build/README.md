# Build assets

Wails build inputs. Everything here is generated from the Wails template originally,
but several files have been changed on purpose — **read this before "fixing" them back
to their defaults**. Deleting a file in this directory makes `wails build` regenerate
the stock template version, which would undo the changes listed below.

| Path | Notes |
| --- | --- |
| `bin/` | Build output. Not committed (`.gitignore`). |
| `appicon.svg` | The icon source, original work for this project. Not read by Wails. |
| `appicon.png` | 1024x1024 render of `appicon.svg`; this is what Wails reads. Regenerate with `rsvg-convert -w 1024 -h 1024 build/appicon.svg -o build/appicon.png`. |
| `darwin/Info.plist` | Used by `wails build`. |
| `darwin/Info.dev.plist` | Same, used by `wails dev`. |
| `windows/wails.exe.manifest` | Application manifest. |
| `windows/info.json` | Version details shown in the exe's Properties → Details tab. |
| `windows/installer/` | NSIS installer template, used by `wails build -nsis`. |

## Intentional deviations from the Wails template

- **`requireAdministrator`** in `windows/wails.exe.manifest`. `netsh` cannot change the
  IP configuration unelevated, and Windows cannot elevate a single child process without
  re-launching, so the whole app requests elevation at launch. See the "Administrator
  rights" section of the top-level [README](../README.md).
- **The bundle identifier is `io.github.yakutozcan.fast-ip-changer`**, not the template's
  `com.wails.*`, in both darwin plists and the Windows manifest. The template value would
  scope this app's preferences and macOS privacy (TCC) entries into the Wails project's
  namespace.
- **`LSMinimumSystemVersion` is `11.0`**, not the template's `10.13.0`. Go 1.25 does not
  produce binaries that run on macOS 10.13, so the template value promises a platform the
  app would only crash on.
- **`NSAppleEventsUsageDescription` was added.** On macOS the privileged commands are
  re-issued through `osascript` (see `pkg/sysexec/elevate_darwin.go`), which is an Apple
  Event; without this key macOS shows an unexplained prompt, or denies the call outright
  under a hardened runtime.
- **`windows/icon.ico` is not committed.** Wails derives it from `appicon.png` at build
  time, so keeping a checked-in copy just means two icons to keep in sync.
- **The `signtool` hooks in `windows/installer/project.nsi` stay commented out.** Releases
  are deliberately unsigned; see the Install section of the top-level README.
