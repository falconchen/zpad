package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// searchApp opens a window with two documents that both hold the query.
func searchApp(t *testing.T) (*App, *ui.Tester) {
	t.Helper()
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "first\nneedle here"))
	a.addDocument(docAt("/tmp/b.txt", "needle there"))
	tt := ui.NewTester(a.View, 900, 700)
	for range 2 {
		tt.Frame()
	}
	return a, tt
}

func openSearch(t *testing.T, a *App, tt *ui.Tester) {
	t.Helper()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyF)
	tt.Frame()
	if !a.searchOpen {
		t.Fatal("the shortcut did not open the search panel")
	}
}

// runSearch types nothing: it presses Enter to search what the field
// holds, as the panel asks.
func runSearch(t *testing.T, a *App, tt *ui.Tester) {
	t.Helper()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !a.searchFresh() {
		t.Fatalf("Enter did not search %q; busy %v err %q", a.searchQuery, a.searchBusy, a.searchErr)
	}
}

func TestSearchShortcutOpensThePanelFocused(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	if !tt.Focused(locales.SearchOpenFilesLabel()) {
		t.Fatalf("the query field is not focused; texts %q", tt.Texts())
	}
}

func TestSearchTypingDoesNotSearch(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)

	tt.Type("needle")
	tt.Frame()
	if len(a.searchRows) != 0 {
		t.Fatalf("typing searched on its own: %d rows", len(a.searchRows))
	}
	if a.searchFresh() {
		t.Fatal("the rows are marked fresh before a search")
	}
	if !tt.HasText(locales.SearchPressEnter()) {
		t.Fatalf("no hint to press Enter; texts %q", tt.Texts())
	}
}

func TestSearchPanelListsEachDocumentOnce(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)

	if !tt.HasText("a.txt") || !tt.HasText("b.txt") {
		t.Fatalf("missing file headers; texts %q", tt.Texts())
	}
	if a.searchFiles != 2 {
		t.Fatalf("files %d, want 2", a.searchFiles)
	}
	hits := 0
	for _, row := range a.searchRows {
		if !row.header {
			hits++
		}
	}
	if hits != 2 {
		t.Fatalf("hits %d, want 2", hits)
	}
}

func TestSearchButtonSearches(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()

	if err := tt.Click(locales.SearchRun()); err != nil {
		t.Fatalf("click the search button: %v; texts %q", err, tt.Texts())
	}
	tt.Frame()
	if !a.searchFresh() || len(a.searchRows) == 0 {
		t.Fatalf("the button did not search; fresh %v rows %d", a.searchFresh(), len(a.searchRows))
	}
	// The button takes the focus on a click; the field gets it back, so
	// typing and Enter go there again.
	tt.Frame()
	if !tt.Focused(locales.SearchOpenFilesLabel()) {
		t.Fatalf("the field did not get the focus back; texts %q", tt.Texts())
	}
}

func TestSearchEnterOpensTheFirstMatch(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)

	tt.Key(0, ui.KeyEnter)
	tt.Frame()

	if a.searchOpen {
		t.Fatal("the panel stayed open after Enter on a match")
	}
	tab := a.tabForDocument(a.left.tabs[0].doc)
	if tab == nil {
		t.Fatal("no tab for the first document")
	}
	if got := tab.state.SelectedText(); got != "needle" {
		t.Fatalf("selected %q, want needle", got)
	}
	if a.searchQuery != "needle" || len(a.searchRows) == 0 {
		t.Fatal("closing after a jump dropped the query or the results")
	}
}

func TestSearchClickOpensTheMatchInItsDocument(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)

	// The second document's match: b.txt line 1, column 1.
	if err := tt.Click("b.txt:1:1"); err != nil {
		t.Fatalf("click a result: %v; texts %q", err, tt.Texts())
	}
	tt.Frame()
	if a.searchOpen {
		t.Fatal("the panel stayed open after a click")
	}
	tab := a.tabForDocument(a.left.tabs[1].doc)
	if tab == nil {
		t.Fatal("no tab for the second document")
	}
	if got := tab.state.SelectedText(); got != "needle" {
		t.Fatalf("selected %q, want needle", got)
	}
}

func TestSearchDownChoosesTheNextMatch(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)

	first := a.searchSel
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if a.searchSel == first {
		t.Fatalf("Down did not change the choice from row %d", first)
	}
	if a.searchRows[a.searchSel].header {
		t.Fatalf("the choice landed on a file header, row %d", a.searchSel)
	}

	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tab := a.tabForDocument(a.left.tabs[1].doc)
	if tab == nil {
		t.Fatal("no tab for the second document")
	}
	if got := tab.state.SelectedText(); got != "needle" {
		t.Fatalf("selected %q in the second document", got)
	}
}

