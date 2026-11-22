package app

import (
	"path/filepath"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/locales"
)

// renameModal asks for a new name of a file or folder.
func (a *App) renameModal(c *ui.Context, t *ui.Theme) {
	if !a.renameOpen {
		return
	}
	ui.DialogBase(c, &a.renameOpen, func(backdrop, panel ui.Element) {
		c.SetTheme(t)
		backdrop.Background(ui.RGBA(0, 0, 0, 0.4))
		panel.Width(420).Padding(20).Gap(12).
			Background(t.Background).Border(1, t.Border).Radius(14).
			Shadow(0, 16, 48, 0, ui.RGBA(0, 0, 0, 0.28))
		panel.Children(func() {
			ui.Text(c, locales.Rename()).FontSize(16).Bold()
			ui.Text(c, filepath.Dir(a.renamePath)).FontSize(11).TextColor(t.TextMuted).
				SingleLine().TruncateStart()
			in := ui.TextInput(c, &a.renameValue).FillWidth().Label(locales.RenameLabel())
			if a.renameFocus {
				in.AutoFocus()
				a.renameFocus = false
			}
			if a.renameErr != "" {
				ui.Text(c, a.renameErr).FontSize(11).TextColor(t.Danger).SingleLine()
			}
			submit := in.Submitted()
			ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
				ui.Spacer(c)
				if ui.Button(c, locales.Cancel()).Cursor(ui.CursorPointer).Clicked() {
					a.renameOpen = false
				}
				if ui.PrimaryButton(c, locales.Rename()).Cursor(ui.CursorPointer).Clicked() {
					submit = true
				}
			})
			if submit {
				a.applyRename()
			}
		})
	})
}
