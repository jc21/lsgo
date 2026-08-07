package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jc21/lsgo/internal/fsx"
)

func TestIconForFile(t *testing.T) {
	dir := t.TempDir()

	byName := mustFile(t, dir, "Makefile")
	if got := IconForFile(byName, nil); got != 0xf489 {
		t.Errorf("IconForFile(Makefile) = %U, want %U", got, rune(0xf489))
	}

	// ".idea" is only in directoryIconsByName (not iconsByName), so this
	// specifically exercises the directory-special-case branch.
	sub := filepath.Join(dir, ".idea")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	dirFile, err := fsx.NewFile(sub, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := IconForFile(dirFile, nil); got != 0xe7b5 {
		t.Errorf("IconForFile(.idea) = %U, want %U", got, rune(0xe7b5))
	}

	plainDir := filepath.Join(dir, "some-random-dir")
	if err := os.Mkdir(plainDir, 0o755); err != nil {
		t.Fatal(err)
	}
	plainDirFile, err := fsx.NewFile(plainDir, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := IconForFile(plainDirFile, nil); got != IconFolder {
		t.Errorf("IconForFile(random dir) = %U, want IconFolder", got)
	}

	byExt := mustFile(t, dir, "photo.jpg")
	if got := IconForFile(byExt, nil); got == IconFallback {
		t.Error("expected a specific icon for a known extension")
	}

	unknownExt := mustFile(t, dir, "file.qqz")
	if got := IconForFile(unknownExt, nil); got != IconFallback {
		t.Errorf("IconForFile(unknown ext) = %U, want IconFallback", got)
	}

	noExt := mustFile(t, dir, "noextension")
	if got := IconForFile(noExt, nil); got != IconFile {
		t.Errorf("IconForFile(no ext) = %U, want IconFile", got)
	}

	// A non-nil iconer takes priority over the extension map.
	withIconer := mustFile(t, dir, "song.mp3")
	if got := IconForFile(withIconer, FileExtensions{}); got != IconAudio {
		t.Errorf("IconForFile with iconer = %U, want IconAudio", got)
	}
}
