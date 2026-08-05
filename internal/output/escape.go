package output

import (
	"fmt"

	"lsgo/internal/style"
)

// escapeInto appends s to cell, painting ordinary printable runs with
// good and escaping any control characters (which would otherwise corrupt
// the terminal display, e.g. embedded newlines) with bad.
func escapeInto(cell *Cell, s string, good, bad style.Style) {
	// Fast path: this covers the overwhelming majority of filenames.
	if isAllPrintable(s) {
		cell.Text(good, s)
		return
	}

	start := 0
	runes := []rune(s)
	flush := func(end int) {
		if end > start {
			cell.Text(good, string(runes[start:end]))
		}
	}

	for i, r := range runes {
		if isPrintable(r) {
			continue
		}
		flush(i)
		cell.Text(bad, escapeControlChar(r))
		start = i + 1
	}
	flush(len(runes))
}

func isAllPrintable(s string) bool {
	for _, r := range s {
		if !isPrintable(r) {
			return false
		}
	}
	return true
}

// isPrintable reports whether r can be shown as-is: anything except the
// C0 control characters and DEL. This deliberately doesn't exclude
// non-ASCII characters, which are printable in a UTF-8 terminal.
func isPrintable(r rune) bool {
	return r >= 0x20 && r != 0x7f
}

// escapeControlChar renders a single non-printable rune the way Rust's
// char::escape_default does: familiar two-letter escapes for the common
// whitespace controls, and a "\u{XX}" hex escape for everything else.
func escapeControlChar(r rune) string {
	switch r {
	case '\t':
		return `\t`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	default:
		return fmt.Sprintf(`\u{%x}`, r)
	}
}
