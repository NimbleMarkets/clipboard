# Phase 2: Browser Web Clipboard API Support Design

**Date:** 2026-05-04  
**Status:** Design Pending Approval  
**Scope:** Add browser Web Clipboard API read/write support to go-clipboard WASM builds

---

## Overview

Extend `clipboard_wasm.go` to support browser Web Clipboard API (`navigator.clipboard.readText()` and `navigator.clipboard.writeText()`) alongside existing OSC 52 support. Use channel-based blocking to present a synchronous API while handling JavaScript's async Promises internally.

### Goals
- Support clipboard read in WASM environments (Phase 1 only had write)
- Browser API preferred over OSC 52 when available
- User-configurable via `CLIPBOARD_MODE` environment variable
- Maintain synchronous public API (`ReadAll()`, `WriteAll()`)
- Graceful fallback to OSC 52 when browser API unavailable
- Clear error messages for permission/security failures

### Why Browser API Reading?

OSC 52 clipboard reading is impractical:
- Requires terminal to echo back via escape sequence
- Not universally supported
- Complex parsing and timeout handling
- Fragile and unreliable

Browser Web Clipboard API provides clean read support, but requires handling async Promises. Solution: **block internally on channels** (same pattern as subprocess shelling out in other platforms).

---

## Architecture

### Clipboard Mode Selection

In `init()`, determine which mechanism to use:

```
1. Check CLIPBOARD_MODE env var
   ├─ "browser" → use browser API only
   ├─ "osc52" → use OSC 52 only
   ├─ "disabled" → set Unsupported = true
   └─ (not set) → auto-detect
   
2. If auto-detect:
   ├─ Check navigator.clipboard exists
   ├─ Check HTTPS or localhost
   ├─ If both OK → use browser API
   └─ Else → fall back to OSC 52
   
3. If OSC 52 mode:
   ├─ Check TERM env var
   ├─ Check stdout is TTY
   └─ If both OK → use OSC 52
```

### Implementation Pattern: Channel-Based Blocking

Bridge JavaScript async/Promise with Go's synchronous API:

```go
func readAll() (string, error) {
    if clipboardMode == "browser" {
        return readBrowserClipboard()
    }
    return readOSC52Clipboard()
}

func readBrowserClipboard() (string, error) {
    // JavaScript navigator.clipboard.readText() returns a Promise
    // Block on channel until Promise resolves or rejects
    
    resultChan := make(chan string, 1)
    errorChan := make(chan error, 1)
    
    // Register Promise callbacks via syscall/js
    // When Promise resolves: send result to resultChan
    // When Promise rejects: send error to errorChan
    
    // Block until one of the channels receives a value
    select {
    case text := <-resultChan:
        return text, nil
    case err := <-errorChan:
        return "", err
    case <-time.After(5 * time.Second):
        return "", errors.New("clipboard read timeout")
    }
}

func writeBrowserClipboard(text string) error {
    // Same pattern: register callbacks, block on channels
}
```

### Modes Supported

| Mode | Read | Write | Notes |
|------|------|-------|-------|
| `browser` | ✓ | ✓ | Requires HTTPS/localhost, user permission |
| `osc52` | ✗ | ✓ | Terminal clipboard via escape sequences |
| `disabled` | ✗ | ✗ | `Unsupported = true` |
| auto-detect | ✓/✗ | ✓ | Browser if available, else OSC 52 |

---

## Error Handling

### Browser API Errors

1. **NotAllowedError** - User denied permission
   - Error: "Clipboard access denied (user permission required)"
   - Fallback: none (browser denied it)

2. **NotSupportedError** - Browser doesn't support API
   - Error: "Clipboard API not supported in this environment"
   - Fallback: to OSC 52 if auto-detect mode

3. **SecurityError** - HTTPS requirement not met
   - Error: "Clipboard requires secure context (HTTPS/localhost)"
   - Fallback: to OSC 52 if auto-detect mode

4. **Timeout** - Promise takes too long to resolve
   - Error: "Clipboard operation timeout"
   - Retry: application responsibility

### OSC 52 Errors

1. **Unavailable** - TERM not set or stdout not TTY
   - Error: "clipboard is not available (disabled or auto-detection failed)"

2. **Write failed** - stdout I/O error
   - Error: underlying system error

3. **Read not supported** - OSC 52 read not implemented
   - Error: "clipboard read not supported in OSC 52 mode"

### User-Facing Errors

