package app

import (
	"fmt"
	"path/filepath"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/caseconv"
	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/internal/theme"
	"github.com/helloxz/zpad/locales"
)

// View builds the window: toolbar, sidebar and content, status bar.
func (a *App) View(c *ui.Context) {
	t := theme.For(a.Dark)
	c.SetTheme(t)

	ui.Themed(c, t).Fill().Children(func() {
		a.toolbar(c, t)
		if a.Settings.SidebarOpen {
			sp := ui.Split(c, &a.Settings.SidebarWidth,
				func() { a.sidebar(c, t) },
				func() { a.content(c, t) },
			)
			sp.Grow(1).FillWidth()
			if sp.Changed() {
				a.saveSettings()
			}
		} else {
			a.content(c, t)
		}
		a.statusBar(c, t)
		a.settingsModal(c, t)
		if a.pendingToast != "" {
			c.Toast(a.pendingToast)
			a.pendingToast = ""
		}
	})
}

func (a *App) toolbar(c *ui.Context, t *ui.Theme) {
	ui.Toolbar(c, func() {
		a.iconButton(c, t, iconSidebar, locales.ToggleSidebar(), a.Settings.SidebarOpen, a.toggleSidebar).
			KeyShortcuts(ui.Cmd, ui.KeyB)
		a.iconButton(c, t, iconPlus, locales.NewFile(), false, a.NewFile).
			KeyShortcuts(ui.Cmd, ui.KeyN)
		a.iconButton(c, t, iconOpen, locales.OpenFile(), false, a.OpenFileDialog).
			KeyShortcuts(ui.Cmd, ui.KeyO)
		a.iconButton(c, t, iconFolder, locales.OpenFolder(), false, a.OpenFolderDialog).
			KeyShortcuts(ui.Cmd|ui.Shift, ui.KeyO)
		a.iconButton(c, t, iconSave, locales.Save(), false, a.SaveActive).
			KeyShortcuts(ui.Cmd, ui.KeyS)
		ui.Box(c).Width(1).Height(18).Background(t.Border).Margin(0, 4)
		a.iconButton(c, t, iconUndo, locales.Undo(), false, a.Undo).
			KeyShortcuts(ui.Cmd, ui.KeyZ)
		a.iconButton(c, t, iconRedo, locales.Redo(), false, a.Redo).
			KeyShortcuts(ui.Cmd|ui.Shift, ui.KeyZ)
		a.iconButton(c, t, iconSearch, locales.Find(), false, func() {
			if tab := a.activeTab(); tab != nil {
				tab.state.Run(editor.Find)
			}
		}).KeyShortcuts(ui.Cmd, ui.KeyF)
		ui.MenuButton(c, "Aa", func(m *ui.Menu) {
			if m.Item(locales.UpperCase()).Chosen() {
				a.ConvertCase(caseconv.Upper)
			}
			if m.Item(locales.LowerCase()).Chosen() {
				a.ConvertCase(caseconv.Lower)
			}
			if m.Item(locales.TitleCase()).Chosen() {
				a.ConvertCase(caseconv.Title)
			}
			if m.Item(locales.InvertCase()).Chosen() {
				a.ConvertCase(caseconv.Invert)
			}
		}).Label(locales.CaseMenu()).Tooltip(locales.CaseMenu())
		ui.Spacer(c)
		a.iconButton(c, t, iconSplit, locales.ToggleSplit(), a.split, a.toggleSplit).
			KeyShortcuts(ui.Cmd, ui.KeyBackslash)
		ui.MenuButton(c, locales.Theme(), func(m *ui.Menu) {
			if m.Item(locales.ThemeLight()).Checked(a.ThemeIndex == 0).Chosen() {
				a.setThemeIndex(0)
			}
			if m.Item(locales.ThemeDark()).Checked(a.ThemeIndex == 1).Chosen() {
				a.setThemeIndex(1)
			}
			if m.Item(locales.ThemeSystem()).Checked(a.ThemeIndex == 2).Chosen() {
				a.setThemeIndex(2)
			}
		}).Label(locales.Theme()).Tooltip(locales.Theme())
		a.iconButton(c, t, iconSettings, locales.Settings(), a.showSettings, a.openSettings).
			KeyShortcuts(ui.Cmd, ui.KeyComma)
	}).Padding(8, 10).Background(t.Surface).
		BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
}

