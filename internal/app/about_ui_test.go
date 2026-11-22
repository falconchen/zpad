package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func TestAboutDialogShowsLicenseAndRepository(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.aboutOpen = true
	tt := ui.NewTester(a.View, 900, 700)
	for _, want := range []string{
		"ZPad",
		locales.AboutTagline(),
		locales.AboutVersion(Version),
		locales.AboutLicense(),
		"AGPL-3.0",
		"github.com/helloxz/zpad",
	} {
		if !tt.HasText(want) {
			t.Errorf("about dialog does not show %q; texts %q", want, tt.Texts())
		}
	}
}

func TestAboutLinksOpenInBrowser(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.aboutOpen = true
	tt := ui.NewTester(a.View, 900, 700)
	if err := tt.Click("AGPL-3.0"); err != nil {
		t.Fatal(err)
	}
	if got := tt.OpenedURLs(); len(got) != 1 || got[0] != aboutLicenseURL {
		t.Fatalf("the license link opened %q, want %q", got, aboutLicenseURL)
	}
	if err := tt.Click("github.com/helloxz/zpad"); err != nil {
		t.Fatal(err)
	}
	if got := tt.OpenedURLs(); len(got) != 2 || got[1] != aboutRepoURL {
		t.Fatalf("the repository link opened %q, want %q after the license link", got, aboutRepoURL)
	}
}

func TestAboutDialogCloses(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.aboutOpen = true
	tt := ui.NewTester(a.View, 900, 700)
	if err := tt.Click(locales.Close()); err != nil {
		t.Fatal(err)
	}
	if a.aboutOpen {
		t.Fatal("the close button did not close the about dialog")
	}

	a.aboutOpen = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.aboutOpen {
		t.Fatal("Escape did not close the about dialog")
	}
}

// The About dialog shows the app icon `mygo build` ships, and keeps the
// letter mark when no resources are there, as in these tests.
func TestAboutIconLoads(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.readAboutIcon("../../resources/icon.png")
	if a.aboutIcon == nil {
		t.Fatal("resources/icon.png did not decode")
	}
	if w, h := ui.ImageSize(a.aboutIcon); w != 1024 || h != 1024 {
		t.Fatalf("the app icon is %vx%v, want 1024x1024", w, h)
	}

	a.aboutOpen = true
	tt := ui.NewTester(a.View, 900, 700)
	tt.Frame()
	if !tt.HasText(locales.AboutVersion(Version)) {
		t.Fatalf("the dialog with an icon does not show the version; texts %q", tt.Texts())
	}

	// A missing file leaves the fallback in place rather than failing.
	missing := New(settings.Default(), t.TempDir())
	missing.readAboutIcon("../../resources/missing.png")
	if missing.aboutIcon != nil {
		t.Fatal("a missing icon was reported as loaded")
	}
}
