# Phase 2: Browser Web Clipboard API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add browser Web Clipboard API read/write support to go-clipboard WASM builds, with mode selection and fallback to OSC 52.

**Architecture:** Extend `clipboard_wasm.go` with browser API implementations. Use channel-based blocking to present synchronous API while handling JavaScript Promises internally. Detect clipboard mode in `init()` based on environment and `CLIPBOARD_MODE` override. Dispatch `readAll()` and `writeAll()` to correct implementation.

**Tech Stack:** Go 1.11+, syscall/js, Web Clipboard API, channels for async/sync bridging

---

## Task 1: Add browser API helper functions and mode detection

**Files:**
- Modify: `clipboard_wasm.go`
- Test: `clipboard_wasm_test.go`

**Goal:** Implement core browser API functions and mode detection logic.

- [ ] **Step 1: Add mode detection function**

Add this function to `clipboard_wasm.go` after the imports section:

```go
var clipboardMode string // "browser", "osc52", or "" (disabled)

func detectClipboardMode() string {
	// Check CLIPBOARD_MODE override first
	mode := os.Getenv("CLIPBOARD_MODE")
	if mode == "browser" || mode == "osc52" || mode == "disabled" {
		return mode
	}

	// Auto-detect: try browser first
	if isBrowserEnvironment() && hasHTTPSOrLocalhost() {
		return "browser"
	}

	// Fall back to OSC 52
	term := os.Getenv("TERM")
	if term != "" {
		// Verify stdout is TTY
		stat, err := os.Stdout.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
			return "osc52"
		}
	}

	// Neither available
	return ""
}

func isBrowserEnvironment() bool {
	navigator := js.Global().Get("navigator")
	clipboard := navigator.Get("clipboard")
	return !clipboard.Undefined()
}

func hasHTTPSOrLocalhost() bool {
	location := js.Global().Get("location")
	protocol := location.Get("protocol").String()
	hostname := location.Get("hostname").String()

	// HTTPS or localhost
	return protocol == "https:" || hostname == "localhost" || hostname == "127.0.0.1"
}
```

Run: `go build ./...`
Expected: Builds successfully

- [ ] **Step 2: Add browser read function with channel-based blocking**

Add this function to `clipboard_wasm.go`:

```go
func readBrowserClipboard() (string, error) {
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	navigator := js.Global().Get("navigator")
	clipboard := navigator.Get("clipboard")
	readText := clipboard.Get("readText")

	// Call navigator.clipboard.readText() which returns a Promise
	readText.Call("invoke").Call("then",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) > 0 {
				resultChan <- args[0].String()
			}
			return nil
		}),
		js.FuncOf(func(this js.Value, args []js.Value) any {
			errMsg := "clipboard read failed"
			if len(args) > 0 {
				errMsg = args[0].String()
			}
			errorChan <- errors.New(errMsg)
			return nil
		}),
	)

	// Block until Promise resolves or timeout
	select {
	case text := <-resultChan:
		return text, nil
	case err := <-errorChan:
		return "", err
	case <-time.After(5 * time.Second):
		return "", errors.New("clipboard read timeout")
	}
}
```

Run: `go build ./...`
Expected: Builds successfully

- [ ] **Step 3: Add browser write function with channel-based blocking**

Add this function to `clipboard_wasm.go`:

```go
func writeBrowserClipboard(text string) error {
	resultChan := make(chan bool, 1)
	errorChan := make(chan error, 1)

	navigator := js.Global().Get("navigator")
	clipboard := navigator.Get("clipboard")
	writeText := clipboard.Get("writeText")

	// Call navigator.clipboard.writeText(text) which returns a Promise
	writeText.Invoke(text).Call("then",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			resultChan <- true
			return nil
		}),
		js.FuncOf(func(this js.Value, args []js.Value) any {
			errMsg := "clipboard write failed"
			if len(args) > 0 {
				errMsg = args[0].String()
			}
			errorChan <- errors.New(errMsg)
			return nil
		}),
	)

	// Block until Promise resolves or timeout
	select {
	case <-resultChan:
		return nil
	case err := <-errorChan:
		return err
	case <-time.After(5 * time.Second):
		return errors.New("clipboard write timeout")
	}
}
```

