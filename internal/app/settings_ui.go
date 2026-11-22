package app

import (
	"sort"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/encoding"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func (a *App) openSettings() {
	a.renameOpen = false
	a.aboutOpen = false
	a.fontSize = float64(a.Settings.FontSize)
	a.lineHeight = float64(a.Settings.LineHeight)
	a.showSettings = true
	a.update()
}

func (a *App) fontNames() []string {
	if a.fontFamilies != nil {
		return a.fontFamilies
	}
	var names []string
	for _, f := range ui.FontFamilies() {
		if f.Monospaced {
			names = append(names, f.Name)
		}
	}
	sort.SliceStable(names, func(i, j int) bool {
		if names[i] == "JetBrains Mono" {
			return true
		}
		if names[j] == "JetBrains Mono" {
			return false
		}
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})
	a.fontFamilies = names
	return names
}

func (a *App) settingsModal(c *ui.Context, t *ui.Theme) {
	if !a.showSettings {
		return
	}
	ui.DialogBase(c, &a.showSettings, func(backdrop, panel ui.Element) {
		c.SetTheme(t)
		w, h := c.Size()
		cw := min(float32(620), w-80)
		if cw < 320 {
			cw = w - 32
		}
		ch := min(float32(660), h-80)
		if ch < 320 {
			ch = h - 32
		}
		backdrop.Background(ui.RGBA(0, 0, 0, 0.4))
		panel.Width(cw).Height(ch).Padding(20).Gap(14).
			Background(t.Background).Border(1, t.Border).Radius(14).
			Shadow(0, 16, 48, 0, ui.RGBA(0, 0, 0, 0.28))
		panel.Children(func() {
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, locales.SettingsTitle()).FontSize(18).Bold()
			})
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Gap(18).Children(func() {
					a.settingsAppearance(c, t)
					a.settingsEditor(c, t)
					a.settingsFiles(c, t)
					a.settingsWorkspace(c, t)
				})
			})
			ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
				if ui.Button(c, locales.RestoreDefaults()).Cursor(ui.CursorPointer).Clicked() {
					go a.confirmRestoreDefaults()
				}
				ui.Spacer(c)
				if ui.Button(c, locales.Close()).Cursor(ui.CursorPointer).Clicked() {
					a.showSettings = false
				}
				if ui.PrimaryButton(c, locales.Done()).Cursor(ui.CursorPointer).Clicked() {
					a.showSettings = false
				}
			})
		})
	})
}

func (a *App) section(c *ui.Context, t *ui.Theme, title string) {
	ui.Text(c, title).FontSize(12).Bold().TextColor(t.TextMuted).Padding(0, 4)
}

func (a *App) settingsAppearance(c *ui.Context, t *ui.Theme) {
	a.section(c, t, locales.Appearance())
	a.selectField(c, locales.Theme(), a.themeLabel(a.Settings.Theme), themeLabels(), func(label string) {
		switch label {
		case locales.ThemeLight():
			a.ThemeIndex = 0
		case locales.ThemeDark():
			a.ThemeIndex = 1
		default:
			a.ThemeIndex = 2
		}
		a.setTheme()
	})
	a.selectField(c, locales.Language(), a.languageLabel(a.Settings.Language), languageLabels(), func(label string) {
		a.setLanguage(languageTag(label))
	})
}

func (a *App) settingsEditor(c *ui.Context, t *ui.Theme) {
	a.section(c, t, locales.EditorSection())

	options := append([]string{locales.SystemMonospace()}, a.fontNames()...)
	current := a.Settings.FontFamily
	if current == "" {
		current = locales.SystemMonospace()
	}
	a.selectField(c, locales.FontFamily(), current, options, func(label string) {
		if label == locales.SystemMonospace() {
			a.Settings.FontFamily = ""
		} else {
			a.Settings.FontFamily = label
		}
	})

	ui.Field(c, locales.FontSize(), func() {
		if ui.NumberInput(c, &a.fontSize, 10, 28, 1).Width(120).Changed() {
			a.Settings.FontSize = float32(a.fontSize)
			a.saveSettings()
			a.update()
		}
	})
	ui.Field(c, locales.LineHeight(), func() {
		if ui.NumberInput(c, &a.lineHeight, 1.2, 2.0, 0.1).Width(120).Changed() {
			a.Settings.LineHeight = float32(a.lineHeight)
			a.saveSettings()
			a.update()
		}
	})
	a.switchField(c, locales.SoftWrap(), &a.Settings.SoftWrap)
	a.selectField(c, locales.LineNumbers(), lineNumbersLabel(a.Settings.LineNumbers),
		[]string{locales.LineNumbersAbsolute(), locales.LineNumbersRelative(), locales.LineNumbersOff()},
		func(label string) {
			switch label {
			case locales.LineNumbersRelative():
				a.Settings.LineNumbers = "relative"
			case locales.LineNumbersOff():
				a.Settings.LineNumbers = "off"
			default:
				a.Settings.LineNumbers = "absolute"
			}
		})
	a.selectField(c, locales.TabSize(), tabSizeLabel(a.Settings.TabSize),
		[]string{"2", "4", "8"}, func(label string) {
			switch label {
			case "2":
				a.Settings.TabSize = 2
			case "8":
				a.Settings.TabSize = 8
			default:
				a.Settings.TabSize = 4
			}
		})
	a.selectField(c, locales.Indent(), indentLabel(a.Settings.UseTabs),
		[]string{locales.IndentSpaces(), locales.IndentTabs()}, func(label string) {
			a.Settings.UseTabs = label == locales.IndentTabs()
		})
	a.switchField(c, locales.IndentGuides(), &a.Settings.ShowGuides)
	a.switchField(c, locales.AutoClose(), &a.Settings.AutoClose)
}

