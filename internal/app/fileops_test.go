package app

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func TestPathWithinAndRebase(t *testing.T) {
	dir := "/work/proj"
	for _, tc := range []struct {
		path string
		in   bool
	}{
		{"/work/proj/a.txt", true},
		{"/work/proj/sub/a.txt", true},
		{"/work/proj", true},
		{"/work/proj2/a.txt", false},
		{"/work/other/a.txt", false},
	} {
		if got := pathWithin(tc.path, dir); got != tc.in {
			t.Errorf("pathWithin(%q, %q) = %v, want %v", tc.path, dir, got, tc.in)
		}
	}
	if got := rebasePath("/work/proj/sub/a.txt", dir, "/work/next"); got != "/work/next/sub/a.txt" {
		t.Errorf("rebasePath = %q", got)
	}
}

func TestRenameFileUpdatesOpenTabs(t *testing.T) {
	old := "/tmp/old.txt"
	newPath := "/tmp/new.md"
	a := New(settings.Default(), t.TempDir())
	doc := docAt(old, "a")
	a.addDocument(doc)
	a.toggleSplit() // the same document in both panes

	a.finishRename(old, newPath)
	if doc.Path != newPath || doc.ID != newPath {
		t.Fatalf("path %q, id %q; want %q", doc.Path, doc.ID, newPath)
	}
	if a.left.tabs[0].doc != doc || a.right.tabs[0].doc != doc {
		t.Fatal("the panes lost the renamed document")
	}
	if doc.Buffer.Language() == nil {
		t.Fatal("the renamed file kept the plain text language")
	}
}

func TestRenameFolderRebasesOpenPaths(t *testing.T) {
	oldDir := "/tmp/proj"
	newDir := "/tmp/proj2"
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/proj/a.txt", "a")
	bDoc := docAt("/tmp/proj/sub/b.txt", "b")
	a.addDocument(aDoc)
	a.addDocument(bDoc)
	a.treeState.Open.Add(oldDir)
	a.treeState.Open.Add("/tmp/proj/sub")

	a.finishRename(oldDir, newDir)
	if aDoc.Path != "/tmp/proj2/a.txt" || bDoc.Path != "/tmp/proj2/sub/b.txt" {
		t.Fatalf("paths %q, %q", aDoc.Path, bDoc.Path)
	}
	if aDoc.ID != "/tmp/proj2/a.txt" || bDoc.ID != "/tmp/proj2/sub/b.txt" {
		t.Fatalf("ids %q, %q", aDoc.ID, bDoc.ID)
	}
	if !a.treeState.Open.Has(newDir) || !a.treeState.Open.Has("/tmp/proj2/sub") {
		t.Fatalf("tree folders %v were not rebased", a.treeState.Open.Len())
	}
	if a.treeState.Open.Has(oldDir) {
		t.Fatal("the old folder is still open")
	}
}

func TestDirtyWithin(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/proj/a.txt", "a")
	a.addDocument(doc)
	if a.dirtyWithin("/tmp/proj") {
		t.Fatal("a clean document is reported dirty")
	}
	doc.Buffer.Set("changed")
	if !a.dirtyWithin("/tmp/proj") {
		t.Fatal("a changed document is not reported dirty")
	}
	if a.dirtyWithin("/tmp/other") {
		t.Fatal("a document outside the folder is reported dirty")
	}
}

func TestCloseTabsWithin(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	inside := docAt("/tmp/proj/a.txt", "a")
	outside := docAt("/tmp/other/b.txt", "b")
	a.addDocument(inside)
	a.addDocument(outside)
	a.toggleSplit() // right duplicates b

	a.closeTabsWithin("/tmp/proj")
	if len(a.left.tabs) != 1 || a.left.tabs[0].doc != outside {
		t.Fatalf("left tabs %d", len(a.left.tabs))
	}
	if len(a.right.tabs) != 1 || a.right.tabs[0].doc != outside {
		t.Fatalf("right tabs %d", len(a.right.tabs))
	}
}

func TestFinishDeleteClosesTabsAndFolders(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sub, "a.txt")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.treeState.Open.Add(sub)
	a.addDocument(docAt(path, "a"))

	a.finishDelete(sub)
	if len(a.left.tabs) != 0 {
		t.Fatalf("left tabs %d, want none", len(a.left.tabs))
	}
	if a.treeState.Open.Has(sub) {
		t.Fatal("the deleted folder is still open in the tree")
	}
	if len(a.pendingToasts) == 0 {
		t.Fatal("no toast after the delete")
	}
}

