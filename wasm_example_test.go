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

// wasmWriteAllExample demonstrates writing to clipboard in WASM mode using OSC 52.
// Not an example function (no Example prefix) because WriteAll has side effects
// that cannot be captured in deterministic test output. See README for usage examples.
func wasmWriteAllExample() {
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

// wasmReadAllExample demonstrates that ReadAll is not yet supported in WASM.
// Not an example function because it documents error behavior. See README for usage examples.
func wasmReadAllExample() {
	text, err := clipboard.ReadAll()

	if err != nil {
		fmt.Printf("ReadAll() is not supported in WASM: %v\n", err)
	}

	if text != "" {
		fmt.Printf("Unexpected text: %s\n", text)
	}
}
