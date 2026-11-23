package app

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/theme"
	"github.com/helloxz/zpad/locales"
)

// searchModal draws the search panel over the window: a field for the
// query, its toggles, the matches grouped by document, and their count.
// DialogBase closes it on Escape and on a click on the backdrop.
func (a *App) searchModal(c *ui.Context, t *ui.Theme) {
	if !a.searchOpen {
		return
	}
	width, height := c.Size()
	panelW := min(float32(640), width-40)
	if panelW < 280 {
		panelW = max(width-24, 240)
	}
	panelH := min(float32(560), height*0.72)
	if panelH < 220 {
		panelH = max(height-80, 200)
	}
	ui.DialogBase(c, &a.searchOpen, func(backdrop, panel ui.Element) {
		c.SetTheme(t)
		backdrop.Background(ui.RGBA(0, 0, 0, 0.35))
		panel.Width(panelW).Height(panelH).Padding(14).Gap(10).
			Background(t.Background).Border(1, t.Border).Radius(14).
			Shadow(0, 16, 48, 0, ui.RGBA(0, 0, 0, 0.28))
		panel.Children(func() {
			a.searchField(c, t)
			a.searchOptions(c, t)
			a.searchBody(c, t)
		})
	})
}

func (a *App) searchField(c *ui.Context, t *ui.Theme) {
	ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
		field := ui.Row(c).Height(32).Grow(1).AlignItems(ui.Center).Gap(4).
			Padding(0, 10).Radius(6).Background(t.Surface).Border(1, t.Border)
		var in ui.Element
		field.Children(func() {
			in = ui.TextInputBase(c, &a.searchQuery).Grow(1).
				Label(locales.SearchOpenFilesLabel()).
				Placeholder(locales.SearchOpenFilesLabel())
			if a.searchFocus {
				in.Focus()
				a.searchFocus = false
			}
			if in.Changed() {
				a.searchInputChanged()
			}
			if in.Submitted() {
				a.submitSearch()
			}
			if in.Shortcut(0, ui.KeyDown) {
				a.moveSearchSelection(1)
			}
			if in.Shortcut(0, ui.KeyUp) {
				a.moveSearchSelection(-1)
			}
			if a.searchQuery != "" {
				a.searchClearButton(c, t)
			}
		})
		if in.FocusWithin() {
			field.BorderColor(t.Focus)
		}
		if field.Clicked() {
			in.Focus()
		}
		tooltipBelow(c, t, in, locales.SearchNavigateHint())

		b := ui.ButtonBase(c).Size(30, 30).Center().Radius(6).
			Background(t.Surface).Border(1, t.Border).
			Label(locales.SearchRun()).Cursor(ui.CursorPointer).
			Disabled(a.searchQuery == "")
		if b.Hovered() {
			b.Background(t.SurfaceHover)
		}
		b.Children(func() {
			ui.Icon(c, iconSearch).FontSize(16).TextColor(t.Text)
		})
		if b.Clicked() {
			a.startSearch()
			// The button takes the focus on a click; give it back, so the
			// next Enter goes to the field and the query stays editable.
			a.searchFocus = true
		}
		tooltipBelow(c, t, b, locales.SearchRun()+" (↵)")
	})
}

// searchClearButton puts the clear button at the end of the search field,
// inside its border. It shows only while the field holds a query.
func (a *App) searchClearButton(c *ui.Context, t *ui.Theme) {
	b := ui.ButtonBase(c).Size(18, 18).Center().Radius(9).
		Label(locales.SearchClear()).Cursor(ui.CursorPointer).KeepFocus()
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() {
		ui.Icon(c, iconClose).FontSize(12).TextColor(t.TextMuted)
	})
	if b.Clicked() {
		a.clearSearchInput()
	}
	tooltipBelow(c, t, b, locales.SearchClear())
}

// searchOptions draws the toggles and what the last search found. A
// toggle searches again at once, as it changes what matches.
func (a *App) searchOptions(c *ui.Context, t *ui.Theme) {
	ui.Row(c).AlignItems(ui.Center).Gap(4).Children(func() {
		a.searchToggle(c, t, "Aa", locales.SearchCaseSensitive(), &a.searchCase)
		a.searchToggle(c, t, "ab", locales.SearchWholeWord(), &a.searchWord)
		a.searchToggle(c, t, ".*", locales.SearchRegex(), &a.searchRegex)
		ui.Spacer(c)
		status, color := a.searchStatus(t)
		if status != "" {
			ui.Text(c, status).FontSize(11).TextColor(color).SingleLine().Tooltip(status)
		}
	})
}

func (a *App) searchToggle(c *ui.Context, t *ui.Theme, text, label string, on *bool) {
	b := ui.ButtonBase(c).Size(24, 22).Center().Radius(4).Label(label).
		Cursor(ui.CursorPointer)
	switch {
	case *on:
		b.Background(t.Accent).Border(1, t.Accent)
	case b.Hovered():
		b.Background(t.SurfaceHover)
	}
	b.Children(func() {
		color := t.TextMuted
		if *on {
			color = t.AccentText
		}
		ui.Text(c, text).FontSize(11).Bold().TextColor(color)
	})
	if b.Clicked() {
		*on = !*on
		// Toggling changes what matches; as typing does, it waits for
		// Enter or the search button rather than searching on its own.
		a.searchInputChanged()
		// The button takes the focus on a click; give it back, so the
		// next Enter goes to the field.
		a.searchFocus = true
	}
	tooltipBelow(c, t, b, label)
}

