package cli

import (
	"reflect"
	"testing"
)

func TestParseBooleanShortFlags(t *testing.T) {
	f, err := Parse([]string{"-la"})
	if err != nil {
		t.Fatal(err)
	}
	if !f.Has("long") || !f.Has("all") {
		t.Errorf("expected both long and all set, got counts=%v", f.counts)
	}
}

func TestParseAllTwiceCounts(t *testing.T) {
	f, err := Parse([]string{"-aa"})
	if err != nil {
		t.Fatal(err)
	}
	if f.Count("all") != 2 {
		t.Errorf("expected count 2, got %d", f.Count("all"))
	}
}

func TestParseShortValueAttached(t *testing.T) {
	f, err := Parse([]string{"-RL4"})
	if err != nil {
		t.Fatal(err)
	}
	if !f.Has("recurse") {
		t.Error("expected recurse to be set")
	}
	v, ok := f.Value("level")
	if !ok || v != "4" {
		t.Errorf("expected level=4, got %q (ok=%v)", v, ok)
	}
}

func TestParseShortValueWithEquals(t *testing.T) {
	f, err := Parse([]string{"-L=5"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := f.Value("level")
	if v != "5" {
		t.Errorf("expected level=5, got %q", v)
	}
}

func TestParseShortValueSeparateArg(t *testing.T) {
	f, err := Parse([]string{"-L", "6"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := f.Value("level")
	if v != "6" {
		t.Errorf("expected level=6, got %q", v)
	}
}

func TestParseLongValueWithEquals(t *testing.T) {
	f, err := Parse([]string{"--sort=" + sortValueSize})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := f.Value("sort")
	if v != sortValueSize {
		t.Errorf("expected sort=%s, got %q", sortValueSize, v)
	}
}

func TestParseLongValueSeparateArg(t *testing.T) {
	f, err := Parse([]string{"--sort", sortValueSize})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := f.Value("sort")
	if v != sortValueSize {
		t.Errorf("expected sort=%s, got %q", sortValueSize, v)
	}
}

func TestParseColourAliasesUnifyToColor(t *testing.T) {
	f, err := Parse([]string{"--colour=never"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := f.Value("color")
	if !ok || v != "never" {
		t.Errorf("expected canonical 'color' key to hold 'never', got %q (ok=%v)", v, ok)
	}
}

func TestParseLastValueWins(t *testing.T) {
	f, err := Parse([]string{"--sort=size", "--sort=name"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := f.Value("sort")
	if v != "name" {
		t.Errorf("expected last value 'name' to win, got %q", v)
	}
}

func TestParseFreeArguments(t *testing.T) {
	f, err := Parse([]string{"-l", "file1.txt", "--all", "dir2"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"file1.txt", "dir2"}
	if !reflect.DeepEqual(f.Free, want) {
		t.Errorf("got free args %v, want %v", f.Free, want)
	}
}

func TestParseDoubleDashStopsFlagParsing(t *testing.T) {
	f, err := Parse([]string{"-l", "--", "-notaflag"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-notaflag"}
	if !reflect.DeepEqual(f.Free, want) {
		t.Errorf("got %v, want %v", f.Free, want)
	}
}

func TestParseBareDashIsFreeArgument(t *testing.T) {
	f, err := Parse([]string{"-"})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Free) != 1 || f.Free[0] != "-" {
		t.Errorf("expected bare '-' to be a free argument, got %v", f.Free)
	}
}

func TestParseUnknownLongFlagErrors(t *testing.T) {
	_, err := Parse([]string{"--nonsense"})
	if err == nil {
		t.Fatal("expected an error for an unknown flag")
	}
}

func TestParseUnknownShortFlagErrors(t *testing.T) {
	_, err := Parse([]string{"-Z"})
	if err == nil {
		t.Fatal("expected an error for an unknown short flag")
	}
}

func TestParseMissingValueErrors(t *testing.T) {
	_, err := Parse([]string{"--sort"})
	if err == nil {
		t.Fatal("expected an error when a value-taking flag has nothing after it")
	}
}

func TestLastOfPicksMostRecentAcrossDifferentFlags(t *testing.T) {
	f, err := Parse([]string{flagOneline, flagLong, flagGrid})
	if err != nil {
		t.Fatal(err)
	}
	winner, ok := f.LastOf("long", "oneline", "grid", "tree")
	if !ok || winner != "grid" {
		t.Errorf("expected 'grid' to win (it's last), got %q (ok=%v)", winner, ok)
	}
}
