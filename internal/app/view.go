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
	})
}

func (a *App) toolbar(c *ui.Context, t *ui.Theme) {
	ui.Toolbar(c, func() {
		a.iconButton(c, t, iconSidebar, "切换侧栏", a.Settings.SidebarOpen, a.toggleSidebar)
		a.iconButton(c, t, iconPlus, "新建文件", false, a.NewFile)
		a.iconButton(c, t, iconOpen, "打开文件", false, a.OpenFileDialog)
		a.iconButton(c, t, iconFolder, "打开文件夹", false, a.OpenFolderDialog)
		a.iconButton(c, t, iconSave, "保存", false, a.SaveActive)
		ui.Box(c).Width(1).Height(18).Background(t.Border).Margin(0, 4)
		a.iconButton(c, t, iconUndo, "撤销", false, a.Undo)
		a.iconButton(c, t, iconRedo, "重做", false, a.Redo)
		a.iconButton(c, t, iconSearch, "查找", false, func() {
			if tab := a.activeTab(); tab != nil {
				tab.state.Run(editor.Find)
			}
		})
		ui.MenuButton(c, "Aa", func(m *ui.Menu) {
			if m.Item("转为大写").Chosen() {
				a.ConvertCase(caseconv.Upper)
			}
			if m.Item("转为小写").Chosen() {
				a.ConvertCase(caseconv.Lower)
			}
			if m.Item("首字母大写").Chosen() {
				a.ConvertCase(caseconv.Title)
			}
			if m.Item("反转大小写").Chosen() {
				a.ConvertCase(caseconv.Invert)
			}
		}).Label("大小写转换").Tooltip("大小写转换")
		ui.Spacer(c)
		ui.MenuButton(c, "主题", func(m *ui.Menu) {
			if m.Item("亮色").Checked(a.ThemeIndex == 0).Chosen() {
				a.setThemeIndex(0)
			}
			if m.Item("暗色").Checked(a.ThemeIndex == 1).Chosen() {
				a.setThemeIndex(1)
			}
			if m.Item("跟随系统").Checked(a.ThemeIndex == 2).Chosen() {
				a.setThemeIndex(2)
			}
		}).Label("主题").Tooltip("主题")
	}).Padding(6, 10).Background(t.Surface).
		BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
}

func (a *App) iconButton(c *ui.Context, t *ui.Theme, icon *ui.SVG, label string, active bool, run func()) {
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
}

