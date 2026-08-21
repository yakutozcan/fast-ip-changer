# Security policy

Fast IP Changer changes the operating system's network configuration, which means it
builds and runs privileged shell commands from values typed into the UI. Vulnerability
reports are taken seriously.

## Supported versions

Only the latest release is supported. Fixes are shipped as a new release rather than as
patches to older tags.

## Reporting a vulnerability

Please **do not open a public issue** for a security problem.

Use GitHub's private vulnerability reporting instead: go to the
[Security tab](https://github.com/yakutozcan/fast-ip-changer/security/advisories/new) and
open a draft advisory. That keeps the report private until a fix exists, and it does not
require an email address on either side.

A useful report includes:

- the platform and OS version,
- what an attacker controls (a UI field, a profile file, a network response, …),
- the concrete effect (command executed, file written, privilege gained),
- and reproduction steps or a proof of concept.

Expect an acknowledgement within a week. The intent is to fix and release within 90 days
of the report, and to credit the reporter in the release notes unless they prefer not to
be named.

## Threat model

### In scope

- **Argument injection through the privileged path.** On macOS every privileged command is
  re-issued through `osascript`'s `do shell script`, so `shellQuote`, `appleScriptQuote`
  and `joinAnd` in [`pkg/sysexec/elevate.go`](pkg/sysexec/elevate.go) are the boundary that
  keeps a UI-typed value from reaching a shell as syntax. Anything that escapes that
  quoting is a vulnerability.
- **Privilege confusion or TOCTOU around `RunPrivilegedBatch`.** The batch collapses
  several commands into a single authorisation prompt; anything that lets an unintended
  command ride along on that one prompt is in scope.
- **Command construction in `pkg/network`.** Adapter names, addresses and DNS lists reach
  `netsh` / `networksetup` argument vectors. Validation lives in
  [`pkg/network/validate.go`](pkg/network/validate.go).
- **The profile store.** `~/.ip_changer_profiles.json` is written `0600` and atomically;
  anything that widens those permissions, follows a symlink out of the home directory or
  lets a crafted profile file influence command construction is in scope.
- **The optional public-IP lookup.** `https://api.ipify.org` is the app's only outbound
  request, it is off by default, and it must stay opt-in. A path that contacts it without
  the user enabling it is in scope, as is anything that leaks more than the request itself.

### Out of scope

- **The app requires administrator rights by design.** On Windows the manifest requests
  elevation at launch; on macOS each change raises the system authorisation dialog. That
  is the intended design, not a privilege-escalation bug.
- **SmartScreen and Gatekeeper warnings on release builds.** Releases are unsigned; see
  the Install section of the [README](README.md).
- **Anything that requires the attacker to already be local administrator / root.** With
  those rights the network configuration can be changed directly, without this app.
- **Linux.** The platform is not supported; network calls return an explicit
  unsupported-platform error.
