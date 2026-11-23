package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// stubOpenInBrowser records the paths the browser would open, in the order
// they were asked for.
func stubOpenInBrowser(t *testing.T, calls *[]string) {
	t.Helper()
	old := openInBrowser
	openInBrowser = func(path string) error {
		*calls = append(*calls, path)
		return nil
	}
	t.Cleanup(func() { openInBrowser = old })
}

func TestIsHTMLPath(t *testing.T) {
	for _, c := range []struct {
		path string
		want bool
	}{
		{"/tmp/index.html", true},
		{"/tmp/index.HTM", true},
		{"/tmp/page.xhtml", true},
		{"/tmp/page.html.bak", false},
		{"/tmp/page.txt", false},
		{"/tmp/page", false},
		{"", false},
	} {
		if got := isHTMLPath(c.path); got != c.want {
			t.Errorf("isHTMLPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestCanOpenInBrowser(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	if a.canOpenInBrowser() {
		t.Fatal("a window with no tabs has nothing to open")
	}

	a.addDocument(docAt("/tmp/notes.txt", "x"))
	if a.canOpenInBrowser() {
		t.Fatal("a text file is not HTML")
	}

	a.addDocument(a.newUntitledDocument())
	if a.canOpenInBrowser() {
		t.Fatal("an untitled document has no path to open")
	}

	a.addDocument(docAt("/tmp/index.HTML", "<h1>x</h1>"))
	if !a.canOpenInBrowser() {
		t.Fatal("an HTML file should open in the browser")
	}
}

func TestOpenActiveInBrowserOpensAnHTMLFile(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/index.html", "<h1>x</h1>"))
	var calls []string
	stubOpenInBrowser(t, &calls)

	a.OpenActiveInBrowser()

	if len(calls) != 1 || calls[0] != "/tmp/index.html" {
		t.Fatalf("opened %q, want [/tmp/index.html]", calls)
	}
}

func TestOpenActiveInBrowserSkipsOtherFiles(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/notes.txt", "x"))
	var calls []string
	stubOpenInBrowser(t, &calls)

	a.OpenActiveInBrowser()

	if len(calls) != 0 {
		t.Fatalf("opened %q, want nothing", calls)
	}
}

// The context menu of a tab opens that tab's file, even when another tab is
// the active one.
func TestOpenTabInBrowserUsesThatTab(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/first.html", "1"))
	a.addDocument(docAt("/tmp/second.html", "2"))
	var calls []string
	stubOpenInBrowser(t, &calls)

	a.openTabInBrowser(a.left.tabs[0])

	if len(calls) != 1 || calls[0] != "/tmp/first.html" {
		t.Fatalf("opened %q, want [/tmp/first.html]", calls)
	}
}

// Unsaved changes go to disk first: the browser shows what the editor
// shows.
func TestOpenActiveInBrowserSavesChangesFirst(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.html")
	if err := os.WriteFile(path, []byte("<h1>old</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := New(settings.Default(), dir)
	doc := docAt(path, "<h1>old</h1>")
	a.addDocument(doc)
	doc.Buffer.Set("<h1>new</h1>")
	if !doc.Buffer.Modified() {
		t.Fatal("the document is not modified")
	}

	opened := make(chan string, 1)
	old := openInBrowser
	openInBrowser = func(path string) error {
		opened <- path
		return nil
	}
	t.Cleanup(func() { openInBrowser = old })

	a.OpenActiveInBrowser()

	select {
	case got := <-opened:
		if got != path {
			t.Fatalf("opened %q, want %q", got, path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the browser was not opened after the save")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "<h1>new</h1>"; got != want {
		t.Fatalf("the saved file is %q, want %q", got, want)
	}
}

func TestTabContextMenuOpensHTMLInBrowser(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/index.html", "<h1>x</h1>"))
	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	if err := tt.RightClick("index.html"); err != nil {
		t.Fatal(err)
	}
	if !menuHas(tt.Menu(), locales.OpenInBrowser()) {
		t.Fatalf("tab menu %q does not show %q", tt.Menu(), locales.OpenInBrowser())
	}

	var calls []string
	stubOpenInBrowser(t, &calls)
	if err := tt.ChooseMenuItem(locales.OpenInBrowser()); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "/tmp/index.html" {
		t.Fatalf("the menu opened %q, want [/tmp/index.html]", calls)
	}
}

func TestOpenInBrowserFailureTellsWithAToast(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/index.html", "<h1>x</h1>"))
	old := openInBrowser
	openInBrowser = func(string) error { return errors.New("no browser") }
	t.Cleanup(func() { openInBrowser = old })

	a.OpenActiveInBrowser()

	if len(a.pendingToasts) != 1 {
		t.Fatalf("%d toasts, want 1", len(a.pendingToasts))
	}
	if got, want := a.pendingToasts[0].Title, locales.OpenInBrowserFailed(); got != want {
		t.Fatalf("toast title %q, want %q", got, want)
	}
	if a.errMsg != "" {
		t.Fatalf("status bar keeps the error %q", a.errMsg)
	}
}

func TestStatusBarBrowserButtonFollowsTheActiveTab(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	tt := ui.NewTester(a.View, 800, 600)
	tt.Frame()

	if info, ok := tt.Access(locales.OpenInBrowser()); ok {
		t.Fatalf("the status bar shows the browser button with no tab open: %+v", info)
	}

	a.addDocument(docAt("/tmp/notes.txt", "x"))
	tt.Frame()
	if info, ok := tt.Access(locales.OpenInBrowser()); ok {
		t.Fatalf("the status bar shows the browser button for a text file: %+v", info)
	}

	a.addDocument(a.newUntitledDocument())
	tt.Frame()
	if info, ok := tt.Access(locales.OpenInBrowser()); ok {
		t.Fatalf("the status bar shows the browser button for an untitled document: %+v", info)
	}

	a.addDocument(docAt("/tmp/index.html", "<h1>x</h1>"))
	tt.Frame()
	info, ok := tt.Access(locales.OpenInBrowser())
	if !ok || info.Disabled {
		t.Fatalf("the status bar has no browser button for HTML: %+v, found %v", info, ok)
	}

	var calls []string
	stubOpenInBrowser(t, &calls)
	if err := tt.Click(locales.OpenInBrowser()); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "/tmp/index.html" {
		t.Fatalf("the button opened %q, want [/tmp/index.html]", calls)
	}
}

func TestOpenInBrowserShortcut(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/index.html", "<h1>x</h1>"))
	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()

	var calls []string
	stubOpenInBrowser(t, &calls)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Frame()
	if len(calls) != 1 || calls[0] != "/tmp/index.html" {
		t.Fatalf("the shortcut opened %q, want [/tmp/index.html]", calls)
	}

	a.addDocument(docAt("/tmp/notes.txt", "x"))
	tt.Frame()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Frame()
	if len(calls) != 1 {
		t.Fatalf("the shortcut opened %q for a text file, want nothing new", calls)
	}
}
