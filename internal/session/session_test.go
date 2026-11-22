package session

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadMissingUsesDefaults(t *testing.T) {
	got, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if want := Default(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := State{
		Root:         "/Users/me/project",
		Split:        true,
		SplitWidth:   720,
		ActivePane:   1,
		Left:         PaneState{Tabs: []TabState{{Path: "/a.md", ScrollY: 12}, {Path: "/b.json"}}, Active: 1},
		Right:        PaneState{Tabs: []TabState{{Path: "/c.css", ScrollY: 4}}, Active: 0},
		SidebarOpen:  true,
		SidebarWidth: 260,
	}
	if err := Save(dir, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestCorruptFileIsBackedUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, fileName)
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := Default(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want defaults %+v", got, want)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("backup not written: %v", err)
	}
}

func TestNormalizeActiveTab(t *testing.T) {
	dir := t.TempDir()
	in := Default()
	in.Left.Active = 9
	in.Right.Active = -3
	if err := Save(dir, in); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Left.Active != 0 || got.Right.Active != 0 {
		t.Fatalf("active tabs %d, %d; want 0, 0", got.Left.Active, got.Right.Active)
	}
}
