package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/internal/session"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// keepApp is an app with the recovery copies on, opened as the launch does.
func keepApp(t *testing.T, dir string) *App {
	t.Helper()
	s := settings.Default()
	s.KeepUnsaved = true
	a := New(s, dir)
	if err := a.openDraftsStore(); err != nil {
		t.Fatal(err)
	}
	a.loadDraftRestoreIndex()
	t.Cleanup(a.stopDraftLoop)
	return a
}

// launchApp starts an app from dir, as the next launch does.
func launchApp(t *testing.T, dir string, keep bool) *App {
	t.Helper()
	s := settings.Default()
	s.KeepUnsaved = keep
	s.RestoreSession = true
	a := New(s, dir)
	a.RestoreSession()
	t.Cleanup(a.stopDraftLoop)
	return a
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDraftCopiesUntitledText(t *testing.T) {
	a := keepApp(t, t.TempDir())
	doc := a.newUntitledDocument()
	doc.Buffer.Set("hello\nworld\n")
	a.addDocument(doc)

	a.flushDrafts()

	entries := a.drafts.Entries()
	if len(entries) != 1 {
		t.Fatalf("%d copies after flushing one untitled document, want 1", len(entries))
	}
	if entries[0].Path != "" || entries[0].Untitled != 1 {
		t.Fatalf("copy %+v, want an untitled copy numbered 1", entries[0])
	}
	text, err := a.drafts.Text(entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if text != "hello\nworld\n" {
		t.Fatalf("copy holds %q, want the untitled text", text)
	}
}

func TestDraftRestoresUntitledOnLaunch(t *testing.T) {
	dir := t.TempDir()
	a := keepApp(t, dir)
	doc := a.newUntitledDocument()
	doc.Buffer.Set("draft text")
	a.addDocument(doc)
	a.flushDrafts()
	a.saveSession()

	b := launchApp(t, dir, true)
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want the untitled document", len(b.left.tabs))
	}
	tab := b.left.tabs[0]
	if got := tab.doc.Buffer.String(); got != "draft text" {
		t.Fatalf("restored %q, want the unsaved text", got)
	}
	if !tab.doc.Buffer.Modified() {
		t.Fatal("the restored document does not look unsaved")
	}
	if tab.doc.Name() != locales.UntitledN(1) {
		t.Fatalf("restored tab %q, want the untitled number", tab.doc.Name())
	}
	if b.untitledSeq != 1 {
		t.Fatalf("next untitled number %d after the restore, want 2", b.untitledSeq+1)
	}
	if len(b.drafts.Entries()) != 1 {
		t.Fatal("the copy was deleted before the restored text was written again")
	}
}

func TestDraftRestoresUnsavedFileText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notes.txt")
	writeFile(t, path, "on disk\n")

	a := keepApp(t, dir)
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	a.addDocument(doc)
	doc.Encoding = encoding.GB18030
	doc.Buffer.Set("draft text\nsecond line\n")
	doc.Buffer.SetLineEnding("\r\n")
	a.flushDrafts()
	a.saveSession()

	b := launchApp(t, dir, true)
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want one", len(b.left.tabs))
	}
	tab := b.left.tabs[0]
	if got := tab.doc.Buffer.String(); got != "draft text\nsecond line\n" {
		t.Fatalf("restored %q, want the unsaved text", got)
	}
	if tab.doc.Path != path {
		t.Fatalf("restored to %q, want %q", tab.doc.Path, path)
	}
	if !tab.doc.Buffer.Modified() {
		t.Fatal("the restored document does not look unsaved")
	}
	if tab.doc.Encoding != encoding.GB18030 {
		t.Fatalf("restored encoding %q, want GB18030", tab.doc.Encoding)
	}
	if got := tab.doc.Buffer.LineEnding(); got != "\r\n" {
		t.Fatalf("restored line ending %q, want CRLF", got)
	}
}

func TestDraftRestoresAnEmptiedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	writeFile(t, path, "one\n")

	a := keepApp(t, dir)
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	a.addDocument(doc)
	doc.Buffer.Set("")
	a.flushDrafts()
	a.saveSession()

	b := launchApp(t, dir, true)
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want one", len(b.left.tabs))
	}
	tab := b.left.tabs[0]
	if got := tab.doc.Buffer.String(); got != "" {
		t.Fatalf("restored %q, want the emptied text", got)
	}
	if !tab.doc.Buffer.Modified() {
		t.Fatal("an emptied file did not come back unsaved")
	}
}

func TestDraftDropsACopyTheFileAlreadyHas(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	writeFile(t, path, "one\n")

	a := keepApp(t, dir)
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	a.addDocument(doc)
	doc.Buffer.Set("two\n")
	a.flushDrafts()
	// The save that happened after the copy was written.
	writeFile(t, path, "two\n")
	a.saveSession()

	b := launchApp(t, dir, true)
	if len(b.drafts.Entries()) != 0 {
		t.Fatal("a copy of text the file already has stayed")
	}
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want one", len(b.left.tabs))
	}
	tab := b.left.tabs[0]
	if tab.doc.Buffer.Modified() {
		t.Fatal("a document the file already has was restored as unsaved")
	}
	if got := tab.doc.Buffer.String(); got != "two\n" {
		t.Fatalf("restored %q, want the file's text", got)
	}
}

func TestDraftGoesWhenTheTextIsSaved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	writeFile(t, path, "one\n")

	a := keepApp(t, dir)
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	a.addDocument(doc)
	doc.Buffer.Set("two\n")
	a.flushDrafts()
	if len(a.drafts.Entries()) != 1 {
		t.Fatal("the copy was not written")
	}

	doc.Buffer.MarkSaved()
	a.flushDrafts()

	if len(a.drafts.Entries()) != 0 {
		t.Fatal("the copy stayed after the text was saved")
	}
}

func TestDraftGoesWhenTheLastTabCloses(t *testing.T) {
	a := keepApp(t, t.TempDir())
	doc := a.newUntitledDocument()
	doc.Buffer.Set("text")
	a.addDocument(doc)
	a.flushDrafts()

	a.removeTab(0, 0)

	if len(a.drafts.Entries()) != 0 {
		t.Fatal("the copy stayed after the last tab of the document closed")
	}
}

func TestDraftStaysWhileTheDocumentIsOpenElsewhere(t *testing.T) {
	a := keepApp(t, t.TempDir())
	doc := a.newUntitledDocument()
	doc.Buffer.Set("text")
	a.addDocument(doc)
	a.flushDrafts()

	a.split = true
	a.right.tabs = append(a.right.tabs, &Tab{
		doc:   doc,
		state: &editor.State{Buffer: doc.Buffer},
	})
	a.right.active = 0
	a.activePane = 0

	a.removeTab(0, 0)

	if len(a.drafts.Entries()) != 1 {
		t.Fatal("the copy went while the document was still open in the other pane")
	}
	a.removeTab(0, 0)
	if len(a.drafts.Entries()) != 0 {
		t.Fatal("the copy stayed after the last tab closed")
	}
}

func TestDraftFollowsARename(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old.txt")
	writeFile(t, old, "one\n")

	a := keepApp(t, dir)
	doc, err := loadDocument(old, 20)
	if err != nil {
		t.Fatal(err)
	}
	a.addDocument(doc)
	doc.Buffer.Set("two\n")
	a.flushDrafts()

	newPath := filepath.Join(dir, "new.txt")
	a.finishRename(old, newPath)

	entries := a.drafts.Entries()
	if len(entries) != 1 {
		t.Fatalf("%d copies after a rename, want 1", len(entries))
	}
	if entries[0].Path != newPath {
		t.Fatalf("copy names %q after a rename, want %q", entries[0].Path, newPath)
	}
}

