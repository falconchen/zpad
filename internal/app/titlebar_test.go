package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// A window that draws its controls over the toolbar (macOS traffic lights)
// keeps the buttons clear of them on both sides, and in full screen, where
// TitleBar is zero, the bar keeps its usual padding.
func TestToolbarKeepsClearOfWindowControls(t *testing.T) {
	const width, height = 1200, 800
	for _, c := range []struct {
		name  string
		bar   ui.TitleBar
		left  float32 // where the toolbar's content starts
		right float32 // where it ends, from the right edge
	}{
		{"traffic lights", ui.TitleBar{Height: 44, Left: 74, Bottom: 29}, 84, 10},
		{"buttons on both sides", ui.TitleBar{Height: 32, Left: 74, Right: 138, Bottom: 32}, 84, 148},
		{"full screen", ui.TitleBar{}, 10, 10},
	} {
		t.Run(c.name, func(t *testing.T) {
			a := New(settings.Default(), t.TempDir())
			tt := ui.NewTester(a.View, width, height)
			tt.SetTitleBar(c.bar)

			first, ok := tt.Find(locales.ToggleSidebar())
			if !ok {
				t.Fatalf("no sidebar button; texts %q", tt.Texts())
			}
			if first.X < c.left {
				t.Errorf("the first button starts at %.0f DIP, want at least %.0f", first.X, c.left)
			}

			last, ok := tt.Find(locales.Settings())
			if !ok {
				t.Fatalf("no settings button; texts %q", tt.Texts())
			}
			if gap := width - (last.X + last.W); gap < c.right {
				t.Errorf("the last button ends %.0f DIP from the right edge, want at least %.0f", gap, c.right)
			}
		})
	}
}
