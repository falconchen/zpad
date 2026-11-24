package app

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// The window's menu bar is the title bar on Windows: its labels keep clear
// of the window buttons at the right, and the toolbar below keeps its own
// padding, not the buttons'.
func TestMenuBarSitsAboveTheToolbar(t *testing.T) {
	const width, height = 1200, 800
	a := New(settings.Default(), t.TempDir())
	a.customMenuBar = true
	a.addDocument(docAt("/tmp/menubar.txt", "text"))

	tt := ui.NewTester(a.View, width, height)
	tt.SetTitleBar(ui.TitleBar{Height: MenuBarHeight, Right: 138})
	tt.Frame()

	labels := a.menuBarLabels()
	if len(labels) == 0 {
		t.Fatal("the menu bar has no labels")
	}
	for _, label := range labels {
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("the menu bar has no %q; texts %q", label, tt.Texts())
		}
		if r.X < 10 {
			t.Errorf("%q starts at %.0f DIP, inside the window's edge", label, r.X)
		}
		if r.Y+r.H > MenuBarHeight {
			t.Errorf("%q ends at %.0f DIP, below the %d DIP menu bar", label, r.Y+r.H, MenuBarHeight)
		}
	}

	first, ok := tt.Find(locales.ToggleSidebar())
	if !ok {
		t.Fatalf("no sidebar button; texts %q", tt.Texts())
	}
	if first.Y < MenuBarHeight {
		t.Errorf("the toolbar starts at %.0f DIP, inside the menu bar", first.Y)
	}

	last, ok := tt.Find(locales.Settings())
	if !ok {
		t.Fatalf("no settings button; texts %q", tt.Texts())
	}
	// The window buttons are in the menu bar above: the toolbar keeps its
	// own 10 DIP padding, not the 138 DIP of the three buttons.
	if gap := width - (last.X + last.W); gap < 8 || gap > 40 {
		t.Errorf("the toolbar ends %.0f DIP from the right edge, want its own padding", gap)
	}
}

// A click on a menu label leaves the keyboard in the editor: the menu's
// Edit items, Undo and Copy in particular, act on the text, not on the
// button.
func TestMenuBarKeepsTheEditorFocus(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.customMenuBar = true
	a.addDocument(docAt("/tmp/menubar.txt", "text"))

	tt := ui.NewTester(a.View, 1200, 800)
	tt.SetTitleBar(ui.TitleBar{Height: MenuBarHeight, Right: 138})
	tt.Frame()

	if !a.activeTab().state.Focused() {
		t.Fatal("the editor did not take the keyboard")
	}
	if err := tt.Click(locales.MenuFile()); err != nil {
		t.Fatalf("click the File menu: %v", err)
	}
	if !a.activeTab().state.Focused() {
		t.Fatal("the menu bar took the keyboard from the editor")
	}
}

// A split divides what is left below the menu bar and the toolbar.
func TestSplitCentersBelowTheMenuBar(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.customMenuBar = true
	a.Settings.SidebarOpen = false
	a.addDocument(docAt("/tmp/menubar.txt", "text"))
	a.openSplit(true)

	tt := ui.NewTester(a.View, 1000, 700)
	tt.SetTitleBar(ui.TitleBar{Height: MenuBarHeight, Right: 138})
	tt.Frame()

	want := halfSplit(700 - toolbarHeight - MenuBarHeight - statusBarHeight)
	if math.Abs(float64(a.splitHeight-want)) > 0.5 {
		t.Fatalf("split height %v, want %v", a.splitHeight, want)
	}
}

// The menu bar shows the application menu's top-level menus, without the
// app menu role, which Windows ignores.
func TestTopMenusSkipsRoles(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	menus := topMenus(a.menuItems())
	if len(menus) == 0 {
		t.Fatal("the application menu has no top-level menus")
	}
	for _, it := range menus {
		if it.Role != "" {
			t.Errorf("%q is a role, which the window's menu bar has no room for", it.Label)
		}
		if len(it.Submenu) == 0 {
			t.Errorf("%q has no items", it.Label)
		}
	}
	if got := menus[0].Label; got != locales.MenuFile() {
		t.Fatalf("first menu %q, want %q", got, locales.MenuFile())
	}
}
