package app

import (
	"testing"

	"github.com/egoist/mygo"
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

// The menu runs outside the view, so a tab it adds owes the frame that shows
// it: CmdOrCtrl+N pressed while Command stayed down used to add its tabs to a
// view nothing rebuilt, and they appeared only once Command was let go.
func TestNewFileMenuAsksForAFrame(t *testing.T) {
	var frames int
	old := askFrame
	askFrame = func(*mygo.Window) { frames++ }
	t.Cleanup(func() { askFrame = old })

	a := New(settings.Default(), t.TempDir())
	item := findMenuItem(a.menuItems(), locales.NewFile())
	if item == nil || item.Click == nil {
		t.Fatalf("the menu has no %q item with a click handler", locales.NewFile())
	}
	item.Click(item, nil)
	if len(a.left.tabs) != 1 || a.left.active != 0 {
		t.Fatalf("the menu added %d tabs, active %d, want one active tab", len(a.left.tabs), a.left.active)
	}
	if frames != 1 {
		t.Fatalf("the menu asked for %d frames, want 1", frames)
	}
}

func findMenuItem(items []*mygo.MenuItem, label string) *mygo.MenuItem {
	for _, it := range items {
		if it.Label == label {
			return it
		}
		if found := findMenuItem(it.Submenu, label); found != nil {
			return found
		}
	}
	return nil
}

// The update item is ours rather than the plugin's: the plugin's label
// would panic in tests, where the plugin is not used, and could not follow
// the app's own language the way the rest of the menu does. Its click is
// the plugin's, so it is not exercised here.
func TestCheckForUpdatesMenuEntry(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	var item *mygo.MenuItem
	var walk func(items []*mygo.MenuItem)
	walk = func(items []*mygo.MenuItem) {
		for _, it := range items {
			if it.Label == locales.CheckForUpdates() {
				item = it
			}
			walk(it.Submenu)
		}
	}
	walk(a.menuItems())
	if item == nil {
		t.Fatalf("the menu has no %q item", locales.CheckForUpdates())
	}
	if item.Click == nil {
		t.Fatalf("%q has no click handler", locales.CheckForUpdates())
	}
}

// The About item is ours rather than the system's, so that the dialog can
// show the license and the repository.
func TestAboutMenuEntry(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	var about *mygo.MenuItem
	var walk func(items []*mygo.MenuItem)
	walk = func(items []*mygo.MenuItem) {
		for _, it := range items {
			if it.Label == locales.AboutApp() {
				about = it
			}
			walk(it.Submenu)
		}
	}
	walk(a.menuItems())
	if about == nil {
		t.Fatalf("the menu has no %q item", locales.AboutApp())
	}
	if about.Click == nil {
		t.Fatalf("%q has no click handler", locales.AboutApp())
	}
	about.Click(about, nil)
	if !a.aboutOpen {
		t.Fatal("choosing About did not open the dialog")
	}
}
