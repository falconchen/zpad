package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func TestThemedMenusOpen(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()

	if err := tt.Click(locales.Theme()); err != nil {
		t.Fatal(err)
	}
	menu := tt.Menu()
	if len(menu) == 0 {
		t.Fatal("the theme menu did not open")
	}
	if r, ok := tt.Find(locales.ThemeLight()); ok {
		tt.Move(r.X+r.W/2, r.Y+r.H/2)
		tt.Frame()
		if got := tt.Cursor(); got != ui.CursorPointer {
			t.Fatalf("cursor over a menu item %v, want the pointing hand", got)
		}
	} else {
		t.Fatalf("theme menu has no %q item", locales.ThemeLight())
	}
	tt.Key(0, ui.KeyEscape)

	if err := tt.Click(encoding.Label(encoding.UTF8)); err != nil {
		t.Fatal(err)
	}
	menu = tt.Menu()
	found := false
	for _, item := range menu {
		if item == locales.ReopenWithEncoding() {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("encoding menu %q has no reopen item", menu)
	}
}
