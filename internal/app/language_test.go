package app

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// TestUntitledLanguageChoice gives a new document a language in the status
// bar: its text is highlighted in it, the label names it, and its save starts
// from the language's extension.
func TestUntitledLanguageChoice(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := a.newUntitledDocument()
	a.addDocument(doc)
	if doc.Language != nil || doc.Buffer.Language() != nil {
		t.Fatalf("a new document has language %v", doc.Language)
	}

	json := editor.LanguageNamed("json")
	a.SelectLanguage(json)
	if doc.Language != json {
		t.Fatalf("document language %v, want JSON", doc.Language)
	}
	if doc.Buffer.Language() != json {
		t.Fatalf("buffer language %v, want JSON", doc.Buffer.Language())
	}
	if got := languageLabel(doc); got != "JSON" {
		t.Fatalf("status bar language %q, want JSON", got)
	}
	if got, want := doc.saveName(), "untitled-1.json"; got != want {
		t.Fatalf("save name %q, want %q", got, want)
	}
	if got := a.saveDialogPath(doc); filepath.Base(got) != "untitled-1.json" {
		t.Fatalf("save dialog path %q, want the JSON name", got)
	}

	tt := ui.NewTester(a.View, 900, 700)
	if _, ok := tt.AccessByRole(ui.RoleMenuButton, "JSON"); !ok {
		t.Errorf("the status bar offers no language menu showing JSON; texts %q", tt.Texts())
	}

	a.SelectLanguage(nil)
	if doc.Language != nil || doc.Buffer.Language() != nil {
		t.Fatalf("plain text leaves language %v", doc.Language)
	}
	if got, want := doc.saveName(), "untitled-1.txt"; got != want {
		t.Fatalf("save name %q after plain text, want %q", got, want)
	}
}

// TestUntitledLanguagesAreSavable keeps the list a document's save can name:
// every language in it saves as an extension the app opens as text, and the
// common ones come first.
func TestUntitledLanguagesAreSavable(t *testing.T) {
	groups := untitledLanguageGroups()
	names := func(langs []*editor.Language) []string {
		out := make([]string, 0, len(langs))
		for _, l := range langs {
			out = append(out, l.Name)
		}
		return out
	}
	all := names(groups.all())
	for _, l := range groups.all() {
		if len(l.Extensions) == 0 {
			t.Fatalf("%s has no extension", l.Name)
		}
		ext := strings.TrimPrefix(l.Extensions[0], ".")
		if !slices.Contains(textExtensions, ext) {
			t.Errorf("%s saves as .%s, which the app does not open as text", l.Name, ext)
		}
	}
	for _, want := range []string{"JSON", "PHP", "YAML", "Go", "Markdown"} {
		if !slices.Contains(all, want) {
			t.Errorf("the list lacks %s: %q", want, all)
		}
	}
	// A language known by a file name alone cannot name the file it saves.
	for _, no := range []string{"Makefile", "Dockerfile"} {
		if slices.Contains(all, no) {
			t.Errorf("%s is offered, but has no extension to save with", no)
		}
	}
	if got := names(groups.common); !slices.Equal(got, commonUntitledLanguages) {
		t.Errorf("the common languages are %q, want %q", got, commonUntitledLanguages)
	}
	if !slices.IsSortedFunc(names(groups.rest), func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	}) {
		t.Errorf("the rest are not ordered by name: %q", names(groups.rest))
	}
	if len(groups.all()) != len(groups.common)+len(groups.rest) {
		t.Errorf("%d languages, but %d common and %d others",
			len(groups.all()), len(groups.common), len(groups.rest))
	}
}

