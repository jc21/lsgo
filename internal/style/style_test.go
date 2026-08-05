package style

import "testing"

func TestPlainStyleDoesNotPaint(t *testing.T) {
	var s Style
	if got := s.Paint("hello"); got != "hello" {
		t.Errorf("expected plain style to leave text untouched, got %q", got)
	}
}

func TestBasicForeground(t *testing.T) {
	got := Red.Normal().Paint("x")
	want := "\x1b[31mx\x1b[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBoldUnderlineCombination(t *testing.T) {
	got := Green.Bold().SetUnderline().Paint("ok")
	want := "\x1b[1;4;32mok\x1b[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFixedColour(t *testing.T) {
	got := Fixed(133).Normal().Paint("y")
	want := "\x1b[38;5;133my\x1b[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRGBBackground(t *testing.T) {
	s := Style{Background: RGB(255, 100, 0)}
	got := s.Paint("z")
	want := "\x1b[48;2;255;100;0mz\x1b[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApplyOverlay(t *testing.T) {
	base := Red.Normal()
	overlay := Style{Underline: true}
	result := ApplyOverlay(base, overlay)

	if !result.Underline {
		t.Error("expected underline to be applied from overlay")
	}
	if result.Foreground != Red {
		t.Error("expected base foreground to be preserved")
	}
}

func TestEmptyTextNeverPainted(t *testing.T) {
	if got := Red.Bold().Paint(""); got != "" {
		t.Errorf("expected empty string to stay empty, got %q", got)
	}
}
