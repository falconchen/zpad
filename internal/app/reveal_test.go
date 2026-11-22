package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// stubShowItemInFolder records the paths the file manager would open in the
// order they were asked for.
func stubShowItemInFolder(t *testing.T, calls *[]string) {
	t.Helper()
	old := showItemInFolder
	showItemInFolder = func(path string) { *calls = append(*calls, path) }
	t.Cleanup(func() { showItemInFolder = old })
}

func TestShowActiveInFileManagerRevealsTheActiveFile(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/reveal.txt", "x"))
	var calls []string
	stubShowItemInFolder(t, &calls)

	a.ShowActiveInFileManager()

	if len(calls) != 1 || calls[0] != "/tmp/reveal.txt" {
		t.Fatalf("revealed %q, want [/tmp/reveal.txt]", calls)
	}
}

func TestShowActiveInFileManagerSkipsAnUntitledDocument(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(a.newUntitledDocument())
	var calls []string
	stubShowItemInFolder(t, &calls)

	if a.canRevealActive() {
		t.Fatal("an untitled document has no path to reveal")
	}
	a.ShowActiveInFileManager()
	if len(calls) != 0 {
		t.Fatalf("revealed %q, want nothing", calls)
	}
}

func TestShowActiveInFileManagerSkipsAnEmptyWindow(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	var calls []string
	stubShowItemInFolder(t, &calls)

	if a.canRevealActive() {
		t.Fatal("a window with no tabs has nothing to reveal")
	}
	a.ShowActiveInFileManager()
	if len(calls) != 0 {
		t.Fatalf("revealed %q, want nothing", calls)
	}
}

// The active pane decides which file the button reveals, so a split with a
// second file on the right reveals that one.
func TestShowActiveInFileManagerUsesTheActivePane(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/left.txt", "x"))
	a.toggleSplit()
	a.addDocument(docAt("/tmp/right.txt", "y"))
	var calls []string
	stubShowItemInFolder(t, &calls)

	a.ShowActiveInFileManager()

	if a.activePane != 1 {
		t.Fatalf("active pane %d, want 1", a.activePane)
	}
	if len(calls) != 1 || calls[0] != "/tmp/right.txt" {
		t.Fatalf("revealed %q, want [/tmp/right.txt]", calls)
	}
}

func TestToolbarRevealButtonFollowsTheActiveTab(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()

	info, ok := tt.Access(locales.RevealInFileManager())
	if !ok {
		t.Fatalf("no reveal button in the toolbar; access %v", tt.AccessAll())
	}
	if !info.Disabled {
		t.Fatal("the reveal button is enabled with no tab open")
	}

	a.addDocument(docAt("/tmp/reveal.txt", "x"))
	tt.Frame()
	info, ok = tt.Access(locales.RevealInFileManager())
	if !ok || info.Disabled {
		t.Fatalf("the reveal button is not enabled for a saved file: %+v, found %v", info, ok)
	}

	var calls []string
	stubShowItemInFolder(t, &calls)
	if err := tt.Click(locales.RevealInFileManager()); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "/tmp/reveal.txt" {
		t.Fatalf("the button revealed %q, want [/tmp/reveal.txt]", calls)
	}

	a.NewFile()
	tt.Frame()
	info, ok = tt.Access(locales.RevealInFileManager())
	if !ok || !info.Disabled {
		t.Fatalf("the reveal button is enabled for an untitled tab: %+v, found %v", info, ok)
	}
}