// TestSavedFileKeepsExtensionLanguage keeps a saved file's language with its
// name: it is not chosen in the status bar, and saving drops the choice an
// untitled document had.
func TestSavedFileKeepsExtensionLanguage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), dir)
	a.addDocument(doc)

	a.SelectLanguage(editor.LanguageNamed("json"))
	if doc.Language == nil || doc.Language.Name != "Go" {
		t.Fatalf("a saved file changed to %v, want Go", doc.Language)
	}
	tt := ui.NewTester(a.View, 900, 700)
	if _, ok := tt.AccessByRole(ui.RoleMenuButton, "Go"); ok {
		t.Error("a saved file offers its language in a menu")
	}
	if !tt.HasText("Go") {
		t.Errorf("the status bar does not name the language; texts %q", tt.Texts())
	}

	// Saving the untitled choice under another name keeps the name's
	// language.
	untitled := a.newUntitledDocument()
	untitled.selectLanguage(editor.LanguageNamed("json"))
	untitled.setLanguage(filepath.Join(dir, "data.php"))
	if untitled.Language == nil || untitled.Language.Name != "PHP" {
		t.Fatalf("language %v after saving as .php, want PHP", untitled.Language)
	}
}

// TestSaveAddsTheLanguageExtension keeps a name typed without an extension
// as the language's file: the highlight and the file agree after the save.
func TestSaveAddsTheLanguageExtension(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct{ name, ext, want string }{
		{"config", ".json", "config.json"},
		{filepath.Join(dir, "config"), ".json", filepath.Join(dir, "config.json")},
		{"config.json", ".json", "config.json"},
		{"config.yaml", ".json", "config.yaml"},
		{".env", ".json", ".env"},
		{"", ".json", ".json"},
	} {
		if got := withExtension(c.name, c.ext); got != c.want {
			t.Errorf("withExtension(%q, %q) = %q, want %q", c.name, c.ext, got, c.want)
		}
	}
	if got := withExtension("config", ""); got != "config" {
		t.Errorf("without an extension the name became %q", got)
	}
}

// TestDraftKeepsUntitledLanguage keeps the language a new document was given
// across a restart, beside the text its copy holds.
func TestDraftKeepsUntitledLanguage(t *testing.T) {
	dir := t.TempDir()
	a := keepApp(t, dir)
	doc := a.newUntitledDocument()
	doc.Buffer.Set(`{"a": 1}`)
	a.addDocument(doc)
	a.SelectLanguage(editor.LanguageNamed("json"))
	a.flushDrafts()
	if entries := a.drafts.Entries(); len(entries) != 1 || entries[0].Language != "JSON" {
		t.Fatalf("copies %+v, want one naming the JSON language", entries)
	}
	a.saveSession()

	b := launchApp(t, dir, true)
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want the untitled document", len(b.left.tabs))
	}
	restored := b.left.tabs[0].doc
	if restored.Language == nil || restored.Language.Name != "JSON" {
		t.Fatalf("restored language %v, want JSON", restored.Language)
	}
	if restored.Buffer.Language() == nil {
		t.Fatal("the restored text is not highlighted")
	}
	if got := restored.saveName(); got != "untitled-1.json" {
		t.Fatalf("restored save name %q, want untitled-1.json", got)
	}
}

// TestEmptyUntitledKeepsItsLanguage keeps the choice of a document whose
// text is still empty: no copy holds it, and the session does.
func TestEmptyUntitledKeepsItsLanguage(t *testing.T) {
	dir := t.TempDir()
	a := keepApp(t, dir)
	a.addDocument(a.newUntitledDocument())
	a.SelectLanguage(editor.LanguageNamed("yaml"))
	a.saveSession()

	b := launchApp(t, dir, true)
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want the empty untitled document", len(b.left.tabs))
	}
	restored := b.left.tabs[0].doc
	if restored.Language == nil || restored.Language.Name != "YAML" {
		t.Fatalf("restored language %v, want YAML", restored.Language)
	}
	if restored.Buffer.Language() == nil {
		t.Fatal("the restored document is not highlighted")
	}
}

// TestLanguageMenuListsAll keeps the menu's items the ones the list decides.
func TestLanguageMenuListsAll(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(a.newUntitledDocument())
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click(locales.PlainText()); err != nil {
		t.Fatalf("click the language: %v; texts %q", err, tt.Texts())
	}
	items := tt.Menu()
	for _, want := range []string{"JSON", "YAML", "PHP"} {
		if !slices.Contains(items, want) {
			t.Fatalf("the menu lacks %s: %q", want, items)
		}
	}
	for _, no := range []string{"Makefile", "Dockerfile"} {
		if slices.Contains(items, no) {
			t.Errorf("%s is in the menu, but has no extension to save with", no)
		}
	}
}
