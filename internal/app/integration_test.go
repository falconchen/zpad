//go:build darwin

package app

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/internal/sysint"
	"github.com/helloxz/zpad/locales"
)

type testHandlers struct {
	current map[string]string
}

func (h *testHandlers) RegisterApp(string) error { return nil }

func (h *testHandlers) GetDefaultHandler(ext string) (string, error) {
	return h.current[ext], nil
}

func (h *testHandlers) SetDefaultHandler(ext, bundleID string) error {
	if h.current == nil {
		h.current = map[string]string{}
	}
	h.current[ext] = bundleID
	return nil
}

func newIntegrationTestApp(t *testing.T) (*App, *testHandlers) {
	t.Helper()
	dir := t.TempDir()
	h := &testHandlers{
		current: map[string]string{
			"txt": "com.apple.TextEdit",
			"md":  "com.apple.TextEdit",
			"cfg": "com.apple.TextEdit",
		},
	}
	a := New(settings.Default(), dir)
	a.sysint = &sysint.Manager{
		AppBundle:            "/Applications/ZPad.app",
		CLIDirs:              []string{t.TempDir()},
		ServicesDir:          t.TempDir(),
		FinderSyncConfigPath: filepath.Join(t.TempDir(), "zpad-sync.json"),
		SkipPBS:              true,
		Handlers:             h,
	}
	return a, h
}

func TestIntegrationTogglesAndRestoreDefaults(t *testing.T) {
	a, h := newIntegrationTestApp(t)

	// Defaults are all off.
	if a.Settings.EnableCLI || a.Settings.FinderFolderMenu || a.Settings.FinderFileMenu || a.Settings.DefaultAppEnabled {
		t.Fatal("expected all integration settings off by default")
	}

	// Enable CLI & Finder menus.
	a.Settings.EnableCLI = true
	a.toggleCLI()
	if _, ok := a.sysint.CLIStatus(); !ok {
		t.Fatal("expected CLI wrapper installed")
	}

	a.Settings.FinderFolderMenu = true
	a.toggleFinderFolderMenu()
	if !a.sysint.HasFinderWorkflow(sysint.FinderKindFolder) {
		t.Fatal("expected Finder folder workflow installed")
	}

	a.Settings.FinderFileMenu = true
	a.toggleFinderFileMenu()
	if !a.sysint.HasFinderWorkflow(sysint.FinderKindFile) {
		t.Fatal("expected Finder file workflow installed")
	}

	// Enable Default App associations and associate preset + custom extensions.
	a.Settings.DefaultAppEnabled = true
	a.toggleDefaultAppEnabled()
	a.toggleDefaultExt("txt")
	a.customExtInput = ".cfg, md"
	a.addCustomDefaultExts()

	if h.current["txt"] != sysint.BundleID || h.current["md"] != sysint.BundleID || h.current["cfg"] != sysint.BundleID {
		t.Fatalf("unexpected handlers after associate: %+v", h.current)
	}
	if !slices.Equal(a.Settings.DefaultAppExts, []string{"txt", "cfg", "md"}) {
		t.Fatalf("DefaultAppExts = %v", a.Settings.DefaultAppExts)
	}

	// Turn off Default App toggle -> restores all handlers.
	a.Settings.DefaultAppEnabled = false
	a.toggleDefaultAppEnabled()
	if h.current["txt"] != "com.apple.TextEdit" || h.current["cfg"] != "com.apple.TextEdit" {
		t.Fatalf("expected handlers restored when toggled off, got %+v", h.current)
	}

	// Turn back on -> re-applies remembered extensions, then restoreDefaults cleans everything up.
	a.Settings.DefaultAppEnabled = true
	a.toggleDefaultAppEnabled()
	if h.current["txt"] != sysint.BundleID {
		t.Fatalf("expected txt re-associated, got %q", h.current["txt"])
	}

	a.restoreDefaults()
	if _, ok := a.sysint.CLIStatus(); ok {
		t.Fatal("expected restoreDefaults to remove CLI wrapper")
	}
	if a.sysint.HasFinderWorkflow(sysint.FinderKindFolder) || a.sysint.HasFinderWorkflow(sysint.FinderKindFile) {
		t.Fatal("expected restoreDefaults to remove Finder workflows")
	}
	if h.current["txt"] != "com.apple.TextEdit" {
		t.Fatalf("expected restoreDefaults to restore txt handler, got %q", h.current["txt"])
	}
}

func TestSettingsModalShowsIntegrationSection(t *testing.T) {
	a, _ := newIntegrationTestApp(t)
	a.openSettings()

	tt := ui.NewTester(a.View, 1200, 800)
	for _, want := range []string{
		locales.IntegrationSection(),
		locales.CliCommand(),
		locales.FinderFolderMenu(),
		locales.FinderFileMenu(),
		locales.DefaultAppAssoc(),
	} {
		if !tt.HasText(want) {
			t.Fatalf("settings modal missing %q", want)
		}
	}

	// Extension picker only appears when DefaultAppEnabled is turned on.
	if tt.HasText(".txt") {
		t.Fatal("extension chips should stay hidden while DefaultAppEnabled is off")
	}
	a.Settings.DefaultAppEnabled = true
	tt.Frame()
	if !tt.HasText(".txt") || !tt.HasText(".md") || !tt.HasText(".json") {
		t.Fatal("extension chips did not appear when DefaultAppEnabled is on")
	}
}

func TestHandleOpenURLQueuesPathsBeforeWindow(t *testing.T) {
	a, _ := newIntegrationTestApp(t)
	a.HandleOpenURL("zpad://open?path=%2Fusr%2Flocal%2Fopenresty&path=%2Ftmp%2Fa+b.conf")
	want := []string{"/usr/local/openresty", "/tmp/a b.conf"}
	if !slices.Equal(a.pendingOpen, want) {
		t.Fatalf("pendingOpen = %q, want %q", a.pendingOpen, want)
	}
	// Non-zpad schemes are ignored.
	a.HandleOpenURL("https://example.com/?path=/etc/passwd")
	if !slices.Equal(a.pendingOpen, want) {
		t.Fatalf("unexpected pendingOpen after foreign URL: %q", a.pendingOpen)
	}
}
