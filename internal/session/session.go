// Package session stores the workspace layout restored at the next launch.
package session

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/helloxz/zpad/internal/atomicfile"
)

const fileName = "session.json"

// TabState is one tab of a pane: the file it shows and where it was
// scrolled.
type TabState struct {
	Path    string  `json:"path"`
	ScrollY float32 `json:"scrollY"`
	// Draft names the recovery copy of the document's unsaved text, when
	// keeping unsaved content is on; Untitled is its number for a document
	// without a file.
	Draft    string `json:"draft,omitempty"`
	Untitled int    `json:"untitled,omitempty"`
	// Language is the language an untitled document was given in the status
	// bar, by name; "" for none. A tab with a path takes its language from
	// the file name.
	Language string `json:"language,omitempty"`
}

// PaneState is the tabs of one pane and the one chosen.
type PaneState struct {
	Tabs   []TabState `json:"tabs"`
	Active int        `json:"active"`
}

// State is the workspace of the last launch.
type State struct {
	Root          string    `json:"root"`
	Split         bool      `json:"split"`
	SplitVertical bool      `json:"splitVertical"`
	SplitWidth    float32   `json:"splitWidth"`
	SplitHeight   float32   `json:"splitHeight"`
	ActivePane    int       `json:"activePane"`
	Left          PaneState `json:"left"`
	Right         PaneState `json:"right"`
	SidebarOpen   bool      `json:"sidebarOpen"`
	SidebarWidth  float32   `json:"sidebarWidth"`
}

// Default is the layout of a first launch.
func Default() State {
	return State{
		SplitWidth:   600,
		SplitHeight:  500,
		SidebarWidth: 240,
	}
}

// Exists reports whether dir has a session file.
func Exists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, fileName))
	return err == nil
}

// Load reads dir/session.json. A missing or corrupt file returns the
// defaults; a corrupt one is kept as session.json.bak.
func Load(dir string) (State, error) {
	s := Default()
	path := filepath.Join(dir, fileName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		_ = os.WriteFile(path+".bak", data, 0o600)
		return Default(), nil
	}
	s.normalize()
	return s, nil
}

// Save writes s atomically to dir/session.json.
func Save(dir string, s State) error {
	s.normalize()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return atomicfile.Write(dir, fileName, data)
}

func (s *State) normalize() {
	if s.SplitWidth < 120 {
		s.SplitWidth = 600
	}
	if s.SplitHeight < 120 {
		s.SplitHeight = 500
	}
	if s.SidebarWidth < 120 {
		s.SidebarWidth = 240
	}
	if s.ActivePane < 0 || s.ActivePane > 1 {
		s.ActivePane = 0
	}
	s.Left = normalizePane(s.Left)
	s.Right = normalizePane(s.Right)
}

func normalizePane(p PaneState) PaneState {
	if p.Active < 0 || p.Active >= len(p.Tabs) {
		if len(p.Tabs) == 0 {
			p.Active = 0
		} else {
			p.Active = len(p.Tabs) - 1
		}
	}
	return p
}
