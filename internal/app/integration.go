package app

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/sysint"
	"github.com/helloxz/zpad/locales"
)

func (a *App) refreshIntegrationStatus() {
	if a.sysint == nil || !a.sysint.Supported() {
		a.sysintStatus = sysint.Status{}
		return
	}
	st := a.sysint.Status()
	a.sysintStatus = st
	if a.running {
		a.Settings.EnableCLI = st.CLIInstalled
		a.Settings.FinderFolderMenu = st.FinderFolderInstalled
		a.Settings.FinderFileMenu = st.FinderFileInstalled
	}
}

func (a *App) toggleCLI() {
	if a.sysint == nil {
		return
	}
	if a.Settings.EnableCLI {
		path, err := a.sysint.InstallCLI()
		if err != nil {
			a.Settings.EnableCLI = false
			a.SetErrorToast(locales.IntegrationFailed(), err.Error())
			return
		}
		a.sysintStatus = a.sysint.Status()
		if path != "" {
			a.sysintStatus.CLIPath = path
			a.sysintStatus.CLIInstalled = true
		}
		a.SetToast(locales.CliCommandEnabled())
		return
	}
	if err := a.sysint.UninstallCLI(); err != nil {
		a.Settings.EnableCLI = true
		a.SetErrorToast(locales.IntegrationFailed(), err.Error())
		return
	}
	a.sysintStatus = a.sysint.Status()
	a.SetToast(locales.CliCommandDisabled())
}

func (a *App) finderMenuOptions() sysint.FinderMenuOptions {
	return sysint.FinderMenuOptions{
		Folder:                 a.Settings.FinderFolderMenu,
		File:                   a.Settings.FinderFileMenu,
		OpenCurrentFolderTitle: locales.FinderOpenCurrentFolderAction(),
		OpenFolderTitle:        locales.FinderOpenFolderAction(),
		OpenFileTitle:          locales.FinderOpenFileAction(),
	}
}

func (a *App) toggleFinderFolderMenu() {
	if a.sysint == nil {
		return
	}
	if err := a.sysint.SyncFinderMenu(a.finderMenuOptions()); err != nil {
		a.Settings.FinderFolderMenu = !a.Settings.FinderFolderMenu
		a.SetErrorToast(locales.IntegrationFailed(), err.Error())
		return
	}
	a.sysintStatus = a.sysint.Status()
	if a.Settings.FinderFolderMenu {
		a.SetToast(locales.FinderFolderMenuEnabled())
	} else {
		a.SetToast(locales.FinderFolderMenuDisabled())
	}
}

func (a *App) toggleFinderFileMenu() {
	if a.sysint == nil {
		return
	}
	if err := a.sysint.SyncFinderMenu(a.finderMenuOptions()); err != nil {
		a.Settings.FinderFileMenu = !a.Settings.FinderFileMenu
		a.SetErrorToast(locales.IntegrationFailed(), err.Error())
		return
	}
	a.sysintStatus = a.sysint.Status()
	if a.Settings.FinderFileMenu {
		a.SetToast(locales.FinderFileMenuEnabled())
	} else {
		a.SetToast(locales.FinderFileMenuDisabled())
	}
}

func (a *App) syncFinderMenuTitles() {
	if a.sysint == nil || !a.sysint.Supported() {
		return
	}
	if a.Settings.FinderFolderMenu || a.Settings.FinderFileMenu {
		_ = a.sysint.SyncFinderMenu(a.finderMenuOptions())
	}
}

func (a *App) toggleDefaultAppEnabled() {
	if a.sysint == nil {
		return
	}
	a.customExtErr = ""
	if a.Settings.DefaultAppEnabled {
		if len(a.Settings.DefaultAppExts) == 0 {
			return
		}
		if a.Settings.PreviousDefaultApps == nil {
			a.Settings.PreviousDefaultApps = map[string]string{}
		}
		for _, ext := range a.Settings.DefaultAppExts {
			if _, err := a.sysint.AssociateExt(ext, a.Settings.PreviousDefaultApps); err != nil {
				a.SetErrorToast(locales.IntegrationFailed(), err.Error())
				break
			}
		}
		return
	}
	if err := a.restoreAllDefaultApps(false); err != nil {
		a.SetErrorToast(locales.IntegrationFailed(), err.Error())
		return
	}
	if len(a.Settings.DefaultAppExts) > 0 {
		a.SetToast(locales.DefaultAppsRestored())
	}
}

