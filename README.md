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
To test `WriteAll()` in WASM mode, you need a WASM runtime (e.g., Node.js with WASM support, wasmtime, or a browser). 

Example with Node.js:
```bash
# Build WASM binary
GOOS=js GOARCH=wasm go build -o clipboard.wasm ./cmd/gocopy

# Run with Node.js (requires WASM runtime support)
node -e "const fs = require('fs'); const buf = fs.readFileSync('clipboard.wasm'); WebAssembly.instantiate(buf, {}).catch(e => console.error(e));"
```

The escape sequence will be written to stdout (visible as `^[]52;c;...^G` or similar in compatible terminals). In terminals supporting OSC 52 (Kitty, iTerm2, Alacritty, WezTerm), the text will be copied to the clipboard.

**Note:** OSC 52 clipboard support depends entirely on the terminal or runtime environment. WASM runtimes like Node.js or browser environments may not support OSC 52 escape sequences.

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



