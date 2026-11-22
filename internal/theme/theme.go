// Package theme builds ZPad's light and dark palettes for the native UI.
package theme

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

// For returns the palette for an appearance.
func For(dark bool) *ui.Theme {
	if dark {
		return darkTheme()
	}
	return lightTheme()
}

// Light returns the light palette.
func Light() *ui.Theme { return lightTheme() }

// Dark returns the dark palette.
func Dark() *ui.Theme { return darkTheme() }

func lightTheme() *ui.Theme {
	t := ui.LightTheme()
	t.Background = ui.Hex("#FFFFFF")
	t.Surface = ui.Hex("#F7F8FA")
	t.SurfaceHover = ui.Hex("#EEF0F4")
	t.SurfacePressed = ui.Hex("#E2E5EA")
	t.Border = ui.Hex("#E5E7EB")
	t.Text = ui.Hex("#111827")
	t.TextMuted = ui.Hex("#6B7280")
	t.Accent = ui.Hex("#4F46E5")
	t.AccentHover = ui.Hex("#4338CA")
	t.AccentPressed = ui.Hex("#3730A3")
	t.AccentText = ui.Hex("#FFFFFF")
	t.Danger = ui.Hex("#DC2626")
	t.Warning = ui.Hex("#D97706")
	t.Success = ui.Hex("#16A34A")
	t.Selection = ui.RGBA(79, 70, 229, 0.16)
	t.Focus = ui.RGBA(79, 70, 229, 0.55)
	t.FocusRingWidth = 2
	t.FocusRingOffset = 1
	t.Scrollbar = ui.RGBA(0, 0, 0, 0.28)
	t.ScrollbarWidth = 10
	t.Radius = 8
	t.Spacing = 4
	t.FontSize = 13
	return t
}

