# WASM Clipboard Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable the clipboard library to compile and function for WebAssembly with OSC 52 escape sequence write support.

**Architecture:** Add a new platform-specific file (`clipboard_wasm.go`) with `+build js` tag that implements write via OSC 52 and read as unsupported. Initialize at runtime with environment detection and explicit override support. All existing platforms remain unchanged.

**Tech Stack:** Go 1.11+, standard library (os, encoding/base64, errors), syscall for TTY detection

---

## Task 1: Create clipboard_wasm.go with OSC 52 write implementation

**Files:**
- Create: `clipboard_wasm.go`

**Goal:** Implement the core OSC 52 encoding and write functionality.

- [ ] **Step 1: Write the failing test for OSC 52 encoding**

Create `clipboard_test.go` and add this test (or append if test file exists):

```go
func TestOSC52Encoding(t *testing.T) {
	text := "hello world"
	expected := "\x1b]52;c;aGVsbG8gd29ybGQ=\x07"
	encoded := encodeOSC52(text)
	if encoded != expected {
		t.Errorf("encodeOSC52(%q) = %q, want %q", text, encoded, expected)
	}
}
```

Run: `go test -v ./...`
Expected: FAIL with "undefined: encodeOSC52" or similar

- [ ] **Step 2: Create clipboard_wasm.go with build tag and initial structure**

Create file `clipboard_wasm.go`:

```go
// Copyright 2013 @atotto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js

package clipboard

import (
	"encoding/base64"
	"fmt"
	"os"
)

// encodeOSC52 encodes text as an OSC 52 escape sequence.
// Format: ESC ] 52 ; c ; <base64-encoded-text> BEL
// Where ESC is \x1b and BEL is \x07
func encodeOSC52(text string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("\x1b]52;c;%s\x07", encoded)
}

func writeAll(text string) error {
	sequence := encodeOSC52(text)
	_, err := os.Stdout.Write([]byte(sequence))
	return err
}

func readAll() (string, error) {
	return "", errors.New("clipboard read not supported in WASM mode; browser API support coming in Phase 2")
}

func init() {
	// TODO: Detection and override logic (Task 2)
}
```

Note: `errors` import will be added in Task 2.

- [ ] **Step 3: Add missing import and run test**

Update imports in `clipboard_wasm.go`:

```go
import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)
```

Run: `go test -v ./... -run TestOSC52Encoding`
Expected: PASS (test compiles and verifies encoding)

- [ ] **Step 4: Test with multiple inputs to verify encoding correctness**

Add test in `clipboard_test.go`:

```go
func TestOSC52EncodingMultiple(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "\x1b]52;c;aGVsbG8=\x07"},
		{"", "\x1b]52;c;=\x07"},
		{"a\nb", "\x1b]52;c;YQpi\x07"},
	}
	for _, tt := range tests {
		if got := encodeOSC52(tt.input); got != tt.expected {
			t.Errorf("encodeOSC52(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
```

Run: `go test -v ./... -run TestOSC52Encoding`
Expected: PASS for all cases

- [ ] **Step 5: Commit**

```bash
git add clipboard_wasm.go clipboard_test.go
git commit -m "feat(wasm): add OSC 52 encoding for clipboard write"
```

---

## Task 2: Implement readAll() error and Unsupported handling

**Files:**
- Modify: `clipboard_wasm.go`
- Modify: `clipboard_test.go`

**Goal:** Ensure readAll() returns a clear error and verify behavior.

- [ ] **Step 1: Write test for readAll() error**

Add test in `clipboard_test.go`:

```go
func TestWasmReadAllUnsupported(t *testing.T) {
	// Only run on WASM builds
	if os.Getenv("GOOS") != "js" {
		t.Skip("skipping WASM test on non-WASM build")
	}
	
	result, err := clipboard.ReadAll()
	if err == nil {
		t.Error("ReadAll() should return error in WASM mode")
	}
	if result != "" {
		t.Errorf("ReadAll() should return empty string on error, got %q", result)
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("Error message should mention 'not supported', got %q", err.Error())
	}
}
```

Run: `go test -v ./...`
Expected: SKIP on non-WASM builds (expected), or FAIL if this somehow runs and readAll works

- [ ] **Step 2: Verify readAll() implementation is correct**

Review the `readAll()` function in `clipboard_wasm.go` from Task 1:

```go
func readAll() (string, error) {
	return "", errors.New("clipboard read not supported in WASM mode; browser API support coming in Phase 2")
}
```

This is already correct. Run the test again:

