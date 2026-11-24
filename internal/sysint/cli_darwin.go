//go:build darwin

package sysint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const cliMarker = "# Managed by ZPad"

// Manager configures macOS integrations. Fields may be overridden in tests to
// point at temporary directories and a mock HandlerBackend.
type Manager struct {
	AppBundle            string
	CLIDirs              []string
	ServicesDir          string
	FinderSyncConfigPath string
	SkipPBS              bool
	Handlers             HandlerBackend

	appRegistered bool
}

// DefaultManager returns a Manager targeting the current macOS user's
// environment.
func DefaultManager() *Manager {
	home, _ := os.UserHomeDir()
	servicesDir := ""
	syncCfgPath := ""
	if home != "" {
		servicesDir = filepath.Join(home, "Library", "Services")
		syncCfgPath = filepath.Join(
			home, "Library", "Containers", FinderSyncBundleID,
			"Data", "Library", "Application Support", "zpad-sync.json",
		)
	}
	return &Manager{
		AppBundle:            ResolveAppBundle(),
		CLIDirs:              defaultCLIDirs(home),
		ServicesDir:          servicesDir,
		FinderSyncConfigPath: syncCfgPath,
		Handlers:             darwinLaunchServices{},
	}
}

func defaultCLIDirs(home string) []string {
	dirs := []string{"/opt/homebrew/bin", "/usr/local/bin"}
	if home != "" {
		dirs = append(dirs, filepath.Join(home, ".local", "bin"))
	}
	return dirs
}

// ResolveAppBundle finds the ZPad.app bundle to target from CLI and Finder
// integrations. When running from a packaged .app outside temporary dev
// directories it uses that bundle; otherwise it prefers /Applications/ZPad.app.
func ResolveAppBundle() string {
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		if bundle := appBundleFromExecutable(exe); bundle != "" && !isTempPath(bundle) {
			return bundle
		}
	}
	if info, err := os.Stat("/Applications/ZPad.app"); err == nil && info.IsDir() {
		return "/Applications/ZPad.app"
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		userApp := filepath.Join(home, "Applications", "ZPad.app")
		if info, err := os.Stat(userApp); err == nil && info.IsDir() {
			return userApp
		}
	}
	return "/Applications/ZPad.app"
}

func appBundleFromExecutable(exe string) string {
	dir := filepath.Dir(exe)
	if filepath.Base(dir) != "MacOS" {
		return ""
	}
	contents := filepath.Dir(dir)
	if filepath.Base(contents) != "Contents" {
		return ""
	}
	bundle := filepath.Dir(contents)
	if !strings.HasSuffix(bundle, ".app") {
		return ""
	}
	return bundle
}

func isTempPath(p string) bool {
	return strings.HasPrefix(p, "/tmp/") ||
		strings.HasPrefix(p, "/private/tmp/") ||
		strings.HasPrefix(p, "/var/folders/") ||
		strings.HasPrefix(p, "/private/var/folders/")
}

// Supported reports whether system integrations are available on this platform.
func (m *Manager) Supported() bool {
	return true
}

// Status inspects the filesystem for installed CLI and Finder integrations.
func (m *Manager) Status() Status {
	cliPath, cliOK := m.CLIStatus()
	return Status{
		Supported:             true,
		CLIInstalled:          cliOK,
		CLIPath:               cliPath,
		FinderFolderInstalled: m.HasFinderWorkflow(FinderKindFolder),
		FinderFileInstalled:   m.HasFinderWorkflow(FinderKindFile),
	}
}

// CLIScript returns the POSIX shell wrapper installed as `zpad`.
func CLIScript(appBundle string) string {
	escaped := shellSingleQuote(appBundle)
	return fmt.Sprintf(`#!/bin/sh
%s
APP=%s
BIN="$APP/Contents/MacOS/ZPad"

case "${1:-}" in
  -h|--help)
    echo "Usage: zpad [path ...]"
    echo "Open files or folders in ZPad."
    exit 0
    ;;
  -v|--version)
    /usr/bin/defaults read "$APP/Contents/Info.plist" CFBundleShortVersionString 2>/dev/null || echo "ZPad"
    exit 0
    ;;
esac

if [ ! -x "$BIN" ]; then
  echo "zpad: ZPad.app not found at $APP" >&2
  exit 1
fi

abs_path() {
  p="$1"
  if [ -d "$p" ]; then
    (cd "$p" 2>/dev/null && pwd -P) || printf '%%s\n' "$p"
  else
    dir=$(dirname "$p")
    base=$(basename "$p")
    if [ -d "$dir" ]; then
      printf '%%s/%%s\n' "$(cd "$dir" 2>/dev/null && pwd -P)" "$base"
    else
      case "$p" in
        /*) printf '%%s\n' "$p" ;;
        *) printf '%%s/%%s\n' "$(pwd -P)" "$p" ;;
      esac
    fi
  fi
}

if [ "$#" -gt 0 ]; then
  first=1
  for arg in "$@"; do
    if [ "$first" -eq 1 ]; then
      set --
      first=0
    fi
    case "$arg" in
      -*) set -- "$@" "$arg" ;;
      *) set -- "$@" "$(abs_path "$arg")" ;;
    esac
  done
fi

if /usr/bin/pgrep -f "^${BIN}" >/dev/null 2>&1; then
  "$BIN" "$@" >/dev/null 2>&1
  /usr/bin/open -a "$APP"
elif [ "$#" -gt 0 ]; then
  /usr/bin/open -a "$APP" --args "$@"
else
  /usr/bin/open -a "$APP"
fi
`, cliMarker, escaped)
}

