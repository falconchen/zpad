package app

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func TestOpenSplitSwitchesDirection(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.addDocument(docAt("/tmp/b.txt", "b"))

	a.openSplit(false) // right duplicates b
	if !a.split || a.splitVertical {
		t.Fatalf("split %v, vertical %v", a.split, a.splitVertical)
	}
	if len(a.left.tabs) != 2 || len(a.right.tabs) != 1 {
		t.Fatalf("left %d, right %d", len(a.left.tabs), len(a.right.tabs))
	}
	rightDoc := a.right.tabs[0].doc
	rightState := a.right.tabs[0].state

	// Changing direction keeps both panes, their tabs and their states.
	a.openSplit(true)
	if !a.split || !a.splitVertical {
		t.Fatalf("split %v, vertical %v", a.split, a.splitVertical)
	}
	if len(a.left.tabs) != 2 || len(a.right.tabs) != 1 {
		t.Fatalf("after switching: left %d, right %d", len(a.left.tabs), len(a.right.tabs))
	}
	if a.right.tabs[0].doc != rightDoc || a.right.tabs[0].state != rightState {
		t.Fatal("switching direction lost the second pane's tab")
	}
	if !a.centerSplit {
		t.Fatal("switching direction did not center the new axis")
	}

	a.closeSplit()
	if a.split || len(a.right.tabs) != 0 {
		t.Fatalf("split %v, right tabs %d after closing", a.split, len(a.right.tabs))
	}
	if len(a.left.tabs) != 2 {
		t.Fatalf("left tabs %d, want 2", len(a.left.tabs))
	}
}

func TestToggleSplitKeepsDirection(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.openSplit(true)

	a.toggleSplit() // closes
	if a.split {
		t.Fatal("the split stayed open")
	}
	a.toggleSplit() // opens again, top and bottom
	if !a.split || !a.splitVertical {
		t.Fatalf("split %v, vertical %v", a.split, a.splitVertical)
	}
}

func TestSplitVerticalCentersInContent(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.Settings.SidebarOpen = false
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.openSplit(true)

	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()
	want := halfSplit(700 - toolbarHeight - statusBarHeight)
	if math.Abs(float64(a.splitHeight-want)) > 0.5 {
		t.Fatalf("split height %v, want %v", a.splitHeight, want)
	}
}

func TestSplitVerticalStacksPanes(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.Settings.SidebarOpen = false
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.openSplit(true) // left: a, right: a
	a.activePane = 1
	a.addDocument(docAt("/tmp/b.txt", "b")) // right: a and b, b active
	a.removeTab(1, 0)                       // right: only b; left: only a

	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()

	top, ok := tt.Find("a.txt")
	if !ok {
		t.Fatalf("no a.txt tab; texts %q", tt.Texts())
	}
	bottom, ok := tt.Find("b.txt")
	if !ok {
		t.Fatalf("no b.txt tab; texts %q", tt.Texts())
	}
	if bottom.Y < top.Y+100 {
		t.Fatalf("the second pane is not below the first: top %v, bottom %v", top, bottom)
	}
	if top.X > 100 || bottom.X > 100 {
		t.Fatalf("panes are not full width: top %v, bottom %v", top, bottom)
	}
}

func TestSessionRestoresVerticalSplit(t *testing.T) {
	workspace := t.TempDir()
	pathA := filepath.Join(workspace, "a.txt")
	pathB := filepath.Join(workspace, "b.txt")
	for _, path := range []string{pathA, pathB} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	settingsDir := t.TempDir()

	a := New(settings.Default(), settingsDir)
	a.setWorkspace(workspace)
	a.addDocument(docAt(pathA, "a"))
	a.addDocument(docAt(pathB, "b"))
	a.openSplit(true)
	a.splitHeight = 420
	a.saveSession()

	bSettings := settings.Default()
	bSettings.RestoreSession = true
	b := New(bSettings, settingsDir)
	b.RestoreSession()
	if !b.split || !b.splitVertical {
		t.Fatalf("restored split %v, vertical %v", b.split, b.splitVertical)
	}
	if b.splitHeight != 420 {
		t.Fatalf("restored split height %v, want 420", b.splitHeight)
	}
	if len(b.left.tabs) != 2 || len(b.right.tabs) != 1 {
		t.Fatalf("restored tabs left %d, right %d", len(b.left.tabs), len(b.right.tabs))
	}
}

func TestPaneLabelFollowsDirection(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	if got, want := a.paneLabel(0), locales.LeftPane(); got != want {
		t.Fatalf("left label %q, want %q", got, want)
	}
	a.splitVertical = true
	if got, want := a.paneLabel(0), locales.TopPane(); got != want {
		t.Fatalf("top label %q, want %q", got, want)
	}
	if got, want := a.paneLabel(1), locales.BottomPane(); got != want {
		t.Fatalf("bottom label %q, want %q", got, want)
	}
}

func TestSplitButtonMenuChoosesDirection(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()

	if err := tt.Click(locales.SplitMenu()); err != nil {
		t.Fatal(err)
	}
	menu := tt.Menu()
	for _, want := range []string{locales.SplitLeftRight(), locales.SplitTopBottom(), locales.CloseSplit()} {
		found := false
		for _, item := range menu {
			if item == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("split menu %q has no %q item", menu, want)
		}
	}
	if err := tt.Click(locales.SplitTopBottom()); err != nil {
		t.Fatal(err)
	}
	if !a.split || !a.splitVertical {
		t.Fatalf("choosing top and bottom gave split %v, vertical %v", a.split, a.splitVertical)
	}
}

func TestSplitDirectionMenuItems(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	var direction *mygo.MenuItem
	var walk func(items []*mygo.MenuItem)
	walk = func(items []*mygo.MenuItem) {
		for _, it := range items {
			if it.Label == locales.SplitDirection() {
				direction = it
			}
			walk(it.Submenu)
		}
	}
	walk(a.menuItems())
	if direction == nil || len(direction.Submenu) != 2 {
		t.Fatalf("split direction submenu %v", direction)
	}
	for _, it := range direction.Submenu {
		if it.Label == locales.SplitTopBottom() {
			it.Click(it, nil)
		}
	}
	if !a.split || !a.splitVertical {
		t.Fatalf("menu click gave split %v, vertical %v", a.split, a.splitVertical)
	}
}
