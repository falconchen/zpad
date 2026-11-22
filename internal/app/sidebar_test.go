package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func TestOpenDocumentsAreUniqueAndOrdered(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	bDoc := docAt("/tmp/b.txt", "b")
	cDoc := docAt("/tmp/c.txt", "c")
	a.addDocument(aDoc)
	a.addDocument(bDoc)
	a.toggleSplit() // right duplicates b
	a.addDocument(cDoc)

	got := a.openDocuments()
	if len(got) != 3 || got[0] != aDoc || got[1] != bDoc || got[2] != cDoc {
		t.Fatalf("open documents %v, want a, b, c once each", got)
	}
}

func TestActivateDocumentChoosesPane(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	bDoc := docAt("/tmp/b.txt", "b")
	a.addDocument(aDoc)
	a.toggleSplit() // right duplicates a; the right pane is active
	a.addDocument(bDoc)

	a.activateDocument(aDoc)
	if a.activePane != 1 || a.right.active != 0 {
		t.Fatalf("active pane %d, tab %d; want the right pane's a", a.activePane, a.right.active)
	}

	a.setActivePane(0)
	a.activateDocument(bDoc)
	if a.activePane != 1 || a.right.tabs[a.right.active].doc != bDoc {
		t.Fatalf("active pane %d, tab %d; want the right pane's b", a.activePane, a.right.active)
	}
}

func TestDocumentDir(t *testing.T) {
	dir := t.TempDir()
	a := New(settings.Default(), t.TempDir())
	a.root = dir
	inside := docAt(filepath.Join(dir, "docs", "a.txt"), "a")
	if got := a.documentDir(inside); got != "docs" {
		t.Fatalf("dir of a file under the workspace %q, want docs", got)
	}
	atRoot := docAt(filepath.Join(dir, "b.txt"), "b")
	if got := a.documentDir(atRoot); got != "" {
		t.Fatalf("dir of a file at the workspace root %q, want none", got)
	}
	outside := docAt(filepath.Join(t.TempDir(), "c.txt"), "c")
	if got := a.documentDir(outside); got != filepath.Dir(outside.Path) {
		t.Fatalf("dir of a file outside the workspace %q, want its own", got)
	}
	if got := a.documentDir(docAt("", "untitled")); got != "" {
		t.Fatalf("dir of an untitled document %q, want none", got)
	}
}

