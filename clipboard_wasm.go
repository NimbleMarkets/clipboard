// Copyright 2013 @atotto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js
// +build js

package clipboard

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"syscall/js"
	"time"
)

const wasmUnsupportedErr = "clipboard is not available (disabled or auto-detection failed)"

var clipboardMode string

// encodeOSC52 encodes text as an OSC 52 escape sequence.
// Format: ESC ] 52 ; c ; <base64-encoded-text> BEL
// Where ESC is \x1b and BEL is \x07
func encodeOSC52(text string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("\x1b]52;c;%s\x07", encoded)
}

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

// writeOSC52 writes text to the terminal using the OSC 52 escape sequence
func writeOSC52(text string) error {
	sequence := encodeOSC52(text)
	_, err := os.Stdout.Write([]byte(sequence))
	return err
}

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

// isBrowserEnvironment checks if navigator.clipboard exists in the JavaScript environment
func isBrowserEnvironment() bool {
	navigator := js.Global().Get("navigator")
	if navigator.IsUndefined() {
		return false
	}
	clipboard := navigator.Get("clipboard")
	return !clipboard.IsUndefined()
}

// hasHTTPSOrLocalhost checks if the location protocol is HTTPS or the hostname is localhost
func hasHTTPSOrLocalhost() bool {
	location := js.Global().Get("location")
	if location.IsUndefined() {
		return false
	}

	protocol := location.Get("protocol").String()
	hostname := location.Get("hostname").String()

	return protocol == "https:" || hostname == "localhost" || hostname == "127.0.0.1"
}

// readBrowserClipboard reads text from the browser clipboard using the Web Clipboard API
func readBrowserClipboard() (string, error) {
	navigator := js.Global().Get("navigator")
	clipboard := navigator.Get("clipboard")

	resultCh := make(chan string)
	errCh := make(chan error)

	// Call navigator.clipboard.readText()
	promise := clipboard.Call("readText")

	// Handle promise resolution
	promise.Call("then",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				resultCh <- args[0].String()
			}
			return nil
		}),
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				errCh <- errors.New(args[0].String())
			} else {
				errCh <- errors.New("clipboard read failed")
			}
			return nil
		}),
	)

	// Wait for result or timeout
	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errCh:
		return "", err
	case <-time.After(5 * time.Second):
		return "", errors.New("clipboard read timeout")
	}
}

// writeBrowserClipboard writes text to the browser clipboard using the Web Clipboard API
func writeBrowserClipboard(text string) error {
	navigator := js.Global().Get("navigator")
	clipboard := navigator.Get("clipboard")

	resultCh := make(chan bool)
	errCh := make(chan error)

	// Call navigator.clipboard.writeText(text)
	promise := clipboard.Call("writeText", text)

	// Handle promise resolution
	promise.Call("then",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resultCh <- true
			return nil
		}),
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				errCh <- errors.New(args[0].String())
			} else {
				errCh <- errors.New("clipboard write failed")
			}
			return nil
		}),
	)

	// Wait for result or timeout
	select {
	case <-resultCh:
		return nil
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Second):
		return errors.New("clipboard write timeout")
	}
}

// detectClipboardMode detects which clipboard mode is available
func detectClipboardMode() string {
	// Check for browser environment first
	if isBrowserEnvironment() && hasHTTPSOrLocalhost() {
		return "browser"
	}

	// Fall back to OSC 52 if TERM is set and stdout is a TTY
	term := os.Getenv("TERM")
	if term != "" {
		stat, err := os.Stdout.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
			return "osc52"
		}
	}

	return ""
}

func init() {
	mode := os.Getenv("CLIPBOARD_MODE")

	// Handle explicit overrides first
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

	// Auto-detect available mode
	detectedMode := detectClipboardMode()
	if detectedMode == "" {
		Unsupported = true
		clipboardMode = ""
		return
	}

	clipboardMode = detectedMode
}
