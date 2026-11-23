package app

import (
	"slices"
	"strings"
	"testing"

	"github.com/helloxz/zpad/internal/settings"
)

func matchStrings(t *testing.T, text, query string, opt searchOptions, limit int) []string {
	t.Helper()
	ms, err := scanText(text, query, opt, limit)
	if err != nil {
		t.Fatalf("scanText(%q): %v", query, err)
	}
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = text[m.start:m.end]
	}
	return out
}

func TestScanTextPlainMatchesCase(t *testing.T) {
	text := "Needle needle NEEDLE"
	if got := matchStrings(t, text, "needle", searchOptions{caseSensitive: true}, 10); !slices.Equal(got, []string{"needle"}) {
		t.Fatalf("case sensitive: %q", got)
	}
	if got := matchStrings(t, text, "needle", searchOptions{}, 10); !slices.Equal(got, []string{"Needle", "needle", "NEEDLE"}) {
		t.Fatalf("case insensitive: %q", got)
	}
}

func TestScanTextUnicodeCaseInsensitive(t *testing.T) {
	text := "Ärger ärger ÄRGER"
	if got := matchStrings(t, text, "ärger", searchOptions{}, 10); len(got) != 3 {
		t.Fatalf("matches %q, want 3", got)
	}
}

func TestScanTextWholeWord(t *testing.T) {
	text := "cat scatter cat_ cat. cat"
	got := matchStrings(t, text, "cat", searchOptions{caseSensitive: true, wholeWord: true}, 10)
	if len(got) != 3 {
		t.Fatalf("whole word matches %d, want 3: %q", len(got), got)
	}
}

func TestScanTextRegex(t *testing.T) {
	text := "cat cot cut"
	if got := matchStrings(t, text, `c[ao]t`, searchOptions{regex: true}, 10); !slices.Equal(got, []string{"cat", "cot"}) {
		t.Fatalf("regex matches %q", got)
	}
	if got := matchStrings(t, "CAT cat", "cat", searchOptions{regex: true}, 10); len(got) != 2 {
		t.Fatalf("case-insensitive regex matches %q, want 2", got)
	}
	if _, err := scanText(text, `(`, searchOptions{regex: true}, 10); err == nil {
		t.Fatal("an invalid regular expression did not fail")
	}
}

func TestScanTextDoesNotOverlap(t *testing.T) {
	text := "aaaa"
	if got := matchStrings(t, text, "aa", searchOptions{}, 10); len(got) != 2 {
		t.Fatalf("matches %d, want 2", len(got))
	}
	if got := matchStrings(t, text, "aa", searchOptions{}, 2); len(got) != 2 {
		t.Fatalf("limit: matches %d, want 2", len(got))
	}
}

func TestPreviewLineClipsAroundMatch(t *testing.T) {
	line := strings.Repeat("x", 200) + "needle" + strings.Repeat("y", 200)
	preview, at, n := previewLine(line, 200, len("needle"))
	if n != len("needle") {
		t.Fatalf("match length %d", n)
	}
	if !strings.HasPrefix(preview, "…") || !strings.HasSuffix(preview, "…") {
		t.Fatalf("preview not clipped: %q", preview)
	}
	if preview[at:at+n] != "needle" {
		t.Fatalf("match at %d is %q", at, preview[at:at+n])
	}
}

func TestPreviewLineExpandsTabs(t *testing.T) {
	preview, at, n := previewLine("\tneedle", 1, len("needle"))
	if preview[at:at+n] != "needle" {
		t.Fatalf("match at %d is %q", at, preview[at:at+n])
	}
	if !strings.HasPrefix(preview, "    ") {
		t.Fatalf("tab not expanded: %q", preview)
	}
}

func TestSearchScansEachOpenDocumentOnce(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "needle and needle")
	a.addDocument(doc)
	a.toggleSplit() // the same document in both panes
	a.searchQuery = "needle"
	a.startSearch()

	if a.searchFiles != 1 {
		t.Fatalf("files %d, want 1", a.searchFiles)
	}
	hits := 0
	for _, row := range a.searchRows {
		if !row.header {
			hits++
		}
	}
	if hits != 2 {
		t.Fatalf("hits %d, want 2 (the document is open in both panes)", hits)
	}
}

func TestSearchGroupsByDocument(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "needle"))
	a.addDocument(docAt("/tmp/b.txt", "no match"))
	untitled := newUntitled(1, "UTF-8")
	untitled.Buffer.Set("needle in an untitled document")
	a.addDocument(untitled)
	a.searchQuery = "needle"
	a.startSearch()

	var headers []string
	for _, row := range a.searchRows {
		if row.header {
			headers = append(headers, row.doc.Name())
		}
	}
	want := []string{"a.txt", untitled.Name()}
	if !slices.Equal(headers, want) {
		t.Fatalf("headers %q, want %q", headers, want)
	}
}

func TestRevealHitSelectsTheMatch(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "first\nneedle here\nlast")
	a.addDocument(doc)
	a.searchQuery = "needle"
	a.startSearch()

	a.openSearchPanel()
	hit, ok := a.selectedSearchHit()
	if !ok {
		t.Fatalf("no hit; rows %#v", a.searchRows)
	}
	a.revealHit(hit)

	if a.searchOpen {
		t.Fatal("the panel stayed open after a jump")
	}
	tab := a.tabForDocument(doc)
	if tab == nil {
		t.Fatal("no tab for the document")
	}
	if got := tab.state.SelectedText(); got != "needle" {
		t.Fatalf("selected %q, want needle", got)
	}
}

