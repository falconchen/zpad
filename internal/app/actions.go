package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/caseconv"
	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/locales"
)

var textExtensions = []string{
	"txt", "md", "markdown", "html", "htm", "css", "json",
	"js", "mjs", "cjs", "ts", "tsx", "jsx", "xml",
	"yaml", "yml", "toml", "ini", "conf", "log", "csv",
}

func textFilters() []mygo.FileFilter {
	return []mygo.FileFilter{
		{Name: locales.TextFiles(), Extensions: textExtensions},
		{Name: locales.AllFiles(), Extensions: []string{"*"}},
	}
}

// OpenFileDialog asks for files and opens them.
func (a *App) OpenFileDialog() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:   a.win,
			Title:    locales.OpenFileTitle(),
			Filters:  textFilters(),
			Multiple: true,
		})
		if err != nil {
			a.fail(locales.OpenDialogFailed(), err)
			return
		}
		a.OpenPaths(paths)
	}()
}

// OpenFolderDialog asks for a workspace directory.
func (a *App) OpenFolderDialog() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:    a.win,
			Title:     locales.OpenFolderTitle(),
			Directory: true,
		})
		if err != nil {
			a.fail(locales.OpenDialogFailed(), err)
			return
		}
		if len(paths) == 0 {
			return
		}
		a.onUI(func() { a.setWorkspace(paths[0]) })
	}()
}

// OpenPaths opens files and directories. It reads files off the UI thread.
func (a *App) OpenPaths(paths []string) {
	a.openPaths(paths, -1)
}

// openPaths opens files and directories in a pane, -1 for the active one.
// Documents already open go at once; the others are read off the UI
// thread.
func (a *App) openPaths(paths []string, paneIdx int) {
	if len(paths) == 0 {
		return
	}
	a.onUI(func() {
		pending := make([]string, 0, len(paths))
		for _, path := range paths {
			if doc := a.findDocument(path); doc != nil {
				a.addDocumentAt(paneIdx, doc)
			} else {
				pending = append(pending, path)
			}
		}
		if len(pending) > 0 {
			a.loadPaths(pending, paneIdx)
		}
	})
}

// loadPaths reads files and opens them, in paneIdx when it is a pane. It
// runs on the UI thread, as openPaths calls it.
func (a *App) loadPaths(paths []string, paneIdx int) {
	maxMB := a.Settings.LargeFileMB
	a.loading++
	defer func() {
		if a.loading > 0 {
			a.loading--
		}
		a.update()
	}()
	go func() {
		var dirs, files []string
		for _, path := range paths {
			info, err := os.Stat(path)
			if err != nil {
				a.fail(locales.OpenFailed(filepath.Base(path)), err)
				continue
			}
			if info.IsDir() {
				dirs = append(dirs, path)
			} else {
				files = append(files, path)
			}
		}
		// A folder first, so files opened with it share its workspace.
		for _, dir := range dirs {
			a.onUI(func() { a.setWorkspace(dir) })
		}
		for _, path := range files {
			doc, err := loadDocument(path, maxMB)
			if err != nil {
				a.fail(locales.OpenFailed(filepath.Base(path)), err)
				continue
			}
			if doc.ReadOnly {
				res, err := mygo.Dialog.Message(mygo.MessageOptions{
					Parent:        a.win,
					Type:          mygo.MessageInfo,
					Message:       locales.LargeFileQuestion(),
					Detail:        filepath.Base(path) + " · " + locales.FileTooLarge(maxMB),
					Buttons:       []string{locales.OpenReadOnly(), locales.Cancel()},
					DefaultButton: 0,
					CancelButton:  1,
				})
				if err != nil || res.Button != 0 {
					continue
				}
			}
			a.onUI(func() { a.addDocumentAt(paneIdx, doc) })
		}
	}()
}

// NewFile adds an untitled document.
func (a *App) NewFile() {
	a.addDocument(a.newUntitledDocument())
}

