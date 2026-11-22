package app

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
)

func TestTreeFolderOpensWithOneClick(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	if err := tt.Click("docs"); err != nil {
		t.Fatal(err)
	}
	if !a.treeState.Open.Has(sub) {
		t.Fatal("one click on a folder did not expand it")
	}
	if err := tt.Click("docs"); err != nil {
		t.Fatal(err)
	}
	if a.treeState.Open.Has(sub) {
		t.Fatal("a second click on a folder did not close it")
	}
}

func TestHalfSplit(t *testing.T) {
	for _, tc := range []struct{ room, want float32 }{
		{room: 100, want: 49.5},
		{room: 101, want: 50},
		{room: 759, want: 379},
		{room: 1, want: 0.5},
	} {
		if got := halfSplit(tc.room); got != tc.want {
			t.Errorf("halfSplit(%v) = %v, want %v", tc.room, got, tc.want)
		}
	}
}

func TestSplitCentersInContent(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.Settings.SidebarOpen = true
	a.Settings.SidebarWidth = 240
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.toggleSplit()

	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()
	want := halfSplit(1000 - 240 - 1)
	if math.Abs(float64(a.splitWidth-want)) > 0.5 {
		t.Fatalf("split width %v, want %v", a.splitWidth, want)
	}
}

func TestSplitCentersWithoutSidebar(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.Settings.SidebarOpen = false
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.toggleSplit()

	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()
	want := halfSplit(1000)
	if math.Abs(float64(a.splitWidth-want)) > 0.5 {
		t.Fatalf("split width %v, want %v", a.splitWidth, want)
	}
}

func TestMoveTabAcrossPanes(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	bDoc := docAt("/tmp/b.txt", "b")
	a.addDocument(aDoc)
	a.addDocument(bDoc)
	a.toggleSplit() // right duplicates b
	a.left.active = 0

	a.moveTab(tabDrag{pane: 0, id: aDoc.ID}, 1)
	if len(a.left.tabs) != 1 || a.left.tabs[0].doc != bDoc {
		t.Fatalf("left tabs %d", len(a.left.tabs))
	}
	if len(a.right.tabs) != 2 || a.right.tabs[1].doc != aDoc {
		t.Fatalf("right tabs %d", len(a.right.tabs))
	}
	if a.activePane != 1 || a.right.active != 1 {
		t.Fatalf("active pane %d, tab %d", a.activePane, a.right.active)
	}
}

func TestMoveTabToExistingCopy(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	bDoc := docAt("/tmp/b.txt", "b")
	a.addDocument(aDoc)
	a.addDocument(bDoc)
	a.toggleSplit() // right duplicates b; left also still has b
	if len(a.right.tabs) != 1 || a.right.tabs[0].doc != bDoc {
		t.Fatal("split setup failed")
	}

	a.moveTab(tabDrag{pane: 0, id: bDoc.ID}, 1)
	if len(a.left.tabs) != 1 || a.left.tabs[0].doc != aDoc {
		t.Fatalf("left tabs %d", len(a.left.tabs))
	}
	if len(a.right.tabs) != 1 || a.right.tabs[0].doc != bDoc {
		t.Fatalf("right tabs %d", len(a.right.tabs))
	}
	if a.activePane != 1 || a.right.active != 0 {
		t.Fatalf("active pane %d, tab %d", a.activePane, a.right.active)
	}
}

func TestMoveLastTabCollapsesSplit(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "a")
	a.addDocument(doc)
	a.toggleSplit() // right duplicates a
	if a.activePane != 1 || len(a.right.tabs) != 1 {
		t.Fatal("split setup failed")
	}

	a.moveTab(tabDrag{pane: 1, id: doc.ID}, 0)
	if a.split {
		t.Fatal("split stayed open with an empty source pane")
	}
	if len(a.left.tabs) != 1 || a.left.tabs[0].doc != doc || len(a.right.tabs) != 0 {
		t.Fatalf("left %d tabs, right %d", len(a.left.tabs), len(a.right.tabs))
	}
	if a.activePane != 0 || a.left.active != 0 {
		t.Fatalf("active pane %d, tab %d", a.activePane, a.left.active)
	}
}

func TestMoveTabSamePaneDoesNothing(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "a")
	a.addDocument(doc)
	a.toggleSplit()
	before := len(a.right.tabs)
	a.moveTab(tabDrag{pane: 1, id: doc.ID}, 1)
	if len(a.right.tabs) != before || !a.split {
		t.Fatal("a drop on its own pane changed the tabs")
	}
}

func TestDragTabAcrossPanes(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	bDoc := docAt("/tmp/b.txt", "b")
	a.addDocument(aDoc)
	a.addDocument(bDoc)
	a.toggleSplit() // right duplicates b

	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()
	src, ok := tt.Find("a.txt")
	if !ok {
		t.Fatalf("no a.txt tab; texts %q", tt.Texts())
	}
	tt.Press(src.X+src.W/2, src.Y+src.H/2)
	tt.Move(src.X+src.W/2+20, src.Y+src.H/2)
	tt.Move(900, 400)
	tt.Release(900, 400)
	tt.Frame()

	if !a.split {
		t.Fatal("the split closed although the left pane still has a tab")
	}
	if len(a.left.tabs) != 1 || a.left.tabs[0].doc != bDoc {
		t.Fatalf("left tabs %d", len(a.left.tabs))
	}
	if len(a.right.tabs) != 2 || a.right.tabs[1].doc != aDoc {
		t.Fatalf("right tabs %d", len(a.right.tabs))
	}
	if a.activePane != 1 || a.right.active != 1 {
		t.Fatalf("active pane %d, tab %d; want the dropped tab", a.activePane, a.right.active)
	}
}
