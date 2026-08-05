package output

// TreePart is one segment of the vertical "trunk" lines drawn to the left
// of a file's name in tree view.
type TreePart int

const (
	// treeEdge is a non-last entry: "├──".
	treeEdge TreePart = iota
	// treeLine continues a still-open ancestor directory: "│  ".
	treeLine
	// treeCorner is the last entry in its directory: "└──".
	treeCorner
	// treeBlank is a closed-off ancestor directory, contributing only
	// blank space to keep later columns aligned: "   ".
	treeBlank
)

func (p TreePart) asciiArt() string {
	switch p {
	case treeEdge:
		return "├──"
	case treeLine:
		return "│  "
	case treeCorner:
		return "└──"
	default:
		return "   "
	}
}

// TreeParams says where in the tree one row sits: how deep it is, and
// whether it's the last entry among its siblings.
type TreeParams struct {
	Depth int
	Last  bool
}

// IsRoot reports whether this row is at the top level (depth 0), which
// gets no tree characters at all.
func (p TreeParams) IsRoot() bool { return p.Depth == 0 }

// TreeTrunk incrementally builds the tree-branch prefix for each row, by
// remembering, for every depth seen so far, whether that ancestor
// directory has more siblings to come.
type TreeTrunk struct {
	stack      []TreePart
	hasLast    bool
	lastParams TreeParams
}

// NewRow returns the sequence of tree parts (one per depth level below the
// root) to draw before this row's own entry.
func (tt *TreeTrunk) NewRow(params TreeParams) []TreePart {
	if tt.hasLast {
		if tt.lastParams.Last {
			tt.stack[tt.lastParams.Depth] = treeBlank
		} else {
			tt.stack[tt.lastParams.Depth] = treeLine
		}
	}

	for len(tt.stack) <= params.Depth {
		tt.stack = append(tt.stack, treeEdge)
	}
	tt.stack = tt.stack[:params.Depth+1]

	if params.Last {
		tt.stack[params.Depth] = treeCorner
	} else {
		tt.stack[params.Depth] = treeEdge
	}

	tt.lastParams = params
	tt.hasLast = true

	// Skip index 0: that's the root level, which never draws a branch
	// of its own (there's nothing above it to connect to).
	if len(tt.stack) <= 1 {
		return nil
	}
	return tt.stack[1:]
}