// Undo takes back the active buffer's last edit.
func (a *App) Undo() {
	if tab := a.activeTab(); tab != nil && !tab.doc.ReadOnly {
		tab.doc.Buffer.Undo()
		a.updateTitle()
		a.update()
	}
}

// Redo makes the active buffer's last undone edit again.
func (a *App) Redo() {
	if tab := a.activeTab(); tab != nil && !tab.doc.ReadOnly {
		tab.doc.Buffer.Redo()
		a.updateTitle()
		a.update()
	}
}

// SaveActive saves the active tab.
func (a *App) SaveActive() {
	if tab := a.activeTab(); tab != nil {
		a.saveTab(tab, false, nil)
	}
}

// SaveAsActive saves the active tab under another name.
func (a *App) SaveAsActive() {
	if tab := a.activeTab(); tab != nil {
		a.saveTab(tab, true, nil)
	}
}

// SaveWithEncoding saves the active tab in kind.
func (a *App) SaveWithEncoding(kind encoding.Kind) {
	tab := a.activeTab()
	if tab == nil {
		return
	}
	tab.doc.Encoding = kind
	a.saveTab(tab, false, nil)
}

// ReopenWithEncoding reads the active file again as kind.
func (a *App) ReopenWithEncoding(kind encoding.Kind) {
	tab := a.activeTab()
	if tab == nil || tab.doc.Path == "" {
		return
	}
	if !tab.doc.Buffer.Modified() {
		a.reloadAs(tab, kind)
		return
	}
	go func() {
		res, err := mygo.Dialog.Message(mygo.MessageOptions{
			Parent:             a.win,
			Type:               mygo.MessageWarning,
			Message:            locales.DiscardChangesQuestion(),
			Detail:             tab.doc.Name(),
			Buttons:            []string{locales.DiscardAndReopen(), locales.Cancel()},
			DefaultButton:      1,
			CancelButton:       1,
			DestructiveButtons: []int{0},
		})
		if err != nil || res.Button != 0 {
			return
		}
		a.reloadAs(tab, kind)
	}()
}

func (a *App) reloadAs(tab *Tab, kind encoding.Kind) {
	path := tab.doc.Path
	go func() {
		doc, err := loadDocumentWith(path, kind, a.Settings.LargeFileMB)
		if err != nil {
			a.fail(locales.ReopenFailed(), err)
			return
		}
		a.onUI(func() {
			tab.doc.Buffer.Reset(doc.Buffer.String())
			tab.doc.Buffer.SetLanguage(editor.LanguageFor(path))
			tab.doc.Encoding = kind
			tab.doc.Raw = doc.Raw
			tab.doc.Size = doc.Size
			tab.doc.ModTime = doc.ModTime
			a.setDocReadOnly(tab.doc, doc.ReadOnly)
			tab.doc.Buffer.MarkSaved()
			a.updateTitle()
		})
	}()
}

// SetLineEnding converts every line of the active tab to eol.
func (a *App) SetLineEnding(eol string) {
	tab := a.activeTab()
	if tab == nil || tab.doc.ReadOnly {
		return
	}
	tab.doc.Buffer.SetLineEnding(eol)
	a.updateTitle()
	a.update()
}

// ConvertCase changes the case of the active tab's selections.
func (a *App) ConvertCase(mode caseconv.Mode) {
	tab := a.activeTab()
	if tab == nil || tab.doc.ReadOnly {
		return
	}
	applyCase(tab, mode)
	a.updateTitle()
	a.update()
}

// CloseActiveTab closes the active tab.
func (a *App) CloseActiveTab() {
	p := a.activePaneObj()
	a.CloseTab(a.activePane, p.active)
}