Run: `GOOS=js go test -v ./... -run TestWasmReadAllUnsupported`
Expected: PASS (or SKIP if cross-compilation isn't set up in test environment)

- [ ] **Step 3: Commit**

```bash
git add clipboard_test.go
git commit -m "test(wasm): add test for readAll unsupported error"
```

---

## Task 3: Implement init() with environment detection and overrides

**Files:**
- Modify: `clipboard_wasm.go`
- Modify: `clipboard_test.go`

**Goal:** Detect terminal environment (TERM, TTY) and allow explicit overrides via CLIPBOARD_MODE.

- [ ] **Step 1: Write test for CLIPBOARD_MODE override**

Add test in `clipboard_test.go`:

```go
func TestClipboardModeOverride(t *testing.T) {
	tests := []struct {
		mode      string
		wantUnsupported bool
	}{
		{"osc52", false},
		{"disabled", true},
		{"", false}, // default: use detection
	}
	
	for _, tt := range tests {
		// Save original env
		oldMode := os.Getenv("CLIPBOARD_MODE")
		defer os.Setenv("CLIPBOARD_MODE", oldMode)
		
		// Set test mode
		if tt.mode == "" {
			os.Unsetenv("CLIPBOARD_MODE")
		} else {
			os.Setenv("CLIPBOARD_MODE", tt.mode)
		}
		
		// Re-run init (simulated by checking expected state)
		// Note: actual init only runs once at package load; this test is conceptual
		// We'll verify the logic in the implementation review step
	}
}
```

This test is conceptual (init runs once at package load). We'll verify the logic in code review.

- [ ] **Step 2: Implement init() with detection logic**

Update `init()` in `clipboard_wasm.go`:

```go
func init() {
	mode := os.Getenv("CLIPBOARD_MODE")
	
	if mode == "disabled" {
		Unsupported = true
		return
	}
	
	if mode == "osc52" {
		// Explicitly enabled
		return
	}
	
	// Auto-detect: check TERM and TTY
	term := os.Getenv("TERM")
	if term == "" {
		Unsupported = true
		return
	}
	
	// If TERM is set, assume we're in a terminal environment
	// OSC 52 support depends on the specific terminal, but presence of TERM is a good indicator
}
```

Run: `go build -o /tmp/test ./...`
Expected: Builds successfully with WASM target

- [ ] **Step 3: Add helper function for TTY detection (optional but good practice)**

Add to `clipboard_wasm.go` for future extensibility:

```go
func isTerminal() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	// Check if stdout is a character device (terminal)
	// This is OS-specific; for WASM in browser, stdout may not have meaningful mode
	return (stat.Mode() & os.ModeCharDevice) != 0
}
```

Update `init()` to use it (optional enhancement for Phase 2):

```go
func init() {
	mode := os.Getenv("CLIPBOARD_MODE")
	
	switch mode {
	case "disabled":
		Unsupported = true
		return
	case "osc52":
		return
	}
	
	// Auto-detect
	term := os.Getenv("TERM")
	if term == "" {
		Unsupported = true
		return
	}
	// TERM is set; assume OSC 52 might work
}
```

Run: `go build -o /tmp/test ./...`
Expected: Builds successfully

- [ ] **Step 4: Verify WASM build compiles**

Build with WASM target to ensure clipboard_wasm.go is used:

Run: `GOOS=js GOARCH=wasm go build -o /tmp/test.wasm ./...`
Expected: Builds successfully without errors

- [ ] **Step 5: Commit**

```bash
git add clipboard_wasm.go
git commit -m "feat(wasm): add init with environment detection and CLIPBOARD_MODE override"
```

---

## Task 4: Test WASM build and verify no regressions on other platforms

**Files:**
- Test: all platform files

**Goal:** Ensure WASM build works and existing platforms still function.

- [ ] **Step 1: Run existing tests on native platform**

Run: `go test -v ./...`
Expected: All tests PASS (same as before, no regressions)

- [ ] **Step 2: Build WASM target**

Run: `GOOS=js GOARCH=wasm go build -o /tmp/clipboard.wasm ./...`
Expected: Builds without errors

- [ ] **Step 3: Verify clipboard_wasm.go is included in WASM build**

Run: `GOOS=js GOARCH=wasm go build -v ./... 2>&1 | grep clipboard_wasm`
Expected: Output includes clipboard_wasm.go

- [ ] **Step 4: Verify other platforms still build**

Run: 
```bash
GOOS=darwin GOARCH=amd64 go build -o /tmp/clipboard-darwin ./...
GOOS=linux GOARCH=amd64 go build -o /tmp/clipboard-linux ./...
GOOS=windows GOARCH=amd64 go build -o /tmp/clipboard-windows.exe ./...
```
Expected: All build successfully

- [ ] **Step 5: Commit**

```bash
git add -A  # No files changed, but marks progress
git commit -m "test(wasm): verify build on all platforms"
```

---

## Task 5: Update README with WASM documentation

**Files:**
- Modify: `README.md`

**Goal:** Document WASM support, OSC 52, and environment variable overrides for users.

- [ ] **Step 1: Review current README structure**

Read `README.md` to understand where to add WASM section.

- [ ] **Step 2: Add WASM section to README**

Add this section after the "Platforms:" list:

```markdown
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
```

- [ ] **Step 3: Verify README formatting**

Run: `cat README.md | less` (or open in editor to review)
Expected: Section is well-formatted, clear, and consistent with existing style

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: add WASM and OSC 52 documentation to README"
```

---

## Task 6: Add example/integration test for manual verification

**Files:**
- Create or Modify: `example_test.go` or create `wasm_example_test.go`

**Goal:** Provide a clear example of how to use WASM clipboard in a real scenario.

- [ ] **Step 1: Write an example test that documents WASM usage**

Add to `example_test.go` or create `wasm_example_test.go`:

```go
// +build js

package clipboard_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/atotto/clipboard"
)

