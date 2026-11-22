package app

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func manyTabs(t *testing.T, n int) (*App, *ui.Tester) {
	t.Helper()
	a := New(settings.Default(), t.TempDir())
	for i := 0; i < n; i++ {
		a.addDocument(docAt(fmt.Sprintf("/tmp/tab-%02d.txt", i), "x"))
	}
	tt := ui.NewTester(a.View, 700, 500)
	// Bounds report the last frame, so the bar needs a frame to learn it
	// overflows, and another to show the button.
	for range 4 {
		tt.Frame()
	}
	return a, tt
}

func TestTabBarOverflowShowsTabList(t *testing.T) {
	a, tt := manyTabs(t, 14)
	if !a.tabOverflow[0] {
		t.Fatal("tab bar did not notice that its tabs overflow")
	}
	if !tt.HasText(locales.AllTabs()) {
		t.Fatalf("no tab list button; texts %q", tt.Texts())
	}
}

// A name longer than the cap must not make one tab as wide as the bar, and
// the tab keeps the name for tooltips and the tab list.
func TestLongTabNameIsCapped(t *testing.T) {
	name := "a-very-long-file-name-that-would-take-the-whole-bar.component.test.tsx"
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/"+name, "x"))
	tt := ui.NewTester(a.View, 700, 500)
	tt.Frame()

	r, ok := tt.Find(name)
	if !ok {
		t.Fatalf("no tab for %q; texts %q", name, tt.Texts())
	}
	// The cap plus the tab's padding, gap and close button, about 240 DIP.
	if maxW := tabTextWidth + 60; r.W > maxW {
		t.Fatalf("tab is %.0f DIP wide, want at most %.0f", r.W, maxW)
	}
	if !tt.HasText(name) {
		t.Fatalf("tab lost the full name; texts %q", tt.Texts())
	}
}

func TestTabListChoosesAHiddenTab(t *testing.T) {
	a, tt := manyTabs(t, 14)
	if err := tt.Click(locales.AllTabs()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	last := a.left.tabs[len(a.left.tabs)-1]
	if err := tt.ChooseMenuItem(last.doc.Name()); err != nil {
		t.Fatal(err)
	}
	if a.left.active != len(a.left.tabs)-1 {
		t.Fatalf("chose tab %d, want %d", a.left.active, len(a.left.tabs)-1)
	}
	if a.revealDoc != "" {
		t.Fatalf("reveal %q still pending after the tab bar was built", a.revealDoc)
	}
}

func TestActivateTabChoosesAndReveals(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.addDocument(docAt("/tmp/b.txt", "b"))
	a.activateTab(0, 0)
	if a.left.active != 0 {
		t.Fatalf("active tab %d, want 0", a.left.active)
	}
	if a.revealPane != 0 || a.revealDoc != a.left.tabs[0].doc.ID {
		t.Fatalf("reveal %d %q, want 0 %q", a.revealPane, a.revealDoc, a.left.tabs[0].doc.ID)
	}
	a.activateTab(0, 5) // out of range: nothing changes
	if a.left.active != 0 {
		t.Fatalf("out-of-range choice moved the active tab to %d", a.left.active)
	}
}

func TestToolbarButtonShowsHandCursor(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	tt := ui.NewTester(a.View, 1200, 800)
	r, ok := tt.Find(locales.NewFile())
	if !ok {
		t.Fatalf("no New File button; texts %q", tt.Texts())
	}
	tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	if got := tt.Cursor(); got != ui.CursorPointer {
		t.Fatalf("cursor over the button %v, want the pointing hand", got)
	}
}
