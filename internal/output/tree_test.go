package output

import (
	"reflect"
	"testing"
)

func params(depth int, last bool) TreeParams { return TreeParams{Depth: depth, Last: last} }

func TestTreeTrunkEmptyAtFirst(t *testing.T) {
	var tt TreeTrunk
	got := tt.NewRow(params(0, true))
	if len(got) != 0 {
		t.Errorf("expected no tree parts at root, got %v", got)
	}
}

func TestTreeTrunkOneChild(t *testing.T) {
	var tt TreeTrunk
	tt.NewRow(params(0, true))
	got := tt.NewRow(params(1, true))
	want := []TreePart{treeCorner}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTreeTrunkTwoChildren(t *testing.T) {
	var tt TreeTrunk
	tt.NewRow(params(0, true))
	got1 := tt.NewRow(params(1, false))
	if !reflect.DeepEqual(got1, []TreePart{treeEdge}) {
		t.Errorf("first child: got %v", got1)
	}
	got2 := tt.NewRow(params(1, true))
	if !reflect.DeepEqual(got2, []TreePart{treeCorner}) {
		t.Errorf("second child: got %v", got2)
	}
}

func TestTreeTrunkTwoTimesTwoChildren(t *testing.T) {
	var tt TreeTrunk
	tt.NewRow(params(0, false))
	check(t, tt.NewRow(params(1, false)), treeEdge)
	check(t, tt.NewRow(params(1, true)), treeCorner)

	tt.NewRow(params(0, true))
	check(t, tt.NewRow(params(1, false)), treeEdge)
	check(t, tt.NewRow(params(1, true)), treeCorner)
}

func TestTreeTrunkNestedChildren(t *testing.T) {
	var tt TreeTrunk
	tt.NewRow(params(0, true))

	check(t, tt.NewRow(params(1, false)), treeEdge)
	check(t, tt.NewRow(params(2, false)), treeLine, treeEdge)
	check(t, tt.NewRow(params(2, true)), treeLine, treeCorner)

	check(t, tt.NewRow(params(1, true)), treeCorner)
	check(t, tt.NewRow(params(2, false)), treeBlank, treeEdge)
	check(t, tt.NewRow(params(2, true)), treeBlank, treeCorner)
}

func check(t *testing.T, got []TreePart, want ...TreePart) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTreePartAsciiArt(t *testing.T) {
	cases := map[TreePart]string{
		treeEdge: "├──", treeLine: "│  ", treeCorner: "└──", treeBlank: "   ",
	}
	for p, want := range cases {
		if got := p.asciiArt(); got != want {
			t.Errorf("%v.asciiArt() = %q, want %q", p, got, want)
		}
	}
}
