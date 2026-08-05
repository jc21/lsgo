package theme

import (
	"testing"

	"lsgo/internal/style"
)

func TestColourPairToStyle(t *testing.T) {
	p := ColourPair{Key: "ex", Value: "31;1"}
	if got := p.ToStyle(); got != style.Red.Bold() {
		t.Errorf("ToStyle() = %+v, want Red.Bold()", got)
	}
}

func TestThemeApplyOverlay(t *testing.T) {
	base := style.Red.Normal()
	overlay := style.Style{Underline: true}
	got := ApplyOverlay(base, overlay)
	if !got.Underline || got.Foreground != style.Red {
		t.Errorf("ApplyOverlay() = %+v, want red foreground with underline", got)
	}
}

func TestBuildWithGlobPattern(t *testing.T) {
	th := Build(ColourAlways, ScaleFixed, true, "*.mp3=35")
	if th.Colourer == nil {
		t.Fatal("expected a colourer to be set")
	}
}

func TestParseColourVarsRecognisedAndGlob(t *testing.T) {
	ui := DefaultUIStyles(ScaleFixed)
	exts := parseColourVars(&ui, "di=31:*.mp3=35")

	if ui.FileKinds.Directory != style.Red.Normal() {
		t.Errorf("expected 'di' to update the UI directly, got %+v", ui.FileKinds.Directory)
	}
	if !exts.nonEmpty() {
		t.Error("expected the glob pattern to be collected as an extension mapping")
	}
}
