package fsx

import "testing"

func TestIgnorePatternsMatching(t *testing.T) {
	ip := NewIgnorePatterns([]string{"*.mp3", "nothing"})

	if !ip.IsIgnored("test.mp3") {
		t.Error("expected *.mp3 to match test.mp3")
	}
	if !ip.IsIgnored("nothing") {
		t.Error("expected exact filename match")
	}
	if ip.IsIgnored("test.wav") {
		t.Error("did not expect test.wav to be ignored")
	}
}

func TestEmptyIgnorePatternsMatchesNothing(t *testing.T) {
	ip := NewIgnorePatterns(nil)
	if ip.IsIgnored("anything") {
		t.Error("expected empty pattern set to match nothing")
	}
}

func TestDotFilterFromCount(t *testing.T) {
	cases := []struct {
		count int
		want  DotFilter
	}{
		{0, JustFiles},
		{1, Dotfiles},
		{2, DotfilesAndDots},
		{5, DotfilesAndDots},
	}

	for _, c := range cases {
		if got := DotFilterFromCount(c.count); got != c.want {
			t.Errorf("DotFilterFromCount(%d) = %v, want %v", c.count, got, c.want)
		}
	}
}

func TestSortSpecCompareByExtension(t *testing.T) {
	a := &File{Name: "b.txt", Ext: "txt"}
	b := &File{Name: "a.zip", Ext: "zip"}

	spec := SortSpec{Field: SortByExtension, Case: CaseInsensitive}
	if spec.Compare(a, b) >= 0 {
		t.Error("expected txt to sort before zip")
	}
}

func TestSortSpecCompareNameMixHiddenStripsDot(t *testing.T) {
	a := &File{Name: ".bashrc"}
	b := &File{Name: "apple"}

	spec := SortSpec{Field: SortByNameMixHidden, Case: CaseInsensitive}
	// ".bashrc" strips to "bashrc", which sorts after "apple".
	if spec.Compare(a, b) <= 0 {
		t.Error("expected .bashrc (stripped to bashrc) to sort after apple")
	}
}
