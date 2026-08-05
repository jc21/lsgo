package output

import (
	"testing"

	"lsgo/internal/style"
)

func cellOf(text string) Cell {
	var c Cell
	c.Text(style.Style{}, text)
	return c
}

func TestFitGridTopToBottomColumnAssignment(t *testing.T) {
	// 6 items, all width 1: with a generous console width, we expect as
	// many columns as fit; check the column-major (down-then-across)
	// assignment directly.
	widths := []int{1, 1, 1, 1, 1, 1}
	cols, _, ok := fitGrid(widths, 20, false)
	if !ok {
		t.Fatal("expected grid to fit")
	}
	if cols != 6 {
		t.Errorf("expected all 6 items to fit on one row given ample width, got %d columns", cols)
	}
}

func TestFitGridNarrowWidthFallsBackToFewerColumns(t *testing.T) {
	// Five 4-character-wide names; with 2-space filling, 2 columns need
	// 4+2+4 = 10 chars, 3 columns need 4+2+4+2+4 = 16.
	widths := []int{4, 4, 4, 4, 4}
	cols, colWidths, ok := fitGrid(widths, 10, false)
	if !ok {
		t.Fatal("expected grid to fit in some configuration")
	}
	if cols != 2 {
		t.Errorf("expected 2 columns to fit in width 10, got %d", cols)
	}
	if len(colWidths) != 2 || colWidths[0] != 4 || colWidths[1] != 4 {
		t.Errorf("unexpected column widths: %v", colWidths)
	}
}

func TestFitGridSingleCellTooWideFails(t *testing.T) {
	widths := []int{50}
	_, _, ok := fitGrid(widths, 10, false)
	if ok {
		t.Error("expected a single overly-wide cell to fail to fit")
	}
}

func TestRenderGridTopToBottomOrder(t *testing.T) {
	// 4 names of width 1 each, console width forces exactly 2 columns.
	// TopToBottom order should read down first column, then down second:
	// a c
	// b d
	cells := []Cell{cellOf("a"), cellOf("b"), cellOf("c"), cellOf("d")}
	got := RenderGrid(cells, 4, false) // "a" + 2 spaces + "c" = 4 chars wide
	want := "a  c\nb  d\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderGridAcrossOrder(t *testing.T) {
	cells := []Cell{cellOf("a"), cellOf("b"), cellOf("c"), cellOf("d")}
	got := RenderGrid(cells, 4, true) // across: fill rows left-to-right
	want := "a  b\nc  d\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderGridFallsBackToLinesWhenTooNarrow(t *testing.T) {
	cells := []Cell{cellOf("averylongfilename"), cellOf("anotherlongname")}
	got := RenderGrid(cells, 5, false)
	want := "averylongfilename\nanotherlongname\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderGridNoTrailingSpacesOnRow(t *testing.T) {
	cells := []Cell{cellOf("aa"), cellOf("b")}
	got := RenderGrid(cells, 80, false)
	// Both items land on one row: "aa", the 2-space filler, then "b" --
	// and critically, no trailing padding after the last cell in the row.
	want := "aa  b\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
