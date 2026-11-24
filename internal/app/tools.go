package app

import (
	"errors"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/texttools"
	"github.com/helloxz/zpad/locales"
)

// textTool is one transformation of the selected text.
type textTool string

const (
	toolBase64Encode textTool = "base64-encode"
	toolBase64Decode textTool = "base64-decode"
	toolMD5Lower32   textTool = "md5-32-lower"
	toolMD5Upper32   textTool = "md5-32-upper"
	toolMD5Lower16   textTool = "md5-16-lower"
	toolMD5Upper16   textTool = "md5-16-upper"
	toolJSONFormat   textTool = "json-format"
	toolJSONMinify   textTool = "json-minify"
)

// maxTextToolBytes is the most text one tool may process, across all
// selections. A tool copies what it converts, and formatting JSON can write
// several times the input, so a whole-file selection would spike the heap.
const maxTextToolBytes = texttools.MaxInput

// runTextTool transforms the selections of tab with tool. With no
// selection, or one over the limit, it says so and changes nothing.
func (a *App) runTextTool(tab *Tab, tool textTool) {
	if tab == nil || tab.doc.ReadOnly || tab.preview {
		return
	}
	n := selectionBytes(tab)
	if n == 0 {
		a.SetErrorToast(locales.TextToolNoSelection(), "")
		return
	}
	if n > maxTextToolBytes {
		a.SetErrorToast(locales.TextToolTooLarge(), formatSize(int64(n))+" / "+formatSize(maxTextToolBytes))
		return
	}

	count, err := transformSelections(tab, func(text string, start int) (string, error) {
		return applyTextTool(tab, tool, text, start)
	})
	if err != nil {
		a.SetErrorToast(textToolErrorTitle(err), err.Error())
		return
	}
	if count == 0 {
		return
	}
	a.updateTitle()
	a.update()
}

// selectionBytes counts the selected UTF-8 bytes of a tab.
func selectionBytes(tab *Tab) int {
	total := 0
	for _, s := range tab.state.Selections() {
		total += s.End() - s.Start()
	}
	return total
}

func applyTextTool(tab *Tab, tool textTool, text string, start int) (string, error) {
	switch tool {
	case toolBase64Encode:
		return texttools.EncodeBase64(text), nil
	case toolBase64Decode:
		return texttools.DecodeBase64(text)
	case toolMD5Lower32:
		return texttools.MD5(text, texttools.MD5Hex32Lower), nil
	case toolMD5Upper32:
		return texttools.MD5(text, texttools.MD5Hex32Upper), nil
	case toolMD5Lower16:
		return texttools.MD5(text, texttools.MD5Hex16Lower), nil
	case toolMD5Upper16:
		return texttools.MD5(text, texttools.MD5Hex16Upper), nil
	case toolJSONFormat:
		out, err := texttools.FormatJSON(text)
		if err != nil {
			return "", err
		}
		return alignToLine(tab.doc.Buffer, start, out), nil
	case toolJSONMinify:
		return texttools.MinifyJSON(text)
	}
	return text, nil
}

// alignToLine indents the lines of a formatted value after the first to
// the indentation the selection starts at, so formatting a value inside a
// larger document does not send it back to column zero.
func alignToLine(buf *editor.Buffer, start int, text string) string {
	if !strings.Contains(text, "\n") {
		return text
	}
	prefix := lineIndent(buf, start)
	if prefix == "" {
		return text
	}
	return strings.ReplaceAll(text, "\n", "\n"+prefix)
}

// lineIndent returns the whitespace before offset on its line, empty when
// the text before it is not all whitespace.
func lineIndent(buf *editor.Buffer, offset int) string {
	p := buf.Pos(offset)
	line := buf.Line(p.Line)
	if p.Col > len(line) {
		p.Col = len(line)
	}
	prefix := line[:p.Col]
	for i := 0; i < len(prefix); i++ {
		if prefix[i] != ' ' && prefix[i] != '\t' {
			return ""
		}
	}
	return prefix
}

func textToolErrorTitle(err error) string {
	switch {
	case errors.Is(err, texttools.ErrNotText):
		return locales.TextToolNotText()
	case errors.Is(err, texttools.ErrEmpty):
		return locales.TextToolEmpty()
	case errors.Is(err, texttools.ErrJSONDepth):
		return locales.TextToolJSONDeep()
	case errors.Is(err, texttools.ErrJSONOutput):
		return locales.TextToolJSONLarge()
	case errors.Is(err, texttools.ErrJSON):
		return locales.TextToolJSONInvalid()
	}
	return locales.TextToolFailed()
}

// toolMenu fills m with the text tools, disabled where tab cannot run
// them. Case conversion is not here: the editor's own context menu has a
// richer Transform submenu, and the app menu keeps Convert Case.
func (a *App) toolMenu(m *ui.Menu, tab *Tab) {
	editable := tab != nil && !tab.doc.ReadOnly && !tab.preview
	selected := 0
	if editable {
		selected = selectionBytes(tab)
	}
	tooLarge := selected > maxTextToolBytes
	hasText := selected > 0

	m.Submenu(locales.Base64Group(), func(mm *ui.Menu) {
		if mm.Item(locales.Base64Encode()).Disabled(!hasText || tooLarge).Chosen() {
			a.runTextTool(tab, toolBase64Encode)
		}
		if mm.Item(locales.Base64Decode()).Disabled(!hasText || tooLarge).Chosen() {
			a.runTextTool(tab, toolBase64Decode)
		}
	}).Disabled(!editable)

	m.Submenu(locales.Md5Group(), func(mm *ui.Menu) {
		if mm.Item(locales.Md5Lower32()).Disabled(!hasText || tooLarge).Chosen() {
			a.runTextTool(tab, toolMD5Lower32)
		}
		if mm.Item(locales.Md5Upper32()).Disabled(!hasText || tooLarge).Chosen() {
			a.runTextTool(tab, toolMD5Upper32)
		}
		if mm.Item(locales.Md5Lower16()).Disabled(!hasText || tooLarge).Chosen() {
			a.runTextTool(tab, toolMD5Lower16)
		}
		if mm.Item(locales.Md5Upper16()).Disabled(!hasText || tooLarge).Chosen() {
			a.runTextTool(tab, toolMD5Upper16)
		}
	}).Disabled(!editable)

	m.Submenu(locales.JsonGroup(), func(mm *ui.Menu) {
		if mm.Item(locales.JsonFormat()).Disabled(!hasText || tooLarge).Chosen() {
			a.runTextTool(tab, toolJSONFormat)
		}
		if mm.Item(locales.JsonMinify()).Disabled(!hasText || tooLarge).Chosen() {
			a.runTextTool(tab, toolJSONMinify)
		}
	}).Disabled(!editable)
}
