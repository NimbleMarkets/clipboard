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

## WebAssembly (WASM) Support

WASM builds (`GOOS=js GOARCH=wasm`) support clipboard operations via **OSC 52 escape sequences**, a terminal-based clipboard protocol supported by Kitty, iTerm2, Alacritty, WezTerm, and other terminals.

### Current Features
- `WriteAll()` encodes text as OSC 52 escape sequences and sends to stdout
- `ReadAll()` is not yet supported (Phase 2 will add browser Web Clipboard API)

### Building for WASM
```bash
GOOS=js GOARCH=wasm go build ./...
```

### Environment Variables
- `CLIPBOARD_MODE=osc52` — Force OSC 52 mode (ignore auto-detection)
- `CLIPBOARD_MODE=disabled` — Disable clipboard (sets `Unsupported = true`)
- No `CLIPBOARD_MODE` env var — Auto-detect based on `TERM` environment variable

### Security Notes
- **Browser Clipboard API** (Phase 2): Gated by browser security policies (HTTPS, user permission, gesture requirements)
- **OSC 52**: Security depends on terminal configuration (e.g., Kitty's clipboard gate, iTerm2 settings)
- **Caller Responsibility**: Applications are responsible for higher-level gating if clipboard access should be restricted

### Testing with OSC 52
To test `WriteAll()` in WASM mode, run in a terminal that supports OSC 52 (e.g., Kitty, iTerm2, Alacritty):

```bash
GOOS=js GOARCH=wasm go run ./cmd/gocopy < input.txt  # Writes OSC 52 to stdout
```

The escape sequence will be visible in the terminal (as `^[]52;c;...^G` or similar). In supported terminals, the text will be copied to the clipboard.

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



