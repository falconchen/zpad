package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// tabMenuApp opens the given files as tabs in the left pane.
func tabMenuApp(t *testing.T, names ...string) (*App, *ui.Tester) {
	t.Helper()
	a := New(settings.Default(), t.TempDir())
	for _, name := range names {
		a.addDocument(docAt("/tmp/"+name+".txt", name))
	}
	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	return a, tt
}

func menuHas(menu []string, want string) bool {
	for _, label := range menu {
		if label == want {
			return true
		}
	}
	return false
}

func TestTabContextMenuShowsTabOperations(t *testing.T) {
	_, tt := tabMenuApp(t, "a", "b", "c")
	if err := tt.RightClick("b.txt"); err != nil {
		t.Fatal(err)
	}
	menu := tt.Menu()
	for _, want := range []string{
		locales.CloseTab(),
		locales.CloseTabsLeft(),
		locales.CloseTabsRight(),
		locales.MoveToMainPane(),
		locales.MoveToOtherPane(),
		locales.CopyToMainPane(),
		locales.CopyToOtherPane(),
	} {
		if !menuHas(menu, want) {
			t.Fatalf("tab menu %q does not show %q", menu, want)
		}
	}
}

func TestTabMenuClosesTabsToTheRight(t *testing.T) {
	a, tt := tabMenuApp(t, "a", "b", "c")
	if err := tt.RightClick("b.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(locales.CloseTabsRight()); err != nil {
		t.Fatal(err)
	}
	if len(a.left.tabs) != 2 || a.left.tabs[0].doc.Name() != "a.txt" || a.left.tabs[1].doc.Name() != "b.txt" {
		t.Fatalf("tabs %v after closing the right ones", tabNames(a.left.tabs))
	}
}

// The two ranges pass a slice of the pane's tab array to closeTabs, the way
// the menu does. The tests call closeTabs directly because the tester has no
// window, so a removal asked for during a frame runs before the frame ends
// instead of after it.
func TestCloseTabsClosesEveryTabToTheRight(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		a.addDocument(docAt("/tmp/"+name+".txt", name))
	}
	a.closeTabs(0, a.left.tabs[2:])
	if got := tabNames(a.left.tabs); len(got) != 2 || got[0] != "a.txt" || got[1] != "b.txt" {
		t.Fatalf("tabs %v after closing every tab to the right", got)
	}
}

func TestCloseTabsClosesEveryTabToTheLeft(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		a.addDocument(docAt("/tmp/"+name+".txt", name))
	}
	a.closeTabs(0, a.left.tabs[:3])
	if got := tabNames(a.left.tabs); len(got) != 2 || got[0] != "d.txt" || got[1] != "e.txt" {
		t.Fatalf("tabs %v after closing every tab to the left", got)
	}
	if active := a.left.activeTab(); active == nil || active.doc.Name() != "e.txt" {
		t.Fatalf("active tab %d after closing every tab to the left, want e.txt", a.left.active)
	}
}

func TestTabMenuClosesTabsToTheLeft(t *testing.T) {
	a, tt := tabMenuApp(t, "a", "b", "c")
	if err := tt.RightClick("b.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(locales.CloseTabsLeft()); err != nil {
		t.Fatal(err)
	}
	if len(a.left.tabs) != 2 || a.left.tabs[0].doc.Name() != "b.txt" || a.left.tabs[1].doc.Name() != "c.txt" {
		t.Fatalf("tabs %v after closing the left ones", tabNames(a.left.tabs))
	}
	if active := a.left.activeTab(); active == nil || active.doc.Name() != "c.txt" {
		t.Fatalf("active tab %d after closing the left ones, want c.txt", a.left.active)
	}
}

func TestTabMenuDisablesEmptyRanges(t *testing.T) {
	_, tt := tabMenuApp(t, "a", "b")
	if err := tt.RightClick("a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(locales.CloseTabsLeft()); err == nil {
		t.Fatal("chose Close Tabs to the Left on the first tab")
	}
	tt.CloseMenu()
	if err := tt.RightClick("b.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(locales.CloseTabsRight()); err == nil {
		t.Fatal("chose Close Tabs to the Right on the last tab")
	}
}

