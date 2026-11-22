package app

import (
	"testing"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func TestSettingsModalOpens(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.showSettings = true
	tt := ui.NewTester(a.View, 900, 760)
	for _, want := range []string{
		locales.SettingsTitle(),
		locales.Appearance(),
		locales.EditorSection(),
		locales.FilesSection(),
		locales.WorkspaceSection(),
		locales.RestoreDefaults(),
		locales.Done(),
	} {
		if !tt.HasText(want) {
			t.Errorf("settings panel does not show %q; texts %q", want, tt.Texts())
		}
	}
}

func TestLanguageSwitchInView(t *testing.T) {
	i18n.SetLocaleFunc(func() string { return "zh-Hans" })
	defer i18n.SetLocaleFunc(func() string { return "en" })

	a := New(settings.Default(), t.TempDir())
	tt := ui.NewTester(a.View, 900, 700)
	if !tt.HasText(locales.In("zh-Hans").AppTagline()) {
		t.Fatalf("view did not switch to Chinese; texts %q", tt.Texts())
	}
}

func TestFontNamesAreMonospaced(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	allowed := map[string]bool{}
	for _, f := range ui.FontFamilies() {
		if f.Monospaced {
			allowed[f.Name] = true
		}
	}
	for _, name := range a.fontNames() {
		if !allowed[name] {
			t.Fatalf("font %q is not monospaced", name)
		}
	}
}

func TestCatalogsHaveBothLanguages(t *testing.T) {
	en := locales.In("en").AppTagline()
	zh := locales.In("zh-Hans").AppTagline()
	if en == "" || zh == "" {
		t.Fatal("empty catalog message")
	}
	if en == zh {
		t.Fatal("English and Chinese app taglines are the same")
	}
	line := locales.In("zh-Hans").LineColumn(3, 7)
	if line != "行 3, 列 7" {
		t.Fatalf("Chinese line/column = %q", line)
	}
}