func TestKeepUnsavedOffLeavesTheFileAsItIs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	writeFile(t, path, "on disk\n")

	a := keepApp(t, dir)
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	a.addDocument(doc)
	doc.Buffer.Set("draft\n")
	a.flushDrafts()
	a.saveSession()

	b := launchApp(t, dir, false)
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want one", len(b.left.tabs))
	}
	tab := b.left.tabs[0]
	if got := tab.doc.Buffer.String(); got != "on disk\n" {
		t.Fatalf("restored %q with keeping off, want the file's text", got)
	}
	if tab.doc.Buffer.Modified() {
		t.Fatal("a document came back unsaved with keeping off")
	}
}

func TestTurningOffPurgesTheCopies(t *testing.T) {
	dir := t.TempDir()
	a := keepApp(t, dir)
	doc := a.newUntitledDocument()
	doc.Buffer.Set("text")
	a.addDocument(doc)
	a.flushDrafts()

	a.Settings.KeepUnsaved = false
	a.keepUnsavedChanged()

	if a.drafts != nil {
		t.Fatal("the store stayed open after keeping was turned off")
	}
	names, err := os.ReadDir(filepath.Join(dir, "drafts"))
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("%d files stayed after keeping was turned off", len(names))
	}
}

func TestQuitKeepsUnsavedText(t *testing.T) {
	dir := t.TempDir()
	a := keepApp(t, dir)
	doc := a.newUntitledDocument()
	doc.Buffer.Set("text")
	a.addDocument(doc)

	a.BeforeQuit(&mygo.QuitEvent{})

	if len(a.drafts.Entries()) != 1 {
		t.Fatal("quitting dropped the unsaved text instead of keeping it")
	}
	if !session.Exists(dir) {
		t.Fatal("quitting did not save the session")
	}
}

func TestDraftSkipsReadOnlyDocuments(t *testing.T) {
	a := keepApp(t, t.TempDir())
	doc := a.newUntitledDocument()
	doc.Buffer.Set("text")
	doc.ReadOnly = true
	a.addDocument(doc)

	a.flushDrafts()

	if len(a.drafts.Entries()) != 0 {
		t.Fatal("a read-only document got a copy")
	}
}

func TestDraftWaitsForAPause(t *testing.T) {
	a := keepApp(t, t.TempDir())
	doc := a.newUntitledDocument()
	doc.Buffer.Set("text")
	a.addDocument(doc)

	now := time.Now()
	if jobs := a.collectDraftJobs(now, false); len(jobs) != 0 {
		t.Fatalf("%d copies on the first look, want none", len(jobs))
	}
	if jobs := a.collectDraftJobs(now.Add(500*time.Millisecond), false); len(jobs) != 0 {
		t.Fatalf("%d copies half a second later, want none", len(jobs))
	}
	if jobs := a.collectDraftJobs(now.Add(1200*time.Millisecond), false); len(jobs) != 1 {
		t.Fatalf("%d copies after the pause, want 1", len(jobs))
	}
}

func TestDraftCopiesDuringLongTyping(t *testing.T) {
	a := keepApp(t, t.TempDir())
	doc := a.newUntitledDocument()
	doc.Buffer.Set("text")
	a.addDocument(doc)

	start := time.Now()
	a.collectDraftJobs(start, false)
	doc.Buffer.Set("text, still typing")

	if jobs := a.collectDraftJobs(start.Add(4900*time.Millisecond), false); len(jobs) != 0 {
		t.Fatalf("%d copies during typing, want none yet", len(jobs))
	}
	if jobs := a.collectDraftJobs(start.Add(5*time.Second), false); len(jobs) != 1 {
		t.Fatalf("%d copies after five seconds of typing, want 1", len(jobs))
	}
}

