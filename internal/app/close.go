package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/editor"
)

// dirtyDocs returns every modified document once, however many panes show
// it.
func (a *App) dirtyDocs() []*Document {
	seen := map[*Document]bool{}
	var out []*Document
	for _, p := range []*Pane{&a.left, &a.right} {
		for _, tab := range p.tabs {
			if tab.doc.Buffer.Modified() && !seen[tab.doc] {
				seen[tab.doc] = true
				out = append(out, tab.doc)
			}
		}
	}
	return out
}

func (a *App) hasDirty() bool { return len(a.dirtyDocs()) > 0 }

func (a *App) onWindowClose(e *mygo.CloseEvent) {
	if a.allowClose {
		a.saveSession()
		return
	}
	if a.closePromptOpen {
		e.PreventDefault()
		return
	}
	if !a.hasDirty() {
		a.saveSession()
		return
	}
	e.PreventDefault()
	a.requestClose("window")
}

// BeforeQuit asks to save before the application quits.
func (a *App) BeforeQuit(e *mygo.QuitEvent) {
	if e.SessionEnding {
		a.saveDirtySync()
		a.saveSession()
		// The session end cannot be canceled, and OnClose must not ask.
		a.allowClose = true
		return
	}
	if a.allowQuit {
		a.saveSession()
		return
	}
	if a.closePromptOpen {
		e.PreventDefault()
		return
	}
	if !a.hasDirty() {
		a.saveSession()
		return
	}
	e.PreventDefault()
	a.requestClose("quit")
}

func (a *App) requestClose(reason string) {
	if a.closePromptOpen || a.busy {
		return
	}
	docs := a.dirtyDocs()
	if len(docs) == 0 {
		a.executeClose(reason)
		return
	}
	a.closePromptOpen = true
	go func() {
		res, err := mygo.Dialog.Message(mygo.MessageOptions{
			Parent:             a.win,
			Type:               mygo.MessageWarning,
			Message:            fmt.Sprintf("有 %d 个文件未保存，是否保存？", len(docs)),
			Detail:             dirtyDetail(docs),
			Buttons:            []string{"保存全部", "不保存", "取消"},
			DefaultButton:      0,
			CancelButton:       2,
			DestructiveButtons: []int{1},
		})
		if err != nil {
			a.onUI(func() {
				a.closePromptOpen = false
				a.SetError("关闭失败: " + err.Error())
			})
			return
		}
		switch res.Button {
		case 0:
			a.onUI(func() {
				a.closePromptOpen = false
				a.saveAllForClose(reason)
			})
		case 1:
			a.onUI(func() {
				a.closePromptOpen = false
				a.executeClose(reason)
			})
		default:
			a.onUI(func() { a.closePromptOpen = false })
		}
	}()
}

func dirtyDetail(docs []*Document) string {
	names := make([]string, 0, len(docs))
	for _, doc := range docs {
		if len(names) == 5 {
			names = append(names, fmt.Sprintf("等 %d 个", len(docs)))
			break
		}
		names = append(names, doc.Name())
	}
	return strings.Join(names, "、")
}

func (a *App) saveAllForClose(reason string) {
	docs := a.dirtyDocs()
	if len(docs) == 0 {
		a.executeClose(reason)
		return
	}
	if a.busy {
		return
	}
	a.busy = true
	go func() {
		type saved struct {
			doc  *Document
			path string
			data []byte
		}
		var results []saved
		for _, doc := range docs {
			path := doc.Path
			if path == "" {
				p, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
					Parent:      a.win,
					Title:       "保存文件",
					DefaultPath: doc.Name(),
					Filters:     textFilters(),
				})
				if err != nil {
					a.onUI(func() {
						a.busy = false
						a.SetError("保存失败: " + err.Error())
					})
					return
				}
				if p == "" {
					a.onUI(func() { a.busy = false })
					return
				}
				path = p
			}
			data, err := writeDocument(doc, path)
			if err != nil {
				a.onUI(func() {
					a.busy = false
					a.SetError("保存失败: " + err.Error())
				})
				mygo.Dialog.Error("保存失败", err.Error())
				return
			}
			results = append(results, saved{doc: doc, path: path, data: data})
		}
		a.onUI(func() {
			for _, r := range results {
				r.doc.Path = r.path
				r.doc.ID = r.path
				r.doc.Size = int64(len(r.data))
				r.doc.Buffer.SetLanguage(editor.LanguageFor(r.path))
				r.doc.Buffer.MarkSaved()
			}
			a.busy = false
			a.updateTitle()
			a.executeClose(reason)
		})
	}()
}

func (a *App) executeClose(reason string) {
	a.saveSession()
	switch reason {
	case "quit":
		a.allowQuit = true
		a.allowClose = true
		mygo.App.Quit()
	default:
		a.allowClose = true
		if a.win != nil {
			a.win.Close()
		}
	}
}

// saveDirtySync writes what it can without dialogs, for the system ending
// the session.
func (a *App) saveDirtySync() {
	for _, doc := range a.dirtyDocs() {
		if doc.Path == "" {
			continue
		}
		if _, err := writeDocument(doc, doc.Path); err != nil {
			log.Printf("save %s on session end: %v", doc.Path, err)
			continue
		}
		doc.Buffer.MarkSaved()
	}
}
