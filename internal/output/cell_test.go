package output

import (
	"testing"

	"lsgo/internal/style"
)

func TestNewCellIsEmpty(t *testing.T) {
	c := NewCell()
	if c.Width != 0 || c.String() != "" {
		t.Errorf("expected empty cell, got width=%d text=%q", c.Width, c.String())
	}
}

func TestCellPlainAndSpaces(t *testing.T) {
	var c Cell
	c.Plain("abc")
	c.Spaces(2)
	c.Spaces(0) // no-op, shouldn't add anything
	c.Spaces(-1)

	if c.Width != 5 {
		t.Errorf("Width = %d, want 5", c.Width)
	}
	if c.String() != "abc  " {
		t.Errorf("String() = %q, want %q", c.String(), "abc  ")
	}
}

func TestCellText(t *testing.T) {
	var c Cell
	c.Text(style.Red.Bold(), "hi")
	if c.Width != 2 {
		t.Errorf("Width = %d, want 2", c.Width)
	}
	if c.String() == "" {
		t.Error("expected styled text to produce non-empty output")
	}
}

func TestCellAppend(t *testing.T) {
	var a, b Cell
	a.Plain("foo")
	b.Plain("bar")
	a.Append(b)

	if a.Width != 6 {
		t.Errorf("Width = %d, want 6", a.Width)
	}
	if a.String() != "foobar" {
		t.Errorf("String() = %q, want %q", a.String(), "foobar")
	}
}

func TestCellPadTo(t *testing.T) {
	var c Cell
	c.Plain("hi")

	if got := c.PadTo(5); got != 3 {
		t.Errorf("PadTo(5) = %d, want 3", got)
	}
	if got := c.PadTo(1); got != 0 {
		t.Errorf("PadTo(1) = %d, want 0 (never negative)", got)
	}
	if got := c.PadTo(2); got != 0 {
		t.Errorf("PadTo(2) = %d, want 0", got)
	}
}