func (a *App) restoreAllDefaultApps(clearList bool) error {
	if a.sysint == nil {
		return nil
	}
	var firstErr error
	for _, ext := range a.Settings.DefaultAppExts {
		if err := a.sysint.RestoreExt(ext, a.Settings.PreviousDefaultApps); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if clearList {
		a.Settings.DefaultAppExts = nil
	}
	if len(a.Settings.PreviousDefaultApps) == 0 {
		a.Settings.PreviousDefaultApps = nil
	}
	return firstErr
}

func (a *App) disableAllIntegrations() {
	if a.sysint == nil || !a.sysint.Supported() {
		return
	}
	if a.Settings.EnableCLI || a.sysint.HasFinderWorkflow(sysint.FinderKindFolder) || a.sysint.HasFinderWorkflow(sysint.FinderKindFile) {
		_ = a.sysint.UninstallCLI()
		_ = a.sysint.UninstallAllFinderWorkflows()
	} else {
		_, installed := a.sysint.CLIStatus()
		if installed {
			_ = a.sysint.UninstallCLI()
		}
	}
	if a.Settings.DefaultAppEnabled && len(a.Settings.DefaultAppExts) > 0 {
		_ = a.restoreAllDefaultApps(true)
	}
}

func (a *App) hasDefaultExt(ext string) bool {
	return slices.Contains(a.Settings.DefaultAppExts, ext)
}

func (a *App) toggleDefaultExt(raw string) {
	if a.sysint == nil || !a.Settings.DefaultAppEnabled {
		return
	}
	ext, err := sysint.NormalizeExt(raw)
	if err != nil {
		return
	}
	a.customExtErr = ""
	if idx := slices.Index(a.Settings.DefaultAppExts, ext); idx >= 0 {
		if err := a.sysint.RestoreExt(ext, a.Settings.PreviousDefaultApps); err != nil {
			a.SetErrorToast(locales.IntegrationFailed(), err.Error())
			return
		}
		a.Settings.DefaultAppExts = slices.Delete(a.Settings.DefaultAppExts, idx, idx+1)
		if len(a.Settings.DefaultAppExts) == 0 {
			a.Settings.DefaultAppExts = nil
		}
		if len(a.Settings.PreviousDefaultApps) == 0 {
			a.Settings.PreviousDefaultApps = nil
		}
		a.saveSettings()
		a.update()
		return
	}
	if a.Settings.PreviousDefaultApps == nil {
		a.Settings.PreviousDefaultApps = map[string]string{}
	}
	if _, err := a.sysint.AssociateExt(ext, a.Settings.PreviousDefaultApps); err != nil {
		a.SetErrorToast(locales.IntegrationFailed(), err.Error())
		return
	}
	a.Settings.DefaultAppExts = append(a.Settings.DefaultAppExts, ext)
	a.saveSettings()
	a.update()
}

func (a *App) applyExtGroup(exts []string, enable bool) {
	if a.sysint == nil || !a.Settings.DefaultAppEnabled {
		return
	}
	a.customExtErr = ""
	if enable && a.Settings.PreviousDefaultApps == nil {
		a.Settings.PreviousDefaultApps = map[string]string{}
	}
	var firstErr error
	for _, ext := range exts {
		active := a.hasDefaultExt(ext)
		if enable && !active {
			if _, err := a.sysint.AssociateExt(ext, a.Settings.PreviousDefaultApps); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			a.Settings.DefaultAppExts = append(a.Settings.DefaultAppExts, ext)
		} else if !enable && active {
			if err := a.sysint.RestoreExt(ext, a.Settings.PreviousDefaultApps); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			idx := slices.Index(a.Settings.DefaultAppExts, ext)
			if idx >= 0 {
				a.Settings.DefaultAppExts = slices.Delete(a.Settings.DefaultAppExts, idx, idx+1)
			}
		}
	}
	if len(a.Settings.DefaultAppExts) == 0 {
		a.Settings.DefaultAppExts = nil
	}
	if len(a.Settings.PreviousDefaultApps) == 0 {
		a.Settings.PreviousDefaultApps = nil
	}
	if firstErr != nil {
		a.SetErrorToast(locales.IntegrationFailed(), firstErr.Error())
	}
	a.saveSettings()
	a.update()
}

func (a *App) clearAllDefaultExts() {
	if a.sysint == nil || !a.Settings.DefaultAppEnabled || len(a.Settings.DefaultAppExts) == 0 {
		return
	}
	if err := a.restoreAllDefaultApps(true); err != nil {
		a.SetErrorToast(locales.IntegrationFailed(), err.Error())
	} else {
		a.SetToast(locales.DefaultAppsRestored())
	}
	a.saveSettings()
	a.update()
}

func (a *App) addCustomDefaultExts() {
	if a.sysint == nil || !a.Settings.DefaultAppEnabled {
		return
	}
	parts := strings.FieldsFunc(a.customExtInput, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '，'
	})
	if len(parts) == 0 {
		return
	}
	var normalized []string
	for _, raw := range parts {
		ext, err := sysint.NormalizeExt(raw)
		if err != nil {
			a.customExtErr = locales.InvalidExtension()
			a.update()
			return
		}
		if !slices.Contains(normalized, ext) {
			normalized = append(normalized, ext)
		}
	}
	a.customExtErr = ""
	if a.Settings.PreviousDefaultApps == nil {
		a.Settings.PreviousDefaultApps = map[string]string{}
	}
	for _, ext := range normalized {
		if a.hasDefaultExt(ext) {
			continue
		}
		if _, err := a.sysint.AssociateExt(ext, a.Settings.PreviousDefaultApps); err != nil {
			a.SetErrorToast(locales.IntegrationFailed(), err.Error())
			a.saveSettings()
			a.update()
			return
		}
		a.Settings.DefaultAppExts = append(a.Settings.DefaultAppExts, ext)
	}
	a.customExtInput = ""
	a.saveSettings()
	a.update()
}