func (a *App) iconButton(c *ui.Context, t *ui.Theme, icon *ui.SVG, label string, active bool, run func()) ui.Element {
	b := ui.ButtonBase(c).Size(28, 28).Center().Radius(6).Label(label).Tooltip(label)
	if active || b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() {
		ui.Icon(c, icon).FontSize(16).TextColor(t.Text)
	})
	if b.Clicked() {
		run()
	}
	return b
}

func (a *App) sidebar(c *ui.Context, t *ui.Theme) {
	ui.Themed(c, t).Fill().Background(theme.Sidebar(t)).Children(func() {
		ui.Row(c).Height(36).Padding(0, 8).AlignItems(ui.Center).Gap(6).Children(func() {
			name := locales.Explorer()
			if a.tree != nil {
				name = a.tree.Name()
			}
			ui.Text(c, name).FontSize(12).Bold().SingleLine().Grow(1)
			menu := ui.ButtonBase(c).Size(26, 24).Center().Radius(6).
				Background(theme.Sidebar(t)).Border(1, t.Border).Label(locales.SidebarMenu())
			if menu.Hovered() {
				menu.Background(t.SurfaceHover)
			}
			menu.Children(func() {
				ui.Text(c, "⋯").FontSize(14).TextColor(t.TextMuted)
			})
			menu.Menu(func(m *ui.Menu) {
				if m.Item(locales.Refresh()).Chosen() {
					a.refreshTree()
				}
				if m.Item(locales.ShowHidden()).Checked(a.Settings.ShowHidden).Chosen() {
					a.toggleHidden()
				}
				if m.Item(locales.CloseSidebar()).Chosen() {
					a.toggleSidebar()
				}
			})
		})

		if a.tree == nil {
			ui.Column(c).Fill().Center().Gap(8).Padding(16).Children(func() {
				ui.Text(c, locales.NoFolder()).FontSize(12).TextColor(t.TextMuted)
				if ui.Button(c, locales.OpenFolder()).Clicked() {
					a.OpenFolderDialog()
				}
			})
			return
		}

		a.treeState.List.Selected = &a.treeSelected
		a.treeState.Generation = a.treeGeneration
		ui.Outline(c, &a.treeState, a.tree.Roots(), a.tree.Children, func(path string) {
			a.treeRow(c, t, path)
		}).Grow(1)
	})
}

func (a *App) treeRow(c *ui.Context, t *ui.Theme, path string) {
	isDir := a.tree.IsDir(path)
	row := ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
		if isDir {
			ui.Icon(c, iconFolder).FontSize(14).TextColor(t.TextMuted)
		} else {
			ui.Icon(c, iconFile).FontSize(14).TextColor(t.TextMuted)
		}
		ui.Text(c, filepath.Base(path)).FontSize(12).SingleLine()
	})
	if row.RowDoubleClicked() {
		if isDir {
			a.toggleFolder(path)
		} else {
			a.OpenPaths([]string{path})
		}
	}
}

func (a *App) toggleFolder(path string) {
	if a.treeState.Open.Has(path) {
		a.treeState.Open.Remove(path)
	} else {
		a.treeState.Open.Add(path)
	}
	a.update()
}

func (a *App) content(c *ui.Context, t *ui.Theme) {
	if a.split {
		sp := ui.Split(c, &a.splitWidth,
			func() { a.paneContent(c, t, 0) },
			func() { a.paneContent(c, t, 1) },
		)
		sp.Grow(1).FillWidth()
		return
	}
	a.paneContent(c, t, 0)
}

