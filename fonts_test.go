package main

import (
	"runtime"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestBundledFontRegistered(t *testing.T) {
	// The list comes from the system's text engine. On Linux it reports the
	// app's font as proportional, or leaves it out, where fontconfig has no
	// fonts of its own, as on a bare CI machine; there registering needs
	// Pango and fontconfig too. The app ships macOS and Windows, where the
	// release workflow runs this test.
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("the font list is only checked where the app ships")
	}
	if err := registerFonts(); err != nil {
		t.Fatal(err)
	}
	for _, f := range ui.FontFamilies() {
		if f.Name == "JetBrains Mono" && f.Monospaced {
			return
		}
	}
	t.Fatal("JetBrains Mono is not listed as a monospaced family")
}
