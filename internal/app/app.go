// Package app holds ZPad's state and builds its window.
package app

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/files"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

// Tab is one open document and the editor state showing it.
type Tab struct {
	doc           *Document
	state         *editor.State
	pendingScroll *float64
}

// Pane is one side of the window: its tabs and the one chosen.
type Pane struct {
	tabs   []*Tab
	active int
}

func (p *Pane) activeTab() *Tab {
	if p.active < 0 || p.active >= len(p.tabs) {
		return nil
	}
	return p.tabs[p.active]
}

// App is the state the window shows.
type App struct {
	Settings    settings.Settings
	SettingsDir string
	ThemeIndex  int // 0 light, 1 dark, 2 system
	Dark        bool

	win     *mygo.Window
	running bool
	errMsg  string

	root           string
	tree           *files.Tree
	treeState      ui.OutlineState[string]
	treeSelected   int
	treeGeneration uint64

	left       Pane
	right      Pane
	split      bool
	splitWidth float32
	activePane int // 0 left, 1 right

	focusEditor      bool
	untitledSeq      int
	busy             bool
	loading          int
	externalChecking bool
	lastDirty        bool
	pendingOpen      []string

	showSettings bool
	fontSize     float64
	lineHeight   float64
	fontFamilies []string
	pendingToast string

	closePromptOpen bool
	allowClose      bool
	allowQuit       bool
}

// New returns an app with saved settings and the directory to save them
// in.
func New(s settings.Settings, settingsDir string) *App {
	idx := 2
	switch s.Theme {
	case "light":
		idx = 0
	case "dark":
		idx = 1
	}
	return &App{
		Settings:     s,
		SettingsDir:  settingsDir,
		ThemeIndex:   idx,
		Dark:         s.Theme == "dark",
		treeSelected: -1,
		left:         Pane{active: -1},
		right:        Pane{active: -1},
		splitWidth:   600,
		fontSize:     float64(s.FontSize),
		lineHeight:   float64(s.LineHeight),
	}
}

// SetThemeSource applies the saved appearance before App.Run.
func SetThemeSource(mode string) {
	switch mode {
	case "light":
		mygo.Theme.SetSource(mygo.ThemeLight)
	case "dark":
		mygo.Theme.SetSource(mygo.ThemeDark)
	default:
		mygo.Theme.SetSource(mygo.ThemeSystem)
	}
}

// Prepare marks the app as running and resolves the system appearance.
// Call it on the main thread once the app is ready, before the window.
func (a *App) Prepare() {
	a.running = true
	if a.Settings.Theme == "system" {
		a.Dark = mygo.Theme.IsDark()
	}
}

// Attach remembers the window, installs the menu and takes dropped files.
func (a *App) Attach(win *mygo.Window) {
	a.win = win
	win.OnClose(a.onWindowClose)
	win.OnFocus(func() { a.checkExternalChanges() })
	win.OnFileDrop(func(e *mygo.FileDropEvent) {
		a.OpenPaths(e.Paths)
	})
	mygo.Theme.OnUpdated(func() {
		if a.Settings.Theme == "system" {
			a.Dark = mygo.Theme.IsDark()
			a.update()
		}
	})
	mygo.App.OnLocaleChanged(func(string) {
		a.InstallMenu()
		a.updateTitle()
		a.update()
	})
	a.InstallMenu()
	a.updateTitle()
	if len(a.pendingOpen) > 0 {
		paths := a.pendingOpen
		a.pendingOpen = nil
		go a.OpenPaths(paths)
	}
}

// HandleOpenFile opens a file the system asks the app to show. Before the
// window exists, it keeps the path for Attach.
func (a *App) HandleOpenFile(path string) {
	if a.win == nil {
		a.pendingOpen = append(a.pendingOpen, path)
		return
	}
	a.OpenPaths([]string{path})
}

// SetError shows an error in the status bar.
func (a *App) SetError(msg string) {
	a.errMsg = msg
}

// SetToast shows a message near the bottom of the window once.
func (a *App) SetToast(msg string) {
	a.pendingToast = msg
	a.update()
}

// onUI runs fn on the main thread, or at once before a window exists.
func (a *App) onUI(fn func()) {
	if a.win == nil {
		fn()
		return
	}
	a.win.Update(fn)
}

func (a *App) update() {
	if a.win != nil {
		a.win.Update(func() {})
	}
}

func (a *App) pane(i int) *Pane {
	if i == 1 {
		return &a.right
	}
	return &a.left
}

func (a *App) activePaneObj() *Pane { return a.pane(a.activePane) }

func (a *App) activeTab() *Tab { return a.activePaneObj().activeTab() }

func (a *App) setActivePane(i int) {
	if i != 0 && (i != 1 || !a.split) {
		return
	}
	if a.activePane == i {
		return
	}
	a.activePane = i
	a.focusEditor = true
	a.updateTitle()
	a.update()
}

func (a *App) setDocReadOnly(doc *Document, readOnly bool) {
	doc.ReadOnly = readOnly
	for _, p := range []*Pane{&a.left, &a.right} {
		for _, tab := range p.tabs {
			if tab.doc == doc {
				tab.state.ReadOnly = readOnly
			}
		}
	}
}

func (a *App) findDocument(path string) *Document {
	for _, p := range []*Pane{&a.left, &a.right} {
		for _, tab := range p.tabs {
			if tab.doc.Path == path {
				return tab.doc
			}
		}
	}
	return nil
}

