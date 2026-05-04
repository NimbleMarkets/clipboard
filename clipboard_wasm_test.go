// Copyright 2013 @atotto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js
// +build js

package clipboard

import (
	"os"
	"strings"
	"testing"
)

func TestOSC52Encoding(t *testing.T) {
	text := "hello world"
	expected := "\x1b]52;c;aGVsbG8gd29ybGQ=\x07"
	encoded := encodeOSC52(text)
	if encoded != expected {
		t.Errorf("encodeOSC52(%q) = %q, want %q", text, encoded, expected)
	}
}

func TestOSC52EncodingMultiple(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "\x1b]52;c;aGVsbG8=\x07"},
		{"", "\x1b]52;c;\x07"},
		{"a\nb", "\x1b]52;c;YQpi\x07"},
	}
	for _, tt := range tests {
		if got := encodeOSC52(tt.input); got != tt.expected {
			t.Errorf("encodeOSC52(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestWasmReadAllUnsupported(t *testing.T) {
	// Only run on WASM builds
	if os.Getenv("GOOS") != "js" {
		t.Skip("skipping WASM test on non-WASM build")
	}

	result, err := readAll()
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
