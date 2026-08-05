package output

import (
	"lsgo/internal/fsx"
	"lsgo/internal/theme"
)

// RenderOneLine renders files one per line (the -1/--oneline view), each
// with its link target shown if it's a symlink.
func RenderOneLine(th *theme.Theme, opts FileNameOptions, files []*fsx.File) string {
	cells := make([]Cell, len(files))
	for i, f := range files {
		cells[i] = RenderFileName(th, opts, f, true)
	}
	return RenderLines(cells)
}
