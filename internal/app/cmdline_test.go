package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helloxz/zpad/internal/settings"
)

func TestCommandLinePathsSkipsFlags(t *testing.T) {
	got := commandLinePaths([]string{"--inspect", "notes.md", "", "-psn_0_1", "/abs/a.txt"}, "")
	want := []string{"notes.md", "/abs/a.txt"}
	if len(got) != len(want) {
		t.Fatalf("paths %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("paths %v, want %v", got, want)
		}
	}
}

func TestCommandLinePathsJoinsTheWorkingDirectory(t *testing.T) {
	got := commandLinePaths([]string{"notes.md", filepath.Join(string(filepath.Separator), "abs", "a.txt")}, "/work")
	want := []string{filepath.Join("/work", "notes.md"), filepath.Join(string(filepath.Separator), "abs", "a.txt")}
	if len(got) != len(want) {
		t.Fatalf("paths %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("paths %v, want %v", got, want)
		}
	}
}

// A second instance and the system may ask for one file twice: the second
// request waits for the read instead of starting another one.
func TestOpenPathsSkipsAFileBeingRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(path, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	if !a.beginLoading(path) {
		t.Fatal("the first claim of a path was refused")
	}
	a.OpenPaths([]string{path})
	if n := len(a.left.tabs); n != 0 {
		t.Fatalf("opened %d tabs for a file being read", n)
	}
	a.endLoading(path)
	if !a.beginLoading(path) {
		t.Fatal("a path was not released once it was read")
	}
}

func TestHandleSecondInstanceBeforeWindowKeepsPaths(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.HandleSecondInstance([]string{"-x", "/tmp/a.txt"}, "")
	if len(a.pendingOpen) != 1 || a.pendingOpen[0] != "/tmp/a.txt" {
		t.Fatalf("pending files %v", a.pendingOpen)
	}
}
