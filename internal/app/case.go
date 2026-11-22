package app

import (
	"sort"
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

	type change struct {
		start, end int
		text       string
		forward    bool
	}
	var changes []change
	for _, s := range sels {
		start, end := s.Start(), s.End()
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
		changes = append(changes, change{
			start:   start,
			end:     end,
			text:    converted,
			forward: s.Anchor <= s.Head,
		})
	}
	if len(changes) == 0 {
		return
	}

	// Apply from the end so earlier offsets stay valid.
	edits := append([]change(nil), changes...)
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	buf.Transact(func() {
		for _, c := range edits {
			buf.Edit(editor.Edit{Start: c.start, End: c.end, Text: c.text})
		}
	})

	// Keep each changed selection pointing at its converted text. An
	// earlier change may have changed the byte length, so shift the
	// selections after it.
	sorted := append([]change(nil), changes...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start < sorted[j].start })
	shifts := make(map[int]int, len(sorted))
	shift := 0
	for _, c := range sorted {
		shifts[c.start] = shift
		shift += len(c.text) - (c.end - c.start)
	}
	sels2 := make([]editor.Selection, 0, len(changes))
	for _, c := range changes {
		start := c.start + shifts[c.start]
		n := start + len(c.text)
		if c.forward {
			sels2 = append(sels2, editor.Selection{Anchor: start, Head: n})
		} else {
			sels2 = append(sels2, editor.Selection{Anchor: n, Head: start})
		}
	}
	tab.state.SetSelections(sels2...)
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
