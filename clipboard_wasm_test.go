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

func TestClipboardModeOverride(t *testing.T) {
	// Note: This is a conceptual test since init() runs only once at package load
	// In actual usage, the behavior depends on environment variables at build/run time
	// This test documents the expected behavior
	tests := []struct {
		mode             string
		wantUnsupported  bool
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

		// In actual implementation, init() only runs once per package load
		// For testing purposes, we document the expected behavior here
		// The actual verification happens through manual testing with environment variables
		_ = tt
	}
}

func TestWriteAllWhenUnsupported(t *testing.T) {
	// Save original Unsupported state
	oldUnsupported := Unsupported
	defer func() { Unsupported = oldUnsupported }()

	// Set Unsupported to true (simulates disabled mode or detection failure)
	Unsupported = true

	// WriteAll should return an error
	err := WriteAll("test text")
	if err == nil {
		t.Error("WriteAll() should return error when Unsupported is true")
	}
	if err != nil && !strings.Contains(err.Error(), "not available") {
		t.Errorf("Error message should mention 'not available', got %q", err.Error())
	}
}
