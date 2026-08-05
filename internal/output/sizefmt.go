package output

import (
	"strconv"
	"strings"

	"lsgo/internal/style"
	"lsgo/internal/theme"
)

// SizeFormat selects how the file-size column's numbers are formatted.
type SizeFormat int

const (
	// DecimalSize uses SI (1000-based) prefixes: 2.1M. This is the
	// default.
	DecimalSize SizeFormat = iota
	// BinarySize uses IEC (1024-based) prefixes: 1.0Mi.
	BinarySize
	// JustBytes shows the exact byte count, comma-grouped.
	JustBytes
)

var decimalPrefixes = []string{"k", "M", "G", "T", "P", "E", "Z", "Y"}
var binaryPrefixes = []string{"Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi", "Yi"}

// magnitude buckets a size into one of the style categories the theme
// distinguishes: kilo, mega, giga, or "huge" (tera and above). A size with
// no prefix at all (under 1000/1024) is "byte".
type magnitude int

const (
	magByte magnitude = iota
	magKilo
	magMega
	magGiga
	magHuge
)

// splitPrefix divides size by the given base repeatedly, returning the
// scaled value and which prefix (if any) was used.
func splitPrefix(size float64, base float64, prefixes []string) (value float64, symbol string, mag magnitude, none bool) {
	value = size
	idx := -1 // counts completed divisions minus one, so 1 division => index 0 ("k")
	for value >= base && idx < len(prefixes)-1 {
		value /= base
		idx++
	}

	if idx == -1 {
		return size, "", magByte, true
	}

	switch idx {
	case 0:
		mag = magKilo
	case 1:
		mag = magMega
	case 2:
		mag = magGiga
	default:
		mag = magHuge
	}

	return value, prefixes[idx], mag, false
}

func sizeStyle(ui *theme.UIStyles, mag magnitude) (number, unit style.Style) {
	switch mag {
	case magKilo:
		return ui.Size.NumberKilo, ui.Size.UnitKilo
	case magMega:
		return ui.Size.NumberMega, ui.Size.UnitMega
	case magGiga:
		return ui.Size.NumberGiga, ui.Size.UnitGiga
	case magHuge:
		return ui.Size.NumberHuge, ui.Size.UnitHuge
	default:
		return ui.Size.NumberByte, ui.Size.UnitByte
	}
}

// RenderSize formats a regular file's byte count according to format.
func RenderSize(ui *theme.UIStyles, format SizeFormat, size uint64) Cell {
	var c Cell

	if format == JustBytes {
		// Still colour it by the magnitude it would have under the
		// binary scale: borrow binary's style, but print the exact
		// byte count rather than a scaled-down number.
		_, _, mag, _ := splitPrefix(float64(size), 1024, binaryPrefixes)
		numberStyle, _ := sizeStyle(ui, mag)
		c.Text(numberStyle, formatThousands(size))
		return c
	}

	base := 1000.0
	prefixes := decimalPrefixes
	if format == BinarySize {
		base = 1024.0
		prefixes = binaryPrefixes
	}

	value, symbol, mag, none := splitPrefix(float64(size), base, prefixes)
	numberStyle, unitStyle := sizeStyle(ui, mag)

	if none {
		c.Text(numberStyle, formatThousands(size))
		return c
	}

	var numberText string
	if value < 10 {
		numberText = strconv.FormatFloat(value, 'f', 1, 64)
	} else {
		numberText = formatThousands(uint64(value + 0.5))
	}

	c.Text(numberStyle, numberText)
	c.Text(unitStyle, symbol)
	return c
}

// RenderNoSize renders the placeholder shown for entries (like
// directories) that don't have a meaningful size.
func RenderNoSize(ui *theme.UIStyles) Cell {
	var c Cell
	c.Text(ui.Punctuation, "-")
	return c
}

// RenderDeviceIDs renders a character or block device's major/minor
// numbers in place of a size.
func RenderDeviceIDs(ui *theme.UIStyles, major, minor uint32) Cell {
	var c Cell
	c.Text(ui.Size.Major, strconv.FormatUint(uint64(major), 10))
	c.Text(ui.Punctuation, ",")
	c.Text(ui.Size.Minor, strconv.FormatUint(uint64(minor), 10))
	return c
}

// formatThousands renders n with ',' as the thousands separator, e.g.
// "1,048,576".
func formatThousands(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if len(s) <= 3 {
		return s
	}

	var b strings.Builder
	lead := len(s) % 3
	if lead == 0 {
		lead = 3
	}
	b.WriteString(s[:lead])
	for i := lead; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// formatInt is a small convenience used by non-size columns (hard link
// counts) that share the same thousands-grouping.
func formatInt(n uint64) string { return formatThousands(n) }