func TestRevealHitFindsTheMatchAgainAfterAnEdit(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "first\nneedle here")
	a.addDocument(doc)
	a.searchQuery = "needle"
	a.startSearch()
	hit, ok := a.selectedSearchHit()
	if !ok {
		t.Fatal("no hit")
	}

	doc.Buffer.Set("an edit\nfirst\nneedle moved\nlast")
	a.openSearchPanel()
	a.revealHit(hit)

	tab := a.tabForDocument(doc)
	if got := tab.state.SelectedText(); got != "needle" {
		t.Fatalf("selected %q after the edit, want needle", got)
	}
}

func TestRevealHitLeavesPreviewForTheSource(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.md", "# needle")
	a.addDocument(doc)
	a.ToggleMarkdownPreview()
	if !a.activeTab().preview {
		t.Fatal("the tab is not in preview")
	}
	a.searchQuery = "needle"
	a.startSearch()
	hit, ok := a.selectedSearchHit()
	if !ok {
		t.Fatal("no hit")
	}

	a.openSearchPanel()
	a.revealHit(hit)
	tab := a.tabForDocument(doc)
	if tab.preview {
		t.Fatal("the tab stayed in preview")
	}
	if got := tab.state.SelectedText(); got != "needle" {
		t.Fatalf("selected %q", got)
	}
}

func TestPruneSearchDocumentDropsAClosedTab(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "needle"))
	a.addDocument(docAt("/tmp/b.txt", "needle"))
	a.searchQuery = "needle"
	a.startSearch()
	a.openSearchPanel()

	a.removeTab(0, 0)
	if a.searchFiles != 1 {
		t.Fatalf("files %d after closing a tab, want 1", a.searchFiles)
	}
	for _, row := range a.searchRows {
		if row.header && row.doc.Path == "/tmp/a.txt" {
			t.Fatal("a closed tab still heads results")
		}
	}
}

func TestSearchPanelEmptyQueryHasNoRows(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "needle"))
	a.searchQuery = "needle"
	a.startSearch()
	if len(a.searchRows) == 0 {
		t.Fatal("no rows before clearing")
	}

	a.searchQuery = ""
	a.startSearch()
	if len(a.searchRows) != 0 || a.searchBusy {
		t.Fatalf("rows %d busy %v after clearing", len(a.searchRows), a.searchBusy)
	}
}

func TestScanDocumentsCapsMatches(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", strings.Repeat("needle ", searchMaxPerFile+50))
	a.addDocument(doc)
	inputs := []searchInput{{doc: doc, snap: doc.Buffer.Snapshot()}}

	rows, files, err, truncated := scanDocuments(inputs, "needle", searchOptions{}, make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	if files != 1 {
		t.Fatalf("files %d", files)
	}
	hits := 0
	for _, row := range rows {
		if !row.header {
			hits++
		}
	}
	if hits != searchMaxPerFile {
		t.Fatalf("hits %d, want %d", hits, searchMaxPerFile)
	}
	if !truncated {
		t.Fatal("the cap did not report that it left matches out")
	}
}

func TestScanDocumentsStopsWhenCancelled(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/a.txt", "needle")
	a.addDocument(doc)
	inputs := []searchInput{{doc: doc, snap: doc.Buffer.Snapshot()}}
	cancel := make(chan struct{})
	close(cancel)

	rows, _, _, _ := scanDocuments(inputs, "needle", searchOptions{}, cancel)
	if len(rows) != 0 {
		t.Fatalf("rows %d after a cancel, want 0", len(rows))
	}
}

func TestFilterOpenRowsDropsAClosedDocument(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "needle")
	bDoc := docAt("/tmp/b.txt", "needle")
	a.addDocument(aDoc)
	a.addDocument(bDoc)
	a.searchQuery = "needle"
	a.startSearch()

	rows := a.filterOpenRows(append([]searchRow(nil), a.searchRows...))
	if len(rows) == 0 {
		t.Fatal("all rows were dropped")
	}
	a.removeTab(0, 0)
	rows = a.filterOpenRows(append([]searchRow(nil), a.searchRows...))
	for _, row := range rows {
		if row.doc == aDoc || (!row.header && row.hit.doc == aDoc) {
			t.Fatal("a closed document still has a row")
		}
	}
}

func TestPruneSearchDocumentRunsWhileClosed(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "needle")
	a.addDocument(aDoc)
	a.addDocument(docAt("/tmp/b.txt", "needle"))
	a.searchQuery = "needle"
	a.startSearch()
	a.closeSearchPanel()
	if a.searchOpen {
		t.Fatal("the panel did not close")
	}

	// Results are kept while closed, but the document that closes must
	// not stay referenced by them: its buffer could be large.
	a.removeTab(0, 0)
	if a.searchFiles != 1 {
		t.Fatalf("files %d after closing a tab while closed, want 1", a.searchFiles)
	}
	for _, row := range a.searchRows {
		if row.header && row.doc == aDoc {
			t.Fatal("a closed document still heads results")
		}
		if !row.header && row.hit.doc == aDoc {
			t.Fatal("a closed document still has a result")
		}
	}
}
