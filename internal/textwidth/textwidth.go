// Package textwidth estimates how many terminal columns a string occupies.
//
// This deliberately ignores ANSI escape sequences (callers pass in the raw,
// unstyled text) and approximates Unicode "wide" characters -- mostly East
// Asian scripts and emoji -- as occupying two columns, which matches how
// most terminal emulators render them.
package textwidth

// Width returns the number of terminal columns the given string occupies,
// treating "wide" runes (East Asian Wide/Fullwidth, plus common emoji
// ranges) as two columns and everything else as one column. Control
// characters are counted as zero width; callers are expected to have
// already replaced them with a printable escape if they want them counted.
func Width(s string) int {
	total := 0
	for _, r := range s {
		total += RuneWidth(r)
	}
	return total
}

// RuneWidth returns the terminal column width of a single rune.
func RuneWidth(r rune) int {
	switch {
	case r == 0:
		return 0
	case r < 0x20 || r == 0x7f:
		// Control characters: callers should have escaped these already,
		// but fall back to zero width rather than guessing.
		return 0
	case isWide(r):
		return 2
	default:
		return 1
	}
}

// isWide reports whether r falls in a Unicode range that's conventionally
// rendered as a double-width glyph. This is a pragmatic subset of the East
// Asian Width "Wide"/"Fullwidth" categories plus common emoji blocks --
// enough to keep column alignment sane for the vast majority of filenames,
// without pulling in a full Unicode properties table.
func isWide(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115F, // Hangul Jamo
		r == 0x2329, r == 0x232A,
		r >= 0x2E80 && r <= 0xA4CF && r != 0x303F, // CJK, Yi, radicals
		r >= 0xAC00 && r <= 0xD7A3,                // Hangul syllables
		r >= 0xF900 && r <= 0xFAFF,                // CJK compatibility ideographs
		r >= 0xFE30 && r <= 0xFE6F,                // CJK compatibility forms
		r >= 0xFF00 && r <= 0xFF60,                // Fullwidth forms
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1FAFF, // emoji blocks
		r >= 0x20000 && r <= 0x3FFFD: // CJK extensions
		return true
	default:
		return false
	}
}
