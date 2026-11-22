package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/caseconv"
	"github.com/helloxz/zpad/internal/encoding"
)

var textExtensions = []string{
	"txt", "md", "markdown", "html", "htm", "css", "json",
	"js", "mjs", "cjs", "ts", "tsx", "jsx", "xml",
	"yaml", "yml", "toml", "ini", "conf", "log", "csv",
}

func textFilters() []mygo.FileFilter {
	return []mygo.FileFilter{
		{Name: "文本文件", Extensions: textExtensions},
		{Name: "所有文件", Extensions: []string{"*"}},
	}
}

// OpenFileDialog asks for files and opens them.
func (a *App) OpenFileDialog() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:   a.win,
			Title:    "打开文件",
			Filters:  textFilters(),
			Multiple: true,
		})
		if err != nil {
			a.fail("打开文件失败", err)
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
			Title:     "打开文件夹",
			Directory: true,
		})
		if err != nil {
			a.fail("打开文件夹失败", err)
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
	if len(paths) == 0 {
		return
	}
	go func() {
		for _, path := range paths {
			info, err := os.Stat(path)
			if err != nil {
				a.fail("打开失败: "+filepath.Base(path), err)
				continue
			}
			if info.IsDir() {
				a.onUI(func() { a.setWorkspace(path) })
				continue
			}
			doc, err := loadDocument(path)
			if err != nil {
				a.fail("打开失败: "+filepath.Base(path), err)
				continue
			}
			a.onUI(func() { a.addDocument(doc) })
		}
	}()
}

// NewFile adds an untitled document.
func (a *App) NewFile() {
	a.addDocument(a.newUntitledDocument())
}

// Undo takes back the active buffer's last edit.
func (a *App) Undo() {
	if tab := a.activeTab(); tab != nil {
		tab.doc.Buffer.Undo()
		a.updateTitle()
		a.update()
	}
}

// Redo makes the active buffer's last undone edit again.
func (a *App) Redo() {
	if tab := a.activeTab(); tab != nil {
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
			Message:            "放弃未保存的修改？",
			Detail:             tab.doc.Name(),
			Buttons:            []string{"放弃修改并重新打开", "取消"},
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
		doc, err := loadDocumentWith(path, kind)
		if err != nil {
			a.fail("重新打开失败", err)
			return
		}
		a.onUI(func() {
			tab.doc.Buffer.Reset(doc.Buffer.String())
			tab.doc.Buffer.SetLanguage(editor.LanguageFor(path))
			tab.doc.Encoding = kind
			tab.doc.Raw = doc.Raw
			tab.doc.Size = doc.Size
			tab.doc.Buffer.MarkSaved()
			a.updateTitle()
		})
	}()
}

// SetLineEnding converts every line of the active tab to eol.
func (a *App) SetLineEnding(eol string) {
	tab := a.activeTab()
	if tab == nil {
		return
	}
	tab.doc.Buffer.SetLineEnding(eol)
	a.updateTitle()
	a.update()
}

// ConvertCase changes the case of the active tab's selections.
func (a *App) ConvertCase(mode caseconv.Mode) {
	tab := a.activeTab()
	if tab == nil {
		return
	}
	applyCase(tab, mode)
	a.updateTitle()
	a.update()
}

// CloseActiveTab closes the active tab.
func (a *App) CloseActiveTab() {
	a.CloseTab(a.active)
}

// CloseTab closes tab i, asking to save when it has changes.
func (a *App) CloseTab(i int) {
	if i < 0 || i >= len(a.tabs) || a.busy {
		return
	}
	tab := a.tabs[i]
	if !tab.doc.Buffer.Modified() {
		// Closing from a tab's button happens during a frame; remove the
		// tab after it, not while its bar is being built.
		a.onUI(func() { a.removeTabByPointer(tab) })
		return
	}
	a.busy = true
	go func() {
		res, err := mygo.Dialog.Message(mygo.MessageOptions{
			Parent:             a.win,
			Type:               mygo.MessageWarning,
			Message:            "是否保存更改？",
			Detail:             tab.doc.DisplayPath(a.root),
			Buttons:            []string{"保存", "不保存", "取消"},
			DefaultButton:      0,
			CancelButton:       2,
			DestructiveButtons: []int{1},
		})
		if err != nil {
			a.onUI(func() { a.busy = false })
			a.fail("关闭标签失败", err)
			return
		}
		switch res.Button {
		case 0:
			a.onUI(func() {
				a.busy = false
				a.saveTab(tab, false, func() { a.removeTabByPointer(tab) })
			})
		case 1:
			a.onUI(func() {
				a.busy = false
				a.removeTabByPointer(tab)
			})
		default:
			a.onUI(func() { a.busy = false })
		}
	}()
}

func (a *App) removeTabByPointer(t *Tab) {
	for i, tab := range a.tabs {
		if tab == t {
			a.removeTab(i)
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

		path := tab.doc.Path
		if saveAs || path == "" {
			defaultPath := path
			if defaultPath == "" {
				defaultPath = tab.doc.Name()
			}
			p, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
				Parent:      a.win,
				Title:       "保存文件",
				DefaultPath: defaultPath,
				Filters:     textFilters(),
			})
			if err != nil {
				a.fail("保存失败", err)
				return
			}
			if p == "" {
				return
			}
			path = p
		}

		data, err := writeDocument(tab.doc, path)
		if err != nil {
			a.fail("保存失败", err)
			mygo.Dialog.Error("保存失败", err.Error())
			return
		}
		a.onUI(func() {
			tab.doc.Path = path
			tab.doc.ID = path
			tab.doc.Size = int64(len(data))
			tab.doc.Buffer.SetLanguage(editor.LanguageFor(path))
			tab.doc.Buffer.MarkSaved()
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