// CloseTab closes tab i of a pane, asking to save when it has changes.
func (a *App) CloseTab(paneIdx, i int) {
	p := a.pane(paneIdx)
	if i < 0 || i >= len(p.tabs) || a.busy {
		return
	}
	tab := p.tabs[i]
	if !tab.doc.Buffer.Modified() {
		// Closing from a tab's button happens during a frame; remove the
		// tab after it, not while its bar is being built.
		a.onUI(func() { a.removeTabByPointer(paneIdx, tab) })
		return
	}
	a.busy = true
	go func() {
		res, err := mygo.Dialog.Message(mygo.MessageOptions{
			Parent:             a.win,
			Type:               mygo.MessageWarning,
			Message:            locales.SaveChangesQuestion(),
			Detail:             tab.doc.DisplayPath(a.root),
			Buttons:            []string{locales.Save(), locales.DontSave(), locales.Cancel()},
			DefaultButton:      0,
			CancelButton:       2,
			DestructiveButtons: []int{1},
		})
		if err != nil {
			a.onUI(func() { a.busy = false })
			a.fail(locales.CloseTabFailed(), err)
			return
		}
		switch res.Button {
		case 0:
			a.onUI(func() {
				a.busy = false
				a.saveTab(tab, false, func() { a.removeTabByPointer(paneIdx, tab) })
			})
		case 1:
			a.onUI(func() {
				a.busy = false
				a.removeTabByPointer(paneIdx, tab)
			})
		default:
			a.onUI(func() { a.busy = false })
		}
	}()
}

func (a *App) removeTabByPointer(paneIdx int, t *Tab) {
	p := a.pane(paneIdx)
	for i, tab := range p.tabs {
		if tab == t {
			a.removeTab(paneIdx, i)
			return
		}
	}
}

func (a *App) saveTab(tab *Tab, saveAs bool, then func()) {
	if a.busy {
		return
	}
	a.busy = true
	go func() {
		defer a.onUI(func() { a.busy = false })

		saveAs = saveAs || tab.doc.ReadOnly
		path := tab.doc.Path
		if saveAs || path == "" {
			defaultPath := path
			if defaultPath == "" {
				defaultPath = tab.doc.Name()
			}
			p, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
				Parent:      a.win,
				Title:       locales.SaveFileTitle(),
				DefaultPath: defaultPath,
				Filters:     textFilters(),
			})
			if err != nil {
				a.fail(locales.SaveFailed(), err)
				return
			}
			if p == "" {
				return
			}
			path = p
		}

		if !saveAs && tab.doc.Path != "" {
			if _, statErr := os.Stat(path); errors.Is(statErr, os.ErrNotExist) {
				res, err := mygo.Dialog.Message(mygo.MessageOptions{
					Parent:        a.win,
					Type:          mygo.MessageWarning,
					Message:       locales.FileMissingQuestion(),
					Detail:        path,
					Buttons:       []string{locales.Recreate(), locales.Cancel()},
					DefaultButton: 0,
					CancelButton:  1,
				})
				if err != nil || res.Button != 0 {
					return
				}
			}
		}

		data, err := writeDocument(tab.doc, path)
		if err != nil {
			a.fail(locales.SaveFailed(), err)
			mygo.Dialog.Error(locales.SaveFailed(), path+": "+err.Error())
			return
		}
		var modTime time.Time
		if info, statErr := os.Stat(path); statErr == nil {
			modTime = info.ModTime()
		}
		a.onUI(func() {
			tab.doc.Path = path
			tab.doc.ID = path
			tab.doc.Size = int64(len(data))
			tab.doc.ModTime = modTime
			tab.doc.Buffer.SetLanguage(editor.LanguageFor(path))
			tab.doc.Buffer.MarkSaved()
			if saveAs {
				a.setDocReadOnly(tab.doc, false)
			}
			a.updateTitle()
			if then != nil {
				then()
			}
		})
	}()
}

func (a *App) fail(title string, err error) {
	if err == nil {
		return
	}
	a.onUI(func() { a.SetError(fmt.Sprintf("%s: %v", title, err)) })
}
