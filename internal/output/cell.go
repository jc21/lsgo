// Package output renders a list of files into one of lsgo's views: a
// flat grid, one-per-line, a detailed table, or a tree.
package output

import (
	"strings"

	"github.com/jc21/lsgo/internal/style"
	"github.com/jc21/lsgo/internal/textwidth"
)

// Cell accumulates styled text fragments alongside their combined display
// width, so that table columns can be padded to line up regardless of how
// much ANSI styling each cell's text carries.
type Cell struct {
	Width int
	buf   strings.Builder
}

// NewCell creates an empty cell.
func NewCell() Cell { return Cell{} }

// Text paints text with the given style and appends it to the cell,
// tracking its contribution to the cell's display width.
func (c *Cell) Text(s style.Style, text string) {
	c.Width += textwidth.Width(text)
	c.buf.WriteString(s.Paint(text))
}

// Plain appends unstyled text.
func (c *Cell) Plain(text string) {
	c.Text(style.Style{}, text)
}

// Spaces appends n literal space characters.
func (c *Cell) Spaces(n int) {
	if n <= 0 {
		return
	}
	c.Width += n
	c.buf.WriteString(strings.Repeat(" ", n))
}

// Append concatenates another cell's contents onto this one.
func (c *Cell) Append(other Cell) {
	c.Width += other.Width
	c.buf.WriteString(other.String())
}

// String returns the fully-styled contents of the cell.
func (c Cell) String() string { return c.buf.String() }

// PadTo returns the number of trailing spaces needed to bring this cell up
// to the given width. Never negative.
func (c Cell) PadTo(width int) int {
	if width <= c.Width {
		return 0
	}
	return width - c.Width
}