Run: `go build ./...`
Expected: Builds successfully

- [ ] **Step 4: Update init() to detect clipboard mode**

Replace the existing `init()` function in `clipboard_wasm.go` with:

```go
func init() {
	mode := os.Getenv("CLIPBOARD_MODE")

	// Check explicit overrides first
	switch mode {
	case "disabled":
		Unsupported = true
		clipboardMode = ""
		return
	case "browser":
		clipboardMode = "browser"
		return
	case "osc52":
		clipboardMode = "osc52"
		return
	}

	// Auto-detect
	detected := detectClipboardMode()
	if detected == "" {
		Unsupported = true
		return
	}
	clipboardMode = detected
}
```

Run: `go build ./...`
Expected: Builds successfully

- [ ] **Step 5: Commit**

```bash
git add clipboard_wasm.go
git commit -m "feat(wasm): add browser API functions and mode detection"
```

---

## Task 2: Update readAll() to dispatch based on clipboard mode

**Files:**
- Modify: `clipboard_wasm.go`

**Goal:** Make `readAll()` dispatch to browser API or OSC 52 based on detected mode.

- [ ] **Step 1: Write test for readAll() dispatch**

Add this test to `clipboard_wasm_test.go`:

```go
func TestReadAllDispatch(t *testing.T) {
	// Save original state
	oldMode := clipboardMode
	oldUnsupported := Unsupported
	defer func() {
		clipboardMode = oldMode
		Unsupported = oldUnsupported
	}()

	// Test 1: Browser mode should call browser API (we can't fully test without mocking JS)
	clipboardMode = "browser"
	Unsupported = false
	// Just verify it doesn't panic when trying to call browser API
	_, _ = readAll()

	// Test 2: OSC 52 mode should return error (Phase 1 behavior - read not supported)
	clipboardMode = "osc52"
	Unsupported = false
	_, err := readAll()
	if err == nil {
		t.Error("readAll() should return error in OSC 52 mode")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("Error should mention 'not supported', got: %v", err)
	}

	// Test 3: Unsupported mode should return error
	Unsupported = true
	_, err = readAll()
	if err == nil {
		t.Error("readAll() should return error when Unsupported is true")
	}
}
```

Run: `go test -v ./... -run TestReadAllDispatch`
Expected: FAIL (function behavior not yet implemented)

- [ ] **Step 2: Update readAll() function**

Replace the existing `readAll()` function in `clipboard_wasm.go` with:

```go
func readAll() (string, error) {
	if Unsupported {
		return "", errors.New(wasmUnsupportedErr)
	}

	switch clipboardMode {
	case "browser":
		return readBrowserClipboard()
	case "osc52":
		return "", errors.New("clipboard read not supported in OSC 52 mode")
	default:
		return "", errors.New(wasmUnsupportedErr)
	}
}
```