func TestDraftKeepsTheStampItWasWrittenWith(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	writeFile(t, path, "on disk\n")
	stamp := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}

	a := keepApp(t, dir)
	doc, err := loadDocument(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	a.addDocument(doc)
	doc.Buffer.Set("draft text\n")
	a.flushDrafts()
	a.saveSession()

	// Another program changes the file while the app is away.
	writeFile(t, path, "changed\n")
	later := time.Now().Add(time.Minute)
	if err := os.Chtimes(path, later, later); err != nil {
		t.Fatal(err)
	}

	b := launchApp(t, dir, true)
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want one", len(b.left.tabs))
	}
	tab := b.left.tabs[0]
	if got := tab.doc.Buffer.String(); got != "draft text\n" {
		t.Fatalf("restored %q, want the unsaved text", got)
	}
	// The document keeps the file's state when the copy was written, so
	// that the external-change check on the next focus asks about it.
	if !tab.doc.ModTime.Equal(stamp) {
		t.Fatalf("restored modification time %v, want the copy's %v", tab.doc.ModTime, stamp)
	}
}

func TestDraftRestoresAnEmptyUntitledTab(t *testing.T) {
	dir := t.TempDir()
	a := keepApp(t, dir)
	doc := a.newUntitledDocument()
	a.addDocument(doc)
	a.saveSession()

	b := launchApp(t, dir, true)
	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored, want the empty untitled one", len(b.left.tabs))
	}
	if got := b.left.tabs[0].doc.Buffer.String(); got != "" {
		t.Fatalf("empty untitled tab restored with %q", got)
	}
	if b.left.tabs[0].doc.Name() != locales.UntitledN(1) {
		t.Fatalf("restored tab %q, want the untitled number", b.left.tabs[0].doc.Name())
	}
}

func TestDraftRestoresWithoutSessionRestore(t *testing.T) {
	dir := t.TempDir()
	a := keepApp(t, dir)
	doc := a.newUntitledDocument()
	doc.Buffer.Set("kept text")
	a.addDocument(doc)
	a.flushDrafts()

	s := settings.Default()
	s.KeepUnsaved = true
	s.RestoreSession = false
	b := New(s, dir)
	b.RestoreSession()
	t.Cleanup(b.stopDraftLoop)

	if len(b.left.tabs) != 1 {
		t.Fatalf("%d tabs restored without session restore, want the unsaved document", len(b.left.tabs))
	}
	if got := b.left.tabs[0].doc.Buffer.String(); got != "kept text" {
		t.Fatalf("restored %q, want the unsaved text", got)
	}
}

func TestTurningOnKeepsWhatIsOpenNow(t *testing.T) {
	dir := t.TempDir()
	s := settings.Default()
	a := New(s, dir)
	t.Cleanup(a.stopDraftLoop)
	doc := a.newUntitledDocument()
	doc.Buffer.Set("typed before turning it on")
	a.addDocument(doc)

	a.Settings.KeepUnsaved = true
	a.keepUnsavedChanged()

	if a.drafts == nil {
		t.Fatal("turning the setting on did not open the store")
	}
	entries := a.drafts.Entries()
	if len(entries) != 1 {
		t.Fatalf("%d copies after turning the setting on, want 1", len(entries))
	}
	if text, err := a.drafts.Text(entries[0].ID); err != nil || text != "typed before turning it on" {
		t.Fatalf("copy = %q, %v", text, err)
	}
}

func TestSettingsShowsKeepUnsaved(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.showSettings = true
	tt := ui.NewTester(a.View, 900, 760)
	if !tt.HasText(locales.KeepUnsaved()) {
		t.Fatalf("settings do not show %q; texts %q", locales.KeepUnsaved(), tt.Texts())
	}
	if !tt.HasText(locales.KeepUnsavedDescription()) {
		t.Fatalf("settings do not describe the switch; texts %q", tt.Texts())
	}
}

func TestConfirmCloseIsDisabledWhileKeepingUnsaved(t *testing.T) {
	s := settings.Default()
	s.KeepUnsaved = true
	a := New(s, t.TempDir())
	a.showSettings = true
	tt := ui.NewTester(a.View, 900, 1100)
	tt.Frame()

	info, ok := tt.AccessByRole(ui.RoleSwitch, locales.ConfirmClose())
	if !ok {
		t.Fatalf("the settings have no %q switch", locales.ConfirmClose())
	}
	if !info.Disabled {
		t.Fatal("Confirm Unsaved Files on Close is still on while keeping unsaved content")
	}
}
