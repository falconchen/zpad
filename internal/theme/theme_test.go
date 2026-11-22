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

func TestScrollbarsAreThin(t *testing.T) {
	for name, th := range map[string]*ui.Theme{"light": Light(), "dark": Dark()} {
		s := th.ScrollbarStyle
		if s.ThumbWidth > 6 || th.ScrollbarWidth > 6 {
			t.Errorf("%s scrollbar thumb %v (width %v), want at most 6", name, s.ThumbWidth, th.ScrollbarWidth)
		}
		if s.TrackWidth <= s.ThumbWidth {
			t.Errorf("%s scrollbar track %v, thumb %v: no room to grab", name, s.TrackWidth, s.ThumbWidth)
		}
		if s.Radius <= 0 {
			t.Errorf("%s scrollbar radius %v, want rounded", name, s.Radius)
		}
		if s.Thumb.A >= 200 {
			t.Errorf("%s scrollbar thumb alpha %d, want translucent", name, s.Thumb.A)
		}
	}
}

func TestFileTreeTheme(t *testing.T) {
	base := Light()
	tree := FileTree(base)
	if tree.Radius != 0 {
		t.Errorf("tree radius %v, want square full-width rows", tree.Radius)
	}
	if tree.Accent == base.Accent {
		t.Error("tree still fills chosen rows with the accent")
	}
	if tree.Accent.A >= 100 {
		t.Errorf("tree wash alpha %d, want faint", tree.Accent.A)
	}
	if tree.AccentText != base.Text {
		t.Error("tree chosen rows do not keep the text color")
	}
	if tree.InactiveAccent != tree.Accent {
		t.Error("tree chosen rows change in an inactive window")
	}
	if base.Radius == 0 || base.AccentText == base.Text {
		t.Error("FileTree changed the theme it copied")
	}
}
