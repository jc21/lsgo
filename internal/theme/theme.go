package theme

import "lsgo/internal/style"

// UseColours controls under what circumstances coloured output is used.
type UseColours int

const (
	// ColourAutomatic shows colours only when standard output is a
	// terminal. This is the default.
	ColourAutomatic UseColours = iota
	ColourAlways
	ColourNever
)

// Theme bundles the resolved UI styles with the filename colourer/iconer
// that together decide how every piece of output gets painted.
type Theme struct {
	UI       UIStyles
	Colourer FileColourer
	Iconer   FileIconer
}

// Build resolves a Theme from the user's colour preferences and the
// LS_COLORS environment variable. isTerminal should reflect whether
// standard output is attached to a terminal, and is only consulted when
// use is ColourAutomatic.
func Build(use UseColours, scale ColourScale, isTerminal bool, lsColors string) Theme {
	if use == ColourNever || (use == ColourAutomatic && !isTerminal) {
		return Theme{UI: Plain(), Colourer: noFileColours{}}
	}

	ui := DefaultUIStyles(scale)
	exts := parseColourVars(&ui, lsColors)

	var colourer FileColourer
	if exts.nonEmpty() {
		colourer = chainedColours{first: exts, second: FileExtensions{}}
	} else {
		colourer = FileExtensions{}
	}

	return Theme{UI: ui, Colourer: colourer, Iconer: FileExtensions{}}
}

// parseColourVars applies LS_COLORS to ui, collecting any glob-based
// filename rules into an ExtensionMappings.
func parseColourVars(ui *UIStyles, lsColors string) *extensionMappings {
	exts := &extensionMappings{}

	if lsColors != "" {
		EachColourPair(lsColors, func(p ColourPair) {
			if !ui.SetLS(p.Key, p.ToStyle()) {
				exts.add(p.Key, p.ToStyle())
			}
		})
	}

	return exts
}

// ApplyOverlay re-exports style.ApplyOverlay for convenience within
// renderers that import theme but not style directly.
func ApplyOverlay(base, overlay style.Style) style.Style {
	return style.ApplyOverlay(base, overlay)
}
