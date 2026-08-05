// Package cli parses lsgo's command-line arguments into a fully
// resolved Options value.
package cli

// valueKind says whether a flag takes an argument.
type valueKind int

const (
	noValue valueKind = iota
	takesValue
)

// flagDef describes one recognised flag: its short form (0 if none), its
// canonical long name, and whether it takes a value.
type flagDef struct {
	short byte
	long  string
	kind  valueKind
}

// Canonical long names for the four per-timestamp long-view flags. These
// are checked directly (via flags.Has) in options.go's deduceTimeTypes, in
// addition to appearing in flagDefs below and as --sort/--time value
// aliases, so they're named constants rather than repeated literals.
const (
	longModified = "modified"
	longChanged  = "changed"
	longAccessed = "accessed"
	longCreated  = "created"
)

// flagDefs is the full set of flags lsgo understands, mirroring the
// reference implementation's option surface.
var flagDefs = []flagDef{
	{short: 'v', long: "version"},
	{short: '?', long: "help"},

	// Display options.
	{short: '1', long: "oneline"},
	{short: 'l', long: "long"},
	{short: 'G', long: "grid"},
	{short: 'x', long: "across"},
	{short: 'R', long: "recurse"},
	{short: 'T', long: "tree"},
	{short: 'F', long: "classify"},
	{long: "color", kind: takesValue},
	{long: "color-scale"},

	// Filtering and sorting.
	{short: 'a', long: "all"},
	{short: 'd', long: "list-dirs"},
	{short: 'L', long: "level", kind: takesValue},
	{short: 'r', long: "reverse"},
	{short: 's', long: "sort", kind: takesValue},
	{short: 'I', long: "ignore-glob", kind: takesValue},
	{long: "git-ignore"},
	{long: "group-directories-first"},
	{short: 'D', long: "only-dirs"},

	// Long-view display options.
	{short: 'b', long: "binary"},
	{short: 'B', long: "bytes"},
	{short: 'g', long: "group"},
	{short: 'n', long: "numeric"},
	{short: 'h', long: "header"},
	{long: "icons"},
	{short: 'i', long: "inode"},
	{short: 'H', long: "links"},
	{short: 'm', long: longModified},
	{long: longChanged},
	{short: 'S', long: "blocks"},
	{short: 't', long: "time", kind: takesValue},
	{short: 'u', long: longAccessed},
	{short: 'U', long: longCreated},
	{long: "time-style", kind: takesValue},

	// Suppressing columns.
	{long: "no-permissions"},
	{long: "no-filesize"},
	{long: "no-user"},
	{long: "no-time"},
	{long: "no-icons"},

	// Optional features.
	{long: "git"},
	{short: '@', long: "extended"},
	{long: "octal-permissions"},
}

// longAliases maps alternate spellings onto the canonical long name used
// in flagDefs and everywhere else in this package. British and American
// spellings of "colour" are both accepted, exactly as in the reference
// implementation.
var longAliases = map[string]string{
	"colour":       "color",
	"colour-scale": "color-scale",
}

func canonicalLong(name string) string {
	if canon, ok := longAliases[name]; ok {
		return canon
	}
	return name
}

func findByShort(ch byte) (flagDef, bool) {
	for _, f := range flagDefs {
		if f.short == ch {
			return f, true
		}
	}
	return flagDef{}, false
}

func findByLong(name string) (flagDef, bool) {
	name = canonicalLong(name)
	for _, f := range flagDefs {
		if f.long == name {
			return f, true
		}
	}
	return flagDef{}, false
}
