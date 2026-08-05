package theme

import (
	"os"
	"path/filepath"
	"testing"

	"lsgo/internal/fsx"
	"lsgo/internal/style"
)

func mustFile(t *testing.T, dir, name string) *fsx.File {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := fsx.NewFile(path, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFileExtensionsColourFileByClass(t *testing.T) {
	dir := t.TempDir()
	fe := FileExtensions{}

	// "Makefile" here (and "class" in TestSourceExtsFor below) recur
	// against the reference tables in icons.go/filetype.go, whose own
	// goconst suppression doesn't cover new literals in this file.
	//
	//nolint:goconst
	cases := []struct {
		name    string
		colours bool
	}{
		{"archive.zip", true},
		{"photo.jpg", true},
		{"movie.mp4", true},
		{"song.mp3", true},
		{"lossless.flac", true},
		{"key.gpg", true},
		{"report.pdf", true},
		{"backup.tmp", true},
		{"object.o", true}, // isCompiled: no sibling source, still in compiledExts
		{"README", true},   // isImmediate via name prefix
		{"Makefile", true}, // isImmediate via explicit name
		{"plain.qqz", false},
	}

	for _, c := range cases {
		f := mustFile(t, dir, c.name)
		_, ok := fe.ColourFile(f)
		if ok != c.colours {
			t.Errorf("ColourFile(%s) ok = %v, want %v", c.name, ok, c.colours)
		}
	}
}

func TestFileExtensionsIsTempVariants(t *testing.T) {
	dir := t.TempDir()
	fe := FileExtensions{}

	names := []string{"file~", "#emacs#", "swap.swp"}
	for _, name := range names {
		f := mustFile(t, dir, name)
		if _, ok := fe.ColourFile(f); !ok {
			t.Errorf("expected %s to be classified as temp", name)
		}
	}
}

func TestFileExtensionsIsCompiledViaParent(t *testing.T) {
	// Exercises isCompiled with a real Parent/sibling set up (main.c next
	// to main.o), even though ".o" is classified as compiled purely by
	// extension -- the sibling lookup only matters for extensions outside
	// compiledExts, none of which sourceExtsFor recognises.
	dir := t.TempDir()
	fe := FileExtensions{}

	if err := os.WriteFile(filepath.Join(dir, "main.c"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.o"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := fsx.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	f, err := fsx.NewFile(filepath.Join(dir, "main.o"), d, "")
	if err != nil {
		t.Fatal(err)
	}

	s, ok := fe.ColourFile(f)
	if !ok {
		t.Fatal("expected main.o to be classified as compiled")
	}
	if s == (style.Style{}) {
		t.Error("expected a non-empty style for a compiled file")
	}
}

func TestFileExtensionsIconFile(t *testing.T) {
	dir := t.TempDir()
	fe := FileExtensions{}

	cases := []struct {
		name string
		want rune
		ok   bool
	}{
		{"song.mp3", IconAudio, true},
		{"lossless.flac", IconAudio, true},
		{"photo.jpg", IconImage, true},
		{"movie.mp4", IconVideo, true},
		{"plain.qqz", 0, false},
	}
	for _, c := range cases {
		f := mustFile(t, dir, c.name)
		icon, ok := fe.IconFile(f)
		if ok != c.ok {
			t.Errorf("IconFile(%s) ok = %v, want %v", c.name, ok, c.ok)
			continue
		}
		if ok && icon != c.want {
			t.Errorf("IconFile(%s) = %U, want %U", c.name, icon, c.want)
		}
	}
}

func TestSourceExtsFor(t *testing.T) {
	//nolint:goconst // see the goconst comment on the cases table above
	cases := map[string][]string{
		"o":       {"c", "cc", "cpp", "cxx", "m"},
		"pyc":     {"py"},
		"class":   {"java"},
		"hi":      {"hs"},
		"elc":     {"el"},
		"unknown": nil,
	}
	for ext, want := range cases {
		got := sourceExtsFor(ext)
		if len(got) != len(want) {
			t.Errorf("sourceExtsFor(%q) = %v, want %v", ext, got, want)
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("sourceExtsFor(%q) = %v, want %v", ext, got, want)
				break
			}
		}
	}
}

func TestExtensionMappingsColourFileMostRecentWins(t *testing.T) {
	em := &extensionMappings{}
	if em.nonEmpty() {
		t.Fatal("expected empty mappings to report nonEmpty=false")
	}

	em.add("*.txt", style.Red.Normal())
	em.add("*.txt", style.Blue.Normal()) // added later, should win

	if !em.nonEmpty() {
		t.Fatal("expected non-empty mappings after add")
	}

	f := &fsx.File{Name: "notes.txt"}
	s, ok := em.ColourFile(f)
	if !ok || s != style.Blue.Normal() {
		t.Errorf("ColourFile() = %+v, %v, want Blue, true (most recent wins)", s, ok)
	}

	nomatch := &fsx.File{Name: "notes.bin"}
	if _, ok := em.ColourFile(nomatch); ok {
		t.Error("expected no match for an unrelated extension")
	}
}

func TestChainedColoursFallsBackToSecond(t *testing.T) {
	first := &extensionMappings{}
	first.add("*.special", style.Purple.Normal())
	chained := chainedColours{first: first, second: FileExtensions{}}

	dir := t.TempDir()
	special := mustFile(t, dir, "a.special")
	if s, ok := chained.ColourFile(special); !ok || s != style.Purple.Normal() {
		t.Errorf("expected first colourer to win for a.special, got %+v, %v", s, ok)
	}

	archive := mustFile(t, dir, "a.zip")
	if _, ok := chained.ColourFile(archive); !ok {
		t.Error("expected fallback to FileExtensions for a.zip")
	}

	plain := mustFile(t, dir, "a.qqz")
	if _, ok := chained.ColourFile(plain); ok {
		t.Error("expected no match from either colourer for a.qqz")
	}
}

func TestNoFileColours(t *testing.T) {
	var n noFileColours
	if _, ok := n.ColourFile(&fsx.File{Name: "anything"}); ok {
		t.Error("expected noFileColours to never match")
	}
}
