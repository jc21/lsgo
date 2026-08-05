package fsx

import "testing"

func TestNaturalCompareNumericOrdering(t *testing.T) {
	const file2 = "file2" // reused below for the "equal to itself" case

	cases := []struct {
		a, b string
		want int
	}{
		{file2, "file10", -1},
		{"file10", file2, 1},
		{file2, file2, 0},
		{"a", "b", -1},
		{"img09.png", "img10.png", -1},
		{"img9.png", "img10.png", -1},
		{"007", "7", 0},
		{"100", "99", 1},
	}

	for _, c := range cases {
		got := naturalCompare(c.a, c.b)
		if sign(got) != sign(c.want) {
			t.Errorf("naturalCompare(%q, %q) = %d, want sign %d", c.a, c.b, got, c.want)
		}
	}
}

func TestNaturalCompareFoldIgnoresCase(t *testing.T) {
	if naturalCompareFold("Apple", "apple") != 0 {
		t.Error("expected case-insensitive compare to treat Apple == apple")
	}
	if naturalCompare("Apple", "apple") == 0 {
		t.Error("expected case-sensitive compare to distinguish Apple from apple")
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
