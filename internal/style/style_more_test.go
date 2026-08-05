package style

import "testing"

func TestColourConstructors(t *testing.T) {
	if got := Blue.Dimmed(); !got.Dim || got.Foreground != Blue {
		t.Errorf("Dimmed() = %+v", got)
	}
	if got := Blue.Italic(); !got.Italic || got.Foreground != Blue {
		t.Errorf("Italic() = %+v", got)
	}
	if got := Blue.Underline(); !got.Underline || got.Foreground != Blue {
		t.Errorf("Underline() = %+v", got)
	}
	if got := Blue.On(Yellow); got.Foreground != Blue || got.Background != Yellow {
		t.Errorf("On() = %+v", got)
	}
	if got := Blue.Normal(); got.Foreground != Blue || got.Bold {
		t.Errorf("Normal() = %+v", got)
	}
}

func TestStyleSetters(t *testing.T) {
	var s Style
	s = s.SetDim()
	s = s.SetItalic()
	s = s.SetBlink()
	s = s.SetReverse()
	s = s.SetHidden()
	s = s.SetStrikethrough()
	s = s.On(Purple)

	if !s.Dim || !s.Italic || !s.Blink || !s.Reverse || !s.Hidden || !s.Strikethrough {
		t.Errorf("expected every attribute set, got %+v", s)
	}
	if s.Background != Purple {
		t.Errorf("expected background Purple, got %+v", s.Background)
	}
}

func TestPaintAllAttributes(t *testing.T) {
	s := Style{
		Bold: true, Dim: true, Italic: true, Underline: true,
		Blink: true, Reverse: true, Hidden: true, Strikethrough: true,
	}
	got := s.Paint("x")
	want := "\x1b[1;2;3;4;5;7;8;9mx\x1b[0m"
	if got != want {
		t.Errorf("Paint() = %q, want %q", got, want)
	}
}

func TestBasicBackground(t *testing.T) {
	s := Style{Background: Green}
	got := s.Paint("x")
	want := "\x1b[42mx\x1b[0m"
	if got != want {
		t.Errorf("Paint() = %q, want %q", got, want)
	}
}

func TestFixedBackground(t *testing.T) {
	s := Style{Background: Fixed(200)}
	got := s.Paint("x")
	want := "\x1b[48;5;200mx\x1b[0m"
	if got != want {
		t.Errorf("Paint() = %q, want %q", got, want)
	}
}
