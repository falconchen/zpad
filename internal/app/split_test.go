package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/internal/session"
	"github.com/helloxz/zpad/internal/settings"
)

func docAt(path, text string) *Document {
	buf := editor.NewBuffer(text)
	buf.SetLanguage(editor.LanguageFor(path))
	return &Document{
		ID:       path,
		Path:     path,
		Buffer:   buf,
		Encoding: encoding.UTF8,
		Size:     int64(len(text)),
	}
}

func TestToggleSplitDuplicatesAndMerges(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "hello")
	a.addDocument(doc)

	a.toggleSplit()
	if !a.split || a.activePane != 1 {
		t.Fatalf("split %v, active pane %d", a.split, a.activePane)
	}
	if len(a.left.tabs) != 1 || len(a.right.tabs) != 1 {
		t.Fatalf("left %d, right %d", len(a.left.tabs), len(a.right.tabs))
	}
	if a.right.tabs[0].doc != a.left.tabs[0].doc {
		t.Fatal("the split does not share the document")
	}

	a.toggleSplit()
	if a.split || a.activePane != 0 {
		t.Fatalf("split %v, active pane %d", a.split, a.activePane)
	}
	if len(a.left.tabs) != 1 || len(a.right.tabs) != 0 {
		t.Fatalf("after merge: left %d, right %d", len(a.left.tabs), len(a.right.tabs))
	}
}

func TestSameFileInBothPanesSharesBuffer(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	bDoc := docAt("/tmp/b.txt", "b")
	a.addDocument(aDoc)
	a.addDocument(bDoc)

	a.toggleSplit() // right duplicates b
	a.activePane = 1
	// Open a.txt in the right pane; the document already open on the left
	// must be reused, so both tabs share one buffer.
	a.addDocument(docAt("/tmp/a.txt", "other content"))
	if len(a.right.tabs) != 2 {
		t.Fatalf("right tabs %d, want 2", len(a.right.tabs))
	}
	var rightA *Tab
	for _, tab := range a.right.tabs {
		if tab.doc.Path == "/tmp/a.txt" {
			rightA = tab
		}
	}
	if rightA == nil || rightA.doc != aDoc {
		t.Fatal("right pane did not reuse the document from the left")
	}

	aDoc.Buffer.Set("changed")
	if rightA.doc.Buffer.String() != "changed" {
		t.Fatal("the panes do not share the buffer")
	}
}

func TestDirtyDocsAreUnique(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	bDoc := docAt("/tmp/b.txt", "b")
	a.addDocument(aDoc)
	a.addDocument(bDoc)
	a.toggleSplit()
	a.activePane = 1
	a.addDocument(docAt("/tmp/a.txt", "again")) // reuses aDoc

	aDoc.Buffer.Set("dirty")
	if got := a.dirtyDocs(); len(got) != 1 || got[0] != aDoc {
		t.Fatalf("dirty docs %v", got)
	}
	bDoc.Buffer.Set("dirty too")
	if got := a.dirtyDocs(); len(got) != 2 {
		t.Fatalf("dirty docs %d, want 2", len(got))
	}
}

func TestRemoveLastTabOfActivePaneSwitches(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.toggleSplit()
	if a.activePane != 1 || len(a.right.tabs) != 1 {
		t.Fatal("split setup failed")
	}
	a.removeTab(1, 0)
	if len(a.right.tabs) != 0 {
		t.Fatal("right tab not removed")
	}
	if a.split {
		t.Fatal("split stayed open with an empty pane")
	}
	if len(a.left.tabs) != 1 {
		t.Fatalf("left tabs %d, want 1", len(a.left.tabs))
	}
	if a.activePane != 0 {
		t.Fatalf("active pane %d, want 0", a.activePane)
	}
}

func TestSessionRoundTrip(t *testing.T) {
	workspace := t.TempDir()
	pathA := filepath.Join(workspace, "a.txt")
	pathB := filepath.Join(workspace, "b.txt")
	if err := os.WriteFile(pathA, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pathB, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	settingsDir := t.TempDir()

	a := New(settings.Default(), settingsDir)
	a.setWorkspace(workspace)
	a.addDocument(docAt(pathA, "a"))
	a.addDocument(docAt(pathB, "b"))
	a.toggleSplit() // right duplicates b
	a.activePane = 1
	a.addDocument(docAt(pathA, "a")) // shares a with the left
	a.splitWidth = 720
	a.left.active = 0
	a.right.active = 1
	a.saveSession()

	bSettings := settings.Default()
	bSettings.RestoreSession = true
	b := New(bSettings, settingsDir)
	b.RestoreSession()
	if b.root != workspace || !b.split || b.splitWidth != 720 || b.activePane != 1 {
		t.Fatalf("restored root %q, split %v, width %v, active %d", b.root, b.split, b.splitWidth, b.activePane)
	}
	if len(b.left.tabs) != 2 || len(b.right.tabs) != 2 {
		t.Fatalf("restored tabs left %d, right %d", len(b.left.tabs), len(b.right.tabs))
	}
	var leftA, rightA *Tab
	for _, tab := range b.left.tabs {
		if tab.doc.Path == pathA {
			leftA = tab
		}
	}
	for _, tab := range b.right.tabs {
		if tab.doc.Path == pathA {
			rightA = tab
		}
	}
	if leftA == nil || rightA == nil || leftA.doc != rightA.doc {
		t.Fatal("restored panes do not share the document")
	}
}

func TestPendingScrollApplied(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/scroll.txt", strings.Repeat("line\n", 300))
	a.addDocument(doc)
	y := 120.0
	a.left.tabs[0].pendingScroll = &y

	tt := ui.NewTester(a.View, 800, 400)
	tt.Frame()
	if a.left.tabs[0].pendingScroll != nil {
		t.Fatal("pending scroll was not applied")
	}
	if got := a.left.tabs[0].state.ScrollY(); got < 100 {
		t.Fatalf("scroll %v, want about 120", got)
	}
}

func TestCloseWithoutDirtySavesSession(t *testing.T) {
	dir := t.TempDir()
	a := New(settings.Default(), dir)
	a.BeforeQuit(&mygo.QuitEvent{})
	if !session.Exists(dir) {
		t.Fatal("quit without changes did not save the session")
	}
	a.onWindowClose(&mygo.CloseEvent{})
}

func TestMissingSessionFileIsSkipped(t *testing.T) {
	dir := t.TempDir()
	st := session.Default()
	st.Root = dir
	st.Left = session.PaneState{
		Tabs:   []session.TabState{{Path: filepath.Join(dir, "missing.txt")}},
		Active: 0,
	}
	if err := session.Save(dir, st); err != nil {
		t.Fatal(err)
	}
	aSettings := settings.Default()
	aSettings.RestoreSession = true
	a := New(aSettings, dir)
	a.RestoreSession()
	if len(a.left.tabs) != 0 {
		t.Fatalf("restored %d missing tabs", len(a.left.tabs))
	}
	if a.errMsg == "" {
		t.Fatal("missing file was not reported")
	}
}

func TestSessionRestoreDisabled(t *testing.T) {
	settingsDir := t.TempDir()
	a := New(settings.Default(), settingsDir)
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.saveSession()

	s := settings.Default()
	s.RestoreSession = false
	b := New(s, settingsDir)
	b.RestoreSession()
	if len(b.left.tabs) != 0 {
		t.Fatalf("restored %d tabs", len(b.left.tabs))
	}
}