func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func isManagedCLIFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), cliMarker)
}

// CLIStatus reports whether a ZPad-managed `zpad` command is currently installed
// and where it lives (or where InstallCLI would place it).
func (m *Manager) CLIStatus() (string, bool) {
	for _, dir := range m.CLIDirs {
		p := filepath.Join(dir, "zpad")
		if isManagedCLIFile(p) {
			return p, true
		}
	}
	if dir, err := m.pickWritableCLIDir(); err == nil {
		return filepath.Join(dir, "zpad"), false
	}
	if len(m.CLIDirs) > 0 {
		return filepath.Join(m.CLIDirs[0], "zpad"), false
	}
	return "", false
}

// InstallCLI writes the `zpad` wrapper into the first writable candidate bin
// directory (updating an existing ZPad-managed wrapper in place).
func (m *Manager) InstallCLI() (string, error) {
	if m.AppBundle == "" {
		return "", errors.New("ZPad.app bundle path is unknown")
	}
	var targetDir string
	// Update in place if already installed.
	for _, dir := range m.CLIDirs {
		if isManagedCLIFile(filepath.Join(dir, "zpad")) {
			targetDir = dir
			break
		}
	}
	if targetDir == "" {
		var err error
		targetDir, err = m.pickWritableCLIDir()
		if err != nil {
			return "", err
		}
	}
	target := filepath.Join(targetDir, "zpad")
	if info, err := os.Lstat(target); err == nil {
		if info.IsDir() || !isManagedCLIFile(target) {
			return "", fmt.Errorf("%s already exists and is not managed by ZPad", target)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := writeExecutableAtomically(targetDir, "zpad", []byte(CLIScript(m.AppBundle))); err != nil {
		return "", err
	}
	return target, nil
}

// UninstallCLI removes every ZPad-managed `zpad` wrapper from candidate dirs.
func (m *Manager) UninstallCLI() error {
	var firstErr error
	for _, dir := range m.CLIDirs {
		p := filepath.Join(dir, "zpad")
		if !isManagedCLIFile(p) {
			continue
		}
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *Manager) pickWritableCLIDir() (string, error) {
	pathEntries := splitPathEnv(os.Getenv("PATH"))
	// Prefer a writable directory that is already on PATH.
	for _, dir := range m.CLIDirs {
		if pathEntries[filepath.Clean(dir)] && canWriteDir(dir, false) {
			return dir, nil
		}
	}
	// Otherwise try any existing writable candidate dir, or create ~/.local/bin.
	for i, dir := range m.CLIDirs {
		allowCreate := i == len(m.CLIDirs)-1
		if canWriteDir(dir, allowCreate) {
			return dir, nil
		}
	}
	return "", fmt.Errorf("no writable bin directory found among %s", strings.Join(m.CLIDirs, ", "))
}

func splitPathEnv(v string) map[string]bool {
	out := map[string]bool{}
	for _, part := range filepath.SplitList(v) {
		if part != "" {
			out[filepath.Clean(part)] = true
		}
	}
	return out
}

func canWriteDir(dir string, create bool) bool {
	info, err := os.Stat(dir)
	if err != nil {
		if !create || !errors.Is(err, os.ErrNotExist) {
			return false
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return false
		}
	} else if !info.IsDir() {
		return false
	}
	f, err := os.CreateTemp(dir, ".zpad-probe-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

func writeExecutableAtomically(dir, name string, data []byte) error {
	f, err := os.CreateTemp(dir, "."+name+".tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmp)
		}
	}()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Chmod(0o755); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, filepath.Join(dir, name)); err != nil {
		return err
	}
	cleanup = false
	return nil
}