func TestOpenInOtherPane(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "a")
	a.addDocument(doc)

	a.openInOtherPane(doc.Path)
	if !a.split || !a.centerSplit {
		t.Fatalf("split %v, centered %v", a.split, a.centerSplit)
	}
	if a.activePane != 1 || len(a.right.tabs) != 1 || a.right.tabs[0].doc != doc {
		t.Fatalf("active pane %d, right tabs %d", a.activePane, len(a.right.tabs))
	}
	if a.left.tabs[0].doc != doc {
		t.Fatal("the left pane lost the document")
	}
}

func TestTreeContextMenu(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.Settings.SidebarOpen = true
	doc := docAt(path, "a")
	a.addDocument(doc)

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	if err := tt.RightClick("a.txt"); err != nil {
		t.Fatal(err)
	}
	menu := tt.Menu()
	for _, want := range []string{locales.OpenInOtherPane(), locales.Rename(), locales.Delete()} {
		found := false
		for _, item := range menu {
			if item == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("menu %q does not show %q", menu, want)
		}
	}
	if r, ok := tt.Find(locales.Rename()); ok {
		tt.Move(r.X+r.W/2, r.Y+r.H/2)
		tt.Frame()
		if got := tt.Cursor(); got != ui.CursorPointer {
			t.Fatalf("cursor over the tree menu %v, want the pointing hand", got)
		}
	}
	if err := tt.ChooseMenuItem(locales.OpenInOtherPane()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.split || a.activePane != 1 || len(a.right.tabs) != 1 || a.right.tabs[0].doc != doc {
		t.Fatalf("split %v, active pane %d, right tabs %d", a.split, a.activePane, len(a.right.tabs))
	}
}

func TestRenameDialogOpens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.Settings.SidebarOpen = true

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	if err := tt.RightClick("a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(locales.Rename()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.renameOpen || a.renamePath != path || a.renameValue != "a.txt" {
		t.Fatalf("rename dialog %v, path %q, value %q", a.renameOpen, a.renamePath, a.renameValue)
	}
	if !tt.HasText(locales.RenameLabel()) {
		t.Fatalf("rename dialog has no name field; texts %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.renameOpen {
		t.Fatal("Escape did not close the rename dialog")
	}
}

func TestRenameValidation(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "a.txt")
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	a := New(settings.Default(), t.TempDir())
	a.renameOpen = true
	a.renamePath = old

	for _, tc := range []struct {
		name  string
		error string
	}{
		{name: "  ", error: locales.NameEmpty()},
		{name: "sub/a.txt", error: locales.NameInvalid()},
		{name: `sub\a.txt`, error: locales.NameInvalid()},
		{name: "b.txt", error: locales.NameExists()},
	} {
		a.renameErr = ""
		a.renameValue = tc.name
		a.applyRename()
		if a.renameErr != tc.error {
			t.Errorf("rename %q error %q, want %q", tc.name, a.renameErr, tc.error)
		}
		if !a.renameOpen {
			t.Errorf("rename %q closed the dialog", tc.name)
		}
	}
}

