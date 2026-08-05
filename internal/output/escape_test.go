package output

import (
	"strings"
	"testing"

	"lsgo/internal/style"
)

func TestEscapeIntoAllPrintable(t *testing.T) {
	var c Cell
	escapeInto(&c, "hello.txt", style.Red.Normal(), style.Yellow.Normal())
	if c.Width != len("hello.txt") {
		t.Errorf("Width = %d, want %d", c.Width, len("hello.txt"))
	}
}

func TestEscapeIntoControlChars(t *testing.T) {
	var c Cell
	escapeInto(&c, "a\tb\nc\rd\x01e", style.Red.Normal(), style.Yellow.Normal())

	got := c.String()
	for _, want := range []string{`\t`, `\n`, `\r`, `\u{1}`} {
		if !strings.Contains(got, want) {
			t.Errorf("expected escaped output to contain %q, got %q", want, got)
		}
	}
}

func TestEscapeIntoLeadingAndTrailingControlChars(t *testing.T) {
	var c Cell
	escapeInto(&c, "\nstart-and-end\n", style.Red.Normal(), style.Yellow.Normal())
	if !strings.Contains(c.String(), "start-and-end") {
		t.Errorf("expected surrounding text to survive, got %q", c.String())
	}
}

func TestIsPrintable(t *testing.T) {
	cases := map[rune]bool{
		'a':    true,
		' ':    true,
		'\x1f': false,
		'\x7f': false,
	}
	for r, want := range cases {
		if got := isPrintable(r); got != want {
			t.Errorf("isPrintable(%q) = %v, want %v", r, got, want)
		}
	}
}

func TestIsAllPrintable(t *testing.T) {
	if !isAllPrintable("plain text") {
		t.Error("expected plain text to be all printable")
	}
	if isAllPrintable("bad\x01char") {
		t.Error("expected control char to make string not all printable")
	}
}

func TestEscapeControlChar(t *testing.T) {
	cases := map[rune]string{
		'\t': `\t`,
		'\n': `\n`,
		'\r': `\r`,
		0x01: `\u{1}`,
	}
	for r, want := range cases {
		if got := escapeControlChar(r); got != want {
			t.Errorf("escapeControlChar(%q) = %q, want %q", r, got, want)
		}
	}
}
