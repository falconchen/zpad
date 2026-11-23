package app

import (
	"log"
	"os"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/locales"
)

// Where the About dialog points: the project and its license.
const (
	aboutRepoURL    = "https://github.com/helloxz/zpad"
	aboutLicenseURL = aboutRepoURL + "/blob/main/LICENSE"
)

// loadAboutIcon decodes the icon the binary embeds once, for the About
// dialog. Without one, as in tests, the dialog keeps the letter mark.
func (a *App) loadAboutIcon() {
	a.readAboutIconData(a.appIconPNG)
}

// readAboutIcon decodes the icon at path, as tests do with the one in the
// repository.
func (a *App) readAboutIcon(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	a.readAboutIconData(data)
}

func (a *App) readAboutIconData(data []byte) {
	if len(data) == 0 {
		return
	}
	icon, err := ui.DecodeBitmap(data)
	if err != nil {
		log.Printf("decode app icon: %v", err)
		return
	}
	a.aboutIcon = icon
}

func (a *App) openAbout() {
	a.showSettings = false
	a.renameOpen = false
	a.aboutOpen = true
	// The icon is a 1024×1024 bitmap: decode it when the dialog first
	// opens, not at every launch.
	if a.aboutIcon == nil {
		a.loadAboutIcon()
	}
	a.update()
}

// aboutModal shows the app's name, version, license, and repository. The
// links open in the browser, and the license and repository are the same
// ones the README and the packaged app carry.
func (a *App) aboutModal(c *ui.Context, t *ui.Theme) {
	if !a.aboutOpen {
		return
	}
	ui.DialogBase(c, &a.aboutOpen, func(backdrop, panel ui.Element) {
		c.SetTheme(t)
		backdrop.Background(ui.RGBA(0, 0, 0, 0.4))
		panel.Width(440).Padding(24).Gap(16).
			Background(t.Background).Border(1, t.Border).Radius(14).
			Shadow(0, 16, 48, 0, ui.RGBA(0, 0, 0, 0.28))
		panel.Children(func() {
			ui.Row(c).AlignItems(ui.Center).Gap(14).Children(func() {
				if a.aboutIcon != nil {
					// The icon art keeps its own rounded corners and
					// margin, so it takes no background or radius here.
					ui.Image(c, a.aboutIcon).Size(64, 64)
				} else {
					ui.Box(c).Size(52, 52).Center().Radius(12).Background(t.Accent).Children(func() {
						ui.Text(c, "Z").FontSize(26).Bold().TextColor(t.AccentText)
					})
				}
				ui.Column(c).Gap(2).Children(func() {
					ui.Text(c, "ZPad").FontSize(20).Bold()
					ui.Text(c, locales.AboutVersion(Version)).FontSize(12).TextColor(t.TextMuted)
				})
			})
			ui.Text(c, locales.AboutTagline()).FontSize(12).TextColor(t.TextMuted)
			ui.Divider(c)
			a.aboutFact(c, t, locales.AboutLicense(), "AGPL-3.0", aboutLicenseURL)
			a.aboutFact(c, t, locales.AboutRepository(), "github.com/helloxz/zpad", aboutRepoURL)
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Spacer(c)
				if ui.PrimaryButton(c, locales.Close()).Cursor(ui.CursorPointer).Clicked() {
					a.aboutOpen = false
				}
			})
		})
	})
}

// aboutFact is one label and the link beside it.
func (a *App) aboutFact(c *ui.Context, t *ui.Theme, label, value, url string) {
	ui.Row(c).AlignItems(ui.Center).Gap(10).Children(func() {
		ui.Text(c, label).FontSize(12).TextColor(t.TextMuted).Width(92)
		ui.Link(c, value, url).FontSize(12)
	})
}