// EditorTheme colors the code editor to match t.
func EditorTheme(t *ui.Theme) *editor.Theme {
	ct := editor.DefaultTheme(t)
	ct.Background = t.Background
	ct.Text = t.Text
	ct.LineNumber = t.TextMuted
	ct.ActiveLineNumber = t.Text
	ct.ActiveLine = t.Surface
	ct.Selection = t.Selection
	ct.InactiveSelection = t.Selection
	ct.Cursor = t.Accent
	ct.MatchingBracket = t.Accent
	ct.IndentGuide = t.Border
	ct.ActiveIndentGuide = t.Accent
	ct.WrapGuide = t.Border
	ct.Whitespace = t.TextMuted
	ct.Error = t.Danger
	ct.Warning = t.Warning
	ct.Info = t.Accent
	ct.Hint = t.TextMuted
	ct.Added = t.Success
	ct.Deleted = t.Danger
	ct.Modified = t.Warning

	if t.Dark {
		ct.Match = ui.RGBA(92, 75, 31, 1)
		ct.CurrentMatch = ui.RGBA(183, 121, 31, 1)
		ct.Tokens[editor.TokenKeyword] = editor.TokenStyle{Color: ui.Hex("#C4A7FF"), Bold: true}
		ct.Tokens[editor.TokenString] = editor.TokenStyle{Color: ui.Hex("#7EE2B8")}
		ct.Tokens[editor.TokenComment] = editor.TokenStyle{Color: ui.Hex("#6B7280"), Italic: true}
		ct.Tokens[editor.TokenNumber] = editor.TokenStyle{Color: ui.Hex("#F5C26B")}
		ct.Tokens[editor.TokenFunction] = editor.TokenStyle{Color: ui.Hex("#8AB4FF")}
		ct.Tokens[editor.TokenType] = editor.TokenStyle{Color: ui.Hex("#7DD3FC")}
		ct.Tokens[editor.TokenVariable] = editor.TokenStyle{Color: t.Text}
		ct.Tokens[editor.TokenConstant] = editor.TokenStyle{Color: ui.Hex("#F5C26B")}
		ct.Tokens[editor.TokenOperator] = editor.TokenStyle{Color: ui.Hex("#9AA4B2")}
		ct.Tokens[editor.TokenPunctuation] = editor.TokenStyle{Color: ui.Hex("#9AA4B2")}
		ct.Tokens[editor.TokenTag] = editor.TokenStyle{Color: ui.Hex("#C4A7FF")}
		ct.Tokens[editor.TokenAttribute] = editor.TokenStyle{Color: ui.Hex("#8AB4FF")}
		ct.Tokens[editor.TokenMeta] = editor.TokenStyle{Color: ui.Hex("#6B7280"), Italic: true}
		ct.Tokens[editor.TokenHeading] = editor.TokenStyle{Color: t.Text, Bold: true}
		ct.Tokens[editor.TokenEmphasis] = editor.TokenStyle{Color: t.Text, Italic: true}
		ct.Tokens[editor.TokenStrong] = editor.TokenStyle{Color: t.Text, Bold: true}
		ct.Tokens[editor.TokenLink] = editor.TokenStyle{Color: ui.Hex("#8AB4FF")}
		ct.Tokens[editor.TokenCode] = editor.TokenStyle{Color: ui.Hex("#7EE2B8")}
		ct.Tokens[editor.TokenInserted] = editor.TokenStyle{Color: ui.Hex("#4ADE80")}
		ct.Tokens[editor.TokenDeleted] = editor.TokenStyle{Color: ui.Hex("#F87171")}
		ct.Tokens[editor.TokenInvalid] = editor.TokenStyle{Color: ui.Hex("#F87171")}
		return ct
	}

	ct.Match = ui.RGBA(253, 230, 138, 1)
	ct.CurrentMatch = ui.RGBA(245, 158, 11, 1)
	ct.Tokens[editor.TokenKeyword] = editor.TokenStyle{Color: ui.Hex("#7C3AED"), Bold: true}
	ct.Tokens[editor.TokenString] = editor.TokenStyle{Color: ui.Hex("#047857")}
	ct.Tokens[editor.TokenComment] = editor.TokenStyle{Color: ui.Hex("#9CA3AF"), Italic: true}
	ct.Tokens[editor.TokenNumber] = editor.TokenStyle{Color: ui.Hex("#B45309")}
	ct.Tokens[editor.TokenFunction] = editor.TokenStyle{Color: ui.Hex("#2563EB")}
	ct.Tokens[editor.TokenType] = editor.TokenStyle{Color: ui.Hex("#0E7490")}
	ct.Tokens[editor.TokenVariable] = editor.TokenStyle{Color: t.Text}
	ct.Tokens[editor.TokenConstant] = editor.TokenStyle{Color: ui.Hex("#B45309")}
	ct.Tokens[editor.TokenOperator] = editor.TokenStyle{Color: ui.Hex("#6B7280")}
	ct.Tokens[editor.TokenPunctuation] = editor.TokenStyle{Color: ui.Hex("#6B7280")}
	ct.Tokens[editor.TokenTag] = editor.TokenStyle{Color: ui.Hex("#7C3AED")}
	ct.Tokens[editor.TokenAttribute] = editor.TokenStyle{Color: ui.Hex("#2563EB")}
	ct.Tokens[editor.TokenMeta] = editor.TokenStyle{Color: ui.Hex("#9CA3AF"), Italic: true}
	ct.Tokens[editor.TokenHeading] = editor.TokenStyle{Color: t.Text, Bold: true}
	ct.Tokens[editor.TokenEmphasis] = editor.TokenStyle{Color: t.Text, Italic: true}
	ct.Tokens[editor.TokenStrong] = editor.TokenStyle{Color: t.Text, Bold: true}
	ct.Tokens[editor.TokenLink] = editor.TokenStyle{Color: ui.Hex("#2563EB")}
	ct.Tokens[editor.TokenCode] = editor.TokenStyle{Color: ui.Hex("#047857")}
	ct.Tokens[editor.TokenInserted] = editor.TokenStyle{Color: ui.Hex("#16A34A")}
	ct.Tokens[editor.TokenDeleted] = editor.TokenStyle{Color: ui.Hex("#DC2626")}
	ct.Tokens[editor.TokenInvalid] = editor.TokenStyle{Color: ui.Hex("#DC2626")}
	return ct
}

func darkTheme() *ui.Theme {
	t := ui.DarkTheme()
	t.Background = ui.Hex("#0F1115")
	t.Surface = ui.Hex("#151922")
	t.SurfaceHover = ui.Hex("#1F2531")
	t.SurfacePressed = ui.Hex("#252C3A")
	t.Border = ui.Hex("#262C38")
	t.Text = ui.Hex("#E6E9EF")
	t.TextMuted = ui.Hex("#9AA4B2")
	t.Accent = ui.Hex("#7C8CFF")
	t.AccentHover = ui.Hex("#93A1FF")
	t.AccentPressed = ui.Hex("#6B7AEB")
	t.AccentText = ui.Hex("#0F1115")
	t.Danger = ui.Hex("#F87171")
	t.Warning = ui.Hex("#FBBF24")
	t.Success = ui.Hex("#4ADE80")
	t.Selection = ui.Hex("#2A3350")
	t.Focus = ui.RGBA(124, 140, 255, 0.60)
	t.FocusRingWidth = 2
	t.FocusRingOffset = 1
	t.Scrollbar = ui.RGBA(255, 255, 255, 0.30)
	t.ScrollbarWidth = 10
	t.Radius = 8
	t.Spacing = 4
	t.FontSize = 13
	return t
}
