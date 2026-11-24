package app

import (
	"sort"

	"github.com/egoist/mygo/ui/editor"
)

// selectionChange replaces the text start to end with text. forward keeps
// the direction of the selection it came from.
type selectionChange struct {
	start, end int
	text       string
	forward    bool
}

// transformSelections rewrites the text of each non-empty selection with fn
// and keeps the selections over the result, in one step of undo. When fn
// fails for any selection, nothing changes and the error is returned.
func transformSelections(tab *Tab, fn func(text string, start int) (string, error)) (int, error) {
	buf := tab.doc.Buffer
	sels := tab.state.Selections()
	if len(sels) == 0 {
		return 0, nil
	}

	changes := make([]selectionChange, 0, len(sels))
	for _, s := range sels {
		start, end := s.Start(), s.End()
		if start >= end {
			continue
		}
		text := buf.Slice(start, end)
		out, err := fn(text, start)
		if err != nil {
			return 0, err
		}
		if out == text {
			continue
		}
		changes = append(changes, selectionChange{
			start:   start,
			end:     end,
			text:    out,
			forward: s.Anchor <= s.Head,
		})
	}
	writeSelectionChanges(tab, changes)
	return len(changes), nil
}

// writeSelectionChanges applies changes from the end so earlier offsets
// stay valid, then keeps each selection pointing at its new text.
func writeSelectionChanges(tab *Tab, changes []selectionChange) {
	if len(changes) == 0 {
		return
	}
	buf := tab.doc.Buffer

	edits := append([]selectionChange(nil), changes...)
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	buf.Transact(func() {
		for _, c := range edits {
			buf.Edit(editor.Edit{Start: c.start, End: c.end, Text: c.text})
		}
	})

	// An earlier change may have changed the byte length, so selections
	// after it move by the same amount.
	sorted := append([]selectionChange(nil), changes...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start < sorted[j].start })
	shifts := make(map[int]int, len(sorted))
	shift := 0
	for _, c := range sorted {
		shifts[c.start] = shift
		shift += len(c.text) - (c.end - c.start)
	}
	sels := make([]editor.Selection, 0, len(changes))
	for _, c := range changes {
		start := c.start + shifts[c.start]
		n := start + len(c.text)
		if c.forward {
			sels = append(sels, editor.Selection{Anchor: start, Head: n})
		} else {
			sels = append(sels, editor.Selection{Anchor: n, Head: start})
		}
	}
	tab.state.SetSelections(sels...)
}