func (a *App) addDocument(doc *Document) {
	a.errMsg = ""
	if doc.Path != "" {
		if existing := a.findDocument(doc.Path); existing != nil {
			doc = existing
		}
	}
	p := a.activePaneObj()
	for i, tab := range p.tabs {
		if tab.doc == doc {
			p.active = i
			a.focusEditor = true
			a.updateTitle()
			return
		}
	}
	st := &editor.State{Buffer: doc.Buffer, ReadOnly: doc.ReadOnly}
	p.tabs = append(p.tabs, &Tab{doc: doc, state: st})
	p.active = len(p.tabs) - 1
	a.focusEditor = true
	a.updateTitle()
}

func (a *App) newUntitledDocument() *Document {
	a.untitledSeq++
	doc := newUntitled(fmt.Sprintf("untitled-%d", a.untitledSeq), a.Settings.DefaultEncoding)
	doc.Buffer.SetLineEnding(a.Settings.DefaultEOL)
	return doc
}

func (a *App) removeTab(paneIdx, i int) {
	p := a.pane(paneIdx)
	if i < 0 || i >= len(p.tabs) {
		return
	}
	p.tabs = append(p.tabs[:i], p.tabs[i+1:]...)
	switch {
	case len(p.tabs) == 0:
		p.active = -1
	case p.active > i:
		p.active--
	case p.active == i && p.active >= len(p.tabs):
		p.active = len(p.tabs) - 1
	}

	if len(p.tabs) == 0 && a.split {
		other := 1 - paneIdx
		if len(a.pane(other).tabs) > 0 {
			a.activePane = other
			a.focusEditor = true
		}
	}
	if paneIdx == a.activePane {
		a.focusEditor = true
	}
	a.updateTitle()
}

func (a *App) toggleSplit() {
	if !a.split {
		a.split = true
		if tab := a.activeTab(); tab != nil {
			found := false
			for i, rt := range a.right.tabs {
				if rt.doc == tab.doc {
					a.right.active = i
					found = true
					break
				}
			}
			if !found {
				a.right.tabs = append(a.right.tabs, &Tab{
					doc:   tab.doc,
					state: &editor.State{Buffer: tab.doc.Buffer, ReadOnly: tab.doc.ReadOnly},
				})
				a.right.active = len(a.right.tabs) - 1
			}
		}
		a.activePane = 1
	} else {
		for _, rt := range a.right.tabs {
			duplicate := false
			for _, lt := range a.left.tabs {
				if lt.doc == rt.doc {
					duplicate = true
					break
				}
			}
			if !duplicate {
				a.left.tabs = append(a.left.tabs, rt)
			}
		}
		if a.left.active < 0 && len(a.left.tabs) > 0 {
			a.left.active = 0
		}
		a.right = Pane{active: -1}
		a.split = false
		a.activePane = 0
	}
	a.focusEditor = true
	a.saveSession()
	a.updateTitle()
	a.update()
}

func (a *App) updateTitle() {
	title := "ZPad"
	dirty := false
	tab := a.activeTab()
	if tab != nil {
		title = tab.doc.Name() + " — ZPad"
		dirty = tab.doc.Buffer.Modified()
	} else if a.root != "" {
		title = filepath.Base(a.root) + " — ZPad"
	}
	a.lastDirty = dirty
	if a.win != nil {
		a.win.SetTitle(title)
		a.win.SetDocumentEdited(dirty)
		if tab != nil {
			a.win.SetRepresentedFile(tab.doc.Path)
		} else {
			a.win.SetRepresentedFile("")
		}
	}
}

func (a *App) setWorkspace(root string) {
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}
	a.errMsg = ""
	a.root = root
	a.tree = files.New(root, a.Settings.ShowHidden)
	a.treeSelected = -1
	a.treeState = ui.OutlineState[string]{}
	a.treeGeneration++
	a.Settings.SidebarOpen = true
	a.saveSettings()
	a.saveSession()
	a.update()
}

func (a *App) toggleSidebar() {
	a.Settings.SidebarOpen = !a.Settings.SidebarOpen
	a.saveSettings()
	a.saveSession()
	a.update()
}

func (a *App) refreshTree() {
	if a.tree != nil {
		a.tree.Refresh()
		a.treeGeneration++
		a.update()
	}
}

func (a *App) toggleHidden() {
	a.Settings.ShowHidden = !a.Settings.ShowHidden
	if a.tree != nil {
		a.tree.SetShowHidden(a.Settings.ShowHidden)
		a.treeGeneration++
	}
	a.saveSettings()
	a.update()
}

func (a *App) setTheme() {
	mode := "system"
	switch a.ThemeIndex {
	case 0:
		mode = "light"
	case 1:
		mode = "dark"
	}
	a.Settings.Theme = mode

	switch mode {
	case "light":
		a.Dark = false
	case "dark":
		a.Dark = true
	}

	if a.running {
		SetThemeSource(mode)
		if mode == "system" {
			a.Dark = mygo.Theme.IsDark()
		}
	}
	a.saveSettings()
	a.update()
}

func (a *App) saveSettings() {
	if a.SettingsDir == "" {
		return
	}
	if err := settings.Save(a.SettingsDir, a.Settings); err != nil {
		a.errMsg = locales.SaveSettingsFailed() + ": " + err.Error()
		log.Printf("save settings: %v", err)
		return
	}
	a.errMsg = ""
}