func (a *App) paneContent(c *ui.Context, t *ui.Theme, index int) {
	p := a.pane(index)
	ui.Column(c).Fill().Background(t.Background).Children(func() {
		if len(p.tabs) == 0 {
			if a.split {
				a.emptyPane(c, t, index)
			} else {
				a.emptyState(c, t)
			}
			return
		}
		a.tabBar(c, t, index, p)
		tab := p.activeTab()
		if tab == nil {
			return
		}
		a.configureEditor(tab, t)
		if index == a.activePane {
			if dirty := tab.doc.Buffer.Modified(); dirty != a.lastDirty {
				a.lastDirty = dirty
				if a.win != nil {
					a.win.SetDocumentEdited(dirty)
				}
			}
		}
		drop := ui.Box(c).Grow(1).FillWidth().Children(func() {
			v := editor.View(c, tab.state).Grow(1).FillWidth()
			if tab.pendingScroll != nil {
				y := *tab.pendingScroll
				tab.pendingScroll = nil
				tab.state.SetScrollY(y)
			}
			if a.focusEditor && index == a.activePane {
				v.AutoFocus()
				a.focusEditor = false
			}
		})
		if files := drop.DroppedFiles(); len(files) > 0 {
			a.setActivePane(index)
			a.OpenPaths(files)
		}
		if drop.FileDragOver() {
			drop.Border(2, t.Accent)
		}
		if tab.state.Focused() && a.activePane != index {
			a.activePane = index
			a.updateTitle()
		}
	})
}

func (a *App) emptyPane(c *ui.Context, t *ui.Theme, index int) {
	ui.Column(c).Fill().Center().Gap(10).Children(func() {
		ui.Text(c, locales.PaneEmpty()).FontSize(13).TextColor(t.TextMuted)
		ui.Row(c).Gap(8).Children(func() {
			if ui.Button(c, locales.OpenFile()).Clicked() {
				a.setActivePane(index)
				a.OpenFileDialog()
			}
			if ui.Button(c, locales.NewFile()).Clicked() {
				a.setActivePane(index)
				a.NewFile()
			}
		})
	})
}

func (a *App) tabBar(c *ui.Context, t *ui.Theme, index int, p *Pane) {
	bg := theme.TabBar(t)
	if a.split && index == a.activePane {
		bg = theme.ActiveTab(t)
	}
	ui.ScrollHorizontal(c).Height(34).FillWidth().Background(bg).
		BorderWidth(0, 0, 1, 0).BorderColor(t.Border).
		Children(func() {
			ui.Row(c).Height(34).Children(func() {
				for i, tab := range p.tabs {
					a.tabItem(c, t, index, p, i, tab)
				}
				a.newTabButton(c, t, index)
			})
		})
}

func (a *App) tabItem(c *ui.Context, t *ui.Theme, paneIdx int, p *Pane, i int, tab *Tab) {
	active := i == p.active
	bg := theme.TabBar(t)
	text := t.TextMuted
	if active {
		bg = theme.ActiveTab(t)
		text = t.Text
	}
	key := fmt.Sprintf("tab-%d-%s", paneIdx, tab.doc.ID)
	b := ui.ButtonBase(c.Key(key)).Height(32).Padding(0, 10).Gap(6).
		Radius(0).Background(bg).Label(tab.doc.Name())
	if !active && b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	closed := false
	b.Children(func() {
		if tab.doc.Buffer.Modified() {
			ui.Box(c).Size(6, 6).Radius(3).Background(t.Accent).Shrink(0)
		}
		ui.Text(c, tab.doc.Name()).FontSize(12).TextColor(text).SingleLine()
		close := ui.ButtonBase(c.Key(key+"-close")).Size(16, 16).Center().
			Radius(4).Label(locales.Close()).Tooltip(locales.Close())
		close.Children(func() {
			ui.Icon(c, iconClose).FontSize(11).TextColor(t.TextMuted)
		})
		if close.Clicked() {
			a.CloseTab(paneIdx, i)
			closed = true
		}
	})
	if active {
		b.BorderWidth(0, 0, 2, 0).BorderColor(t.Accent)
	}
	if b.Clicked() && !closed {
		p.active = i
		a.activePane = paneIdx
		a.focusEditor = true
		a.updateTitle()
		a.update()
	}
}

func (a *App) newTabButton(c *ui.Context, t *ui.Theme, paneIdx int) {
	b := ui.ButtonBase(c.Key(fmt.Sprintf("new-tab-%d", paneIdx))).Size(28, 32).Center().Radius(0).
		Label(locales.NewFile()).Tooltip(locales.NewFile())
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() {
		ui.Icon(c, iconPlus).FontSize(14).TextColor(t.TextMuted)
	})
	if b.Clicked() {
		a.setActivePane(paneIdx)
		a.NewFile()
	}
}