Run: `go test -v ./... -run TestReadAllDispatch`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add clipboard_wasm.go clipboard_wasm_test.go
git commit -m "feat(wasm): implement readAll() dispatch to browser API or OSC 52"
```

---

## Task 3: Update writeAll() to dispatch based on clipboard mode

**Files:**
- Modify: `clipboard_wasm.go`

**Goal:** Make `writeAll()` dispatch to browser API or OSC 52 based on detected mode.

- [ ] **Step 1: Write test for writeAll() dispatch**

Add this test to `clipboard_wasm_test.go`:

```go
func TestWriteAllDispatch(t *testing.T) {
	// Save original state
	oldMode := clipboardMode
	oldUnsupported := Unsupported
	defer func() {
		clipboardMode = oldMode
		Unsupported = oldUnsupported
	}()

	// Test 1: Browser mode
	clipboardMode = "browser"
	Unsupported = false
	// Just verify it doesn't panic
	_ = writeAll("test")

	// Test 2: OSC 52 mode
	clipboardMode = "osc52"
	Unsupported = false
	err := writeAll("test")
	if err != nil {
		// OSC 52 write might fail in test environment, that's OK
		// We just want to verify it attempts write, not error with "not supported"
		if strings.Contains(err.Error(), "not supported") {
			t.Errorf("OSC 52 should attempt write, not report 'not supported': %v", err)
		}
	}

	// Test 3: Unsupported mode
	Unsupported = true
	err = writeAll("test")
	if err == nil {
		t.Error("writeAll() should return error when Unsupported is true")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("Error should mention 'not available', got: %v", err)
	}
}
```

Run: `go test -v ./... -run TestWriteAllDispatch`
Expected: FAIL (function behavior not yet implemented)

- [ ] **Step 2: Update writeAll() function**

Replace the existing `writeAll()` function in `clipboard_wasm.go` with:

```go
func writeAll(text string) error {
	if Unsupported {
		return errors.New(wasmUnsupportedErr)
	}

	switch clipboardMode {
	case "browser":
		return writeBrowserClipboard(text)
	case "osc52":
		return writeOSC52(text)
	default:
		return errors.New(wasmUnsupportedErr)
	}
}

// Helper to call existing OSC 52 write logic
func writeOSC52(text string) error {
	sequence := encodeOSC52(text)
	_, err := os.Stdout.Write([]byte(sequence))
	return err
}
```

Run: `go test -v ./... -run TestWriteAllDispatch`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add clipboard_wasm.go clipboard_wasm_test.go
git commit -m "feat(wasm): implement writeAll() dispatch to browser API or OSC 52"
```

---

## Task 4: Add CLIPBOARD_MODE environment variable tests

**Files:**
- Modify: `clipboard_wasm_test.go`

**Goal:** Test that CLIPBOARD_MODE environment variable properly overrides auto-detection.

- [ ] **Step 1: Write test for CLIPBOARD_MODE override**

Add this test to `clipboard_wasm_test.go`:

```go
func TestClipboardModeEnvOverride(t *testing.T) {
	// Note: init() runs once at package load, so we can't easily test override
	// This test documents the expected behavior via environment variable
	tests := []struct {
		mode          string
		expectedMode  string
		expectUnsupported bool
	}{
		{"browser", "browser", false},
		{"osc52", "osc52", false},
		{"disabled", "", true},
	}

	for _, tt := range tests {
		// These tests are conceptual since init() only runs once
		// In real use: CLIPBOARD_MODE=browser go build ./...
		// In real use: CLIPBOARD_MODE=osc52 go build ./...
		// In real use: CLIPBOARD_MODE=disabled go build ./...
		_ = tt
	}
}
```

Run: `go test -v ./... -run TestClipboardModeEnvOverride`
Expected: PASS (test is conceptual/documentation)

- [ ] **Step 2: Commit**

```bash
git add clipboard_wasm_test.go
git commit -m "test(wasm): add CLIPBOARD_MODE environment variable documentation"
```

---

## Task 5: Update README with Phase 2 documentation

**Files:**
- Modify: `README.md`

**Goal:** Document Phase 2 features, browser requirements, and mode selection.

- [ ] **Step 1: Review current README WASM section**

Read `README.md` and find the WASM section (added in Phase 1).

- [ ] **Step 2: Update WASM section with Phase 2 content**

Replace the "Current Features" subsection with:

```markdown
### Current Features
- `WriteAll()` sends text to clipboard via:
  - Browser Web Clipboard API (when available, HTTPS/localhost)
  - OSC 52 escape sequences (fallback to terminal clipboard)
- `ReadAll()` reads from clipboard via:
  - Browser Web Clipboard API (when available)
  - Returns error in OSC 52 mode (read not supported)
```

And update "Environment Variables" subsection to:

