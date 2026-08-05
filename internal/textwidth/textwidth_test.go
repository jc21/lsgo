package textwidth

import "testing"

func TestRuneWidth(t *testing.T) {
	cases := []struct {
		r    rune
		want int
	}{
		{0, 0},
		{'\t', 0},
		{0x7f, 0},
		{'a', 1},
		{'あ', 2}, // Hiragana, CJK range
		{'가', 2}, // Hangul syllable
		{'😀', 2}, // emoji block
		{'１', 2}, // fullwidth form
		{'!', 1},
	}
	for _, c := range cases {
		if got := RuneWidth(c.r); got != c.want {
			t.Errorf("RuneWidth(%q) = %d, want %d", c.r, got, c.want)
		}
	}
}

func TestWidth(t *testing.T) {
	if got := Width("hello"); got != 5 {
		t.Errorf("Width(hello) = %d, want 5", got)
	}
	if got := Width("あい"); got != 4 {
		t.Errorf("Width(あい) = %d, want 4", got)
	}
	if got := Width(""); got != 0 {
		t.Errorf("Width(\"\") = %d, want 0", got)
	}
	if got := Width("a\tb"); got != 2 {
		t.Errorf("Width(a\\tb) = %d, want 2 (tab is zero width)", got)
	}
}