func (a *App) configureEditor(tab *Tab, t *ui.Theme) {
	st := tab.state
	st.SoftWrap = a.Settings.SoftWrap
	switch a.Settings.LineNumbers {
	case "relative":
		st.LineNumbers = editor.LineNumbersRelative
	case "off":
		st.LineNumbers = editor.LineNumbersOff
	default:
		st.LineNumbers = editor.LineNumbersOn
	}
	st.TabSize = a.Settings.TabSize
	if a.Settings.UseTabs {
		st.Indent = editor.IndentTabs
	} else {
		st.Indent = editor.IndentSpaces
	}
	st.NoIndentGuides = !a.Settings.ShowGuides
	st.NoAutoClose = !a.Settings.AutoClose
	st.Font = ui.Font{Family: a.Settings.FontFamily, Size: a.Settings.FontSize}
	st.LineHeight = a.Settings.LineHeight
	st.Theme = theme.EditorTheme(t)
	st.Label = tab.doc.Name()
	st.Placeholder = locales.EditorPlaceholder()
	st.RenderFindBar = func(c *ui.Context, bar ui.Element, f *editor.FindBar) {
		editor.DefaultFindBar(c, bar, f)
		bar.Background(t.Surface).Border(1, t.Border).Radius(10).
			Shadow(0, 6, 20, 0, ui.RGBA(0, 0, 0, 0.18)).Padding(6)
	}
	st.RenderGoToLine = func(c *ui.Context, bar ui.Element, g *editor.GoToLineBar) {
		editor.DefaultGoToLine(c, bar, g)
		bar.Background(t.Surface).Border(1, t.Border).Radius(10).
			Shadow(0, 6, 20, 0, ui.RGBA(0, 0, 0, 0.18)).Padding(6)
	}
	st.ContextMenu = func(m *ui.Menu) {
		a.editorContextMenu(m)
	}
}

func (a *App) editorContextMenu(m *ui.Menu) {
	m.Separator()
	m.Submenu(locales.CaseMenu(), func(mm *ui.Menu) {
		if mm.Item(locales.UpperCase()).Chosen() {
			a.ConvertCase(caseconv.Upper)
		}
		if mm.Item(locales.LowerCase()).Chosen() {
			a.ConvertCase(caseconv.Lower)
		}
		if mm.Item(locales.TitleCase()).Chosen() {
			a.ConvertCase(caseconv.Title)
		}
		if mm.Item(locales.InvertCase()).Chosen() {
			a.ConvertCase(caseconv.Invert)
		}
	})
	m.Submenu(locales.LineEnding(), func(mm *ui.Menu) {
		if mm.Item("LF").Chosen() {
			a.SetLineEnding("\n")
		}
		if mm.Item("CRLF").Chosen() {
			a.SetLineEnding("\r\n")
		}
		if mm.Item("CR").Chosen() {
			a.SetLineEnding("\r")
		}
	})
	m.Submenu(locales.MenuEncoding(), func(mm *ui.Menu) {
		mm.Submenu(locales.ReopenWithEncoding(), func(m2 *ui.Menu) {
			for _, kind := range encoding.All {
				k := kind
				if m2.Item(encoding.Label(k)).Chosen() {
					a.ReopenWithEncoding(k)
				}
			}
		})
		mm.Submenu(locales.SaveWithEncoding(), func(m2 *ui.Menu) {
			for _, kind := range encoding.All {
				k := kind
				if m2.Item(encoding.Label(k)).Chosen() {
					a.SaveWithEncoding(k)
				}
			}
		})
	})
	if m.Item(locales.CloseTab()).Chosen() {
		a.CloseActiveTab()
	}
}

func (a *App) emptyState(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Fill().Center().Gap(14).Children(func() {
		ui.Text(c, "ZPad").FontSize(30).Bold()
		ui.Text(c, locales.AppTagline()).FontSize(13).TextColor(t.TextMuted)
		ui.Row(c).Gap(8).Children(func() {
			if ui.PrimaryButton(c, locales.OpenFile()).Clicked() {
				a.OpenFileDialog()
			}
			if ui.Button(c, locales.OpenFolder()).Clicked() {
				a.OpenFolderDialog()
			}
			if ui.Button(c, locales.NewFile()).Clicked() {
				a.NewFile()
			}
		})
		ui.Text(c, locales.ShortcutsHint()).FontSize(11).TextColor(t.TextMuted)
	})
}

