package app

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/egoist/mygo"
	"github.com/helloxz/zpad/locales"
)

// openInBrowser opens a file with the system's default application for it,
// a browser for HTML. It is a variable so tests can record the path without
// launching a browser.
var openInBrowser = mygo.Shell.OpenPath

// isHTMLPath reports whether path names a file a browser renders, which is
// what the browser button opens.
func isHTMLPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".html", ".htm", ".xhtml":
		return true
	}
	return false
}

// openInBrowserShortcut is the shortcut that opens the active HTML file in
// the browser, as its tooltip shows it.
func openInBrowserShortcut() string {
	if runtime.GOOS == "darwin" {
		return "⌘⇧B"
	}
	return "Ctrl+Shift+B"
}

// canOpenInBrowser reports whether the active tab has an HTML file to open
// in the browser. An untitled document has no path, and other files are not
// HTML.
func (a *App) canOpenInBrowser() bool {
	tab := a.activeTab()
	return tab != nil && isHTMLPath(tab.doc.Path)
}

// OpenActiveInBrowser opens the active tab's file in the system browser.
// Unsaved changes are saved first, so the browser shows what the editor
// shows.
func (a *App) OpenActiveInBrowser() {
	if tab := a.activeTab(); tab != nil {
		a.openTabInBrowser(tab)
	}
}

// openTabInBrowser opens a tab's file in the browser, saving it first when
// it has unsaved changes. The tab need not be the active one, as a menu on
// a tab's own header opens it.
func (a *App) openTabInBrowser(tab *Tab) {
	if tab == nil || tab.doc.Path == "" || !isHTMLPath(tab.doc.Path) {
		return
	}
	if tab.doc.Buffer.Modified() {
		a.saveTab(tab, false, func() {
			// A save-as may have turned it into another kind of file.
			if isHTMLPath(tab.doc.Path) {
				a.launchBrowser(tab.doc.Path)
			}
		})
		return
	}
	a.launchBrowser(tab.doc.Path)
}

// launchBrowser opens path in the system's default application for it,
// telling in a toast when there is no such application.
func (a *App) launchBrowser(path string) {
	if err := openInBrowser(path); err != nil {
		a.SetErrorToast(locales.OpenInBrowserFailed(), err.Error())
	}
}
