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

func TestStatusBarRevealButtonFollowsTheActiveTab(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()

	if info, ok := tt.Access(locales.RevealInFileManager()); ok {
		t.Fatalf("the status bar shows the reveal button with no tab open: %+v", info)
	}

	a.addDocument(docAt("/tmp/reveal.txt", "x"))
	tt.Frame()
	info, ok := tt.Access(locales.RevealInFileManager())
	if !ok || info.Disabled {
		t.Fatalf("the status bar has no reveal button for a saved file: %+v, found %v", info, ok)
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
	if info, ok := tt.Access(locales.RevealInFileManager()); ok {
		t.Fatalf("the status bar shows the reveal button for an untitled tab: %+v", info)
	}
}

func TestTabContextMenuRevealsSavedFile(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/reveal.txt", "x"))
	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	if err := tt.RightClick("reveal.txt"); err != nil {
		t.Fatal(err)
	}
	if !menuHas(tt.Menu(), locales.RevealInFileManager()) {
		t.Fatalf("tab menu %q does not show %q", tt.Menu(), locales.RevealInFileManager())
	}

	var calls []string
	stubShowItemInFolder(t, &calls)
	if err := tt.ChooseMenuItem(locales.RevealInFileManager()); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "/tmp/reveal.txt" {
		t.Fatalf("the menu revealed %q, want [/tmp/reveal.txt]", calls)
	}
}