func (a *App) customDefaultExts() []string {
	var out []string
	for _, ext := range a.Settings.DefaultAppExts {
		if !sysint.IsPresetExt(ext) {
			out = append(out, ext)
		}
	}
	return out
}

func extGroupLabel(id string) string {
	switch id {
	case "text":
		return locales.ExtGroupText()
	case "config":
		return locales.ExtGroupConfig()
	case "web":
		return locales.ExtGroupWeb()
	case "code":
		return locales.ExtGroupCode()
	default:
		return id
	}
}

func (a *App) settingsIntegration(c *ui.Context, t *ui.Theme) {
	if a.sysint == nil || !a.sysint.Supported() {
		return
	}
	a.settingsSection(c, t, iconTerminal, locales.IntegrationSection(), func() {
		cliDesc := locales.CliCommandDescription()
		if a.Settings.EnableCLI && a.sysintStatus.CLIPath != "" {
			cliDesc = locales.CliCommandInstalledDescription(a.sysintStatus.CLIPath)
		}
		a.switchSettingAction(c, t, locales.CliCommand(), cliDesc,
			&a.Settings.EnableCLI, false, false, a.toggleCLI)

		a.switchSettingAction(c, t, locales.FinderFolderMenu(), locales.FinderFolderMenuDescription(),
			&a.Settings.FinderFolderMenu, false, false, a.toggleFinderFolderMenu)

		a.switchSettingAction(c, t, locales.FinderFileMenu(), locales.FinderFileMenuDescription(),
			&a.Settings.FinderFileMenu, false, false, a.toggleFinderFileMenu)

		a.switchSettingAction(c, t, locales.DefaultAppAssoc(), locales.DefaultAppAssocDescription(),
			&a.Settings.DefaultAppEnabled, !a.Settings.DefaultAppEnabled, false, a.toggleDefaultAppEnabled)

		if a.Settings.DefaultAppEnabled {
			a.settingsDefaultExtPicker(c, t)
		}
	})
}