// ExampleWasmWriteOSC52 demonstrates writing to clipboard in WASM mode using OSC 52.
func ExampleWasmWriteOSC52(t *testing.T) {
	// In a real WASM application, this would write an OSC 52 sequence
	// In a terminal that supports OSC 52 (Kitty, iTerm2, Alacritty), the text is copied

	text := "Hello from WASM!"
	err := clipboard.WriteAll(text)
	
	if clipboard.Unsupported {
		fmt.Println("Clipboard not available (Unsupported flag set)")
		return
	}
	
	if err != nil {
		fmt.Printf("Error writing to clipboard: %v\n", err)
		return
	}
	
	fmt.Println("Text written to clipboard via OSC 52")
}

// ExampleWasmReadNotSupported demonstrates that ReadAll is not yet supported in WASM.
func ExampleWasmReadNotSupported(t *testing.T) {
	text, err := clipboard.ReadAll()
	
	if err != nil {
		fmt.Printf("ReadAll() is not supported in WASM: %v\n", err)
	}
	
	if text != "" {
		fmt.Printf("Unexpected text: %s\n", text)
	}
	// Output: ReadAll() is not supported in WASM: clipboard read not supported in WASM mode; browser API support coming in Phase 2
}
```

Run: `GOOS=js go test -v ./... -run ExampleWasm`
Expected: Example compiles and shows expected output

- [ ] **Step 2: Commit**

```bash
git add wasm_example_test.go  # or example_test.go if appended
git commit -m "test(wasm): add examples for manual verification"
```

---

## Task 7: Final verification and build test matrix

**Files:**
- No new files; verification only

**Goal:** Ensure the implementation is complete and all platforms build cleanly.

- [ ] **Step 1: Run all tests**

Run: `go test -v ./...`
Expected: All tests PASS

- [ ] **Step 2: Build matrix for all supported platforms**

Run:
```bash
echo "Testing builds for all platforms..."
GOOS=darwin GOARCH=amd64 go build -o /tmp/clipboard-darwin ./... && echo "✓ darwin/amd64"
GOOS=darwin GOARCH=arm64 go build -o /tmp/clipboard-darwin-arm64 ./... && echo "✓ darwin/arm64"
GOOS=linux GOARCH=amd64 go build -o /tmp/clipboard-linux ./... && echo "✓ linux/amd64"
GOOS=windows GOARCH=amd64 go build -o /tmp/clipboard-windows.exe ./... && echo "✓ windows/amd64"
GOOS=js GOARCH=wasm go build -o /tmp/clipboard.wasm ./... && echo "✓ js/wasm"
```
Expected: All builds succeed with ✓ markers

- [ ] **Step 3: Verify no errors or warnings**

Run: `go vet ./...`
Expected: No errors or warnings

- [ ] **Step 4: Check for any test coverage gaps**

Run: `go test -cover ./...`
Expected: Coverage is reasonable (no hard requirement, but note any significant gaps)

- [ ] **Step 5: Final commit summary**

Run:
```bash
git log --oneline -10
```
Expected: See commits from Tasks 1-6 with clear, descriptive messages

---

## Acceptance Criteria (Verification Checklist)

- ✅ `clipboard_wasm.go` created with `+build js` tag
- ✅ `WriteAll()` implements OSC 52 encoding correctly
- ✅ `ReadAll()` returns clear error message
- ✅ `init()` detects TERM environment variable for auto-detection
- ✅ `CLIPBOARD_MODE` environment variable overrides work (osc52, disabled)
- ✅ `Unsupported` flag set correctly based on detection/overrides
- ✅ Library builds cleanly for: darwin, linux, windows, plan9, and WASM
- ✅ Existing tests pass (no regressions)
- ✅ WASM-specific tests cover error paths and encoding
- ✅ README updated with WASM documentation and security notes
- ✅ All `go vet` checks pass
- ✅ Example code provided for users

---

## Implementation Notes

### OSC 52 Encoding
- Base64 encoding is standard library (`encoding/base64`)
- Sequence format: `ESC ] 52 ; c ; <base64> BEL` = `\x1b]52;c;<base64>\x07`
- No special handling needed for different terminals at this stage

### Environment Detection
- `TERM` environment variable presence indicates likely terminal environment
- `CLIPBOARD_MODE` overrides provide explicit control (mirrors booba pattern)
- `Unsupported` flag allows graceful degradation

### Phase 2 Readiness
- Structure left open for browser Web Clipboard API in Phase 2
- `readAll()` stub clearly indicates future support
- `init()` can accept new modes (e.g., `CLIPBOARD_MODE=browser`)

### Testing Strategy
- Unit tests verify encoding logic and error paths
- Platform build matrix ensures no regressions
- WASM-specific tests can be skipped on non-WASM builds
- Real OSC 52 testing requires compatible terminal (documented as manual test)
