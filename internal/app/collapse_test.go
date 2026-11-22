package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helloxz/zpad/internal/session"
	"github.com/helloxz/zpad/internal/settings"
)

func TestClosingLastTabCollapsesLeftPane(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	bDoc := docAt("/tmp/b.txt", "b")
	a.addDocument(aDoc)
	a.toggleSplit() // right duplicates a
	a.addDocument(bDoc)

	a.removeTab(0, 0) // the left pane loses its only tab
	if a.split {
		t.Fatal("split stayed open with an empty left pane")
	}
	if len(a.left.tabs) != 2 || a.left.tabs[0].doc != aDoc || a.left.tabs[1].doc != bDoc {
		t.Fatalf("left panes did not take the right tabs: %d", len(a.left.tabs))
	}
	if len(a.right.tabs) != 0 {
		t.Fatalf("right pane kept %d tabs", len(a.right.tabs))
	}
	if a.activePane != 0 || a.left.active != 1 {
		t.Fatalf("active pane %d, tab %d; want the right pane's b", a.activePane, a.left.active)
	}
}

func TestCloseLastTabWithoutSplitShowsEmptyState(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.removeTab(0, 0)
	if a.split {
		t.Fatal("a single pane became a split")
	}
	if len(a.left.tabs) != 0 || a.left.active != -1 {
		t.Fatalf("left tabs %d, active %d; want the empty state", len(a.left.tabs), a.left.active)
	}
}

func TestRestoreCollapsesEmptyPane(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	st := session.Default()
	st.Root = dir
	st.Split = true
	st.Left = session.PaneState{Tabs: []session.TabState{{Path: path}}, Active: 0}
	st.Right = session.PaneState{Active: -1}
	if err := session.Save(dir, st); err != nil {
		t.Fatal(err)
	}

	s := settings.Default()
	s.RestoreSession = true
	a := New(s, dir)
	a.RestoreSession()
	if a.split {
		t.Fatal("restored a split with an empty pane")
	}
	if len(a.left.tabs) != 1 || len(a.right.tabs) != 0 {
		t.Fatalf("left %d tabs, right %d", len(a.left.tabs), len(a.right.tabs))
	}
	if a.activePane != 0 || a.left.active != 0 {
		t.Fatalf("active pane %d, tab %d", a.activePane, a.left.active)
	}
}
