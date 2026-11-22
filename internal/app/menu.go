package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/caseconv"
	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/locales"
)

// InstallMenu builds the application menu.
func (a *App) InstallMenu() {
	fileMenu := []*mygo.MenuItem{
		{Label: locales.NewFile(), Accelerator: "CmdOrCtrl+N", Click: func(*mygo.MenuItem, *mygo.Window) { a.NewFile() }},
		{Label: locales.OpenFileEllipsis(), Accelerator: "CmdOrCtrl+O", Click: func(*mygo.MenuItem, *mygo.Window) { a.OpenFileDialog() }},
		{Label: locales.OpenFolderEllipsis(), Accelerator: "CmdOrCtrl+Shift+O", Click: func(*mygo.MenuItem, *mygo.Window) { a.OpenFolderDialog() }},
		mygo.Separator(),
		{Label: locales.Save(), Accelerator: "CmdOrCtrl+S", Click: func(*mygo.MenuItem, *mygo.Window) { a.SaveActive() }},
		{Label: locales.SaveAs(), Accelerator: "CmdOrCtrl+Shift+S", Click: func(*mygo.MenuItem, *mygo.Window) { a.SaveAsActive() }},
		mygo.Separator(),
		{Label: locales.CloseTab(), Accelerator: "CmdOrCtrl+W", Click: func(*mygo.MenuItem, *mygo.Window) { a.CloseActiveTab() }},
	}

	editMenu := []*mygo.MenuItem{
		{Role: mygo.RoleUndo},
		{Role: mygo.RoleRedo},
		mygo.Separator(),
		{Role: mygo.RoleCut},
		{Role: mygo.RoleCopy},
		{Role: mygo.RolePaste},
		{Role: mygo.RoleSelectAll},
		mygo.Separator(),
		{Label: locales.CaseMenu(), Submenu: []*mygo.MenuItem{
			{Label: locales.UpperCase(), Click: func(*mygo.MenuItem, *mygo.Window) { a.ConvertCase(caseconv.Upper) }},
			{Label: locales.LowerCase(), Click: func(*mygo.MenuItem, *mygo.Window) { a.ConvertCase(caseconv.Lower) }},
			{Label: locales.TitleCase(), Click: func(*mygo.MenuItem, *mygo.Window) { a.ConvertCase(caseconv.Title) }},
			{Label: locales.InvertCase(), Click: func(*mygo.MenuItem, *mygo.Window) { a.ConvertCase(caseconv.Invert) }},
		}},
	}

	viewMenu := []*mygo.MenuItem{
		{Label: locales.ToggleSidebar(), Accelerator: "CmdOrCtrl+B", Click: func(*mygo.MenuItem, *mygo.Window) { a.toggleSidebar() }},
		{Label: locales.ToggleSplit(), Accelerator: "CmdOrCtrl+\\", Click: func(*mygo.MenuItem, *mygo.Window) { a.toggleSplit() }},
		{Label: locales.Refresh(), Click: func(*mygo.MenuItem, *mygo.Window) { a.refreshTree() }},
		{Label: locales.FindEllipsis(), Click: func(*mygo.MenuItem, *mygo.Window) {
			if tab := a.activeTab(); tab != nil {
				tab.state.Run(editor.Find)
			}
		}},
		{Label: locales.ReplaceEllipsis(), Click: func(*mygo.MenuItem, *mygo.Window) {
			if tab := a.activeTab(); tab != nil {
				tab.state.Run(editor.FindReplace)
			}
		}},
		{Label: locales.ShowHidden(), Type: mygo.MenuItemCheckbox, Checked: a.Settings.ShowHidden, Click: func(it *mygo.MenuItem, _ *mygo.Window) {
			a.toggleHidden()
			it.SetChecked(a.Settings.ShowHidden)
		}},
		mygo.Separator(),
		{Label: locales.Theme(), Submenu: []*mygo.MenuItem{
			{Label: locales.ThemeLight(), Click: func(*mygo.MenuItem, *mygo.Window) { a.setThemeIndex(0) }},
			{Label: locales.ThemeDark(), Click: func(*mygo.MenuItem, *mygo.Window) { a.setThemeIndex(1) }},
			{Label: locales.ThemeSystem(), Click: func(*mygo.MenuItem, *mygo.Window) { a.setThemeIndex(2) }},
		}},
		mygo.Separator(),
		{Label: locales.SettingsEllipsis(), Accelerator: "CmdOrCtrl+,", Click: func(*mygo.MenuItem, *mygo.Window) { a.openSettings() }},
	}

	encMenu := []*mygo.MenuItem{
		{Label: locales.ReopenWithEncoding(), Submenu: encodingItems(func(kind encoding.Kind) {
			a.ReopenWithEncoding(kind)
		})},
		{Label: locales.SaveWithEncoding(), Submenu: encodingItems(func(kind encoding.Kind) {
			a.SaveWithEncoding(kind)
		})},
		mygo.Separator(),
		{Label: locales.LineEnding(), Submenu: []*mygo.MenuItem{
			{Label: "LF", Click: func(*mygo.MenuItem, *mygo.Window) { a.SetLineEnding("\n") }},
			{Label: "CRLF", Click: func(*mygo.MenuItem, *mygo.Window) { a.SetLineEnding("\r\n") }},
			{Label: "CR", Click: func(*mygo.MenuItem, *mygo.Window) { a.SetLineEnding("\r") }},
		}},
	}

	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: locales.MenuFile(), Submenu: fileMenu},
		{Label: locales.MenuEdit(), Submenu: editMenu},
		{Label: locales.MenuView(), Submenu: viewMenu},
		{Label: locales.MenuEncoding(), Submenu: encMenu},
	})
	mygo.App.SetMenu(menu)
}

func (a *App) setThemeIndex(i int) {
	a.ThemeIndex = i
	a.setTheme()
}

func encodingItems(run func(encoding.Kind)) []*mygo.MenuItem {
	items := make([]*mygo.MenuItem, 0, len(encoding.All))
	for _, kind := range encoding.All {
		k := kind
		items = append(items, &mygo.MenuItem{
			Label: encoding.Label(k),
			Click: func(*mygo.MenuItem, *mygo.Window) { run(k) },
		})
	}
	return items
}
