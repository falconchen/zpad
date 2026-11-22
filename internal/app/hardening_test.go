package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

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
