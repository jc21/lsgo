package theme

import (
	"strconv"
	"strings"

	"github.com/jc21/lsgo/internal/style"
)

// ColourPair is one "key=value" entry from an LS_COLORS string, before
// it's been decided whether the key names a UI element or a filename
// glob.
type ColourPair struct {
	Key   string
	Value string
}

// EachColourPair splits an LS_COLORS-style string on ':' and calls fn for
// each syntactically valid "key=value" entry. Entries with
// zero or more than one '=', or an empty key/value, are silently skipped,
// matching the reference implementation's tolerant parsing.
func EachColourPair(s string, fn func(ColourPair)) {
	for _, entry := range strings.Split(s, ":") {
		parts := strings.SplitN(entry, "=", 3)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			fn(ColourPair{Key: parts[0], Value: parts[1]})
		}
	}
}

// ToStyle parses this pair's value as a sequence of ';'-separated SGR
// parameters (the same format LS_COLORS/dircolors and terminal escape
// codes use) into a style.Style.
func (p ColourPair) ToStyle() style.Style {
	return ParseSGR(p.Value)
}

// ParseSGR interprets a ';'-separated list of SGR parameter codes (as used
// by LS_COLORS entries and raw ANSI escapes, without the leading "\x1b["
// or trailing "m") into a Style.
func ParseSGR(value string) style.Style {
	var s style.Style

	parts := strings.Split(value, ";")
	for i := 0; i < len(parts); i++ {
		trimmed := trimLeadingZeros(parts[i])

		switch trimmed {
		case "1":
			s.Bold = true
		case "2":
			s.Dim = true
		case "3":
			s.Italic = true
		case "4":
			s.Underline = true
		case "5":
			s.Blink = true
		case "7":
			s.Reverse = true
		case "8":
			s.Hidden = true
		case "9":
			s.Strikethrough = true

		case "30":
			s.Foreground = style.Black
		case "31":
			s.Foreground = style.Red
		case "32":
			s.Foreground = style.Green
		case "33":
			s.Foreground = style.Yellow
		case "34":
			s.Foreground = style.Blue
		case "35":
			s.Foreground = style.Purple
		case "36":
			s.Foreground = style.Cyan
		case "37":
			s.Foreground = style.White
		case "38":
			c, consumed, ok := parseHighColour(parts[i+1:])
			if ok {
				s.Foreground = c
			}
			i += consumed

		case "40":
			s.Background = style.Black
		case "41":
			s.Background = style.Red
		case "42":
			s.Background = style.Green
		case "43":
			s.Background = style.Yellow
		case "44":
			s.Background = style.Blue
		case "45":
			s.Background = style.Purple
		case "46":
			s.Background = style.Cyan
		case "47":
			s.Background = style.White
		case "48":
			c, consumed, ok := parseHighColour(parts[i+1:])
			if ok {
				s.Background = c
			}
			i += consumed

		default:
			// Unrecognised codes (including out-of-range numbers) are
			// ignored rather than treated as an error.
		}
	}

	return s
}

// parseHighColour parses the parameters following a "38" or "48" code:
// either "5;N" for a 256-colour palette entry, or "2;R;G;B" for true
// colour. It returns how many of the following fields were consumed by
// this attempt, even when the value turned out to be invalid -- matching
// the reference parser, which reads the tokens off the stream before
// discovering they're out of range.
func parseHighColour(rest []string) (style.Colour, int, bool) {
	if len(rest) == 0 {
		return style.Colour{}, 0, false
	}

	switch rest[0] {
	case "5":
		if len(rest) < 2 {
			return style.Colour{}, 1, false
		}
		n, err := strconv.Atoi(rest[1])
		if err != nil || n < 0 || n > 255 {
			return style.Colour{}, 2, false
		}
		return style.Fixed(uint8(n)), 2, true

	case "2":
		consumed := 1
		parseByte := func(idx int) (uint8, bool) {
			if len(rest) <= idx {
				return 0, false
			}
			consumed = idx + 1
			n, err := strconv.Atoi(rest[idx])
			if err != nil || n < 0 || n > 255 {
				return 0, false
			}
			return uint8(n), true
		}

		r, rOK := parseByte(1)
		g, gOK := parseByte(2)
		b, bOK := parseByte(3)

		if consumed < 4 || !rOK || !gOK || !bOK {
			return style.Colour{}, consumed, false
		}
		return style.RGB(r, g, b), 4, true

	default:
		return style.Colour{}, 0, false
	}
}

func trimLeadingZeros(s string) string {
	i := 0
	for i < len(s)-1 && s[i] == '0' {
		i++
	}
	return s[i:]
}
