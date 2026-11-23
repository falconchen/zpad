package app

import (
	"errors"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// A failure opening a file tells in a toast that goes away on its own, and
// leaves nothing in the status bar to sit there.
func TestOpenFailureToastExpires(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.reportOpenFailures([]openFailure{{"photo.png", locales.BinaryFile()}})

	tt := ui.NewTester(a.View, 1200, 800)
	title := locales.OpenFailed("photo.png")
	if !tt.HasText(title) {
		t.Fatalf("view does not show %q; texts %q", title, tt.Texts())
	}
	if !tt.HasText(locales.BinaryFile()) {
		t.Fatalf("the toast does not say why; texts %q", tt.Texts())
	}
	if a.errMsg != "" {
		t.Fatalf("status bar keeps the error %q", a.errMsg)
	}
	if len(a.pendingToasts) != 0 {
		t.Fatalf("%d toasts still queued", len(a.pendingToasts))
	}

	tt.Advance(errorToastTime + time.Second)
	tt.Frame()
	if tt.HasText(title) {
		t.Fatal("the toast did not go away")
	}
}

// A batch of files that cannot open reports once, with the count and each
// reason, rather than stacking one toast per file.
func TestOpenFailuresOfABatchShareOneToast(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.reportOpenFailures([]openFailure{
		{"a.png", locales.BinaryFile()},
		{"b.png", locales.BinaryFile()},
		{"c.txt", locales.UnknownEncoding()},
	})

	if len(a.pendingToasts) != 1 {
		t.Fatalf("%d toasts for a batch", len(a.pendingToasts))
	}
	toast := a.pendingToasts[0]
	if want := locales.FilesFailedToOpen(3); toast.Title != want {
		t.Fatalf("title %q, want %q", toast.Title, want)
	}
	want := locales.BinaryFile() + locales.ReasonSeparator() + locales.UnknownEncoding()
	if toast.Description != want {
		t.Fatalf("description %q, want %q", toast.Description, want)
	}
	if toast.Type != "error" || toast.Timeout != errorToastTime {
		t.Fatalf("toast type %q timeout %s", toast.Type, toast.Timeout)
	}
}

// An action that fails tells with a toast, not with a status bar error
// that would stay until the next file opens.
func TestFailTellsWithAToast(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.fail(locales.SaveFailed(), errors.New("disk full"))

	if a.errMsg != "" {
		t.Fatalf("status bar keeps the error %q", a.errMsg)
	}
	if len(a.pendingToasts) != 1 {
		t.Fatalf("%d toasts", len(a.pendingToasts))
	}
	if toast := a.pendingToasts[0]; toast.Title != locales.SaveFailed() ||
		toast.Description != "disk full" {
		t.Fatalf("toast %q / %q", toast.Title, toast.Description)
	}
}