func (a *App) statusBar(c *ui.Context, t *ui.Theme) {
	ui.Row(c).Height(26).FillWidth().Padding(0, 10).AlignItems(ui.Center).Gap(10).
		Background(t.Surface).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).
		Children(func() {
			if a.split {
				ui.Text(c, paneLabel(a.activePane)).FontSize(11).TextColor(t.Accent)
			}
			tab := a.activeTab()
			if tab == nil {
				msg, col := locales.Ready(), t.TextMuted
				if a.errMsg != "" {
					msg, col = a.errMsg, t.Danger
				}
				ui.Text(c, msg).FontSize(11).TextColor(col).SingleLine()
				ui.Spacer(c)
				ui.Text(c, "ZPad 0.1.0").FontSize(11).TextColor(t.TextMuted)
				return
			}

			doc := tab.doc
			if a.errMsg != "" {
				ui.Text(c, a.errMsg).FontSize(11).TextColor(t.Danger).MaxWidth(420).SingleLine()
			} else {
				ui.Text(c, doc.DisplayPath(a.root)).FontSize(11).TextColor(t.TextMuted).
					MaxWidth(320).SingleLine()
			}
			ui.Text(c, languageName(doc)).FontSize(11).TextColor(t.TextMuted)

			ui.MenuButton(c, encoding.Label(doc.Encoding), func(m *ui.Menu) {
				m.Submenu(locales.ReopenWithEncoding(), func(mm *ui.Menu) {
					for _, kind := range encoding.All {
						k := kind
						if mm.Item(encoding.Label(k)).Chosen() {
							a.ReopenWithEncoding(k)
						}
					}
				})
				m.Submenu(locales.SaveWithEncoding(), func(mm *ui.Menu) {
					for _, kind := range encoding.All {
						k := kind
						if mm.Item(encoding.Label(k)).Chosen() {
							a.SaveWithEncoding(k)
						}
					}
				})
			}).FontSize(11).Padding(0, 3)

			currentEOL := doc.Buffer.LineEnding()
			ui.MenuButton(c, eolLabel(currentEOL), func(m *ui.Menu) {
				for _, item := range []struct{ label, value string }{
					{"LF", "\n"}, {"CRLF", "\r\n"}, {"CR", "\r"},
				} {
					if m.Item(item.label).Checked(currentEOL == item.value).Chosen() {
						a.SetLineEnding(item.value)
					}
				}
			}).FontSize(11).Padding(0, 3)

			ui.Spacer(c)
			ui.Text(c, cursorLabel(tab)).FontSize(11).TextColor(t.TextMuted)
			if sel := selectionLabel(tab); sel != "" {
				ui.Text(c, sel).FontSize(11).TextColor(t.TextMuted)
			}
			ui.Text(c, formatSize(doc.Size)).FontSize(11).TextColor(t.TextMuted)
			if doc.Buffer.Modified() {
				ui.Box(c).Size(6, 6).Radius(3).Background(t.Accent)
			}
		})
}

func paneLabel(i int) string {
	if i == 1 {
		return locales.RightPane()
	}
	return locales.LeftPane()
}

func languageName(doc *Document) string {
	if lang := doc.Buffer.Language(); lang != nil {
		return lang.Name
	}
	return locales.PlainText()
}

func eolLabel(eol string) string {
	switch eol {
	case "\r\n":
		return "CRLF"
	case "\r":
		return "CR"
	default:
		return "LF"
	}
}

func cursorLabel(tab *Tab) string {
	p := tab.state.CursorPos()
	line := 0
	if p.Line >= 0 && p.Line < tab.doc.Buffer.LineCount() {
		line = p.Line
	}
	text := tab.doc.Buffer.Line(line)
	col := p.Col
	if col > len(text) {
		col = len(text)
	}
	return locales.LineColumn(line+1, utf8.RuneCountInString(text[:col])+1)
}

func selectionLabel(tab *Tab) string {
	n := 0
	for _, s := range tab.state.Selections() {
		if !s.Empty() {
			n += utf8.RuneCountInString(tab.doc.Buffer.Slice(s.Start(), s.End()))
		}
	}
	if n == 0 {
		return ""
	}
	return locales.SelectedCount(n)
}

func formatSize(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
