//go:build darwin

package sysint

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type mockHandlers struct {
	registered string
	current    map[string]string
}

func (m *mockHandlers) RegisterApp(appBundle string) error {
	m.registered = appBundle
	return nil
}

func (m *mockHandlers) GetDefaultHandler(ext string) (string, error) {
	return m.current[ext], nil
}

func (m *mockHandlers) SetDefaultHandler(ext, bundleID string) error {
	if m.current == nil {
		m.current = map[string]string{}
	}
	m.current[ext] = bundleID
	return nil
}

func TestNormalizeExt(t *testing.T) {
	valid := map[string]string{
		"txt":       "txt",
		".MD":       "md",
		"  ..json ": "json",
		"c++":       "c++",
		"file_1-a":  "file_1-a",
	}
	for in, want := range valid {
		got, err := NormalizeExt(in)
		if err != nil {
			t.Fatalf("NormalizeExt(%q) failed: %v", in, err)
		}
		if got != want {
			t.Fatalf("NormalizeExt(%q) = %q, want %q", in, got, want)
		}
	}
	for _, bad := range []string{"", "...", "a/b", "a b", "-foo", "abcdefghijklmnopq"} {
		if _, err := NormalizeExt(bad); err == nil {
			t.Fatalf("NormalizeExt(%q) succeeded, want error", bad)
		}
	}
}

func TestCLIInstallUninstallAndRefuseForeignFile(t *testing.T) {
	binDir := t.TempDir()
	mgr := &Manager{
		AppBundle: "/Applications/ZPad.app",
		CLIDirs:   []string{binDir},
		SkipPBS:   true,
	}

	if _, ok := mgr.CLIStatus(); ok {
		t.Fatal("CLI should not be installed initially")
	}

	installedPath, err := mgr.InstallCLI()
	if err != nil {
		t.Fatalf("InstallCLI failed: %v", err)
	}
	wantPath := filepath.Join(binDir, "zpad")
	if installedPath != wantPath {
		t.Fatalf("installed path = %q, want %q", installedPath, wantPath)
	}
	info, err := os.Stat(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("zpad script is not executable: mode %o", info.Mode().Perm())
	}
	if out, err := exec.Command("/bin/sh", "-n", wantPath).CombinedOutput(); err != nil {
		t.Fatalf("generated CLI script has shell syntax error: %v (%s)", err, out)
	}
	if gotPath, ok := mgr.CLIStatus(); !ok || gotPath != wantPath {
		t.Fatalf("CLIStatus = (%q, %v), want (%q, true)", gotPath, ok, wantPath)
	}

	if err := mgr.UninstallCLI(); err != nil {
		t.Fatalf("UninstallCLI failed: %v", err)
	}
	if _, err := os.Stat(wantPath); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, stat err = %v", wantPath, err)
	}

	// Must not overwrite a foreign script not managed by ZPad.
	if err := os.WriteFile(wantPath, []byte("#!/bin/sh\necho custom\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.InstallCLI(); err == nil {
		t.Fatal("InstallCLI overwrote a non-ZPad file")
	}
	// UninstallCLI must also leave foreign files untouched.
	if err := mgr.UninstallCLI(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("UninstallCLI removed a foreign file: %v", err)
	}
}

