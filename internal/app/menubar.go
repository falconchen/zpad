package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// MenuBarHeight is the height of the menu bar Windows draws as its title
// bar. main.go passes it as mygo.WindowOptions.TitleBarHeight, so the
// system's window buttons fill the same band.
const MenuBarHeight = 32

// chromeHeight is the room the window's own chrome takes above the content:
// the toolbar and, on Windows, the menu bar that doubles as the title bar.
func (a *App) chromeHeight() float32 {
	if a.customMenuBar {
		return toolbarHeight + MenuBarHeight
	}
	return toolbarHeight
}

// topMenus returns the menu items a window's own menu bar shows: the
// application menu's top-level menus, without the roles (the app menu) and
// without the hidden ones.
func topMenus(items []*mygo.MenuItem) []*mygo.MenuItem {
	out := make([]*mygo.MenuItem, 0, len(items))
	for _, it := range items {
		if it == nil || it.Role != "" || it.Hidden || it.Label == "" || len(it.Submenu) == 0 {
			continue
		}
		out = append(out, it)
	}
	return out
}

// topMenuLabels returns the labels of topMenus, in order.
func topMenuLabels(items []*mygo.MenuItem) []string {
	menus := topMenus(items)
	labels := make([]string, len(menus))
	for i, it := range menus {
		labels[i] = it.Label
	}
	return labels
}

// menuBarLabels returns the menu bar's labels, building them before
// InstallMenu ran, as in the view tests.
func (a *App) menuBarLabels() []string {
	if a.menuBarNames == nil {
		a.menuBarNames = topMenuLabels(a.menuItems())
	}
	return a.menuBarNames
}

// menuBar draws the menu bar that doubles as the title bar on Windows. Its
// menus are the application menu's top-level menus, opened in native
// popups, so labels, shortcuts, roles and check marks follow the same
// definitions, and the native menu bar remains installed for its
// accelerators and for Alt and F10 on a window without one.
func (a *App) menuBar(c *ui.Context, t *ui.Theme) {
	bar := c.TitleBar()
	row := ui.Row(c).Height(MenuBarHeight).AlignItems(ui.Center).Gap(2).
		Padding(0, 10+bar.Right, 0, 10+bar.Left).
		Background(t.Surface).
		BorderWidth(0, 0, 1, 0).BorderColor(t.Border).
		DragWindow().KeepFocus().Role(ui.RoleMenuBar)
	row.Children(func() {
		for i, label := range a.menuBarLabels() {
			b := ui.ButtonBase(c).Height(MenuBarHeight-8).Padding(0, 8).
				Center().Radius(4).Label(label).HasPopup(ui.PopupMenu).
				Cursor(ui.CursorPointer)
			if b.Hovered() {
				b.Background(t.SurfaceHover)
			}
			b.Children(func() {
				ui.Text(c, label).FontSize(13).TextColor(t.Text)
			})
			if b.Clicked() {
				r := b.Bounds()
				a.openTopMenu(i, r.X, MenuBarHeight)
			}
		}
	})
}

// openTopMenu opens menu i of the menu bar below (x, y), in the window's
// client area. The menus are built now, so a check mark shows what the app
// holds now, and they are closed with the window or the menu.
func (a *App) openTopMenu(i int, x, y float32) {
	win := a.win
	if win == nil {
		return
	}
	menus := topMenus(a.menuItems())
	if i < 0 || i >= len(menus) {
		return
	}
	menu := mygo.NewMenu(menus[i].Submenu)
	go menu.PopupAt(win, int(x), int(y))
}