All errors returned as `error` from `ReadAll()` and `WriteAll()`. Applications handle with:
- Try/catch-like patterns
- Graceful fallbacks
- User-friendly error messages

---

## Implementation Details

### File Structure

**clipboard_wasm.go (extended):**
- Keep existing OSC 52 code (Phase 1)
- Add browser API implementation functions
- Add mode detection in `init()`
- Update `readAll()` and `writeAll()` to dispatch to correct implementation

### Key Functions

```go
// Determine which clipboard mechanism to use
func detectClipboardMode() string

// Browser API implementations
func readBrowserClipboard() (string, error)
func writeBrowserClipboard(text string) error

// Mode dispatch
func readAll() (string, error)
func writeAll(text string) error
```

### Imports

Add to existing imports:
- `"time"` - for timeout on Promise wait
- `"syscall/js"` - already used, for calling navigator.clipboard

### Constants & Variables

```go
var clipboardMode string // Set in init(): "browser", "osc52", or ""(disabled)
const browserReadTimeout = 5 * time.Second
```

---

## Testing Strategy

### Unit Tests

1. **Mode detection**
   - `CLIPBOARD_MODE=browser` forces browser mode
   - `CLIPBOARD_MODE=osc52` forces OSC 52 mode
   - `CLIPBOARD_MODE=disabled` sets Unsupported=true
   - Auto-detect chooses correctly based on environment

2. **Browser read/write (WASM environment)**
   - Mock `navigator.clipboard` calls
   - Test Promise resolution paths
   - Test error paths (permission denied, timeout, etc.)

3. **OSC 52 read/write (existing)**
   - Verify Phase 1 tests still pass
   - Ensure fallback works when browser unavailable

### Integration Tests

1. **Mode priority**
   - Browser available + `CLIPBOARD_MODE=osc52` → uses OSC 52
   - Browser unavailable + auto-detect → falls back to OSC 52
   - Both unavailable → `Unsupported = true`

2. **Full read/write cycle**
   - Write text, verify it reaches go-clipboard
   - Read text, verify correct data returned

### Manual Testing

1. **Browser environment**
   - Test with Firefox, Chrome, Safari
   - Verify permission prompt appears (if needed)
   - Verify HTTPS requirement enforced
   - Test on localhost (should work without HTTPS)

2. **Terminal OSC 52 environment**
   - Test with Kitty, iTerm2, Alacritty
   - Verify OSC 52 fallback works

3. **Mixed scenarios**
   - Browser + force OSC 52 mode
   - Terminal + try browser mode (should gracefully fail)

---

## Documentation Updates

### README.md

Update Phase 1 section with Phase 2 additions:
- Browser Web Clipboard API now supported
- Read operations supported via browser API
- Mode selection (`CLIPBOARD_MODE=browser` vs `CLIPBOARD_MODE=osc52`)
- Browser security requirements (HTTPS, permissions)
- Updated testing instructions

### Code Comments

- Explain channel-based blocking pattern
- Document Promise error handling
- Note timeout behavior

---

## Acceptance Criteria

- ✅ `ReadAll()` works via browser Web Clipboard API
- ✅ `WriteAll()` works via browser Web Clipboard API
- ✅ `CLIPBOARD_MODE` controls which mechanism is used
- ✅ Auto-detection tries browser first, falls back to OSC 52
- ✅ Synchronous API from caller perspective (async hidden internally)
- ✅ Clear error messages for all failure modes
- ✅ Phase 1 OSC 52 support continues working
- ✅ Timeout prevents indefinite hangs on Promise
- ✅ All tests pass (unit, integration, manual)
- ✅ README updated with Phase 2 features and requirements

---

## Phase Progression

**Phase 1 (Complete):** OSC 52 write-only support
- ✅ Basic OSC 52 encoding
- ✅ Environment detection (TERM, TTY)
- ✅ `CLIPBOARD_MODE` override
- ✅ Tests and documentation

**Phase 2 (This):** Browser Web Clipboard API
- Add browser read/write
- Add mode selection and fallback
- Update documentation

**Future:** Enhanced features
- Clipboard change detection
- Format detection (plain text vs rich text)
- Browser Web Clipboard API read-only mode
- Integration with go-booba for BubbleTea apps

---

## References

- **Web Clipboard API:** https://developer.mozilla.org/en-US/docs/Web/API/Clipboard_API
- **syscall/js:** https://pkg.go.dev/syscall/js
- **Phase 1 Design:** 2026-05-04-wasm-clipboard-design.md
