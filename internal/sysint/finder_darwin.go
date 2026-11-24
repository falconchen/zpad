//go:build darwin

package sysint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FinderSyncBundleID is the bundle identifier of ZPad's embedded Finder Sync
// App Extension (ZPad.app/Contents/PlugIns/ZPadFinderSync.appex).
const FinderSyncBundleID = "com.xiaoz.zpad.FinderSync"

// FinderKind identifies which Finder context menu capability is enabled.
type FinderKind string

const (
	FinderKindFolder FinderKind = "folder"
	FinderKindFile   FinderKind = "file"
)

// FinderMenuOptions configures ZPad's Finder Sync extension.
type FinderMenuOptions struct {
	Folder                 bool
	File                   bool
	OpenCurrentFolderTitle string
	OpenFolderTitle        string
	OpenFileTitle          string
}

type finderSyncConfig struct {
	EnableFolder           bool   `json:"enableFolder"`
	EnableFile             bool   `json:"enableFile"`
	OpenCurrentFolderTitle string `json:"openCurrentFolderTitle"`
	OpenFolderTitle        string `json:"openFolderTitle"`
	OpenFileTitle          string `json:"openFileTitle"`
}

const legacyWorkflowMarker = ".zpad-kind"

// FinderSyncAppexPath returns the expected path to the embedded FinderSync
// extension inside m.AppBundle.
func (m *Manager) FinderSyncAppexPath() string {
	if m.AppBundle == "" {
		return ""
	}
	return filepath.Join(m.AppBundle, "Contents", "PlugIns", "ZPadFinderSync.appex")
}

// HasFinderWorkflow reports whether Finder context menu integration for kind
// is currently enabled.
func (m *Manager) HasFinderWorkflow(kind FinderKind) bool {
	cfg, ok := m.readFinderSyncConfig()
	if !ok {
		return false
	}
	if !m.SkipPBS && !m.isFinderSyncPluginEnabled() {
		return false
	}
	switch kind {
	case FinderKindFolder:
		return cfg.EnableFolder
	case FinderKindFile:
		return cfg.EnableFile
	default:
		return false
	}
}

// SyncFinderMenu configures ZPad's native FinderSync extension (supporting both
// blank-space container right-click and item right-click) and cleans up any
// legacy Automator .workflow bundles left over from earlier builds.
func (m *Manager) SyncFinderMenu(opts FinderMenuOptions) error {
	m.cleanupLegacyWorkflows()

	cfgPath := m.FinderSyncConfigPath
	if cfgPath == "" {
		if !opts.Folder && !opts.File {
			return nil
		}
		return errors.New("FinderSync configuration path is unavailable")
	}

	curTitle := strings.TrimSpace(opts.OpenCurrentFolderTitle)
	if curTitle == "" {
		curTitle = "Open Current Folder in ZPad"
	}
	folderTitle := strings.TrimSpace(opts.OpenFolderTitle)
	if folderTitle == "" {
		folderTitle = "Open Folder in ZPad"
	}
	fileTitle := strings.TrimSpace(opts.OpenFileTitle)
	if fileTitle == "" {
		fileTitle = "Open with ZPad"
	}

	cfg := finderSyncConfig{
		EnableFolder:           opts.Folder,
		EnableFile:             opts.File,
		OpenCurrentFolderTitle: curTitle,
		OpenFolderTitle:        folderTitle,
		OpenFileTitle:          fileTitle,
	}
	if err := m.writeFinderSyncConfig(cfgPath, cfg); err != nil {
		return err
	}

	if m.SkipPBS {
		return nil
	}

	if opts.Folder || opts.File {
		appex := m.FinderSyncAppexPath()
		if _, err := os.Stat(appex); err != nil {
			return fmt.Errorf("ZPadFinderSync.appex not found at %s (rebuild with scripts/build-macos.sh)", appex)
		}
		if out, err := exec.Command("/usr/bin/pluginkit", "-a", appex).CombinedOutput(); err != nil {
			return fmt.Errorf("register Finder extension failed: %w (%s)", err, strings.TrimSpace(string(out)))
		}
		if out, err := exec.Command("/usr/bin/pluginkit", "-e", "use", "-i", FinderSyncBundleID).CombinedOutput(); err != nil {
			return fmt.Errorf("enable Finder extension failed: %w (%s)", err, strings.TrimSpace(string(out)))
		}
		return nil
	}

	_ = exec.Command("/usr/bin/pluginkit", "-e", "ignore", "-i", FinderSyncBundleID).Run()
	return nil
}

// UninstallAllFinderWorkflows disables ZPad's FinderSync extension and removes
// any legacy Automator workflows.
func (m *Manager) UninstallAllFinderWorkflows() error {
	return m.SyncFinderMenu(FinderMenuOptions{})
}

func (m *Manager) readFinderSyncConfig() (finderSyncConfig, bool) {
	if m.FinderSyncConfigPath == "" {
		return finderSyncConfig{}, false
	}
	data, err := os.ReadFile(m.FinderSyncConfigPath)
	if err != nil {
		return finderSyncConfig{}, false
	}
	var cfg finderSyncConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return finderSyncConfig{}, false
	}
	return cfg, true
}

func (m *Manager) writeFinderSyncConfig(path string, cfg finderSyncConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (m *Manager) isFinderSyncPluginEnabled() bool {
	out, err := exec.Command("/usr/bin/pluginkit", "-m", "-i", FinderSyncBundleID).Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, FinderSyncBundleID) && strings.HasPrefix(strings.TrimSpace(line), "+") {
			return true
		}
	}
	return false
}

func (m *Manager) cleanupLegacyWorkflows() {
	if m.ServicesDir == "" {
		return
	}
	entries, err := os.ReadDir(m.ServicesDir)
	if err != nil {
		return
	}
	removed := false
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".workflow") {
			continue
		}
		wfPath := filepath.Join(m.ServicesDir, entry.Name())
		if _, err := os.Stat(filepath.Join(wfPath, "Contents", legacyWorkflowMarker)); err != nil {
			continue
		}
		if err := os.RemoveAll(wfPath); err == nil {
			removed = true
		}
	}
	if removed && !m.SkipPBS {
		const pbs = "/System/Library/CoreServices/pbs"
		if _, err := os.Stat(pbs); err == nil {
			_ = exec.Command(pbs, "-flush").Run()
			_ = exec.Command(pbs, "-update").Run()
		}
	}
}