func TestFinderMenuConfigAndLegacyCleanup(t *testing.T) {
	servicesDir := t.TempDir()
	cfgPath := filepath.Join(t.TempDir(), "zpad-sync.json")

	// Seed a legacy Automator workflow managed by an earlier ZPad build.
	legacyDir := filepath.Join(servicesDir, "使用 ZPad 打开.workflow", "Contents")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, legacyWorkflowMarker), []byte("both\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := &Manager{
		AppBundle:            "/Applications/ZPad.app",
		ServicesDir:          servicesDir,
		FinderSyncConfigPath: cfgPath,
		SkipPBS:              true,
	}

	if err := mgr.SyncFinderMenu(FinderMenuOptions{
		Folder:                 true,
		File:                   false,
		OpenCurrentFolderTitle: "使用 ZPad 打开当前目录",
		OpenFolderTitle:        "使用 ZPad 打开目录",
		OpenFileTitle:          "使用 ZPad 打开",
	}); err != nil {
		t.Fatalf("SyncFinderMenu(folder only): %v", err)
	}
	if _, err := os.Stat(filepath.Join(servicesDir, "使用 ZPad 打开.workflow")); !os.IsNotExist(err) {
		t.Fatal("legacy workflow was not cleaned up")
	}
	if !mgr.HasFinderWorkflow(FinderKindFolder) || mgr.HasFinderWorkflow(FinderKindFile) {
		t.Fatal("expected only folder menu enabled")
	}

	if err := mgr.SyncFinderMenu(FinderMenuOptions{
		Folder:                 true,
		File:                   true,
		OpenCurrentFolderTitle: "Open Current Folder in ZPad",
		OpenFolderTitle:        "Open Folder in ZPad",
		OpenFileTitle:          "Open with ZPad",
	}); err != nil {
		t.Fatalf("SyncFinderMenu(both): %v", err)
	}
	if !mgr.HasFinderWorkflow(FinderKindFolder) || !mgr.HasFinderWorkflow(FinderKindFile) {
		t.Fatal("expected both folder and file menus enabled")
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg finderSyncConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.OpenCurrentFolderTitle != "Open Current Folder in ZPad" || cfg.OpenFileTitle != "Open with ZPad" {
		t.Fatalf("unexpected saved config: %+v", cfg)
	}

	if err := mgr.UninstallAllFinderWorkflows(); err != nil {
		t.Fatal(err)
	}
	if mgr.HasFinderWorkflow(FinderKindFolder) || mgr.HasFinderWorkflow(FinderKindFile) {
		t.Fatal("expected both menus disabled after uninstall")
	}
}

func TestDefaultAppAssociateAndRestore(t *testing.T) {
	mock := &mockHandlers{
		current: map[string]string{
			"txt": "com.apple.TextEdit",
			"md":  "com.microsoft.VSCode",
		},
	}
	mgr := &Manager{
		AppBundle: "/Applications/ZPad.app",
		Handlers:  mock,
		SkipPBS:   true,
	}

	prev := map[string]string{}
	ext, err := mgr.AssociateExt(".TXT", prev)
	if err != nil || ext != "txt" {
		t.Fatalf("AssociateExt(.TXT) = (%q, %v)", ext, err)
	}
	if mock.registered != "/Applications/ZPad.app" {
		t.Fatalf("RegisterApp called with %q", mock.registered)
	}
	if mock.current["txt"] != BundleID {
		t.Fatalf("txt handler = %q, want %q", mock.current["txt"], BundleID)
	}
	if prev["txt"] != "com.apple.TextEdit" {
		t.Fatalf("saved previous handler = %q, want com.apple.TextEdit", prev["txt"])
	}

	// Re-associating while already set to ZPad must preserve the original backup.
	if _, err := mgr.AssociateExt("txt", prev); err != nil {
		t.Fatal(err)
	}
	if prev["txt"] != "com.apple.TextEdit" {
		t.Fatalf("previous handler overwritten: %q", prev["txt"])
	}

	// Restoring reverts to TextEdit and clears the backup map entry.
	if err := mgr.RestoreExt("txt", prev); err != nil {
		t.Fatal(err)
	}
	if mock.current["txt"] != "com.apple.TextEdit" {
		t.Fatalf("restored txt handler = %q, want com.apple.TextEdit", mock.current["txt"])
	}
	if _, exists := prev["txt"]; exists {
		t.Fatal("expected prev[txt] to be deleted after restore")
	}

	// Restoring an extension with no backup falls back to TextEdit.
	if err := mgr.RestoreExt("log", prev); err != nil {
		t.Fatal(err)
	}
	if mock.current["log"] != FallbackTextHandler {
		t.Fatalf("fallback handler = %q, want %q", mock.current["log"], FallbackTextHandler)
	}
}

func TestDarwinLaunchServicesReadOnly(t *testing.T) {
	mgr := DefaultManager()
	handler, err := mgr.GetDefaultHandler("txt")
	if err != nil {
		t.Fatalf("GetDefaultHandler(txt) failed: %v", err)
	}
	if handler == "" {
		t.Fatal("expected a non-empty macOS default handler for .txt")
	}
}
