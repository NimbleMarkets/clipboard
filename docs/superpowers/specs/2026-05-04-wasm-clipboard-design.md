# WASM Clipboard Support Design

**Date:** 2026-05-04  
**Status:** Design Approved  
**Scope:** Phase 1 — OSC 52 write-only clipboard support for WebAssembly builds

---

## Overview

Enable the `atotto/clipboard` library to compile and function when built for WebAssembly (`GOOS=js GOARCH=wasm`). Phase 1 focuses on **OSC 52 escape sequence support for terminal-based clipboard writes**. Phase 2 will add browser Web Clipboard API support.

### Goals
- Library compiles cleanly with GOOS=js (no build errors)
- WASM builds can write to clipboard via OSC 52 (terminal clipboard protocol)
- Clear error messages when operations are unsupported
- Foundation for future browser API support
- Documented security model (caller responsibility for gating)

---

## Architecture

### Public API (Unchanged)
The public interface remains identical:
```go
func ReadAll() (string, error)
func WriteAll(text string) error
var Unsupported bool
```

### Platform Implementations
Current structure continues:
- `clipboard_darwin.go` (+build darwin)
- `clipboard_unix.go` (+build freebsd linux netbsd openbsd solaris dragonfly)
- `clipboard_windows.go` (+build windows)
- `clipboard_plan9.go` (+build plan9)
- `clipboard.go` (platform-agnostic public API)

### New: WASM Implementation
**File:** `clipboard_wasm.go`  
**Build tag:** `+build js`

Implements the internal functions:
- `readAll()` — returns error (not supported in Phase 1)
- `writeAll(text)` — encodes text and sends OSC 52 escape sequence to stdout

---

## Phase 1: OSC 52 Terminal Clipboard

### What is OSC 52?
OSC 52 is a terminal escape sequence that instructs compatible terminals to write data to the clipboard:
```
ESC ] 52 ; c ; <base64-encoded-text> BEL
```

Supported terminals: Kitty, iTerm2, xterm (with configuration), Alacritty, WezTerm, and others.

### Implementation Details

#### `writeAll(text string) error`
1. Encode `text` as base64
2. Construct OSC 52 sequence: `\x1b]52;c;<base64>\x07`
3. Write to stdout via `os.Stdout.Write()`
4. Return error if write fails

#### `readAll() string, error`
Return error: "clipboard read not supported in WASM mode" (deferred to Phase 2)

#### Initialization (`init()`)
Detect environment and set up clipboard mode:

1. **Detection:** Check if OSC 52 might be viable
   - Check `TERM` environment variable: if set (e.g., "xterm", "screen", "tmux", "alacritty"), assume terminal environment
   - Optionally check if stdout is a TTY using `os.Stdout.Stat()` and `IsCharDevice()`
   - If both checks pass, assume OSC 52 is likely viable
   - Set default mode based on detection (OSC 52 if viable, else Unsupported)

2. **Explicit override:** Allow environment variable control
   - `CLIPBOARD_MODE=osc52` — force OSC 52 mode (ignore detection)
   - `CLIPBOARD_MODE=disabled` — disable clipboard (set `Unsupported = true`)
   - No `CLIPBOARD_MODE` env var — use detection result

3. **Fallback:** If detection fails and no override, set `Unsupported = true`

This approach mirrors booba's pattern of detection + explicit override, allowing auto-detection while giving users control when needed.

---

## Error Handling

### ReadAll() in WASM
```go
func readAll() (string, error) {
    return "", errors.New("clipboard read not supported in WASM mode; browser API support coming in Phase 2")
}
```

### WriteAll() Failures
Natural propagation of `os.Stdout.Write()` errors (e.g., stdout closed, permission denied).

### Unsupported Flag
Set `Unsupported = true` when:
- OSC 52 detection fails and no override is set
- User explicitly sets `CLIPBOARD_MODE=disabled`

Callers can check `Unsupported` and offer fallback UI or skip clipboard features.

---

## File Structure

```
clipboard.go                              # Public API (unchanged)
clipboard_darwin.go                       # Darwin (unchanged)
clipboard_unix.go                         # Unix/Linux (unchanged)
clipboard_windows.go                      # Windows (unchanged)
clipboard_plan9.go                        # Plan 9 (unchanged)
clipboard_wasm.go                         # NEW: WASM + OSC 52
clipboard_test.go                         # Existing tests
example_test.go                           # Existing tests
docs/superpowers/specs/
  2026-05-04-wasm-clipboard-design.md    # This document
```

---

## Documentation & Security Model

### README Changes
Add section on WASM support:

> **WebAssembly (WASM) Support**
>
> WASM builds (GOOS=js) support clipboard operations via OSC 52 escape sequences, a terminal-based clipboard protocol supported by Kitty, iTerm2, Alacritty, WezTerm, and others.
>
> **Current limitations:**
> - `WriteAll()` sends OSC 52 sequences to stdout
> - `ReadAll()` is not supported (returning an error); browser API support is planned
>
> **Environment control:**
> - `CLIPBOARD_MODE=osc52` — force OSC 52 mode
> - `CLIPBOARD_MODE=disabled` — disable clipboard
>
> **Security:** 
> - Browser Clipboard API is gated by browser security policies (HTTPS, user permission, gesture requirements)
> - OSC 52 security depends on terminal configuration (e.g., Kitty's clipboard gate, iTerm2 settings)
> - **Callers are responsible for higher-level gating** if clipboard access should be restricted in their application

### Future Phase 2 Note
Document that Phase 2 will add browser Web Clipboard API support via `syscall/js` for browser environments.

---

## Testing Strategy

### Phase 1 Testing
1. **Compile test:** Verify library builds with `GOOS=js GOARCH=wasm`
2. **Sequence format test:** Unit test that verifies OSC 52 sequence encoding is correct (without actual terminal I/O)
3. **Error path test:** Verify `ReadAll()` returns expected error message
4. **Manual/integration test:** Document how to test with a compatible terminal (e.g., Kitty, iTerm2)

### Limitations
- CI/unit testing cannot verify actual OSC 52 clipboard interaction (requires compatible terminal)
- Real-world testing requires running in/against a compatible terminal

---

## Phase 2 Considerations (Future)

Structure Phase 1 to accommodate Phase 2 additions:
- Browser Web Clipboard API detection will be another mode in `init()`
- `readAll()` and `writeAll()` can be extended to support browser API without refactoring
- Example: `CLIPBOARD_MODE=browser` for explicit browser-only mode

---

## Build & Deployment

### WASM Build Command
```bash
GOOS=js GOARCH=wasm go build ./cmd/...
```

### CI/CD
- Add WASM build target to CI (verify compilation, don't require terminal test pass)
- Existing platform tests (darwin, linux, windows) continue unchanged

---

## Acceptance Criteria

- ✅ Library compiles cleanly with `GOOS=js GOARCH=wasm`
- ✅ `WriteAll()` encodes and writes valid OSC 52 sequences
- ✅ `ReadAll()` returns clear error message
- ✅ `Unsupported` flag set correctly based on detection/overrides
- ✅ README documents WASM support, security model, and environment variables
- ✅ Tests pass for all existing platforms (darwin, linux, windows, plan9)
- ✅ New tests cover OSC 52 format and WASM-specific error paths

---

## References

- **Issue #61:** Support GOOS=js
- **Issue #73:** OSC 52 support discussion
- **Reference:** go-booba clipboard handling (detection + override pattern)
- **OSC 52 Spec:** Terminal escape sequence for clipboard operations
