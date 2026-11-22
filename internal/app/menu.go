package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/caseconv"
	"github.com/helloxz/zpad/internal/encoding"
)

// InstallMenu builds the application menu.
func (a *App) InstallMenu() {
	fileMenu := []*mygo.MenuItem{
		{Label: "新建文件", Accelerator: "CmdOrCtrl+N", Click: func(*mygo.MenuItem, *mygo.Window) { a.NewFile() }},
		{Label: "打开文件…", Accelerator: "CmdOrCtrl+O", Click: func(*mygo.MenuItem, *mygo.Window) { a.OpenFileDialog() }},
		{Label: "打开文件夹…", Accelerator: "CmdOrCtrl+Shift+O", Click: func(*mygo.MenuItem, *mygo.Window) { a.OpenFolderDialog() }},
		mygo.Separator(),
		{Label: "保存", Accelerator: "CmdOrCtrl+S", Click: func(*mygo.MenuItem, *mygo.Window) { a.SaveActive() }},
		{Label: "另存为…", Accelerator: "CmdOrCtrl+Shift+S", Click: func(*mygo.MenuItem, *mygo.Window) { a.SaveAsActive() }},
		mygo.Separator(),
		{Label: "关闭标签", Accelerator: "CmdOrCtrl+W", Click: func(*mygo.MenuItem, *mygo.Window) { a.CloseActiveTab() }},
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
		{Label: "大小写转换", Submenu: []*mygo.MenuItem{
			{Label: "转为大写", Click: func(*mygo.MenuItem, *mygo.Window) { a.ConvertCase(caseconv.Upper) }},
			{Label: "转为小写", Click: func(*mygo.MenuItem, *mygo.Window) { a.ConvertCase(caseconv.Lower) }},
			{Label: "首字母大写", Click: func(*mygo.MenuItem, *mygo.Window) { a.ConvertCase(caseconv.Title) }},
			{Label: "反转大小写", Click: func(*mygo.MenuItem, *mygo.Window) { a.ConvertCase(caseconv.Invert) }},
		}},
	}

	viewMenu := []*mygo.MenuItem{
		{Label: "切换侧栏", Accelerator: "CmdOrCtrl+B", Click: func(*mygo.MenuItem, *mygo.Window) { a.toggleSidebar() }},
		{Label: "刷新文件树", Click: func(*mygo.MenuItem, *mygo.Window) { a.refreshTree() }},
		{Label: "查找…", Click: func(*mygo.MenuItem, *mygo.Window) {
			if tab := a.activeTab(); tab != nil {
				tab.state.Run(editor.Find)
			}
		}},
		{Label: "替换…", Click: func(*mygo.MenuItem, *mygo.Window) {
			if tab := a.activeTab(); tab != nil {
				tab.state.Run(editor.FindReplace)
			}
		}},
		{Label: "显示隐藏文件", Type: mygo.MenuItemCheckbox, Checked: a.Settings.ShowHidden, Click: func(it *mygo.MenuItem, _ *mygo.Window) {
			a.toggleHidden()
			it.SetChecked(a.Settings.ShowHidden)
		}},
		mygo.Separator(),
		{Label: "主题", Submenu: []*mygo.MenuItem{
			{Label: "亮色", Click: func(*mygo.MenuItem, *mygo.Window) { a.setThemeIndex(0) }},
			{Label: "暗色", Click: func(*mygo.MenuItem, *mygo.Window) { a.setThemeIndex(1) }},
			{Label: "跟随系统", Click: func(*mygo.MenuItem, *mygo.Window) { a.setThemeIndex(2) }},
		}},
	}

	encMenu := []*mygo.MenuItem{
		{Label: "重新以…编码打开", Submenu: encodingItems(func(kind encoding.Kind) {
			a.ReopenWithEncoding(kind)
		})},
		{Label: "以…编码保存", Submenu: encodingItems(func(kind encoding.Kind) {
			a.SaveWithEncoding(kind)
		})},
		mygo.Separator(),
		{Label: "换行符", Submenu: []*mygo.MenuItem{
			{Label: "LF", Click: func(*mygo.MenuItem, *mygo.Window) { a.SetLineEnding("\n") }},
			{Label: "CRLF", Click: func(*mygo.MenuItem, *mygo.Window) { a.SetLineEnding("\r\n") }},
			{Label: "CR", Click: func(*mygo.MenuItem, *mygo.Window) { a.SetLineEnding("\r") }},
		}},
	}

	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: "文件", Submenu: fileMenu},
		{Label: "编辑", Submenu: editMenu},
		{Label: "显示", Submenu: viewMenu},
		{Label: "编码", Submenu: encMenu},
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
