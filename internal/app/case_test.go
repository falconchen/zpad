package app

import (
	"testing"

	"github.com/egoist/mygo/ui/editor"
	"github.com/helloxz/zpad/internal/caseconv"
	"github.com/helloxz/zpad/internal/encoding"
)

func testTab(text string) *Tab {
	doc := &Document{ID: "test", Buffer: editor.NewBuffer(text), Encoding: encoding.UTF8}
	return &Tab{doc: doc, state: &editor.State{Buffer: doc.Buffer}}
}

func TestApplyCaseSelection(t *testing.T) {
	tab := testTab("hello world")
	tab.state.SetSelections(editor.Selection{Anchor: 0, Head: 5})
	applyCase(tab, caseconv.Upper)
	if got := tab.doc.Buffer.String(); got != "HELLO world" {
		t.Fatalf("text %q", got)
	}
	sel := tab.state.Selections()
	if len(sel) != 1 || tab.doc.Buffer.Slice(sel[0].Start(), sel[0].End()) != "HELLO" {
		t.Fatalf("selection %+v", sel)
	}
}

func TestApplyCaseMultipleSelections(t *testing.T) {
	tab := testTab("foo bar")
	tab.state.SetSelections(
		editor.Selection{Anchor: 0, Head: 3},
		editor.Selection{Anchor: 4, Head: 7},
	)
	applyCase(tab, caseconv.Upper)
	if got := tab.doc.Buffer.String(); got != "FOO BAR" {
		t.Fatalf("text %q", got)
	}
	if len(tab.state.Selections()) != 2 {
		t.Fatalf("selections %+v", tab.state.Selections())
	}
}

func TestApplyCaseWordAtCaret(t *testing.T) {
	tab := testTab("hello world")
	tab.state.SetSelections(editor.Selection{Anchor: 1, Head: 1})
	applyCase(tab, caseconv.Upper)
	if got := tab.doc.Buffer.String(); got != "HELLO world" {
		t.Fatalf("text %q", got)
	}
}

func TestApplyCaseChineseUnchanged(t *testing.T) {
	tab := testTab("中文 English")
	tab.state.SetSelections(editor.Selection{Anchor: 0, Head: len("中文 English")})
	applyCase(tab, caseconv.Upper)
	if got := tab.doc.Buffer.String(); got != "中文 ENGLISH" {
		t.Fatalf("text %q", got)
	}
}
