// Package style provides a small ANSI text-styling primitive used throughout
// lsgo's output renderers. It intentionally mirrors the ergonomics of a
// "paint this text with this style" API: colours and attributes are built up
// by chaining methods on a Colour or Style value, and turned into an escaped
// string with Style.Paint.
//
// A zero-value Style paints text unchanged, which makes "no colour" (for
// --color=never, non-terminal output, etc.) the natural default.
package style

import (
	"strconv"
	"strings"
)

// colourKind distinguishes the three ways a terminal colour can be specified.
type colourKind uint8

const (
	kindNone colourKind = iota
	kindBasic
	kindFixed
	kindRGB
)

// Colour is a single foreground or background colour. The zero value means
// "no colour set".
type Colour struct {
	kind    colourKind
	basic   uint8 // 0-7, for the eight standard ANSI colours
	fixed   uint8 // 0-255, for the 256-colour palette
	r, g, b uint8
}

// The eight standard ANSI colours.
var (
	Black  = Colour{kind: kindBasic, basic: 0}
	Red    = Colour{kind: kindBasic, basic: 1}
	Green  = Colour{kind: kindBasic, basic: 2}
	Yellow = Colour{kind: kindBasic, basic: 3}
	Blue   = Colour{kind: kindBasic, basic: 4}
	Purple = Colour{kind: kindBasic, basic: 5}
	Cyan   = Colour{kind: kindBasic, basic: 6}
	White  = Colour{kind: kindBasic, basic: 7}
)

// Fixed returns one of the 256 extended-palette colours.
func Fixed(n uint8) Colour { return Colour{kind: kindFixed, fixed: n} }

// RGB returns a 24-bit true colour.
func RGB(r, g, b uint8) Colour { return Colour{kind: kindRGB, r: r, g: g, b: b} }

func (c Colour) set() bool { return c.kind != kindNone }

// fgCodes/bgCodes return the SGR parameter codes needed to select this
// colour as a foreground or background colour, respectively.
func (c Colour) fgCodes() []string {
	switch c.kind {
	case kindBasic:
		return []string{strconv.Itoa(30 + int(c.basic))}
	case kindFixed:
		return []string{"38", "5", strconv.Itoa(int(c.fixed))}
	case kindRGB:
		return []string{"38", "2", strconv.Itoa(int(c.r)), strconv.Itoa(int(c.g)), strconv.Itoa(int(c.b))}
	default:
		return nil
	}
}

func (c Colour) bgCodes() []string {
	switch c.kind {
	case kindBasic:
		return []string{strconv.Itoa(40 + int(c.basic))}
	case kindFixed:
		return []string{"48", "5", strconv.Itoa(int(c.fixed))}
	case kindRGB:
		return []string{"48", "2", strconv.Itoa(int(c.r)), strconv.Itoa(int(c.g)), strconv.Itoa(int(c.b))}
	default:
		return nil
	}
}

// Style is a combination of an optional foreground colour, an optional
// background colour, and a set of text attributes. The zero value is
// "plain text, no styling".
type Style struct {
	Foreground    Colour
	Background    Colour
	Bold          bool
	Dim           bool
	Italic        bool
	Underline     bool
	Blink         bool
	Reverse       bool
	Hidden        bool
	Strikethrough bool
}

// Convenience constructors on Colour, mirroring the chainable style-builder
// pattern used throughout the renderers (e.g. style.Yellow.Bold()).

// Normal returns a Style using this foreground colour with no attributes.
func (c Colour) Normal() Style { return Style{Foreground: c} }

// Bold returns a Style using this foreground colour, bolded.
func (c Colour) Bold() Style { return Style{Foreground: c, Bold: true} }

// Dimmed returns a Style using this foreground colour, dimmed.
func (c Colour) Dimmed() Style { return Style{Foreground: c, Dim: true} }

// Italic returns a Style using this foreground colour, italicised.
func (c Colour) Italic() Style { return Style{Foreground: c, Italic: true} }

// Underline returns a Style using this foreground colour, underlined.
func (c Colour) Underline() Style { return Style{Foreground: c, Underline: true} }

// On returns a Style using this foreground colour on the given background.
func (c Colour) On(bg Colour) Style { return Style{Foreground: c, Background: bg} }

// Chain methods on Style, so attributes can be layered onto an existing
// style: e.g. Green.Bold().Underline().

// SetBold returns a copy of s with Bold set.
func (s Style) SetBold() Style { s.Bold = true; return s }

// SetDim returns a copy of s with Dim set.
func (s Style) SetDim() Style { s.Dim = true; return s }

// SetItalic returns a copy of s with Italic set.
func (s Style) SetItalic() Style { s.Italic = true; return s }

// SetUnderline returns a copy of s with Underline set.
func (s Style) SetUnderline() Style { s.Underline = true; return s }

// SetBlink returns a copy of s with Blink set.
func (s Style) SetBlink() Style { s.Blink = true; return s }

// SetReverse returns a copy of s with Reverse set.
func (s Style) SetReverse() Style { s.Reverse = true; return s }

// SetHidden returns a copy of s with Hidden set.
func (s Style) SetHidden() Style { s.Hidden = true; return s }

// SetStrikethrough returns a copy of s with Strikethrough set.
func (s Style) SetStrikethrough() Style { s.Strikethrough = true; return s }

// On returns a copy of s with Background set to bg.
func (s Style) On(bg Colour) Style { s.Background = bg; return s }

// IsPlain reports whether this style has no colours or attributes set, in
// which case painting is a no-op.
func (s Style) IsPlain() bool {
	return s == Style{}
}

// Paint wraps text in the ANSI escape codes for this style, if it has any
// effect. Plain (zero-value) styles are returned unchanged.
func (s Style) Paint(text string) string {
	if s.IsPlain() || text == "" {
		return text
	}

	var codes []string
	if s.Bold {
		codes = append(codes, "1")
	}
	if s.Dim {
		codes = append(codes, "2")
	}
	if s.Italic {
		codes = append(codes, "3")
	}
	if s.Underline {
		codes = append(codes, "4")
	}
	if s.Blink {
		codes = append(codes, "5")
	}
	if s.Reverse {
		codes = append(codes, "7")
	}
	if s.Hidden {
		codes = append(codes, "8")
	}
	if s.Strikethrough {
		codes = append(codes, "9")
	}
	codes = append(codes, s.Foreground.fgCodes()...)
	codes = append(codes, s.Background.bgCodes()...)

	if len(codes) == 0 {
		return text
	}

	return "\x1b[" + strings.Join(codes, ";") + "m" + text + "\x1b[0m"
}

// ApplyOverlay layers the attributes and colours of overlay on top of base,
// keeping anything from base that overlay doesn't set. This is used for
// "overlay" styles such as the underline applied to a broken symlink's
// target path.
func ApplyOverlay(base, overlay Style) Style {
	if overlay.Foreground.set() {
		base.Foreground = overlay.Foreground
	}
	if overlay.Background.set() {
		base.Background = overlay.Background
	}
	base.Bold = base.Bold || overlay.Bold
	base.Dim = base.Dim || overlay.Dim
	base.Italic = base.Italic || overlay.Italic
	base.Underline = base.Underline || overlay.Underline
	base.Blink = base.Blink || overlay.Blink
	base.Reverse = base.Reverse || overlay.Reverse
	base.Hidden = base.Hidden || overlay.Hidden
	base.Strikethrough = base.Strikethrough || overlay.Strikethrough
	return base
}
