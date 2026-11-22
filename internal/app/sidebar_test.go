package app

import (
	"path/filepath"
	"testing"

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
