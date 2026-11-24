package app

import (
	"unicode"
	"unicode/utf8"

	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/caseconv"
)

func applyCase(tab *Tab, mode caseconv.Mode) {
	buf := tab.doc.Buffer
	sels := tab.state.Selections()
	if len(sels) == 0 {
		return
	}

	var changes []selectionChange
	for _, s := range sels {
		start, end := s.Start(), s.End()
		// An empty selection converts the word at the caret, as editors
		// do: the caret alone should not make the command do nothing.
		if start == end {
			start, end = wordRange(buf, s.Head)
		}
		if start >= end {
			continue
		}
		text := buf.Slice(start, end)
		converted := caseconv.Apply(text, mode)
		if converted == text {
			continue
		}
		changes = append(changes, selectionChange{
			start:   start,
			end:     end,
			text:    converted,
			forward: s.Anchor <= s.Head,
		})
	}
	writeSelectionChanges(tab, changes)
}

func wordRange(b *editor.Buffer, offset int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > b.Len() {
		offset = b.Len()
	}
	p := b.Pos(offset)
	line := b.Line(p.Line)
	if p.Col > len(line) {
		p.Col = len(line)
	}

	start, end := p.Col, p.Col
	for start > 0 {
		r, size := utf8.DecodeLastRuneInString(line[:start])
		if !isWordRune(r) {
			break
		}
		start -= size
	}
	for end < len(line) {
		r, size := utf8.DecodeRuneInString(line[end:])
		if !isWordRune(r) {
			break
		}
		end += size
	}
	base := b.Offset(editor.Pos{Line: p.Line, Col: 0})
	return base + start, base + end
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
