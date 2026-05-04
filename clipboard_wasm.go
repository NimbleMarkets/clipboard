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

func init() {
	// TODO: Detection and override logic (Task 2)
}
