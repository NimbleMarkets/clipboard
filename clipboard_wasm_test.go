// Copyright 2013 @atotto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js

package clipboard

import (
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
		{"", "\x1b]52;c;=\x07"},
		{"a\nb", "\x1b]52;c;YQpi\x07"},
	}
	for _, tt := range tests {
		if got := encodeOSC52(tt.input); got != tt.expected {
			t.Errorf("encodeOSC52(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