func TestSidebarTabs(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.Settings.SidebarOpen = true
	a.addDocument(docAt("/tmp/a.txt", "a"))

	tt := ui.NewTester(a.View, 900, 600)
	for _, want := range []string{locales.OpenFiles(), locales.Workspace(), "a.txt"} {
		if !tt.HasText(want) {
			t.Errorf("sidebar does not show %q; texts %q", want, tt.Texts())
		}
	}
	if err := tt.Click(locales.Workspace()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText(locales.NoFolder()) {
		t.Errorf("workspace tab shows no empty state; texts %q", tt.Texts())
	}
	if err := tt.Click(locales.OpenFiles()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("a.txt") {
		t.Errorf("open files tab does not show the file; texts %q", tt.Texts())
	}
}

func TestSidebarEmptyOpenFiles(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.Settings.SidebarOpen = true
	tt := ui.NewTester(a.View, 900, 600)
	if !tt.HasText(locales.NoOpenFiles()) {
		t.Fatalf("sidebar does not show the empty state; texts %q", tt.Texts())
	}
}

func TestWorkspaceTabOnOpenFolder(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(t.TempDir())
	if a.sidebarTab != sidebarWorkspace {
		t.Fatalf("sidebar tab %d, want the workspace", a.sidebarTab)
	}
	if !a.Settings.SidebarOpen {
		t.Fatal("sidebar did not open")
	}
}

func TestTreeClickSelectsRow(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.Settings.SidebarOpen = true

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	if err := tt.Click("docs"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.treeSelected < 0 || a.treeSelected >= a.treeState.Rows() || a.treeState.Item(a.treeSelected) != sub {
		t.Fatalf("tree selection %d is not the docs folder", a.treeSelected)
	}
}

func TestTreeFileClickSelectsRow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.Settings.SidebarOpen = true

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	row, ok := tt.Find("a.txt")
	if !ok {
		t.Fatal("no a.txt in the tree")
	}
	// Open the file first, at the row's place, so the click takes the
	// fast path and changes no document off the UI thread.
	a.addDocument(docAt(path, "a"))
	a.treeSelected = -1
	tt.ClickAt(row.X+row.W/2, row.Y+row.H/2)
	tt.Frame()
	if a.treeSelected < 0 || a.treeSelected >= a.treeState.Rows() || a.treeState.Item(a.treeSelected) != path {
		t.Fatalf("tree selection %d is not a.txt", a.treeSelected)
	}
}

// TestTreeRowClickAnywhereOpens clicks a tree row left of its name, on the
// indentation, and right of it, on the row's blank end: the whole row opens
// the file, not only its text, where the outline's own row would just
// choose it.
func TestTreeRowClickAnywhereOpens(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sub, "a.txt")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.Settings.SidebarOpen = true
	a.treeState.Open.Add(sub) // shows a.txt one level deep
	// Both are open already, so a click takes the fast path and changes no
	// document off the UI thread.
	doc := docAt(path, "a")
	other := docAt(filepath.Join(dir, "b.txt"), "b")
	a.addDocument(doc)
	a.addDocument(other)

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	row, ok := tt.Find("a.txt")
	if !ok {
		t.Fatalf("no a.txt in the tree; texts %q", tt.Texts())
	}
	y := row.Y + row.H/2
	for _, x := range []float32{4, 20, a.Settings.SidebarWidth - 12} {
		a.activateDocument(other)
		if tab := a.activeTab(); tab == nil || tab.doc != other {
			t.Fatal("setup: b is not the active tab")
		}
		tt.ClickAt(x, y)
		tt.Frame()
		if tab := a.activeTab(); tab == nil || tab.doc != doc {
			t.Fatalf("a click at x=%v did not open a.txt; the active tab is %v", x, a.activeTab())
		}
	}
}

// TestTreeFolderGutterClickToggles opens and closes a folder with clicks
// left of its name, where its arrow is drawn.
func TestTreeFolderGutterClickToggles(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.Settings.SidebarOpen = true

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	row, ok := tt.Find("docs")
	if !ok {
		t.Fatalf("no docs in the tree; texts %q", tt.Texts())
	}
	y := row.Y + row.H/2
	tt.ClickAt(4, y)
	tt.Frame()
	if !a.treeState.Open.Has(sub) {
		t.Fatal("a click left of the name did not open the folder")
	}
	tt.Advance(600 * time.Millisecond)
	tt.ClickAt(4, y)
	tt.Frame()
	if a.treeState.Open.Has(sub) {
		t.Fatal("a second click left of the name did not close the folder")
	}
}

// TestTreeOptionClickOpensAll opens and closes a folder with the folders
// inside it, as Option-clicking a row does.
func TestTreeOptionClickOpensAll(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	deeper := filepath.Join(sub, "notes")
	if err := os.MkdirAll(deeper, 0o755); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.Settings.SidebarOpen = true

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	row, ok := tt.Find("docs")
	if !ok {
		t.Fatalf("no docs in the tree; texts %q", tt.Texts())
	}
	tt.ClickAtWith(ui.Alt, 4, row.Y+row.H/2)
	tt.Frame()
	if !a.treeState.Open.Has(sub) || !a.treeState.Open.Has(deeper) {
		t.Fatalf("an Option-click did not open docs and notes: %v", a.treeState.Open)
	}
	tt.Advance(600 * time.Millisecond)
	if row, ok = tt.Find("docs"); !ok {
		t.Fatal("docs left the tree")
	}
	tt.ClickAtWith(ui.Alt, 4, row.Y+row.H/2)
	tt.Frame()
	if a.treeState.Open.Has(sub) || a.treeState.Open.Has(deeper) {
		t.Fatalf("a second Option-click did not close docs and notes: %v", a.treeState.Open)
	}
}