func (a *App) settingsDefaultExtPicker(c *ui.Context, t *ui.Theme) {
	ui.Column(c).FillWidth().Padding(t.Space(3.5), t.Space(4.5)).Gap(t.Space(3.5)).Children(func() {
		for _, group := range sysint.DefaultExtGroups() {
			allSelected := true
			anySelected := false
			for _, ext := range group.Exts {
				if a.hasDefaultExt(ext) {
					anySelected = true
				} else {
					allSelected = false
				}
			}
			ui.Column(c).FillWidth().Gap(t.Space(2)).Children(func() {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(8).Children(func() {
					ui.Text(c, extGroupLabel(group.ID)).FontSize(12).Bold().TextColor(t.TextMuted)
					ui.Spacer(c)
					if !allSelected {
						a.extSmallAction(c, t, locales.SelectAll(), func() {
							a.applyExtGroup(group.Exts, true)
						})
					}
					if anySelected {
						a.extSmallAction(c, t, locales.ClearAll(), func() {
							a.applyExtGroup(group.Exts, false)
						})
					}
				})
				ui.Row(c).FillWidth().Wrap().Gap(6).Children(func() {
					for _, ext := range group.Exts {
						a.extChip(c, t, ext, a.hasDefaultExt(ext), false)
					}
				})
			})
		}

		// Custom extensions & input row.
		ui.Column(c).FillWidth().Gap(t.Space(2)).Children(func() {
			ui.Text(c, locales.ExtGroupCustom()).FontSize(12).Bold().TextColor(t.TextMuted)
			customs := a.customDefaultExts()
			if len(customs) > 0 {
				ui.Row(c).FillWidth().Wrap().Gap(6).Children(func() {
					for _, ext := range customs {
						a.extChip(c, t, ext, true, true)
					}
				})
			}
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(8).Children(func() {
				in := ui.TextInput(c, &a.customExtInput).Grow(1).
					Label(locales.ExtGroupCustom()).
					Placeholder(locales.CustomExtPlaceholder())
				submit := in.Submitted()
				if in.Changed() && a.customExtErr != "" {
					a.customExtErr = ""
				}
				if ui.Button(c, locales.Add()).Cursor(ui.CursorPointer).
					Disabled(strings.TrimSpace(a.customExtInput) == "").Clicked() {
					submit = true
				}
				if submit {
					a.addCustomDefaultExts()
				}
			})
			if a.customExtErr != "" {
				ui.Text(c, a.customExtErr).FontSize(11).TextColor(t.Danger)
			}
		})

		if len(a.Settings.DefaultAppExts) > 0 {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(8).Children(func() {
				ui.Text(c, locales.DefaultAppCount(len(a.Settings.DefaultAppExts))).
					FontSize(12).TextColor(t.TextMuted)
				ui.Spacer(c)
				if ui.Button(c, locales.ClearAll()).Cursor(ui.CursorPointer).Clicked() {
					a.clearAllDefaultExts()
				}
			})
		}
	})
}

func (a *App) extChip(c *ui.Context, t *ui.Theme, ext string, active, removable bool) {
	label := "." + ext
	if removable {
		label += " ×"
	}
	b := ui.ButtonBase(c).Padding(4, 10).Radius(6).Cursor(ui.CursorPointer).Label("." + ext)
	switch {
	case active:
		b.Background(t.Accent).Border(1, t.Accent)
	case b.Hovered():
		b.Background(t.SurfaceHover).Border(1, t.Border)
	default:
		b.Background(t.Background).Border(1, t.Border)
	}
	b.Children(func() {
		color := t.Text
		if active {
			color = t.AccentText
		}
		ui.Text(c, label).FontSize(12).TextColor(color).SingleLine()
	})
	if b.Clicked() {
		a.toggleDefaultExt(ext)
	}
}

func (a *App) extSmallAction(c *ui.Context, t *ui.Theme, label string, onClick func()) {
	b := ui.ButtonBase(c).Padding(2, 6).Radius(4).Cursor(ui.CursorPointer).Label(label)
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() {
		ui.Text(c, label).FontSize(11).TextColor(t.Accent).SingleLine()
	})
	if b.Clicked() {
		onClick()
	}
}
