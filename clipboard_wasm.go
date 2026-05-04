// Copyright 2013 @atotto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js

package clipboard

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

const wasmUnsupportedErr = "clipboard is not available (disabled or auto-detection failed)"

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
	sequence := encodeOSC52(text)
	_, err := os.Stdout.Write([]byte(sequence))
	return err
}

func readAll() (string, error) {
	return "", errors.New("clipboard read not supported in WASM mode; browser API support coming in Phase 2")
}

func init() {
	mode := os.Getenv("CLIPBOARD_MODE")

	switch mode {
	case "disabled":
		Unsupported = true
		return
	case "osc52":
		// Explicitly enabled
		return
	}

	// Auto-detect: check TERM and verify stdout is a terminal
	term := os.Getenv("TERM")
	if term == "" {
		Unsupported = true
		return
	}

	// Verify stdout is actually a terminal (TTY) to avoid writing escape sequences to logs/files
	stat, err := os.Stdout.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		Unsupported = true
		return
	}

	// Both TERM is set and stdout is a terminal; OSC 52 should work
}
