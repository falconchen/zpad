package main

import (
	"embed"
	"fmt"

	"github.com/egoist/mygo/ui"
)

//go:embed fonts/*.ttf
var fontFS embed.FS

// registerFonts adds the bundled JetBrains Mono faces. Call it before any
// window exists.
func registerFonts() error {
	for _, face := range []string{"Regular", "Bold", "Italic", "BoldItalic"} {
		data, err := fontFS.ReadFile("fonts/JetBrainsMono-" + face + ".ttf")
		if err != nil {
			return err
		}
		if err := ui.RegisterFont(data, "JetBrains Mono"); err != nil {
			return fmt.Errorf("register JetBrains Mono %s: %w", face, err)
		}
	}
	return nil
}
