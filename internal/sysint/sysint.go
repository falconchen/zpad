// Package sysint manages OS-level integrations for ZPad: the terminal CLI
// wrapper, Finder right-click Quick Actions, and default file associations.
package sysint

import (
	"errors"
	"strings"
)

// BundleID is ZPad's macOS bundle identifier, matching mygo.json.
const BundleID = "com.xiaoz.zpad"

// FallbackTextHandler is used when restoring an extension whose previous
// default handler was not recorded.
const FallbackTextHandler = "com.apple.TextEdit"

// ErrUnsupported is returned on platforms where an integration is not
// implemented yet.
var ErrUnsupported = errors.New("not supported on this platform")

// ExtGroup groups common file extensions in the Settings UI, similar to
// Notepad++'s File Association panel.
type ExtGroup struct {
	ID   string
	Exts []string
}

// DefaultExtGroups returns the built-in extension groups offered in Settings.
func DefaultExtGroups() []ExtGroup {
	return []ExtGroup{
		{
			ID:   "text",
			Exts: []string{"txt", "md", "markdown", "log", "csv"},
		},
		{
			ID:   "config",
			Exts: []string{"json", "yaml", "yml", "toml", "xml", "ini", "conf", "env", "properties"},
		},
		{
			ID:   "web",
			Exts: []string{"html", "htm", "css", "scss", "less", "js", "ts", "jsx", "tsx", "vue", "svelte"},
		},
		{
			ID:   "code",
			Exts: []string{"go", "py", "rs", "c", "cpp", "h", "java", "swift", "rb", "php", "sh", "zsh", "sql"},
		},
	}
}

// AllPresetExts returns every extension in DefaultExtGroups in display order.
func AllPresetExts() []string {
	var all []string
	for _, g := range DefaultExtGroups() {
		all = append(all, g.Exts...)
	}
	return all
}

// IsPresetExt reports whether ext (already normalized) belongs to a built-in group.
func IsPresetExt(ext string) bool {
	for _, g := range DefaultExtGroups() {
		for _, e := range g.Exts {
			if e == ext {
				return true
			}
		}
	}
	return false
}

// NormalizeExt trims spaces and leading dots, lowercases raw, and verifies it
// is a safe filename extension such as "txt", "md", or "c++".
func NormalizeExt(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	s = strings.TrimLeft(s, ".")
	s = strings.ToLower(s)
	if s == "" {
		return "", errors.New("extension is empty")
	}
	if len(s) > 16 {
		return "", errors.New("extension is too long")
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
		if i > 0 {
			ok = ok || c == '_' || c == '-' || c == '+'
		}
		if !ok {
			return "", errors.New("extension contains invalid characters")
		}
	}
	return s, nil
}

// HandlerBackend abstracts OS default-application queries and updates so tests
// can exercise association logic without modifying host LaunchServices state.
type HandlerBackend interface {
	RegisterApp(appBundle string) error
	GetDefaultHandler(ext string) (string, error)
	SetDefaultHandler(ext, bundleID string) error
}

// Status summarizes the live state of ZPad's system integrations on disk.
type Status struct {
	Supported             bool
	CLIInstalled          bool
	CLIPath               string
	FinderFolderInstalled bool
	FinderFileInstalled   bool
}
