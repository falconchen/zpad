package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/locales"
)

// pathWithin reports whether path is dir or somewhere inside it.
func pathWithin(path, dir string) bool {
	if path == dir {
		return true
	}
	rel, err := filepath.Rel(dir, path)
	return err == nil && filepath.IsLocal(rel)
}

// rebasePath moves path from under oldDir to the same place under newDir.
func rebasePath(path, oldDir, newDir string) string {
	rel, err := filepath.Rel(oldDir, path)
	if err != nil {
		return path
	}
	return filepath.Join(newDir, rel)
}

// selectTreePath marks the row of path chosen.
func (a *App) selectTreePath(path string) {
	for i := 0; i < a.treeState.Rows(); i++ {
		if a.treeState.Item(i) == path {
			a.treeSelected = i
			return
		}
	}
}

// beginRename opens the rename dialog for path.
func (a *App) beginRename(path string) {
	a.renameOpen = true
	a.renamePath = path
	a.renameValue = filepath.Base(path)
	a.renameErr = ""
	a.renameFocus = true
	a.update()
}

// applyRename checks the typed name and starts the rename.
func (a *App) applyRename() {
	name := strings.TrimSpace(a.renameValue)
	old := a.renamePath
	switch {
	case name == "":
		a.renameErr = locales.NameEmpty()
		a.update()
		return
	case name == "." || name == ".." || strings.ContainsAny(name, `/\`):
		a.renameErr = locales.NameInvalid()
		a.update()
		return
	}
	newPath := filepath.Join(filepath.Dir(old), name)
	if newPath == old {
		a.renameOpen = false
		return
	}
	if !strings.EqualFold(newPath, old) {
		if _, err := os.Stat(newPath); err == nil {
			a.renameErr = locales.NameExists()
			a.update()
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			a.renameErr = err.Error()
			a.update()
			return
		}
		if a.findDocument(newPath) != nil {
			a.renameErr = locales.NameExists()
			a.update()
			return
		}
	}
	a.renameOpen = false
	go a.renameOnDisk(old, newPath)
}

// renameOnDisk moves a file or folder and updates the app.
func (a *App) renameOnDisk(old, newPath string) {
	if err := os.Rename(old, newPath); err != nil {
		a.fail(locales.RenameFailed(filepath.Base(old)), err)
		return
	}
	a.onUI(func() { a.finishRename(old, newPath) })
}

// finishRename updates open documents, the tree and the session after a
// rename.
func (a *App) finishRename(old, newPath string) {
	for _, p := range []*Pane{&a.left, &a.right} {
		for _, tab := range p.tabs {
			if tab.doc.Path == "" || !pathWithin(tab.doc.Path, old) {
				continue
			}
			oldID := tab.doc.ID
			tab.doc.Path = rebasePath(tab.doc.Path, old, newPath)
			tab.doc.ID = tab.doc.Path
			tab.doc.Buffer.SetLanguage(editor.LanguageFor(tab.doc.Path))
			if a.revealDoc == oldID {
				a.revealDoc = tab.doc.ID
			}
		}
	}
	var open []string
	for p := range a.treeState.Open.All() {
		if pathWithin(p, old) {
			open = append(open, rebasePath(p, old, newPath))
		} else {
			open = append(open, p)
		}
	}
	a.treeState.Open.Clear()
	for _, p := range open {
		a.treeState.Open.Add(p)
	}
	a.treeSelected = -1
	if a.tree != nil {
		a.tree.Refresh()
		a.treeGeneration++
	}
	if a.activeTab() != nil {
		a.focusEditor = true
	}
	a.updateTitle()
	a.saveSession()
	a.update()
}

// dirtyWithin reports whether a document at path, or under a folder, has
// unsaved changes.
func (a *App) dirtyWithin(path string) bool {
	for _, p := range []*Pane{&a.left, &a.right} {
		for _, tab := range p.tabs {
			if tab.doc.Path != "" && pathWithin(tab.doc.Path, path) && tab.doc.Buffer.Modified() {
				return true
			}
		}
	}
	return false
}

// confirmDelete asks before deleting a file or folder.
func (a *App) confirmDelete(path string, isDir bool) {
	name := filepath.Base(path)
	if a.dirtyWithin(path) {
		go func() {
			_, _ = mygo.Dialog.Message(mygo.MessageOptions{
				Parent:  a.win,
				Type:    mygo.MessageWarning,
				Message: locales.DeleteUnsavedTitle(),
				Detail:  locales.DeleteUnsavedDetail(name),
			})
		}()
		return
	}
	go func() {
		question := locales.DeleteFileQuestion(name)
		if isDir {
			question = locales.DeleteFolderQuestion(name)
		}
		res, err := mygo.Dialog.Message(mygo.MessageOptions{
			Parent:             a.win,
			Type:               mygo.MessageWarning,
			Message:            question,
			Detail:             path + "\n" + locales.DeletePermanent(),
			Buttons:            []string{locales.Cancel(), locales.Delete()},
			DefaultButton:      0,
			CancelButton:       0,
			DestructiveButtons: []int{1},
		})
		if err != nil || res.Button != 1 {
			return
		}
		a.deletePath(path, isDir)
	}()
}

// deletePath removes a file or folder and updates the app.
func (a *App) deletePath(path string, isDir bool) {
	var err error
	if isDir {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}
	if err != nil {
		a.fail(locales.DeleteFailed(filepath.Base(path)), err)
		return
	}
	a.onUI(func() { a.finishDelete(path) })
}

// finishDelete closes the tabs of a deleted file or folder and refreshes
// the tree.
func (a *App) finishDelete(path string) {
	a.closeTabsWithin(path)
	a.forgetTreePaths(path)
	if a.tree != nil {
		a.tree.Refresh()
		a.treeGeneration++
	}
	a.updateTitle()
	a.saveSession()
	a.SetToast(locales.Deleted(filepath.Base(path)))
	a.update()
}

// closeTabsWithin closes the tabs of path, or of the documents under it
// when it is a folder.
func (a *App) closeTabsWithin(path string) {
	var targets []*Tab
	for _, p := range []*Pane{&a.left, &a.right} {
		for _, tab := range p.tabs {
			if tab.doc.Path != "" && pathWithin(tab.doc.Path, path) {
				targets = append(targets, tab)
			}
		}
	}
	for _, target := range targets {
		for paneIdx := 0; paneIdx < 2; paneIdx++ {
			p := a.pane(paneIdx)
			for i, tab := range p.tabs {
				if tab == target {
					a.removeTab(paneIdx, i)
					break
				}
			}
		}
	}
}

// forgetTreePaths closes any tree folder within path in the outline.
func (a *App) forgetTreePaths(path string) {
	var keep []string
	for p := range a.treeState.Open.All() {
		if !pathWithin(p, path) {
			keep = append(keep, p)
		}
	}
	a.treeState.Open.Clear()
	for _, p := range keep {
		a.treeState.Open.Add(p)
	}
	a.treeSelected = -1
}

// openInOtherPane opens path in the pane that is not active, creating the
// split with just that file when there is none.
func (a *App) openInOtherPane(path string) {
	target := 1 - a.activePane
	if !a.split {
		a.ensureSplit()
		a.activePane = 1
		a.focusEditor = true
		target = 1
	}
	a.openInPane(path, target)
	a.saveSession()
	a.update()
}
