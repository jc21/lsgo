package output

import "strings"

// gridFilling is the number of spaces left between adjacent grid columns.
const gridFilling = 2

// fitGrid finds the largest number of columns that a set of cells (given
// only by their display widths) can be arranged into without any row
// exceeding consoleWidth, and the width each of those columns needs to be
// padded to.
//
// Column widths are computed for candidate column counts from most to
// fewest, so the first configuration that fits is also the widest one
// that does -- mirroring the classic multi-column `ls` layout algorithm.
// If even a single column doesn't fit (i.e. the widest cell alone exceeds
// consoleWidth), ok is false and the caller should fall back to printing
// one file per line.
func fitGrid(widths []int, consoleWidth int, across bool) (numColumns int, colWidths []int, ok bool) {
	n := len(widths)
	if n == 0 {
		return 0, nil, true
	}

	for cols := n; cols >= 1; cols-- {
		rows := (n + cols - 1) / cols
		candidate := make([]int, cols)

		for i, w := range widths {
			col := columnOf(i, rows, cols, across)
			if w > candidate[col] {
				candidate[col] = w
			}
		}

		total := (cols - 1) * gridFilling
		for _, w := range candidate {
			total += w
		}

		if total <= consoleWidth || cols == 1 {
			return cols, candidate, total <= consoleWidth
		}
	}

	// Unreachable: the cols==1 branch above always returns.
	return 1, nil, false
}

// columnOf returns which column item i falls into for a grid with the
// given row count, either filled top-to-bottom column by column (the
// default) or left-to-right row by row (--across).
func columnOf(i, rows, cols int, across bool) int {
	if across {
		return i % cols
	}
	return i / rows
}

// RenderGrid lays cells out into a multi-column grid that fits within
// consoleWidth, falling back to one cell per line if even a single column
// would be too wide.
func RenderGrid(cells []Cell, consoleWidth int, across bool) string {
	if len(cells) == 0 {
		return ""
	}

	widths := make([]int, len(cells))
	for i, c := range cells {
		widths[i] = c.Width
	}

	cols, colWidths, ok := fitGrid(widths, consoleWidth, across)
	if !ok {
		return RenderLines(cells)
	}
	if cols <= 1 {
		return RenderLines(cells)
	}

	n := len(cells)
	rows := (n + cols - 1) / cols

	var b strings.Builder
	for r := 0; r < rows; r++ {
		type placed struct {
			idx, col int
		}
		var line []placed

		for c := 0; c < cols; c++ {
			var idx int
			if across {
				idx = r*cols + c
			} else {
				idx = c*rows + r
			}
			if idx >= n {
				if across {
					break // end of the last, short row
				}
				continue // TopToBottom: this column has no entry in this row
			}
			line = append(line, placed{idx, c})
		}

		for i, p := range line {
			cell := cells[p.idx]
			b.WriteString(cell.String())
			if i < len(line)-1 {
				b.WriteString(strings.Repeat(" ", colWidths[p.col]-cell.Width))
				b.WriteString(strings.Repeat(" ", gridFilling))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderLines prints one cell per line, with no column alignment. Used
// both by the dedicated one-per-line (-1) view, and as the grid view's
// fallback when nothing fits in columns.
func RenderLines(cells []Cell) string {
	var b strings.Builder
	for _, c := range cells {
		b.WriteString(c.String())
		b.WriteByte('\n')
	}
	return b.String()
}
