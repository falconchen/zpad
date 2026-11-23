// Package app holds ZPad's state and builds its window.
package app

import (
	"log"
	"path/filepath"
	"sync"
	"time"

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

// dragKind tells what a paneDrag carries.
type dragKind uint8

const (
	dragTab dragKind = iota
	dragFile
)

// paneDrag is something dropped on a pane: a tab to move there, or a file
// to open there.
type paneDrag struct {
	kind dragKind
	pane int
	id   string
	path string
}

// wantsPane reports whether d can land on a pane: a file on any, a tab
// only on one it does not come from.
func (d paneDrag) wantsPane(pane int) bool {
	return d.kind == dragFile || d.pane != pane
}

func (p *Pane) activeTab() *Tab {
	if p.active < 0 || p.active >= len(p.tabs) {
		return nil
	}
	return p.tabs[p.active]
}

// The sidebar's tabs.
const (
	sidebarFiles = iota
	sidebarWorkspace
)

// Version is the release version shown in the status bar and the About
// dialog. Keep it equal to the version in mygo.json.
const Version = "0.1.0"

// App is the state the window shows.
type App struct {
	Settings    settings.Settings
	SettingsDir string
	// LogPath is where SetupLogging writes, "" when logging is off.
	LogPath    string
	ThemeIndex int // 0 light, 1 dark, 2 system
	Dark       bool

	win     *mygo.Window
	running bool
	errMsg  string

	root           string
	tree           *files.Tree
	treeState      ui.OutlineState[string]
	treeSelected   int
	treeGeneration uint64
	sidebarTab     int

	left       Pane
	right      Pane
	split      bool
	splitWidth float32
	// splitVertical divides the content top and bottom rather than left
	// and right; splitHeight is the top pane's height then.
	splitVertical bool
	splitHeight   float32
	centerSplit   bool
	activePane    int // 0 left/top, 1 right/bottom

	// tabOverflow says whether a pane's tab bar is wider than the pane,
	// which shows the tab list button; revealPane and revealDoc ask the
	// tab about to build to scroll into view.
	tabOverflow [2]bool
	revealPane  int
	revealDoc   string

	focusEditor bool
	untitledSeq int
	busy        bool
	loading     int
	// loadingPaths holds the files being read, so that one asked for twice,
	// as a second instance and the system may, is read and opened once.
	loadingMu        sync.Mutex
	loadingPaths     map[string]bool
	externalChecking bool
	lastDirty        bool
	pendingOpen      []string

	showSettings bool
	aboutOpen    bool
	aboutIcon    *ui.Bitmap
	fontSize     float64
	lineHeight   float64
	fontFamilies []string
	// pendingToasts are the toasts the next frame shows, in order.
	pendingToasts []ui.Toast

	renameOpen  bool
	renamePath  string
	renameValue string
	renameErr   string
	renameFocus bool

	closePromptOpen bool
	allowClose      bool
	allowQuit       bool
	// viewLogged says the window size was logged, once, on the first frame.
	viewLogged bool
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
		LogPath:      filepath.Join(settingsDir, LogName),
		ThemeIndex:   idx,
		Dark:         s.Theme == "dark",
		treeSelected: -1,
		left:         Pane{active: -1},
		right:        Pane{active: -1},
		splitWidth:   600,
		splitHeight:  500,
		fontSize:     float64(s.FontSize),
		lineHeight:   float64(s.LineHeight),
		loadingPaths: map[string]bool{},
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

// SetError shows an error in the status bar, for a failure that has no
// better home: one the user must act on while its document is open.
//
// Actions that end, as opening a file does, tell with SetErrorToast
// instead, so that a message the user cannot act on does not stay.
func (a *App) SetError(msg string) {
	a.errMsg = msg
}

// errorToastTime is how long an error toast shows: enough to read a reason
// under its title. The pointer resting on it holds it, and Escape closes
// it.
const errorToastTime = 6 * time.Second

// SetErrorToast tells near the bottom of the window that something the
// user asked for did not happen: title says what, detail why, and it goes
// away on its own.
func (a *App) SetErrorToast(title, detail string) {
	a.pendingToasts = append(a.pendingToasts, ui.Toast{
		ID:          title,
		Title:       title,
		Description: detail,
		Type:        "error",
		Timeout:     errorToastTime,
	})
	a.update()
}

// SetToast shows a message near the bottom of the window once.
func (a *App) SetToast(msg string) {
	a.pendingToasts = append(a.pendingToasts, ui.Toast{ID: msg, Title: msg})
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

// activateTab chooses tab i of a pane, as a click on it does, and sees it
// into view once the tab bar is laid out.
func (a *App) activateTab(paneIdx, i int) {
	p := a.pane(paneIdx)
	if i < 0 || i >= len(p.tabs) {
		return
	}
	p.active = i
	a.activePane = paneIdx
	a.focusEditor = true
	a.revealPane, a.revealDoc = paneIdx, p.tabs[i].doc.ID
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

// openDocuments returns the documents open in either pane, once each, in
// tab order: the left pane first, then the right.
func (a *App) openDocuments() []*Document {
	n := len(a.left.tabs) + len(a.right.tabs)
	docs := make([]*Document, 0, n)
	seen := make(map[*Document]bool, n)
	for _, p := range []*Pane{&a.left, &a.right} {
		for _, tab := range p.tabs {
			if !seen[tab.doc] {
				seen[tab.doc] = true
				docs = append(docs, tab.doc)
			}
		}
	}
	return docs
}

// activateDocument shows doc, in the active pane when it is open there and
// in the other pane otherwise.
func (a *App) activateDocument(doc *Document) {
	for _, paneIdx := range []int{a.activePane, 1 - a.activePane} {
		if paneIdx == 1 && !a.split {
			continue
		}
		p := a.pane(paneIdx)
		for i, tab := range p.tabs {
			if tab.doc == doc {
				a.activateTab(paneIdx, i)
				return
			}
		}
	}
}

func (a *App) addDocument(doc *Document) {
	a.addDocumentAt(-1, doc)
}

// addDocumentAt adds doc to a pane, -1 for the active one, sharing the
// document already open elsewhere.
func (a *App) addDocumentAt(paneIdx int, doc *Document) {
	a.errMsg = ""
	if doc.Path != "" {
		if existing := a.findDocument(doc.Path); existing != nil {
			doc = existing
		}
	}
	if (paneIdx != 0 && paneIdx != 1) || (paneIdx == 1 && !a.split) {
		paneIdx = a.activePane
	}
	p := a.pane(paneIdx)
	for i, tab := range p.tabs {
		if tab.doc == doc {
			p.active = i
			a.activePane = paneIdx
			a.focusEditor = true
			a.updateTitle()
			return
		}
	}
	st := &editor.State{Buffer: doc.Buffer, ReadOnly: doc.ReadOnly}
	p.tabs = append(p.tabs, &Tab{doc: doc, state: st})
	p.active = len(p.tabs) - 1
	a.activePane = paneIdx
	a.revealPane, a.revealDoc = paneIdx, doc.ID
	a.focusEditor = true
	a.updateTitle()
}

func (a *App) newUntitledDocument() *Document {
	a.untitledSeq++
	doc := newUntitled(a.untitledSeq, a.Settings.DefaultEncoding)
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
		a.collapseSplit(paneIdx)
	}
	if paneIdx == a.activePane {
		a.focusEditor = true
	}
	a.updateTitle()
}

// collapseSplit closes the split when one of its panes has no tabs left,
// keeping the other pane's tabs, states and choice.
func (a *App) collapseSplit(emptyPane int) {
	if !a.split {
		return
	}
	if emptyPane == 0 {
		a.left = a.right
	}
	a.right = Pane{active: -1}
	a.split = false
	a.activePane = 0
	a.revealPane, a.revealDoc = 0, ""
	if tab := a.left.activeTab(); tab != nil {
		a.revealDoc = tab.doc.ID
	}
	a.focusEditor = true
}

// dropOnPane takes what was dropped on a pane: it moves a tab there, or
// opens a file from the sidebar there.
func (a *App) dropOnPane(d paneDrag, toPane int) {
	if d.kind == dragFile {
		a.openInPane(d.path, toPane)
		return
	}
	a.moveTab(d, toPane)
}

// ensureSplit opens the split with an empty right pane.
func (a *App) ensureSplit() {
	if a.split {
		return
	}
	a.split = true
	a.centerSplit = true
	a.right = Pane{active: -1}
}

// openInPane opens path as a tab in a pane, sharing the document already
// open elsewhere.
func (a *App) openInPane(path string, paneIdx int) {
	a.openPaths([]string{path}, paneIdx)
}

// moveTab moves the tab dragged from one pane into the other, at the end,
// or shows the copy the target pane already has. The source pane's split
// closes when its last tab leaves.
func (a *App) moveTab(from paneDrag, toPane int) {
	if from.kind != dragTab {
		return
	}
	if from.pane == toPane || from.pane < 0 || from.pane > 1 || toPane < 0 || toPane > 1 {
		return
	}
	src := a.pane(from.pane)
	for i, tab := range src.tabs {
		if tab.doc.ID == from.id {
			a.moveTabIndex(from.pane, i, toPane)
			return
		}
	}
}

// moveTabIndex moves a pane's tab into the other pane, at the end, or
// shows the copy the target pane already has. The source pane's split
// closes when its last tab leaves.
func (a *App) moveTabIndex(srcPane, idx, dstPane int) {
	if srcPane == dstPane || srcPane < 0 || srcPane > 1 || dstPane < 0 || dstPane > 1 {
		return
	}
	if dstPane == 1 && !a.split {
		return
	}
	src, dst := a.pane(srcPane), a.pane(dstPane)
	if idx < 0 || idx >= len(src.tabs) {
		return
	}
	tab := src.tabs[idx]
	target := -1
	for i, other := range dst.tabs {
		if other.doc == tab.doc {
			target = i
			break
		}
	}

	src.tabs = append(src.tabs[:idx], src.tabs[idx+1:]...)
	switch {
	case len(src.tabs) == 0:
		src.active = -1
	case src.active > idx:
		src.active--
	case src.active == idx && src.active >= len(src.tabs):
		src.active = len(src.tabs) - 1
	}

	if target < 0 {
		dst.tabs = append(dst.tabs, tab)
		target = len(dst.tabs) - 1
	}
	dst.active = target
	a.activePane = dstPane

	if a.split && len(src.tabs) == 0 {
		a.collapseSplit(srcPane)
	}
	a.focusEditor = true
	a.revealPane, a.revealDoc = a.activePane, tab.doc.ID
	a.updateTitle()
	a.saveSession()
	a.update()
}

// copyTabIndex shows a pane's tab in the other pane, sharing its document,
// or activates the copy that pane already has.
func (a *App) copyTabIndex(srcPane, idx, dstPane int) {
	if srcPane == dstPane || srcPane < 0 || srcPane > 1 || dstPane < 0 || dstPane > 1 {
		return
	}
	if dstPane == 1 && !a.split {
		return
	}
	src, dst := a.pane(srcPane), a.pane(dstPane)
	if idx < 0 || idx >= len(src.tabs) {
		return
	}
	tab := src.tabs[idx]
	target := -1
	for i, other := range dst.tabs {
		if other.doc == tab.doc {
			target = i
			break
		}
	}
	if target < 0 {
		dst.tabs = append(dst.tabs, &Tab{
			doc:   tab.doc,
			state: &editor.State{Buffer: tab.doc.Buffer, ReadOnly: tab.doc.ReadOnly},
		})
		target = len(dst.tabs) - 1
	}
	dst.active = target
	a.activePane = dstPane
	a.focusEditor = true
	a.revealPane, a.revealDoc = dstPane, tab.doc.ID
	a.updateTitle()
	a.saveSession()
	a.update()
}

// openSplit opens the split in the given direction, duplicating the active
// tab into the second pane. A split that is already open only changes
// direction, keeping both panes' tabs, states and choice.
func (a *App) openSplit(vertical bool) {
	if a.split && a.splitVertical == vertical {
		return
	}
	a.splitVertical = vertical
	a.centerSplit = true
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
	}
	a.focusEditor = true
	a.saveSession()
	a.updateTitle()
	a.update()
}

// closeSplit merges the second pane into the first, keeping the tabs only
// it had.
func (a *App) closeSplit() {
	if !a.split {
		return
	}
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
	a.focusEditor = true
	a.saveSession()
	a.updateTitle()
	a.update()
}

func (a *App) toggleSplit() {
	if a.split {
		a.closeSplit()
		return
	}
	a.openSplit(a.splitVertical)
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
	a.sidebarTab = sidebarWorkspace
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
