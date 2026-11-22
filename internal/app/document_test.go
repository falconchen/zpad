package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/locales"
)

func TestLoadAndSaveKeepsEncoding(t *testing.T) {
	for _, kind := range encoding.All {
		dir := t.TempDir()
		path := filepath.Join(dir, "note.txt")
		want := "第一行\nsecond line\n"
		data, err := encoding.Encode(want, kind)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}

		doc, err := loadDocument(path, 20)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if got := doc.Buffer.String(); got != want {
			t.Fatalf("%s: text %q, want %q", kind, got, want)
		}
		if doc.Encoding != kind {
			t.Fatalf("%s: encoding %s", kind, doc.Encoding)
		}

		doc.Buffer.Set("更新后的内容\nnew text\n")
		saved, err := writeDocument(doc, path)
		if err != nil {
			t.Fatalf("%s: save: %v", kind, err)
		}
		back, err := encoding.Decode(saved, kind)
		if err != nil {
			t.Fatalf("%s: decode saved: %v", kind, err)
		}
		if back != "更新后的内容\nnew text\n" {
			t.Fatalf("%s: saved text %q", kind, back)
		}
	}
}

func TestLineEndingsPreservedAndConverted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "crlf.txt")
	if err := os.WriteFile(path, []byte("a\r\nb\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Buffer.LineEnding(); got != "\r\n" {
		t.Fatalf("line ending %q, want CRLF", got)
	}
	data, err := writeDocument(doc, path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "a\r\nb\r\n" {
		t.Fatalf("saved bytes %q", data)
	}

	doc.Buffer.SetLineEnding("\n")
	data, err = writeDocument(doc, path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "a\nb\n" {
		t.Fatalf("converted bytes %q", data)
	}
}

func TestLoadBinaryRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(path, []byte{0x89, 'P', 'N', 'G', 0, 0, 0}, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDocument(path, 20); err == nil {
		t.Fatal("binary file loaded")
	}
}

func TestDisplayPath(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "workspace")
	doc := &Document{Path: filepath.Join(root, "src", "main.go")}
	if got := doc.DisplayPath(root); got != filepath.Join("src", "main.go") {
		t.Fatalf("relative path %q", got)
	}
	outside := &Document{Path: filepath.Join(string(filepath.Separator), "elsewhere", "a.txt")}
	if got := outside.DisplayPath(root); got != outside.Path {
		t.Fatalf("outside path %q", got)
	}
}

func TestNewUntitled(t *testing.T) {
	doc := newUntitled("untitled-1", "gb18030")
	if doc.Name() != locales.Untitled() || doc.Encoding != encoding.GB18030 {
		t.Fatalf("untitled %+v", doc)
	}
	if doc.Buffer.Modified() {
		t.Fatal("new untitled document is modified")
	}
}