func TestOpenFilesContextMenu(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.Settings.SidebarOpen = true
	a.sidebarTab = sidebarFiles
	doc := docAt(path, "a")
	a.addDocument(doc)

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	if err := tt.RightClick("a.txt"); err != nil {
		t.Fatal(err)
	}
	menu := tt.Menu()
	for _, want := range []string{locales.OpenInOtherPane(), locales.Rename(), locales.Delete()} {
		found := false
		for _, item := range menu {
			if item == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("open files menu %q does not show %q", menu, want)
		}
	}
	if err := tt.ChooseMenuItem(locales.OpenInOtherPane()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.split || a.activePane != 1 || len(a.right.tabs) != 1 || a.right.tabs[0].doc != doc {
		t.Fatalf("split %v, active pane %d, right tabs %d", a.split, a.activePane, len(a.right.tabs))
	}
}

// eventually waits for ok, which file work settles into on its goroutine.
func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestWorkspaceBlankMenu(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	// The space under the tree's only row.
	tt.RightClickAt(60, 420)
	tt.Frame()
	menu := tt.Menu()
	for _, want := range []string{locales.NewFile(), locales.NewFolder()} {
		if !slices.Contains(menu, want) {
			t.Fatalf("blank menu %q does not show %q", menu, want)
		}
	}
	if err := tt.ChooseMenuItem(locales.NewFolder()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.createOpen || a.createDir != dir || !a.createIsDir {
		t.Fatalf("create dialog %v, dir %q, folder %v", a.createOpen, a.createDir, a.createIsDir)
	}
	if !tt.HasText(locales.NameLabel()) {
		t.Fatalf("create dialog has no name field; texts %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.createOpen {
		t.Fatal("Escape did not close the create dialog")
	}

	// A row keeps its own menu: no creating beside a file from its row.
	tt.RightClickAt(60, 420)
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if err := tt.RightClick("a.txt"); err != nil {
		t.Fatal(err)
	}
	if menu := tt.Menu(); slices.Contains(menu, locales.NewFolder()) {
		t.Fatalf("file menu %q offers to create a folder", menu)
	}
}

func TestCreateFileFromBlankMenu(t *testing.T) {
	dir := t.TempDir()
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	a.beginCreate(dir, false)
	a.createValue = "new.txt"
	a.applyCreate()
	if a.createOpen {
		t.Fatal("the create dialog stayed open")
	}
	path := filepath.Join(dir, "new.txt")
	eventually(t, "the new file to open", func() bool {
		return a.findDocument(path) != nil
	})
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the file was not created: %v", err)
	}
	if len(a.left.tabs) != 1 || a.left.tabs[0].doc.Path != path {
		t.Fatalf("left tabs %d, want the new file", len(a.left.tabs))
	}
	tt.Frame()
	if roots := a.tree.Roots(); len(roots) != 1 || roots[0] != path {
		t.Fatalf("tree roots %q, want %q", roots, path)
	}
}

func TestFolderMenuCreatesInsideFolder(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	if err := tt.RightClick("docs"); err != nil {
		t.Fatal(err)
	}
	menu := tt.Menu()
	for _, want := range []string{locales.NewFile(), locales.NewFolder()} {
		if !slices.Contains(menu, want) {
			t.Fatalf("folder menu %q does not show %q", menu, want)
		}
	}
	if err := tt.ChooseMenuItem(locales.NewFolder()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.createOpen || a.createDir != sub || !a.createIsDir {
		t.Fatalf("create dialog %v, dir %q, folder %v", a.createOpen, a.createDir, a.createIsDir)
	}
	a.createValue = "notes"
	a.applyCreate()
	newDir := filepath.Join(sub, "notes")
	eventually(t, "the new folder", func() bool {
		info, err := os.Stat(newDir)
		return err == nil && info.IsDir()
	})
	tt.Frame()
	if !a.treeState.Open.Has(sub) {
		t.Fatal("the folder holding the new entry is not expanded")
	}
	if len(a.left.tabs) != 0 {
		t.Fatalf("left tabs %d, a new folder opened one", len(a.left.tabs))
	}
}

func TestCreateValidation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), t.TempDir())
	a.setWorkspace(dir)
	a.createOpen = true
	a.createDir = dir

	for _, tc := range []struct {
		name string
		want string
	}{
		{"  ", locales.NameEmpty()},
		{"sub/x.txt", locales.NameInvalid()},
		{`sub\x.txt`, locales.NameInvalid()},
		{"a.txt", locales.NameExists()},
	} {
		a.createErr = ""
		a.createValue = tc.name
		a.applyCreate()
		if a.createErr != tc.want {
			t.Errorf("create %q error %q, want %q", tc.name, a.createErr, tc.want)
		}
		if !a.createOpen {
			t.Errorf("create %q closed the dialog", tc.name)
		}
	}

	// A document still open under a name deleted from disk keeps the name.
	gone := filepath.Join(dir, "gone.txt")
	a.addDocument(docAt(gone, "x"))
	a.createErr = ""
	a.createValue = "gone.txt"
	a.applyCreate()
	if a.createErr != locales.NameExists() {
		t.Fatalf("open document name error %q, want %q", a.createErr, locales.NameExists())
	}
}
