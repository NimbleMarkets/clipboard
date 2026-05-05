[![Build Status](https://travis-ci.com/atotto/clipboard.svg?branch=master)](https://travis-ci.com/atotto/clipboard)

[![GoDoc](https://godoc.org/github.com/atotto/clipboard?status.svg)](http://godoc.org/github.com/atotto/clipboard)

# Clipboard for Go

Provide copying and pasting to the Clipboard for Go.

Build:

    $ go get github.com/atotto/clipboard

Platforms:

* OSX
* Windows 7 (probably work on other Windows)
* Linux, Unix (requires 'xclip' or 'xsel' command to be installed)
* WebAssembly (WASM, `GOOS=js GOARCH=wasm`, requires support for **OSC 52 escape sequences**) 

### Current Features
- `WriteAll()` sends text to clipboard via:
  - Browser Web Clipboard API (when available, HTTPS/localhost)
  - OSC 52 escape sequences (fallback to terminal clipboard)
- `ReadAll()` reads from clipboard via:
  - Browser Web Clipboard API (when available)
  - Returns error in OSC 52 mode (read not supported)

### Building for WASM
```bash
GOOS=js GOARCH=wasm go build ./...
```

### Browser Requirements
- **HTTPS or localhost** - Browser Clipboard API only works in secure contexts
- **User permission** - Browser may require user gesture or permission (depends on browser/site)
- **Modern browser** - Chrome 66+, Firefox 53+, Safari 13.1+, Edge 79+

### Environment Variables
- `CLIPBOARD_MODE=browser` — Force browser Web Clipboard API (requires HTTPS/localhost)
- `CLIPBOARD_MODE=osc52` — Force OSC 52 terminal clipboard
- `CLIPBOARD_MODE=disabled` — Disable clipboard (sets `Unsupported = true`)
- No `CLIPBOARD_MODE` env var — Auto-detect:
  1. Try browser Web Clipboard API if available (HTTPS or localhost)
  2. Fall back to OSC 52 if TERM and TTY detected
  3. Return unsupported if neither available

### Security Notes
- **Browser Clipboard API** (Phase 2): Gated by browser security policies (HTTPS, user permission, gesture requirements)
- **OSC 52**: Security depends on terminal configuration (e.g., Kitty's clipboard gate, iTerm2 settings)
- **Caller Responsibility**: Applications are responsible for higher-level gating if clipboard access should be restricted

### Testing Clipboard Operations

#### Testing with Browser Web Clipboard API

In a browser environment (HTTPS or localhost):

```bash
# Build WASM binary with browser clipboard support
GOOS=js GOARCH=wasm go build -o clipboard.wasm ./cmd/gocopy
```

The browser Web Clipboard API works in:
- Local development (`http://localhost:3000`)
- HTTPS deployments
- Secure contexts only

#### Testing with OSC 52 (Terminal)

In a terminal that supports OSC 52:

```bash
# Force OSC 52 mode
CLIPBOARD_MODE=osc52 GOOS=js GOARCH=wasm go build -o clipboard.wasm ./cmd/gocopy
```

Compatible terminals: Kitty, iTerm2, Alacritty, WezTerm, Konsole (with settings)

#### Testing in Mixed Environments

```bash
# Auto-detect (tries browser first, falls back to OSC 52)
GOOS=js GOARCH=wasm go build -o clipboard.wasm ./cmd/gocopy

# Disable clipboard
CLIPBOARD_MODE=disabled GOOS=js GOARCH=wasm go build -o clipboard.wasm ./cmd/gocopy
```

Document: 

* http://godoc.org/github.com/atotto/clipboard

Notes:

* Text string only
* UTF-8 text encoding only (no conversion)

TODO:

* Clipboard watcher(?)

## Commands:

paste shell command:

    $ go install github.com/atotto/clipboard/cmd/gopaste@latest
    $ # example:
    $ gopaste > document.txt

copy shell command:

    $ go install github.com/atotto/clipboard/cmd/gocopy@latest
    $ # example:
    $ cat document.txt | gocopy



