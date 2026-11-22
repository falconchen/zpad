package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
)

func TestViewShowsSkeleton(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	tt := ui.NewTester(a.View, 1200, 800)

	if !tt.HasText("ZPad") {
		t.Fatalf("view does not show the app name; texts %q", tt.Texts())
	}
	if !tt.HasText("打开文件或文件夹开始编辑") {
		t.Fatalf("view does not show the empty state; texts %q", tt.Texts())
	}
	if !tt.HasText("ZPad 0.1.0") {
		t.Fatalf("status bar does not show the version; texts %q", tt.Texts())
	}
}

func TestEditorView(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	src := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), dir)
	a.addDocument(doc)

	tt := ui.NewTester(a.View, 1200, 800)
	for _, want := range []string{"main.go", "Go", "UTF-8", "行 1, 列 1"} {
		if !tt.HasText(want) {
			t.Errorf("view does not show %q; texts %q", want, tt.Texts())
		}
	}
}

func TestTabs(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.NewFile()
	a.NewFile()
	if len(a.left.tabs) != 2 || a.left.active != 1 {
		t.Fatalf("tabs %d, active %d", len(a.left.tabs), a.left.active)
	}
	a.removeTab(0, 1)
	if len(a.left.tabs) != 1 || a.left.active != 0 {
		t.Fatalf("after remove: tabs %d, active %d", len(a.left.tabs), a.left.active)
	}
	if a.left.tabs[0].doc.Name() != "未命名" {
		t.Fatalf("tab name %q", a.left.tabs[0].doc.Name())
	}
}

func TestWorkspaceOpensSidebar(t *testing.T) {
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(workspace)
	if !a.Settings.SidebarOpen {
		t.Fatal("sidebar did not open with the workspace")
	}
	if a.tree == nil {
		t.Fatal("no tree")
	}
	if roots := a.tree.Roots(); len(roots) != 1 || filepath.Base(roots[0]) != "a.txt" {
		t.Fatalf("roots %v", roots)
	}
}

func TestThemeSwitchPersists(t *testing.T) {
	dir := t.TempDir()
	a := New(settings.Default(), dir)

	a.setThemeIndex(1)
	if a.ThemeIndex != 1 {
		t.Fatalf("theme index %d, want 1", a.ThemeIndex)
	}
	if a.Settings.Theme != "dark" {
		t.Fatalf("settings theme %q, want dark", a.Settings.Theme)
	}
	if !a.Dark {
		t.Fatal("app is not dark after choosing the dark theme")
	}

	if _, err := os.Stat(filepath.Join(dir, "settings.json")); err != nil {
		t.Fatalf("settings were not saved: %v", err)
	}
}
