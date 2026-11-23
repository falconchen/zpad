package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// A new tab owes the keyboard to its editor: the editor element outlives the
// tab, so AutoFocus did nothing once a tab was open, and typing needed a
// click in the text first.
func TestNewTabTakesTheKeyboard(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/one.txt", "one"))
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()

	info, ok := tt.AccessByRole(ui.RoleButton, locales.NewFile())
	if !ok {
		t.Fatalf("the toolbar has no %q button", locales.NewFile())
	}
	tt.ClickAt(info.Bounds.X+info.Bounds.W/2, info.Bounds.Y+info.Bounds.H/2)

	if len(a.left.tabs) != 2 || a.left.active != 1 {
		t.Fatalf("tabs %d, active %d, want the new one chosen", len(a.left.tabs), a.left.active)
	}
	if !a.left.activeTab().state.Focused() {
		t.Fatal("the new tab did not take the keyboard")
	}
}

// Choosing a tab gives its editor the keyboard too, from the tab click.
func TestChoosingATabTakesTheKeyboard(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/one.txt", "one"))
	a.addDocument(docAt("/tmp/two.txt", "two"))
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()

	if err := tt.Click("one.txt"); err != nil {
		t.Fatal(err)
	}
	if a.left.active != 0 {
		t.Fatalf("active tab %d, want 0", a.left.active)
	}
	if !a.left.tabs[0].state.Focused() {
		t.Fatal("the chosen tab did not take the keyboard")
	}
}
