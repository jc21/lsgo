package output

import (
	"fmt"
	"io"

	"github.com/jc21/lsgo/internal/fsx"
	"github.com/jc21/lsgo/internal/theme"
	"github.com/jc21/lsgo/internal/xattr"
)

// RecurseOptions controls how far (and whether, as a tree) the details
// view recurses into subdirectories.
type RecurseOptions struct {
	// Tree draws directory contents nested under their parent with tree
	// branch characters, rather than as separate directory sections.
	Tree bool
	// MaxDepth limits recursion; nil means unlimited.
	MaxDepth *int
}

// IsTooDeep reports whether depth (1-based, the root directory's children
// are depth 1) has passed the configured limit.
func (r RecurseOptions) IsTooDeep(depth int) bool {
	return r.MaxDepth != nil && depth > *r.MaxDepth
}

// DetailsOptions configures the long/tree view.
type DetailsOptions struct {
	// Table is nil for a plain --tree listing (no -l); otherwise it
	// drives the metadata columns shown before each name.
	Table      *TableOptions
	Header     bool
	ShowXattrs bool
}

// DetailsRenderer renders the long (-l) and/or tree (-T) views, which
// share a table/tree-trunk-based layout engine.
type DetailsRenderer struct {
	Theme           *theme.Theme
	FileNameOptions FileNameOptions
	Options         DetailsOptions
	Recurse         *RecurseOptions
	Filter          fsx.FileFilter
	Git             *fsx.GitCache
	GitIgnoring     bool
}

type detailsRow struct {
	tree  TreeParams
	cells []Cell // nil for rows with no table (xattr/error rows, or tree-only mode)
	name  Cell
}

// Render writes the details view for files (which the caller has already
// filtered and sorted) to w.
func (r *DetailsRenderer) Render(w io.Writer, files []*fsx.File) error {
	var table *Table
	var rows []detailsRow

	if r.Options.Table != nil {
		table = NewTable(r.Theme, r.Git, *r.Options.Table)

		if r.Options.Header {
			rows = append(rows, detailsRow{
				tree:  TreeParams{},
				cells: table.HeaderRow(),
				name:  headerNameCell(r.Theme),
			})
		}
	}

	r.addFiles(table, &rows, files, 1)

	if table != nil {
		for _, row := range rows {
			table.AddWidths(row.cells)
		}
	}

	return r.writeRows(w, table, rows)
}

func headerNameCell(th *theme.Theme) Cell {
	var c Cell
	c.Text(th.UI.Header, "Name")
	return c
}

// addFiles appends one row per file (recursing into subdirectories when
// tree mode is active) to rows.
func (r *DetailsRenderer) addFiles(table *Table, rows *[]detailsRow, files []*fsx.File, depth int) {
	type fileWithXattrs struct {
		file      *fsx.File
		hasXattrs bool
		attrs     []xattr.Attribute
	}

	prepared := make([]fileWithXattrs, len(files))
	for i, f := range files {
		fx := fileWithXattrs{file: f}
		if r.Options.ShowXattrs && xattr.Enabled {
			if attrs, err := xattr.List(f.Path); err == nil && len(attrs) > 0 {
				fx.hasXattrs = true
				fx.attrs = attrs
			}
		}
		prepared[i] = fx
	}

	// Sorting happens here, on every call (not just once at the top),
	// because each directory's contents -- including ones discovered
	// deeper into a tree recursion -- need to be in their final order
	// before "is this the last entry" can be determined below.
	r.Filter.SortFiles(files)

	for i, item := range prepared {
		f := item.file
		isLast := i == len(prepared)-1

		var cells []Cell
		if table != nil {
			cells = table.RowForFile(f, item.hasXattrs)
		}

		name := RenderFileName(r.Theme, r.FileNameOptions, f, true)

		*rows = append(*rows, detailsRow{
			tree:  TreeParams{Depth: depth - 1, Last: isLast && len(item.attrs) == 0},
			cells: cells,
			name:  name,
		})

		for ai, attr := range item.attrs {
			*rows = append(*rows, detailsRow{
				tree: TreeParams{Depth: depth, Last: isLast && ai == len(item.attrs)-1},
				name: xattrCell(r.Theme, attr),
			})
		}

		if r.Recurse != nil && f.IsDirectory() && !f.IsDotEntry && r.Recurse.Tree && !r.Recurse.IsTooDeep(depth) {
			sub, err := fsx.ReadDir(f.Path)
			if err != nil {
				continue
			}

			gitIgnore := r.Filter.GitIgnore == fsx.GitIgnoreCheckAndIgnore
			children, _ := sub.Files(r.Filter.DotFilter, r.Git, gitIgnore)
			children = r.Filter.FilterChildFiles(children)
			r.Filter.SortFiles(children)

			if len(children) > 0 {
				r.addFiles(table, rows, children, depth+1)
			}
		}
	}
}

func xattrCell(th *theme.Theme, a xattr.Attribute) Cell {
	var c Cell
	c.Text(th.UI.Permissions.Attribute, fmt.Sprintf("%s (len %d)", a.Name, a.Size))
	return c
}

func (r *DetailsRenderer) writeRows(w io.Writer, table *Table, rows []detailsRow) error {
	var trunk TreeTrunk
	totalWidth := 0
	if table != nil {
		for _, width := range table.widths {
			totalWidth += width + 1
		}
	}

	for _, row := range rows {
		var line Cell

		if table != nil {
			if row.cells != nil {
				line.Append(table.Render(row.cells))
			} else {
				line.Spaces(totalWidth)
			}
		}

		// Tree branch segments butt directly up against each other (each
		// is already 3 characters wide, e.g. "│  " or "├──"); exactly one
		// separating space goes between the last of them and the name.
		for _, part := range trunk.NewRow(row.tree) {
			line.Text(r.Theme.UI.Punctuation, part.asciiArt())
		}
		if !row.tree.IsRoot() {
			line.Spaces(1)
		}

		line.Append(row.name)

		if _, err := fmt.Fprintln(w, line.String()); err != nil {
			return err
		}
	}

	return nil
}
