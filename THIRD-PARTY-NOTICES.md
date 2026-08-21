# Third-party notices

Fast IP Changer is distributed under the MIT License (see [LICENSE](LICENSE)). It
incorporates third-party material as described below.

## Wails

The application is built with [Wails v2](https://github.com/wailsapp/wails), which is
used in two distinct ways:

1. As a Go module dependency — `github.com/wailsapp/wails/v2`, declared in
   [`go.mod`](go.mod).
2. As **vendored JavaScript source** in [`frontend/wailsjs/`](frontend/wailsjs). That
   directory holds the generated Go bindings (`frontend/wailsjs/go/`) together with the
   Wails JavaScript runtime (`frontend/wailsjs/runtime/`), which is copied from the
   Wails distribution and is checked into this repository because the frontend imports
   it at build time.

The complete license text follows.

```
MIT License

Copyright (c) 2018-Present Lea Anthony

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Other dependencies

The remaining Go dependencies are declared in [`go.mod`](go.mod) and pinned in
[`go.sum`](go.sum); the frontend dependencies are declared in
[`frontend/package.json`](frontend/package.json) and pinned in
`frontend/package-lock.json`. Those files are the authoritative, machine-readable
manifest of what this project depends on, and each dependency ships its own license
in its own distribution.

No third-party fonts, icons or images are bundled with this application. The
application icon is original work, generated from
[`build/appicon.svg`](build/appicon.svg).
