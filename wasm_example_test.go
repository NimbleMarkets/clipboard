// Copyright 2013 @atotto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js
// +build js

package clipboard_test

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// ExampleWriteAll demonstrates writing to clipboard in WASM mode using OSC 52.
func ExampleWriteAll_wasm() {
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
	// Output: Text written to clipboard via OSC 52
}

// ExampleReadAll demonstrates that ReadAll is not yet supported in WASM.
func ExampleReadAll_wasm() {
	text, err := clipboard.ReadAll()

	if err != nil {
		fmt.Printf("ReadAll() is not supported in WASM: %v\n", err)
	}

	if text != "" {
		fmt.Printf("Unexpected text: %s\n", text)
	}
	// Output: ReadAll() is not supported in WASM: clipboard read not supported in WASM mode; browser API support coming in Phase 2
}