func (a *App) sidebar(c *ui.Context, t *ui.Theme) {
	ui.Themed(c, t).Fill().Background(t.Surface).Children(func() {
		ui.Row(c).Height(36).Padding(0, 8).AlignItems(ui.Center).Gap(6).Children(func() {
			name := "资源管理器"
			if a.tree != nil {
				name = a.tree.Name()
			}
			ui.Text(c, name).FontSize(12).Bold().SingleLine().Grow(1)
			menu := ui.ButtonBase(c).Size(26, 24).Center().Radius(6).
				Background(t.Surface).Border(1, t.Border).Label("侧栏菜单")
			if menu.Hovered() {
				menu.Background(t.SurfaceHover)
			}
			menu.Children(func() {
				ui.Text(c, "⋯").FontSize(14).TextColor(t.TextMuted)
			})
			menu.Menu(func(m *ui.Menu) {
				if m.Item("刷新").Chosen() {
					a.refreshTree()
				}
				if m.Item("显示隐藏文件").Checked(a.Settings.ShowHidden).Chosen() {
					a.toggleHidden()
				}
				if m.Item("关闭侧栏").Chosen() {
					a.toggleSidebar()
				}
			})
		})

		if a.tree == nil {
			ui.Column(c).Fill().Center().Gap(8).Padding(16).Children(func() {
				ui.Text(c, "未打开文件夹").FontSize(12).TextColor(t.TextMuted)
				if ui.Button(c, "打开文件夹").Clicked() {
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
	ui.Column(c).Grow(1).FillWidth().Background(t.Background).Children(func() {
		if len(a.tabs) == 0 {
			a.emptyState(c, t)
			return
		}
		a.tabBar(c, t)
		tab := a.activeTab()
		if tab == nil {
			return
		}
		a.configureEditor(tab, t)
		if dirty := tab.doc.Buffer.Modified(); dirty != a.lastDirty {
			a.lastDirty = dirty
			if a.win != nil {
				a.win.SetDocumentEdited(dirty)
			}
		}
		drop := ui.Box(c).Grow(1).FillWidth().Children(func() {
			v := editor.View(c, tab.state).Grow(1).FillWidth()
			if a.focusEditor {
				v.AutoFocus()
				a.focusEditor = false
			}
		})
		if files := drop.DroppedFiles(); len(files) > 0 {
			a.OpenPaths(files)
		}
		if drop.FileDragOver() {
			drop.Border(2, t.Accent)
		}
	})
}

func (a *App) tabBar(c *ui.Context, t *ui.Theme) {
	ui.ScrollHorizontal(c).Height(38).FillWidth().Background(t.Surface).
		BorderWidth(0, 0, 1, 0).BorderColor(t.Border).
		Children(func() {
			ui.Row(c).Height(38).Children(func() {
				for i, tab := range a.tabs {
					a.tabItem(c, t, i, tab)
				}
				a.newTabButton(c, t)
			})
		})
}

func (a *App) tabItem(c *ui.Context, t *ui.Theme, i int, tab *Tab) {
	active := i == a.active
	bg := t.Surface
	text := t.TextMuted
	if active {
		bg = t.Background
		text = t.Text
	}
	b := ui.ButtonBase(c.Key(tab.doc.ID)).Height(36).Padding(0, 10).Gap(6).
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
		close := ui.ButtonBase(c.Key(tab.doc.ID+"-close")).Size(16, 16).Center().
			Radius(4).Label("关闭").Tooltip("关闭")
		close.Children(func() {
			ui.Icon(c, iconClose).FontSize(11).TextColor(t.TextMuted)
		})
		if close.Clicked() {
			a.CloseTab(i)
			closed = true
		}
	})
	if active {
		b.BorderWidth(0, 0, 2, 0).BorderColor(t.Accent)
	}
	if b.Clicked() && !closed {
		a.active = i
		a.focusEditor = true
		a.updateTitle()
		a.update()
	}
}

func (a *App) newTabButton(c *ui.Context, t *ui.Theme) {
	b := ui.ButtonBase(c.Key("new-tab")).Size(28, 36).Center().Radius(0).
		Label("新建文件").Tooltip("新建文件")
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() {
		ui.Icon(c, iconPlus).FontSize(14).TextColor(t.TextMuted)
	})
	if b.Clicked() {
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
	st.Placeholder = "开始输入…"
	st.ContextMenu = func(m *ui.Menu) {
		a.editorContextMenu(m)
	}
}

func (a *App) editorContextMenu(m *ui.Menu) {
	m.Separator()
	m.Submenu("大小写转换", func(mm *ui.Menu) {
		if mm.Item("转为大写").Chosen() {
			a.ConvertCase(caseconv.Upper)
		}
		if mm.Item("转为小写").Chosen() {
			a.ConvertCase(caseconv.Lower)
		}
		if mm.Item("首字母大写").Chosen() {
			a.ConvertCase(caseconv.Title)
		}
		if mm.Item("反转大小写").Chosen() {
			a.ConvertCase(caseconv.Invert)
		}
	})
	m.Submenu("换行符", func(mm *ui.Menu) {
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
	m.Submenu("编码", func(mm *ui.Menu) {
		mm.Submenu("重新以…编码打开", func(m2 *ui.Menu) {
			for _, kind := range encoding.All {
				k := kind
				if m2.Item(encoding.Label(k)).Chosen() {
					a.ReopenWithEncoding(k)
				}
			}
		})
		mm.Submenu("以…编码保存", func(m2 *ui.Menu) {
			for _, kind := range encoding.All {
				k := kind
				if m2.Item(encoding.Label(k)).Chosen() {
					a.SaveWithEncoding(k)
				}
			}
		})
	})
	if m.Item("关闭标签").Chosen() {
		a.CloseActiveTab()
	}
}

func (a *App) emptyState(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Fill().Center().Gap(14).Children(func() {
		ui.Text(c, "ZPad").FontSize(30).Bold()
		ui.Text(c, "打开文件或文件夹开始编辑").FontSize(13).TextColor(t.TextMuted)
		ui.Row(c).Gap(8).Children(func() {
			if ui.PrimaryButton(c, "打开文件").Clicked() {
				a.OpenFileDialog()
			}
			if ui.Button(c, "打开文件夹").Clicked() {
				a.OpenFolderDialog()
			}
			if ui.Button(c, "新建文件").Clicked() {
				a.NewFile()
			}
		})
	})
}

func (a *App) statusBar(c *ui.Context, t *ui.Theme) {
	ui.Row(c).Height(28).FillWidth().Padding(0, 10).AlignItems(ui.Center).Gap(10).
		Background(t.Surface).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).
		Children(func() {
			tab := a.activeTab()
			if tab == nil {
				msg, col := "就绪", t.TextMuted
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
				m.Submenu("重新以…编码打开", func(mm *ui.Menu) {
					for _, kind := range encoding.All {
						k := kind
						if mm.Item(encoding.Label(k)).Chosen() {
							a.ReopenWithEncoding(k)
						}
					}
				})
				m.Submenu("以…编码保存", func(mm *ui.Menu) {
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

func languageName(doc *Document) string {
	if lang := doc.Buffer.Language(); lang != nil {
		return lang.Name
	}
	return "纯文本"
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
	return fmt.Sprintf("行 %d, 列 %d", line+1, utf8.RuneCountInString(text[:col])+1)
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
	return fmt.Sprintf("已选 %d", n)
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
