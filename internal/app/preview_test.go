package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func TestIsMarkdownPath(t *testing.T) {
	for _, c := range []struct {
		path string
		want bool
	}{
		{"/tmp/README.md", true},
		{"/tmp/notes.MARKDOWN", true},
		{"/tmp/notes.mkd", true},
		{"/tmp/notes.mdx", false},
		{"/tmp/notes.txt", false},
		{"/tmp/notes.md.bak", false},
		{"", false},
	} {
		if got := isMarkdownPath(c.path); got != c.want {
			t.Errorf("isMarkdownPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestCanPreviewTab(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	if a.canPreviewTab(nil) {
		t.Fatal("no tab can be previewed")
	}

	a.addDocument(docAt("/tmp/notes.txt", "x"))
	if a.canPreviewTab(a.activeTab()) {
		t.Fatal("a text file is not Markdown")
	}

	a.addDocument(docAt("/tmp/README.md", "# Hi"))
	tab := a.activeTab()
	if !a.canPreviewTab(tab) {
		t.Fatal("a Markdown file should be previewable")
	}

	big := docAt("/tmp/big.md", strings.Repeat("x\n", markdownPreviewMaxLines+1))
	a.addDocument(big)
	bigTab := a.activeTab()
	if a.canPreviewTab(bigTab) {
		t.Fatal("a Markdown file over the line limit should not start a preview")
	}
	bigTab.preview = true
	if !a.canPreviewTab(bigTab) {
		t.Fatal("a preview showing should always return to the source")
	}
}

func TestClearPreviewWhenPathStopsBeingMarkdown(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/README.md", "# Hi")
	a.addDocument(doc)
	tab := a.activeTab()
	a.ToggleMarkdownPreview()
	if !tab.preview {
		t.Fatal("the tab is not previewing")
	}

	doc.Path = "/tmp/README.txt"
	a.clearPreviewIfNotMarkdown(doc)
	if tab.preview {
		t.Fatal("a tab kept previewing after its file stopped being Markdown")
	}

	doc.Path = "/tmp/README.md"
	tab.preview = true
	a.clearPreviewIfNotMarkdown(doc)
	if !tab.preview {
		t.Fatal("a Markdown tab lost its preview")
	}
}

func TestToggleMarkdownPreview(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/README.md", "# Hi"))
	tab := a.activeTab()

	a.ToggleMarkdownPreview()
	if !tab.preview {
		t.Fatal("the Markdown tab is not previewing")
	}
	a.ToggleMarkdownPreview()
	if tab.preview {
		t.Fatal("the Markdown tab did not return to the source")
	}

	a.addDocument(docAt("/tmp/notes.txt", "x"))
	text := a.activeTab()
	a.ToggleMarkdownPreview()
	if text.preview {
		t.Fatal("a text tab started a Markdown preview")
	}
}

func TestStatusBarPreviewButtonFollowsTheActiveTab(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	tt := ui.NewTester(a.View, 800, 700)
	tt.Frame()

	if info, ok := tt.Access(locales.PreviewMarkdown()); ok {
		t.Fatalf("the status bar shows the preview button with no tab open: %+v", info)
	}

	a.addDocument(docAt("/tmp/notes.txt", "x"))
	tt.Frame()
	if info, ok := tt.Access(locales.PreviewMarkdown()); ok {
		t.Fatalf("the status bar shows the preview button for a text file: %+v", info)
	}

	a.addDocument(docAt("/tmp/README.md", "# Hi"))
	tt.Frame()
	info, ok := tt.Access(locales.PreviewMarkdown())
	if !ok || info.Disabled {
		t.Fatalf("the status bar has no preview button for Markdown: %+v, found %v", info, ok)
	}
	if err := tt.Click(locales.PreviewMarkdown()); err != nil {
		t.Fatal(err)
	}
	tab := a.activeTab()
	if !tab.preview {
		t.Fatal("the preview button did not start the preview")
	}

	tt.Frame()
	info, ok = tt.Access(locales.BackToSource())
	if !ok || info.Disabled {
		t.Fatalf("the status bar has no back-to-source button: %+v, found %v", info, ok)
	}
	if err := tt.Click(locales.BackToSource()); err != nil {
		t.Fatal(err)
	}
	if tab.preview {
		t.Fatal("the button did not return to the source")
	}

	a.addDocument(docAt("/tmp/index.html", "<h1>x</h1>"))
	tt.Frame()
	if info, ok := tt.Access(locales.PreviewMarkdown()); ok {
		t.Fatalf("the status bar shows the preview button for HTML: %+v", info)
	}
}

func TestMarkdownPreviewShortcut(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/README.md", "# Hi"))
	tt := ui.NewTester(a.View, 900, 700)
	tt.Frame()
	tab := a.activeTab()

	tt.Key(ui.Cmd|ui.Shift, ui.KeyV)
	tt.Frame()
	if !tab.preview {
		t.Fatal("the shortcut did not start the preview")
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyV)
	tt.Frame()
	if tab.preview {
		t.Fatal("the shortcut did not return to the source")
	}

	a.addDocument(docAt("/tmp/notes.txt", "x"))
	tt.Frame()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyV)
	tt.Frame()
	if a.activeTab().preview {
		t.Fatal("the shortcut previewed a text file")
	}
}

func TestTabContextMenuTogglesMarkdownPreview(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/README.md", "# Hi"))
	tt := ui.NewTester(a.View, 900, 700)
	tt.Frame()

	if err := tt.RightClick("README.md"); err != nil {
		t.Fatal(err)
	}
	if !menuHas(tt.Menu(), locales.PreviewMarkdown()) {
		t.Fatalf("tab menu %q does not show %q", tt.Menu(), locales.PreviewMarkdown())
	}
	if err := tt.ChooseMenuItem(locales.PreviewMarkdown()); err != nil {
		t.Fatal(err)
	}
	tab := a.activeTab()
	if !tab.preview {
		t.Fatal("the menu did not start the preview")
	}

	tt.Frame()
	if err := tt.RightClick("README.md"); err != nil {
		t.Fatal(err)
	}
	if !menuHas(tt.Menu(), locales.BackToSource()) {
		t.Fatalf("tab menu %q does not show %q", tt.Menu(), locales.BackToSource())
	}
}

func TestMarkdownPreviewRendersText(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/README.md", "# Title\n\nHello **world**\n"))
	tt := ui.NewTester(a.View, 1100, 700)
	tt.Frame()
	a.ToggleMarkdownPreview()
	tt.Frame()

	for _, want := range []string{"Title", "Hello", "world", locales.PreviewMode()} {
		if !tt.HasText(want) {
			t.Fatalf("preview does not show %q; texts %q", want, tt.Texts())
		}
	}
}

// A split shows the source on one side and the preview on the other, both
// reading the same document as it changes.
func TestMarkdownPreviewInSplitSharesTheDocument(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/README.md", "# One"))
	a.toggleSplit()
	a.ToggleMarkdownPreview()

	right := a.right.activeTab()
	left := a.left.activeTab()
	if right == nil || left == nil || !right.preview || left.preview {
		t.Fatalf("split tabs: left preview %v, right preview %v", left != nil && left.preview, right != nil && right.preview)
	}
	if right.doc != left.doc {
		t.Fatal("the split does not share the document")
	}

	right.doc.Buffer.Set("# Two")
	if got := right.doc.PreviewText(); got != "# Two" {
		t.Fatalf("preview text %q, want %q", got, "# Two")
	}
}

func TestLargeMarkdownCannotPreview(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/big.md", strings.Repeat("x\n", markdownPreviewMaxLines+1)))
	tab := a.activeTab()
	if a.canPreviewTab(tab) {
		t.Fatal("a Markdown file over the line limit is previewable")
	}
	a.ToggleMarkdownPreview()
	if tab.preview {
		t.Fatal("the preview started for a file over the line limit")
	}
	if len(a.pendingToasts) != 1 || a.pendingToasts[0].Title != locales.PreviewTooLarge() {
		t.Fatalf("toasts %+v, want the size limit", a.pendingToasts)
	}

	tt := ui.NewTester(a.View, 900, 700)
	tt.Frame()
	info, ok := tt.Access(locales.PreviewMarkdown())
	if !ok || !info.Disabled {
		t.Fatalf("the preview button is not disabled: %+v, found %v", info, ok)
	}

	if err := tt.RightClick("big.md"); err != nil {
		t.Fatal(err)
	}
	if !menuHas(tt.Menu(), locales.PreviewTooLarge()) {
		t.Fatalf("tab menu %q does not explain the size limit", tt.Menu())
	}
	if err := tt.ChooseMenuItem(locales.PreviewTooLarge()); err == nil {
		t.Fatal("chose the disabled preview entry")
	}
}

func TestPreviewIgnoresUndo(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/README.md", "# One")
	a.addDocument(doc)
	doc.Buffer.Set("# Two")
	a.ToggleMarkdownPreview()

	a.Undo()

	if got := doc.Buffer.String(); got != "# Two" {
		t.Fatalf("undo changed the buffer while previewing: %q", got)
	}
}

func TestMarkdownPreviewScrollsWithKeyboard(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/README.md", strings.Repeat("## Section\n\nSome text.\n\n", 400)))
	tt := ui.NewTester(a.View, 900, 700)
	tt.Frame()
	a.ToggleMarkdownPreview()
	tt.Frame()

	tab := a.activeTab()
	if tab.previewScroll.Y != 0 {
		t.Fatalf("preview starts at %v, want the top", tab.previewScroll.Y)
	}
	tt.Key(0, ui.KeyPageDown)
	tt.Frame()
	if tab.previewScroll.Y == 0 {
		t.Fatal("Page Down did not scroll the preview")
	}
}

func TestMarkdownPreviewLoadsEncodedImagePaths(t *testing.T) {
	dir := t.TempDir()
	image := filepath.Join(dir, "my image.svg")
	if err := os.WriteFile(image, []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := docAt(filepath.Join(dir, "README.md"), "![x](my%20image.svg)")
	a := New(settings.Default(), dir)
	a.addDocument(doc)

	opts := a.markdownOptions(doc, ui.LightTheme())
	if opts.LoadImage == nil {
		t.Fatal("the preview has no image loader")
	}
	src, err := opts.LoadImage("my%20image.svg")
	if err != nil {
		t.Fatalf("load image: %v", err)
	}
	if src == nil {
		t.Fatal("the encoded image path loaded nothing")
	}
}

func TestMarkdownPreviewTableWidthFollowsFontSize(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/README.md", "| Date | Text |\n| --- | --- |\n| 2026-11-23 | Hello |")
	a.addDocument(doc)

	for _, size := range []float32{13, 20} {
		a.Settings.FontSize = size
		opts := a.markdownOptions(doc, ui.LightTheme())
		font := size + 2
		if got := opts.Theme.FontSize; got != font {
			t.Fatalf("preview font %v, want %v", got, font)
		}
		if got, want := opts.Theme.TableMaxColumn, markdownTableMaxPixels/font; got != want {
			t.Fatalf("TableMaxColumn at font %v = %v, want %v", font, got, want)
		}
	}
}

func TestMarkdownPreviewLinkHandling(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	doc := docAt("/tmp/README.md", "# One")
	a.addDocument(doc)
	tt := ui.NewTester(a.View, 900, 700)
	tt.Frame()
	a.ToggleMarkdownPreview()
	tt.Frame()
	if doc.md == nil {
		t.Fatal("the preview has no options")
	}

	var opened []string
	old := openExternal
	openExternal = func(raw string) error {
		opened = append(opened, raw)
		return nil
	}
	t.Cleanup(func() { openExternal = old })

	a.openMarkdownLink(doc, "https://example.com")
	a.openMarkdownLink(doc, "javascript:alert(1)")
	a.openMarkdownLink(doc, "#one")

	if len(opened) != 1 || opened[0] != "https://example.com" {
		t.Fatalf("opened %q, want only https://example.com", opened)
	}
	if doc.md.Anchor != "one" || doc.md.AnchorVersion == 0 {
		t.Fatalf("anchor %q version %d, want one", doc.md.Anchor, doc.md.AnchorVersion)
	}
}
