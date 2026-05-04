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

func isTerminal() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	// Check if stdout is a character device (terminal)
	// This is OS-specific; for WASM in browser, stdout may not have meaningful mode
	return (stat.Mode() & os.ModeCharDevice) != 0
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

	// Auto-detect: check TERM environment variable
	term := os.Getenv("TERM")
	if term == "" {
		Unsupported = true
		return
	}

	// If TERM is set, assume we're in a terminal environment
	// OSC 52 support depends on the specific terminal, but presence of TERM is a good indicator
}