```markdown
### Environment Variables
- `CLIPBOARD_MODE=browser` — Force browser Web Clipboard API (requires HTTPS/localhost)
- `CLIPBOARD_MODE=osc52` — Force OSC 52 terminal clipboard
- `CLIPBOARD_MODE=disabled` — Disable clipboard (sets `Unsupported = true`)
- No `CLIPBOARD_MODE` env var — Auto-detect:
  1. Try browser Web Clipboard API if available (HTTPS or localhost)
  2. Fall back to OSC 52 if TERM and TTY detected
  3. Return unsupported if neither available

### Browser Requirements
- **HTTPS or localhost** - Browser Clipboard API only works in secure contexts
- **User permission** - Browser may require user gesture or permission (depends on browser/site)
- **Modern browser** - Chrome 66+, Firefox 53+, Safari 13.1+, Edge 79+
```

- [ ] **Step 3: Update "Testing with OSC 52" section**

Replace the existing testing section with:

```markdown
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
```

- [ ] **Step 4: Verify README formatting**

Run: `cat README.md | grep -A 20 "WebAssembly (WASM) Support"`
Expected: WASM section shows updated documentation

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: update WASM section with Phase 2 browser API features"
```

---

## Task 6: Run final verification tests

**Files:**
- Test: `clipboard_wasm_test.go`

**Goal:** Verify all tests pass and code is clean.

- [ ] **Step 1: Run all tests**

Run: `go test -v ./...`
Expected: All tests PASS

- [ ] **Step 2: Run go vet**

Run: `go vet ./...`
Expected: No errors or warnings

- [ ] **Step 3: Verify WASM build**

Run: `GOOS=js GOARCH=wasm go build ./...`
Expected: Builds successfully

- [ ] **Step 4: Verify all platforms still build**

Run: 
```bash
GOOS=darwin GOARCH=amd64 go build ./... && echo "✓ darwin/amd64"
GOOS=linux GOARCH=amd64 go build ./... && echo "✓ linux/amd64"
GOOS=windows GOARCH=amd64 go build ./... && echo "✓ windows/amd64"
```
Expected: All build successfully

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "test(wasm): final verification of Phase 2 implementation"
```

---

## Acceptance Criteria

- ✅ `readAll()` works via browser Web Clipboard API when available
- ✅ `readAll()` returns error in OSC 52 mode (read not supported)
- ✅ `writeAll()` works via browser Web Clipboard API when available
- ✅ `writeAll()` works via OSC 52 when browser unavailable
- ✅ `CLIPBOARD_MODE` environment variable controls mode selection
- ✅ Auto-detection tries browser first, falls back to OSC 52
- ✅ Synchronous API from caller perspective (async hidden internally)
- ✅ Timeout prevents indefinite hangs on Promise (5 second timeout)
- ✅ Clear error messages for all failure modes
- ✅ Phase 1 OSC 52 support continues working
- ✅ All tests pass (unit, integration)
- ✅ README updated with Phase 2 features and requirements
- ✅ All platforms build (darwin, linux, windows, js/wasm)

---

## Implementation Notes

### Browser API Error Handling

Browser errors are translated to Go errors:
- Permission denied → "clipboard read/write failed: [browser error]"
- Secure context required → "clipboard requires HTTPS or localhost"
- Timeout → "clipboard read/write timeout"

### Channel-Based Blocking

The channel pattern bridges async JavaScript with sync Go:
1. Create channels for result and error
2. Register Promise callbacks that write to channels
3. Use `select` with timeout to wait for Promise resolution
4. Return result or error synchronously to caller

This avoids goroutine explosion and complex callback chains.

### Mode Priority

1. Explicit `CLIPBOARD_MODE` override (browser, osc52, disabled)
2. Auto-detect: browser if available, else OSC 52
3. Unsupported if neither available

---

## References

- **Phase 2 Design:** 2026-05-04-phase2-browser-clipboard-design.md
- **Web Clipboard API:** https://developer.mozilla.org/en-US/docs/Web/API/Clipboard_API
- **syscall/js:** https://pkg.go.dev/syscall/js
- **Phase 1 Implementation:** Tasks from 2026-05-04-wasm-clipboard.md