// searchStatus says what the panel found, and in which color: an error in
// the danger color, a scan on its way or a current set of rows in the
// muted one. A query or toggles not searched yet leave it empty, the
// panel's hint saying what to do.
func (a *App) searchStatus(t *ui.Theme) (string, ui.Color) {
	switch {
	case a.searchErr != "":
		return a.searchErr, t.Danger
	case a.searchBusy:
		return locales.Searching(), t.TextMuted
	case !a.searchFresh():
		return "", t.TextMuted
	}
	results := 0
	for _, row := range a.searchRows {
		if !row.header {
			results++
		}
	}
	if results == 0 {
		return "", t.TextMuted
	}
	status := locales.SearchResults(results) + " · " + locales.SearchFiles(a.searchFiles)
	if a.searchTruncated {
		status += " · " + locales.SearchTruncated()
	}
	return status, t.TextMuted
}

func (a *App) searchBody(c *ui.Context, t *ui.Theme) {
	if len(a.searchRows) == 0 {
		msg := locales.SearchHint()
		switch {
		case a.searchBusy:
			msg = locales.Searching()
		case a.searchErr != "":
			msg = a.searchErr
		case a.searchQuery == "":
			msg = locales.SearchHint()
		case !a.searchFresh():
			msg = locales.SearchPressEnter()
		default:
			msg = locales.SearchNoResults()
			if len(a.openDocuments()) == 0 {
				msg = locales.SearchNoOpenFiles()
			}
		}
		ui.Column(c).Grow(1).Center().Children(func() {
			ui.Text(c, msg).FontSize(12).TextColor(t.TextMuted).SingleLine()
		})
		return
	}

	a.searchList.Selected = &a.searchSel
	a.searchList.Header = func(i int) bool {
		return i >= 0 && i < len(a.searchRows) && a.searchRows[i].header
	}
	a.searchList.Label = func(i int) string {
		if i < 0 || i >= len(a.searchRows) {
			return ""
		}
		row := a.searchRows[i]
		if row.header {
			return row.doc.Name()
		}
		return fmt.Sprintf("%s:%d:%d", row.hit.doc.Name(), row.hit.line+1, row.hit.col+1)
	}
	a.searchList.Key = func(i int) any {
		if i < 0 || i >= len(a.searchRows) {
			return i
		}
		row := a.searchRows[i]
		if row.header {
			return "file:" + row.doc.ID
		}
		return fmt.Sprintf("match:%s:%d:%d", row.hit.doc.ID, row.hit.line, row.hit.col)
	}
	var list ui.Element
	ui.Themed(c, theme.ResultList(t)).Grow(1).Background(ui.Transparent).Children(func() {
		list = ui.List(c, &a.searchList, len(a.searchRows), func(i int) {
			if i < 0 || i >= len(a.searchRows) {
				return
			}
			row := a.searchRows[i]
			if row.header {
				a.searchHeaderRow(c, t, row)
				return
			}
			a.searchHitRow(c, t, row.hit)
		}).Grow(1)
	})
	if list.Submitted() {
		a.openSelectedHit()
	}
}

func (a *App) searchHeaderRow(c *ui.Context, t *ui.Theme, row searchRow) {
	h := ui.Row(c).FillWidth().Height(26).Padding(0, 8).Gap(6).
		AlignItems(ui.Center).Background(t.Surface)
	h.Children(func() {
		ui.Icon(c, iconFile).FontSize(12).TextColor(t.TextMuted).Shrink(0)
		ui.Text(c, row.doc.Name()).FontSize(12).Bold().SingleLine().TruncateMiddle()
		ui.Spacer(c)
		ui.Text(c, fmt.Sprintf("%d", row.count)).FontSize(11).TextColor(t.TextMuted).Shrink(0)
	})
	if row.doc.Path != "" {
		h.Tooltip(row.doc.Path)
	}
}

func (a *App) searchHitRow(c *ui.Context, t *ui.Theme, hit searchHit) {
	before, match, after := splitPreview(hit)
	row := ui.Row(c).FillWidth().Height(24).Padding(0, 8).Gap(8).
		AlignItems(ui.Center).Cursor(ui.CursorPointer)
	row.Children(func() {
		ui.Text(c, fmt.Sprintf("%d", hit.line+1)).FontSize(11).TextColor(t.TextMuted).
			Width(36).TextAlign(ui.End).SingleLine().Shrink(0)
		ui.RichText(c,
			ui.Span{Text: before},
			ui.Span{Text: match, Background: t.Selection, Radius: 2},
			ui.Span{Text: after},
		).FontSize(12).SingleLine().Grow(1)
	})
	if row.RowClicked() {
		a.revealHit(hit)
	}
}

func (a *App) openSelectedHit() {
	if hit, ok := a.selectedSearchHit(); ok {
		a.revealHit(hit)
	}
}

// moveSearchSelection steps the chosen match by dir, round the ends, and
// scrolls it into view. The keys that move it from the field skip the
// file headers.
func (a *App) moveSearchSelection(dir int) {
	n := len(a.searchRows)
	if n == 0 {
		return
	}
	start := a.searchSel
	if start < 0 || start >= n {
		start = 0
	}
	for i := 1; i <= n; i++ {
		row := (start + dir*i) % n
		if row < 0 {
			row += n
		}
		if !a.searchRows[row].header {
			a.searchSel = row
			a.searchList.ScrollIntoView(row)
			a.update()
			return
		}
	}
}
