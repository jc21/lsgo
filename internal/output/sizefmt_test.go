package output

import (
	"strings"
	"testing"

	"lsgo/internal/theme"
)

func TestFormatThousands(t *testing.T) {
	cases := map[uint64]string{
		0:         "0",
		5:         "5",
		999:       "999",
		1000:      "1,000",
		1048576:   "1,048,576",
		2100000:   "2,100,000",
		123456789: "123,456,789",
	}
	for n, want := range cases {
		if got := formatThousands(n); got != want {
			t.Errorf("formatThousands(%d) = %q, want %q", n, got, want)
		}
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestRenderSizeDecimal(t *testing.T) {
	ui := theme.DefaultUIStyles(theme.ScaleFixed)
	cell := RenderSize(&ui, DecimalSize, 2_100_000)
	if got := stripANSI(cell.String()); got != "2.1M" {
		t.Errorf("got %q, want 2.1M", got)
	}
}

func TestRenderSizeBinary(t *testing.T) {
	ui := theme.DefaultUIStyles(theme.ScaleFixed)
	cell := RenderSize(&ui, BinarySize, 1_048_576)
	if got := stripANSI(cell.String()); got != "1.0Mi" {
		t.Errorf("got %q, want 1.0Mi", got)
	}
}

func TestRenderSizeJustBytes(t *testing.T) {
	ui := theme.DefaultUIStyles(theme.ScaleFixed)
	cell := RenderSize(&ui, JustBytes, 1_048_576)
	if got := stripANSI(cell.String()); got != "1,048,576" {
		t.Errorf("got %q, want 1,048,576", got)
	}
}

func TestRenderSizeSmallValueNoUnit(t *testing.T) {
	ui := theme.DefaultUIStyles(theme.ScaleFixed)
	cell := RenderSize(&ui, DecimalSize, 512)
	if got := stripANSI(cell.String()); got != "512" {
		t.Errorf("got %q, want 512", got)
	}
}