func TestTabMenuCannotMoveOnlyTabIntoNewSplit(t *testing.T) {
	a, tt := tabMenuApp(t, "a")
	if err := tt.RightClick("a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(locales.MoveToOtherPane()); err == nil {
		t.Fatal("moved the only tab into a split it would have to create")
	}
	if a.split {
		t.Fatal("opened a split while the item was disabled")
	}
}

func TestTabMenuMovesTabIntoNewSplit(t *testing.T) {
	a, tt := tabMenuApp(t, "a", "b")
	if err := tt.RightClick("a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(locales.MoveToOtherPane()); err != nil {
		t.Fatal(err)
	}
	if !a.split {
		t.Fatal("moving a tab did not open the split")
	}
	if len(a.left.tabs) != 1 || a.left.tabs[0].doc.Name() != "b.txt" {
		t.Fatalf("left tabs %v after the move", tabNames(a.left.tabs))
	}
	if len(a.right.tabs) != 1 || a.right.tabs[0].doc.Name() != "a.txt" {
		t.Fatalf("right tabs %v after the move", tabNames(a.right.tabs))
	}
	if a.activePane != 1 || a.right.active != 0 {
		t.Fatalf("active pane %d, right active %d", a.activePane, a.right.active)
	}
}

func TestTabMenuCopiesTabIntoNewSplit(t *testing.T) {
	a, tt := tabMenuApp(t, "a")
	if err := tt.RightClick("a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(locales.CopyToOtherPane()); err != nil {
		t.Fatal(err)
	}
	if !a.split || len(a.left.tabs) != 1 || len(a.right.tabs) != 1 {
		t.Fatalf("split %v, left %d, right %d", a.split, len(a.left.tabs), len(a.right.tabs))
	}
	if a.left.tabs[0].doc != a.right.tabs[0].doc {
		t.Fatal("the copy does not share the document")
	}
	if a.left.tabs[0].state == a.right.tabs[0].state {
		t.Fatal("the copy shares the editor state")
	}
	if a.activePane != 1 {
		t.Fatalf("active pane %d after the copy, want 1", a.activePane)
	}
}

func TestCopyTabFromRightToMain(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.toggleSplit()
	b := docAt("/tmp/b.txt", "b")
	a.addDocumentAt(1, b)

	a.copyTabIndex(1, 1, 0)
	if len(a.left.tabs) != 2 || len(a.right.tabs) != 2 {
		t.Fatalf("left %d, right %d after the copy", len(a.left.tabs), len(a.right.tabs))
	}
	if a.left.active != 1 || a.left.tabs[1].doc != b {
		t.Fatalf("left active %d, copied %v", a.left.active, tabNames(a.left.tabs))
	}
	if a.activePane != 0 {
		t.Fatalf("active pane %d after copying to the main pane, want 0", a.activePane)
	}
}

func TestCopyTabActivatesTheCopyAlreadyThere(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.toggleSplit()
	a.left.active = 0
	a.activePane = 1

	a.copyTabIndex(1, 0, 0)
	if len(a.left.tabs) != 1 || len(a.right.tabs) != 1 {
		t.Fatalf("left %d, right %d after copying a document the pane has", len(a.left.tabs), len(a.right.tabs))
	}
	if a.activePane != 0 || a.left.active != 0 {
		t.Fatalf("active pane %d, left active %d", a.activePane, a.left.active)
	}
}

func TestDirtyTabDocsCountsEachDocumentOnce(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "a")
	a.addDocument(doc)
	a.toggleSplit()
	doc.Buffer.Edit(editor.Edit{Start: 0, End: 0, Text: "x"})

	docs := dirtyTabDocs([]*Tab{a.left.tabs[0], a.right.tabs[0]})
	if len(docs) != 1 || docs[0] != doc {
		t.Fatalf("dirty docs %d, want the shared document once", len(docs))
	}
}

func tabNames(tabs []*Tab) []string {
	names := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		names = append(names, tab.doc.Name())
	}
	return names
}
