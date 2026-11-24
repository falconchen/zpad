package draft

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func write(text string) func(io.Writer) error {
	return func(w io.Writer) error {
		_, err := io.WriteString(w, text)
		return err
	}
}

func TestSaveReadReopenAndRemove(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	id := NewID()
	entry := Entry{ID: id, Path: "/tmp/notes.md", Encoding: "gb18030", Size: 12, ModTime: 99}
	if ok, err := s.Save(entry, write("one\r\ntwo\r\n"), nil); err != nil || !ok {
		t.Fatalf("save = %v, %v", ok, err)
	}
	if got, err := s.Text(id); err != nil || got != "one\r\ntwo\r\n" {
		t.Fatalf("text = %q, %v", got, err)
	}

	// The index survives a restart.
	again, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	entries := again.Entries()
	if len(entries) != 1 {
		t.Fatalf("entries %d after reopening, want 1", len(entries))
	}
	if entries[0].Path != entry.Path || entries[0].Encoding != entry.Encoding ||
		entries[0].Size != entry.Size || entries[0].ModTime != entry.ModTime {
		t.Fatalf("entry %+v, want %+v", entries[0], entry)
	}

	// Saving the same id replaces the text.
	if _, err := again.Save(entry, write("new"), nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := again.Text(id); got != "new" {
		t.Fatalf("text %q after rewriting, want new", got)
	}

	if err := again.Remove(id); err != nil {
		t.Fatal(err)
	}
	if len(again.Entries()) != 0 {
		t.Fatal("an entry stayed after Remove")
	}
	if _, err := os.Stat(filepath.Join(dir, id)); !os.IsNotExist(err) {
		t.Fatalf("the copy's file stayed: %v", err)
	}
	if _, err := again.Text(id); err == nil {
		t.Fatal("the copy's text could still be read")
	}
}

func TestSaveSkipsWhenNoLongerWanted(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id := NewID()
	ok, err := s.Save(Entry{ID: id}, write("text"), func() bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("a copy nobody wants was written")
	}
	if len(s.Entries()) != 0 {
		t.Fatal("a skipped copy was indexed")
	}
}

func TestSetPathFollowsARename(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id := NewID()
	if _, err := s.Save(Entry{ID: id, Path: "/a/old.txt"}, write("text"), nil); err != nil {
		t.Fatal(err)
	}
	if err := s.SetPath(id, "/a/new.txt"); err != nil {
		t.Fatal(err)
	}
	if got := s.Entries()[0].Path; got != "/a/new.txt" {
		t.Fatalf("path %q after rename, want /a/new.txt", got)
	}
}

func TestPurgeRemovesEverything(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := s.Save(Entry{ID: NewID()}, write("text"), nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Purge(); err != nil {
		t.Fatal(err)
	}
	if len(s.Entries()) != 0 {
		t.Fatal("entries stayed after Purge")
	}
	names, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("%d files stayed in the draft directory after Purge", len(names))
	}
}

func TestCorruptIndexIsKeptBeside(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, indexName), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Entries()) != 0 {
		t.Fatal("a corrupt index produced entries")
	}
	if _, err := os.Stat(filepath.Join(dir, indexName+".bak")); err != nil {
		t.Fatalf("the corrupt index was not kept: %v", err)
	}
}

func TestNewIDIsAFilename(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id := NewID()
		if !ValidID(id) {
			t.Fatalf("NewID returned %q, which is not a usable file name", id)
		}
		if seen[id] {
			t.Fatalf("NewID repeated %q", id)
		}
		seen[id] = true
	}
}
