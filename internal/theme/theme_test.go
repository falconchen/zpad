package theme

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

func TestFor(t *testing.T) {
	light, dark := For(false), For(true)
	if light.Dark {
		t.Fatal("light theme is dark")
	}
	if !dark.Dark {
		t.Fatal("dark theme is not dark")
	}
	if light.Background == dark.Background {
		t.Fatal("light and dark share a background")
	}
	if light.Radius != 8 || dark.Radius != 8 {
		t.Fatalf("radii %v, %v; want 8", light.Radius, dark.Radius)
	}
	ct := EditorTheme(light)
	if ct.Background != light.Background || ct.Text != light.Text {
		t.Fatal("editor theme does not follow the UI theme")
	}
	if got := ct.Tokens[editor.TokenKeyword].Color; got != ui.Hex("#7C3AED") {
		t.Fatalf("keyword color %v, want #7C3AED", got)
	}
	if got := ct.Tokens[editor.TokenString].Color; got != ui.Hex("#047857") {
		t.Fatalf("string color %v, want #047857", got)
	}
}
