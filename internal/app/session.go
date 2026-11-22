package app

import (
	"log"
	"os"

	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/files"
	"github.com/helloxz/zpad/internal/session"
	"github.com/helloxz/zpad/locales"
)

// RestoreSession loads the workspace of the last launch, when the user
// asked for it.
func (a *App) RestoreSession() {
	if !a.Settings.RestoreSession || a.SettingsDir == "" || !session.Exists(a.SettingsDir) {
		return
	}
	st, err := session.Load(a.SettingsDir)
	if err != nil {
		log.Printf("load session: %v", err)
		return
	}
	a.restoreSession(st)
}

func (a *App) restoreSession(st session.State) {
	if st.Root != "" {
		if info, err := os.Stat(st.Root); err == nil && info.IsDir() {
			a.root = st.Root
			a.tree = files.New(st.Root, a.Settings.ShowHidden)
			a.treeGeneration++
		}
	}
	if st.SidebarWidth >= 160 {
		a.Settings.SidebarWidth = st.SidebarWidth
	}
	a.Settings.SidebarOpen = st.SidebarOpen
	a.split = st.Split
	a.splitVertical = st.SplitVertical
	if st.SplitWidth >= 120 {
		a.splitWidth = st.SplitWidth
	}
	if st.SplitHeight >= 120 {
		a.splitHeight = st.SplitHeight
	}
	a.activePane = st.ActivePane

	docs := map[string]*Document{}
	missing := 0
	loadPane := func(ps session.PaneState) Pane {
		p := Pane{active: -1}
		for _, ts := range ps.Tabs {
			doc := docs[ts.Path]
			if doc == nil {
				d, err := loadDocument(ts.Path, a.Settings.LargeFileMB)
				if err != nil {
					missing++
					continue
				}
				doc = d
				docs[ts.Path] = doc
			}
			// Do not show one document twice in one pane.
			duplicate := false
			for _, tab := range p.tabs {
				if tab.doc == doc {
					duplicate = true
					break
				}
			}
			if duplicate {
				continue
			}
			y := float64(ts.ScrollY)
			p.tabs = append(p.tabs, &Tab{
				doc:           doc,
				state:         &editor.State{Buffer: doc.Buffer, ReadOnly: doc.ReadOnly},
				pendingScroll: &y,
			})
		}
		p.active = ps.Active
		if p.active < 0 || p.active >= len(p.tabs) {
			if len(p.tabs) == 0 {
				p.active = -1
			} else {
				p.active = len(p.tabs) - 1
			}
		}
		return p
	}
	a.left = loadPane(st.Left)
	a.right = loadPane(st.Right)

	if a.split && (len(a.left.tabs) == 0 || len(a.right.tabs) == 0) {
		if len(a.left.tabs) == 0 {
			a.collapseSplit(0)
		} else {
			a.collapseSplit(1)
		}
	}
	if !a.split {
		a.activePane = 0
	}
	if a.activePane == 0 && len(a.left.tabs) == 0 && len(a.right.tabs) > 0 {
		a.activePane = 1
	}
	if a.activePane == 1 && len(a.right.tabs) == 0 && len(a.left.tabs) > 0 {
		a.activePane = 0
	}
	if a.root != "" {
		a.sidebarTab = sidebarWorkspace
	}
	a.focusEditor = len(a.left.tabs)+len(a.right.tabs) > 0
	if tab := a.pane(a.activePane).activeTab(); tab != nil {
		a.revealPane, a.revealDoc = a.activePane, tab.doc.ID
	}
	if missing > 0 {
		a.errMsg = locales.MissingFiles(missing)
	}
}

// saveSession stores the workspace for the next launch.
func (a *App) saveSession() {
	if a.SettingsDir == "" {
		return
	}
	st := session.State{
		Root:          a.root,
		Split:         a.split,
		SplitWidth:    a.splitWidth,
		SplitVertical: a.splitVertical,
		SplitHeight:   a.splitHeight,
		ActivePane:    a.activePane,
		SidebarOpen:   a.Settings.SidebarOpen,
		SidebarWidth:  a.Settings.SidebarWidth,
		Left:          a.paneState(a.left),
		Right:         a.paneState(a.right),
	}
	if err := session.Save(a.SettingsDir, st); err != nil {
		log.Printf("save session: %v", err)
	}
}

func (a *App) paneState(p Pane) session.PaneState {
	ps := session.PaneState{Active: p.active}
	for _, tab := range p.tabs {
		if tab.doc.Path == "" {
			continue
		}
		ps.Tabs = append(ps.Tabs, session.TabState{
			Path:    tab.doc.Path,
			ScrollY: float32(tab.state.ScrollY()),
		})
	}
	return ps
}
