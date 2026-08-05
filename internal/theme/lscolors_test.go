package theme

import (
	"testing"

	"lsgo/internal/style"
)

func TestParseSGRBasic(t *testing.T) {
	cases := []struct {
		in   string
		want style.Style
	}{
		{"1", style.Style{Bold: true}},
		{"01", style.Style{Bold: true}},
		{"4", style.Style{Underline: true}},
		{"1;4", style.Style{Bold: true, Underline: true}},
		{"31", style.Red.Normal()},
		{"43", style.Style{Background: style.Yellow}},
		{"31;43", style.Red.On(style.Yellow)},
		{"43;31;1;4", style.Red.On(style.Yellow).SetBold().SetUnderline()},
		{"", style.Style{}},
		{";;;;;;", style.Style{}},
		{"99999999", style.Style{}},
		{"GREEN", style.Style{}},
		{"38;5;149", style.Fixed(149).Normal()},
		{"48;5;1", style.Style{Background: style.Fixed(1)}},
		{"48;5;1;1", style.Style{Background: style.Fixed(1), Bold: true}},
		{"4;48;5;1", style.Style{Background: style.Fixed(1), Underline: true}},
		{"38;2;255;100;0", style.Style{Foreground: style.RGB(255, 100, 0)}},
		{"48;5;999", style.Style{}},
	}

	for _, c := range cases {
		got := ParseSGR(c.in)
		if got != c.want {
			t.Errorf("ParseSGR(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestEachColourPair(t *testing.T) {
	var got []ColourPair
	EachColourPair("di=31:ex=32", func(p ColourPair) { got = append(got, p) })

	if len(got) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(got))
	}
	if got[0].Key != "di" || got[0].Value != "31" {
		t.Errorf("first pair = %+v", got[0])
	}
	if got[1].Key != "ex" || got[1].Value != "32" {
		t.Errorf("second pair = %+v", got[1])
	}
}

func TestEachColourPairSkipsMalformed(t *testing.T) {
	var got []ColourPair
	EachColourPair("=di:id=:ok=1", func(p ColourPair) { got = append(got, p) })

	if len(got) != 1 || got[0].Key != "ok" {
		t.Errorf("expected only the valid pair, got %+v", got)
	}
}

func TestUIStylesSetLS(t *testing.T) {
	var ui UIStyles
	if !ui.SetLS("di", style.Red.Normal()) {
		t.Fatal("expected 'di' to be recognised")
	}
	if ui.FileKinds.Directory != style.Red.Normal() {
		t.Error("directory style not applied")
	}

	if ui.SetLS("zz", style.Style{}) {
		t.Error("expected unrecognised key to return false")
	}
}

func TestBuildThemeNeverUsesPlainStyles(t *testing.T) {
	th := Build(ColourNever, ScaleFixed, true, "di=31")
	if th.UI.FileKinds.Directory != (style.Style{}) {
		t.Error("expected ColourNever to produce entirely plain styles")
	}
}

func TestBuildThemeAutomaticRespectsTerminal(t *testing.T) {
	th := Build(ColourAutomatic, ScaleFixed, false, "")
	if th.UI.FileKinds.Directory != (style.Style{}) {
		t.Error("expected non-terminal output with Automatic to be plain")
	}

	th2 := Build(ColourAutomatic, ScaleFixed, true, "")
	if th2.UI.FileKinds.Directory == (style.Style{}) {
		t.Error("expected terminal output with Automatic to be coloured")
	}
}