func (a *App) settingsFiles(c *ui.Context, t *ui.Theme) {
	a.section(c, t, locales.FilesSection())
	a.selectField(c, locales.DefaultEncoding(), encoding.Label(encoding.Parse(a.Settings.DefaultEncoding)),
		encodingLabels(), func(label string) {
			a.Settings.DefaultEncoding = string(encodingFromLabel(label))
		})
	a.selectField(c, locales.DefaultLineEnding(), eolLabel(a.Settings.DefaultEOL),
		[]string{"LF", "CRLF", "CR"}, func(label string) {
			switch label {
			case "CRLF":
				a.Settings.DefaultEOL = "\r\n"
			case "CR":
				a.Settings.DefaultEOL = "\r"
			default:
				a.Settings.DefaultEOL = "\n"
			}
		})
	a.selectField(c, locales.LargeFileThreshold(), largeFileLabel(a.Settings.LargeFileMB),
		[]string{"10 MB", "20 MB", "50 MB"}, func(label string) {
			switch label {
			case "10 MB":
				a.Settings.LargeFileMB = 10
			case "50 MB":
				a.Settings.LargeFileMB = 50
			default:
				a.Settings.LargeFileMB = 20
			}
		})
}

func (a *App) settingsWorkspace(c *ui.Context, t *ui.Theme) {
	a.section(c, t, locales.WorkspaceSection())
	a.switchField(c, locales.SidebarDefault(), &a.Settings.SidebarOpen)
	a.switchField(c, locales.RestoreSession(), &a.Settings.RestoreSession)
	a.switchField(c, locales.ConfirmClose(), &a.Settings.ConfirmClose)
}

func (a *App) selectField(c *ui.Context, label, current string, options []string, apply func(string)) {
	ui.Field(c, label, func() {
		sel := current
		if ui.Select(c, &sel, options).Width(300).Cursor(ui.CursorPointer).Changed() {
			apply(sel)
			a.saveSettings()
			a.update()
		}
	})
}

func (a *App) switchField(c *ui.Context, label string, v *bool) {
	ui.Field(c, label, func() {
		if ui.Switch(c, v).Cursor(ui.CursorPointer).Changed() {
			a.saveSettings()
			a.update()
		}
	})
}

func (a *App) setLanguage(tag string) {
	a.Settings.Language = tag
	a.saveSettings()
	mygo.App.SetLocale(tag)
	a.InstallMenu()
	a.updateTitle()
	a.update()
}

func (a *App) confirmRestoreDefaults() {
	res, err := mygo.Dialog.Message(mygo.MessageOptions{
		Parent:             a.win,
		Type:               mygo.MessageWarning,
		Message:            locales.RestoreDefaultsQuestion(),
		Buttons:            []string{locales.Cancel(), locales.Restore()},
		DefaultButton:      1,
		CancelButton:       0,
		DestructiveButtons: []int{1},
	})
	if err != nil || res.Button != 1 {
		return
	}
	a.onUI(func() { a.restoreDefaults() })
}

func (a *App) restoreDefaults() {
	a.Settings = settings.Default()
	a.ThemeIndex = 2
	a.fontSize = float64(a.Settings.FontSize)
	a.lineHeight = float64(a.Settings.LineHeight)
	if a.tree != nil {
		a.tree.SetShowHidden(a.Settings.ShowHidden)
		a.treeGeneration++
	}
	a.saveSettings()
	SetThemeSource("system")
	if a.running {
		a.Dark = mygo.Theme.IsDark()
	}
	mygo.App.SetLocale("")
	a.InstallMenu()
	a.updateTitle()
	a.SetToast(locales.SettingsReset())
	a.update()
}

func (a *App) themeLabel(mode string) string {
	switch mode {
	case "light":
		return locales.ThemeLight()
	case "dark":
		return locales.ThemeDark()
	default:
		return locales.ThemeSystem()
	}
}

func themeLabels() []string {
	return []string{locales.ThemeLight(), locales.ThemeDark(), locales.ThemeSystem()}
}

func (a *App) languageLabel(tag string) string {
	switch tag {
	case "zh-Hans":
		return locales.LanguageChinese()
	case "en":
		return locales.LanguageEnglish()
	default:
		return locales.LanguageSystem()
	}
}

func languageLabels() []string {
	return []string{locales.LanguageSystem(), locales.LanguageChinese(), locales.LanguageEnglish()}
}

func languageTag(label string) string {
	switch label {
	case locales.LanguageChinese():
		return "zh-Hans"
	case locales.LanguageEnglish():
		return "en"
	default:
		return ""
	}
}

func lineNumbersLabel(v string) string {
	switch v {
	case "relative":
		return locales.LineNumbersRelative()
	case "off":
		return locales.LineNumbersOff()
	default:
		return locales.LineNumbersAbsolute()
	}
}

func tabSizeLabel(v int) string {
	switch v {
	case 2:
		return "2"
	case 8:
		return "8"
	default:
		return "4"
	}
}

func indentLabel(useTabs bool) string {
	if useTabs {
		return locales.IndentTabs()
	}
	return locales.IndentSpaces()
}

func encodingLabels() []string {
	labels := make([]string, 0, len(encoding.All))
	for _, kind := range encoding.All {
		labels = append(labels, encoding.Label(kind))
	}
	return labels
}

func encodingFromLabel(label string) encoding.Kind {
	for _, kind := range encoding.All {
		if encoding.Label(kind) == label {
			return kind
		}
	}
	return encoding.UTF8
}

func largeFileLabel(v int) string {
	switch v {
	case 10:
		return "10 MB"
	case 50:
		return "50 MB"
	default:
		return "20 MB"
	}
}
