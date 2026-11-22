package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestBundledFontRegistered(t *testing.T) {
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