func TestSearchEscapeKeepsQueryAndResults(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)

	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.searchOpen {
		t.Fatal("Escape did not close the panel")
	}
	if a.searchQuery != "needle" || len(a.searchRows) == 0 || !a.searchFresh() {
		t.Fatalf("closing dropped the search: query %q rows %d fresh %v",
			a.searchQuery, len(a.searchRows), a.searchFresh())
	}

	// Reopening shows the same results without searching again.
	openSearch(t, a, tt)
	if !a.searchFresh() {
		t.Fatal("reopening searched again or lost the results")
	}
	if !tt.HasText("a.txt") || !tt.HasText("b.txt") {
		t.Fatalf("results not shown again; texts %q", tt.Texts())
	}
}

func TestSearchClearButtonForgetsQueryAndResults(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)

	if err := tt.Click(locales.SearchClear()); err != nil {
		t.Fatalf("click the clear button: %v; texts %q", err, tt.Texts())
	}
	tt.Frame()
	if a.searchQuery != "" || len(a.searchRows) != 0 || a.searchFresh() {
		t.Fatalf("clear kept the search: query %q rows %d", a.searchQuery, len(a.searchRows))
	}
	if !tt.HasText(locales.SearchHint()) {
		t.Fatalf("no empty hint after clearing; texts %q", tt.Texts())
	}
	tt.Frame()
	if !tt.Focused(locales.SearchOpenFilesLabel()) {
		t.Fatalf("the field lost the focus after clearing; texts %q", tt.Texts())
	}

	// Closing and reopening keeps the cleared state.
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	openSearch(t, a, tt)
	if a.searchQuery != "" || len(a.searchRows) != 0 {
		t.Fatalf("reopening restored a cleared search: query %q rows %d", a.searchQuery, len(a.searchRows))
	}
}

func TestSearchToggleWaitsForEnter(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)
	if a.searchCase {
		t.Fatal("case sensitivity starts on")
	}

	if err := tt.Click(locales.SearchCaseSensitive()); err != nil {
		t.Fatalf("click the case toggle: %v", err)
	}
	tt.Frame()
	if !a.searchCase {
		t.Fatal("the case toggle did not turn on")
	}
	if a.searchFresh() || len(a.searchRows) != 0 {
		t.Fatalf("the toggle searched on its own: fresh %v rows %d", a.searchFresh(), len(a.searchRows))
	}

	runSearch(t, a, tt)
	if len(a.searchRows) == 0 {
		t.Fatal("no rows after Enter searched with the new toggle")
	}
}

func TestSearchInvalidRegexSaysSo(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("(")
	tt.Frame()
	if err := tt.Click(locales.SearchRegex()); err != nil {
		t.Fatalf("click the regex toggle: %v", err)
	}
	tt.Frame()
	runSearch(t, a, tt)

	if a.searchErr == "" {
		t.Fatal("an invalid regular expression did not report an error")
	}
	if !tt.HasText(locales.SearchInvalidRegex()) {
		t.Fatalf("no error text; texts %q", tt.Texts())
	}
}

func TestSearchEditBeforeJumpFindsTheMatchAgain(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)

	doc := a.left.tabs[0].doc
	doc.Buffer.Set("an edit\nfirst\nneedle moved")
	// The panel is modal; the test edits the buffer directly, as a file
	// reload or another view could.
	if err := tt.Click("a.txt:2:1"); err != nil {
		t.Fatalf("click a stale result: %v", err)
	}
	tt.Frame()
	tab := a.tabForDocument(doc)
	if got := tab.state.SelectedText(); got != "needle" {
		t.Fatalf("selected %q after the edit", got)
	}
}

func TestSearchPreviewTabJumpsToSource(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/notes.md", "# needle"))
	a.ToggleMarkdownPreview()
	tt := ui.NewTester(a.View, 900, 700)
	tt.Frame()

	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()
	runSearch(t, a, tt)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()

	tab := a.tabForDocument(a.left.tabs[0].doc)
	if tab.preview {
		t.Fatal("the tab stayed in preview")
	}
	if got := tab.state.SelectedText(); got != "needle" {
		t.Fatalf("selected %q", got)
	}
}

func TestSearchButtonTooltipShowsBelow(t *testing.T) {
	a, tt := searchApp(t)
	openSearch(t, a, tt)
	tt.Type("needle")
	tt.Frame()

	btn, ok := tt.Find(locales.SearchRun())
	if !ok {
		t.Fatalf("no search button; texts %q", tt.Texts())
	}
	tt.Move(btn.X+btn.W/2, btn.Y+btn.H/2)
	tt.Advance(700 * time.Millisecond)
	tt.Frame()

	tip, ok := tt.Find(locales.SearchRun() + " (↵)")
	if !ok {
		t.Fatalf("no tooltip after resting on the button; texts %q", tt.Texts())
	}
	if tip.Y < btn.Y+btn.H {
		t.Fatalf("the tooltip is at %v, above the button %v", tip, btn)
	}
}
