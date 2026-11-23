package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/settings"
)

func sameTabNames(got []string, want ...string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestReorderTabWithinPane(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc, bDoc, cDoc := docAt("/tmp/a.txt", "a"), docAt("/tmp/b.txt", "b"), docAt("/tmp/c.txt", "c")
	a.addDocument(aDoc)
	a.addDocument(bDoc)
	a.addDocument(cDoc)
	a.left.active = 0

	a.moveTabAt(0, 0, 0, 3)
	if got := tabNames(a.left.tabs); !sameTabNames(got, "b.txt", "c.txt", "a.txt") {
		t.Fatalf("tabs %q", got)
	}
	if a.left.active != 2 || a.left.tabs[2].doc != aDoc {
		t.Fatalf("active %d, want the dragged tab", a.left.active)
	}

	a.moveTabAt(0, 2, 0, 0)
	if got := tabNames(a.left.tabs); !sameTabNames(got, "a.txt", "b.txt", "c.txt") {
		t.Fatalf("tabs %q", got)
	}
	if a.left.active != 0 {
		t.Fatalf("active %d, want the dragged tab", a.left.active)
	}
}

func TestReorderTabKeepsChosenTab(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	bDoc := docAt("/tmp/b.txt", "b")
	a.addDocument(bDoc)
	a.addDocument(docAt("/tmp/c.txt", "c"))
	a.left.active = 1

	a.moveTabAt(0, 0, 0, 3)
	if got := tabNames(a.left.tabs); !sameTabNames(got, "b.txt", "c.txt", "a.txt") {
		t.Fatalf("tabs %q", got)
	}
	if a.left.active != 0 || a.left.activeTab().doc != bDoc {
		t.Fatalf("active %d, want b.txt still chosen", a.left.active)
	}
}

func TestReorderTabOntoItselfDoesNothing(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.addDocument(docAt("/tmp/b.txt", "b"))
	a.left.active = 1

	a.moveTabAt(0, 0, 0, 0)
	a.moveTabAt(0, 1, 0, 1)
	a.moveTabAt(0, 1, 0, 2)
	if got := tabNames(a.left.tabs); !sameTabNames(got, "a.txt", "b.txt") {
		t.Fatalf("tabs %q", got)
	}
	if a.left.tabs[1].doc != a.left.activeTab().doc {
		t.Fatal("the chosen tab moved")
	}
}

func TestDropOnOwnPaneWithoutPlaceDoesNothing(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc := docAt("/tmp/a.txt", "a")
	a.addDocument(aDoc)
	a.addDocument(docAt("/tmp/b.txt", "b"))
	a.left.active = 1

	a.dropOnPane(paneDrag{kind: dragTab, pane: 0, id: aDoc.ID}, 0)
	if got := tabNames(a.left.tabs); !sameTabNames(got, "a.txt", "b.txt") {
		t.Fatalf("tabs %q", got)
	}
}

func TestDragTabReordersWithinPane(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	a.addDocument(docAt("/tmp/a.txt", "a"))
	a.addDocument(docAt("/tmp/b.txt", "b"))
	a.addDocument(docAt("/tmp/c.txt", "c"))

	tt := ui.NewTester(a.View, 900, 600)
	tt.Frame()
	src, ok := tt.Find("a.txt")
	if !ok {
		t.Fatalf("no a.txt tab; texts %q", tt.Texts())
	}
	last, ok := tt.Find("c.txt")
	if !ok {
		t.Fatalf("no c.txt tab; texts %q", tt.Texts())
	}
	y := src.Y + src.H/2
	tt.Press(src.X+src.W/2, y)
	tt.Move(src.X+src.W/2+20, y)
	tt.Move(last.X+last.W+20, y)
	tt.Release(last.X+last.W+20, y)
	tt.Frame()

	if got := tabNames(a.left.tabs); !sameTabNames(got, "b.txt", "c.txt", "a.txt") {
		t.Fatalf("tabs %q", got)
	}
}

func TestDragTabAcrossPanesInsertsWhereDropped(t *testing.T) {
	a := New(settings.Default(), t.TempDir())
	aDoc, bDoc, cDoc := docAt("/tmp/a.txt", "a"), docAt("/tmp/b.txt", "b"), docAt("/tmp/c.txt", "c")
	a.addDocument(aDoc)
	a.addDocument(cDoc)
	a.left.active = 0
	a.ensureSplit()
	a.addDocumentAt(1, bDoc)
	if got := tabNames(a.right.tabs); !sameTabNames(got, "b.txt") {
		t.Fatalf("right tabs %q", got)
	}

	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()
	src, ok := tt.Find("a.txt")
	if !ok {
		t.Fatalf("no a.txt tab; texts %q", tt.Texts())
	}
	target, ok := tt.Find("b.txt")
	if !ok {
		t.Fatalf("no b.txt tab; texts %q", tt.Texts())
	}
	y := target.Y + target.H/2
	// Inside b.txt, left of its middle and past the divider's grip.
	x := target.X + 12
	tt.Press(src.X+src.W/2, src.Y+src.H/2)
	tt.Move(src.X+src.W/2+20, src.Y+src.H/2)
	tt.Move(x, y)
	tt.Release(x, y)
	tt.Frame()

	if got := tabNames(a.right.tabs); !sameTabNames(got, "a.txt", "b.txt") {
		t.Fatalf("right tabs %q", got)
	}
	if got := tabNames(a.left.tabs); !sameTabNames(got, "c.txt") {
		t.Fatalf("left tabs %q", got)
	}
	if a.activePane != 1 || a.right.active != 0 {
		t.Fatalf("active pane %d, tab %d; want the dropped tab", a.activePane, a.right.active)
	}
}

func TestTabDropAt(t *testing.T) {
	rects := []ui.Rect{{X: 0, Y: 0, W: 100, H: 34}, {X: 100, W: 50, H: 34}}
	for _, tc := range []struct {
		x    float32
		want int
		at   float32
	}{
		{x: 10, want: 0, at: 0},
		{x: 49, want: 0, at: 0},
		{x: 51, want: 1, at: 100},
		{x: 120, want: 1, at: 100},
		{x: 200, want: 2, at: 150},
	} {
		got, at := tabDropAt(rects, tc.x)
		if got != tc.want || at != tc.at {
			t.Errorf("tabDropAt(%v) = %d at %v, want %d at %v", tc.x, got, at, tc.want, tc.at)
		}
	}
}
