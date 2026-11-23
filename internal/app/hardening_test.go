package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func TestLargeFileOpensReadOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "many-lines.txt")
	data := bytes.Repeat([]byte("a\n"), 600_000)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	if !doc.ReadOnly {
		t.Fatal("a file over 500,000 lines is not read-only")
	}
	a := New(settings.Default(), dir)
	a.addDocument(doc)
	if !a.activeTab().state.ReadOnly {
		t.Fatal("the tab is not read-only")
	}
	tt := ui.NewTester(a.View, 900, 700)
	if !tt.HasText(locales.ReadOnly()) {
		t.Fatalf("status bar does not show read-only; texts %q", tt.Texts())
	}
}

// TestLargeFileWithoutHighlight opens a code file too long to highlight:
// the buffer keeps its text as plain text, while the status bar still names
// the language and says it is not highlighted.
func TestLargeFileWithoutHighlight(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "many.go")
	if err := os.WriteFile(path, bytes.Repeat([]byte("x\n"), 600_000), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Language == nil || doc.Language.Name != "Go" {
		t.Fatalf("document language %v, want Go", doc.Language)
	}
	if doc.Buffer.Language() != nil {
		t.Fatal("a file over 500,000 lines is highlighted")
	}
	a := New(settings.Default(), dir)
	a.addDocument(doc)
	tt := ui.NewTester(a.View, 900, 700)
	if want := locales.LanguageNoHighlight("Go"); !tt.HasText(want) {
		t.Fatalf("status bar does not show %q; texts %q", want, tt.Texts())
	}
}

// TestSmallFileKeepsHighlight keeps the regression the other way: a file
// below the limit still gets its language.
func TestSmallFileKeepsHighlight(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Buffer.Language() == nil {
		t.Fatal("a short Go file is not highlighted")
	}
	if got := languageLabel(doc); got != "Go" {
		t.Fatalf("status bar language %q, want Go", got)
	}
}

// TestNoteFrameAllocatesNothing guards the per-frame cost of the frame log:
// the hot path must not allocate, and must not write unless a frame ran
// slow.
func TestNoteFrameAllocatesNothing(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() { NoteFrame(time.Millisecond) })
	if allocs > 0 {
		t.Fatalf("NoteFrame allocates %v per frame, want 0", allocs)
	}
}

func TestLargeFileBySizeReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "big.txt")
	data := bytes.Repeat([]byte("a"), 1<<20+1)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !doc.ReadOnly {
		t.Fatal("a file over the size limit is not read-only")
	}
}

func TestSaveToReadOnlyDirectoryFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	doc.Buffer.Set("two")
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755)
	if _, err := writeDocument(doc, path); err == nil {
		t.Fatal("saving into a read-only directory succeeded")
	}
	if doc.Buffer.String() != "two" {
		t.Fatal("the buffer changed after a failed save")
	}
}

func TestReloadExternal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(path, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.reloadExternal(doc)
	if doc.Buffer.String() != "two" {
		t.Fatalf("buffer %q after reload", doc.Buffer.String())
	}
	if doc.Buffer.Modified() {
		t.Fatal("reloaded document is dirty")
	}
	if doc.Size != 3 {
		t.Fatalf("size %d after reload", doc.Size)
	}
}

func TestHandleOpenFileBeforeWindow(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.HandleOpenFile("/tmp/a.txt")
	a.HandleOpenFile("/tmp/b.txt")
	if len(a.pendingOpen) != 2 {
		t.Fatalf("pending files %v", a.pendingOpen)
	}
}
