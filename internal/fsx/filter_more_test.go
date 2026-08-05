package fsx

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultSortSpec(t *testing.T) {
	spec := DefaultSortSpec()
	if spec.Field != SortByName || spec.Case != CaseInsensitive {
		t.Errorf("DefaultSortSpec() = %+v, want name/case-insensitive", spec)
	}
}

// sortFixture builds two real files with distinct sizes, mtimes, and
// inodes, so every metadata-based SortField has something meaningful to
// compare.
func sortFixture(t *testing.T) (a, b *File) {
	t.Helper()
	dir := t.TempDir()

	smallPath := filepath.Join(dir, "small.txt")
	must(t, os.WriteFile(smallPath, []byte("x"), 0o644))
	bigPath := filepath.Join(dir, "big.txt")
	must(t, os.WriteFile(bigPath, []byte("xxxxxxxxxx"), 0o644))

	older := time.Now().Add(-time.Hour)
	must(t, os.Chtimes(smallPath, older, older))

	a, err := NewFile(smallPath, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err = NewFile(bigPath, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	return a, b
}

func TestSortSpecCompareEveryField(t *testing.T) {
	small, big := sortFixture(t)

	cases := []struct {
		name  string
		field SortField
	}{
		{"size", SortBySize},
		{"inode", SortByInode},
		{"modified", SortByModified},
		{"modifiedAge", SortByModifiedAge},
		{"accessed", SortByAccessed},
		{"changed", SortByChanged},
		{"created", SortByCreated},
		{"type", SortByType},
		{"unsorted", SortUnsorted},
	}
	for _, c := range cases {
		spec := SortSpec{Field: c.field}
		// Just exercise every branch without panicking; only a handful
		// have a guaranteed direction (size, modified/modifiedAge).
		got := spec.Compare(small, big)
		switch c.field {
		case SortBySize:
			if got >= 0 {
				t.Errorf("%s: expected small.txt to sort before big.txt, got %d", c.name, got)
			}
		case SortByModified:
			if got >= 0 {
				t.Errorf("%s: expected the older file to sort first, got %d", c.name, got)
			}
		case SortByModifiedAge:
			if got <= 0 {
				t.Errorf("%s: expected age order to reverse modified order, got %d", c.name, got)
			}
		case SortUnsorted:
			if got != 0 {
				t.Errorf("%s: expected 0, got %d", c.name, got)
			}
		}
	}
}

func TestCompareHelpersTieBreak(t *testing.T) {
	if compareInt64(5, 5) != 0 || compareInt64(1, 2) != -1 || compareInt64(2, 1) != 1 {
		t.Error("compareInt64 ordering incorrect")
	}
	if compareUint64(5, 5) != 0 || compareUint64(1, 2) != -1 || compareUint64(2, 1) != 1 {
		t.Error("compareUint64 ordering incorrect")
	}
	now := time.Now()
	later := now.Add(time.Second)
	if compareTime(now, now) != 0 || compareTime(now, later) != -1 || compareTime(later, now) != 1 {
		t.Error("compareTime ordering incorrect")
	}
}

func TestFilterChildFiles(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "keep.txt"), nil, 0o644))
	must(t, os.WriteFile(filepath.Join(dir, "skip.log"), nil, 0o644))
	must(t, os.Mkdir(filepath.Join(dir, "subdir"), 0o755))

	d, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	// FilterChildFiles compacts its input slice in place, so each
	// sub-case needs its own freshly-read slice rather than reusing one
	// that's already been filtered.
	files, _ := d.Files(JustFiles, nil, false)
	ff := FileFilter{IgnorePatterns: NewIgnorePatterns([]string{"*.log"})}
	filtered := ff.FilterChildFiles(files)
	for _, f := range filtered {
		if f.Name == "skip.log" {
			t.Error("expected skip.log to be filtered out")
		}
	}

	files2, _ := d.Files(JustFiles, nil, false)
	onlyDirs := FileFilter{OnlyDirs: true}
	dirsOnly := onlyDirs.FilterChildFiles(files2)
	for _, f := range dirsOnly {
		if !f.IsDirectory() {
			t.Errorf("expected only directories, got %s", f.Name)
		}
	}
	if len(dirsOnly) != 1 {
		t.Errorf("expected exactly 1 directory, got %d", len(dirsOnly))
	}
}

func TestFilterArgumentFilesNeverAppliesOnlyDirs(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "keep.txt"), nil, 0o644))
	must(t, os.WriteFile(filepath.Join(dir, "skip.log"), nil, 0o644))

	d, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	files, _ := d.Files(JustFiles, nil, false)

	ff := FileFilter{IgnorePatterns: NewIgnorePatterns([]string{"*.log"}), OnlyDirs: true}
	filtered := ff.FilterArgumentFiles(files)
	if len(filtered) != 1 || filtered[0].Name != "keep.txt" {
		t.Errorf("expected only keep.txt (OnlyDirs ignored), got %v", names(filtered))
	}
}

func TestSortFilesReverseAndDirsFirst(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "b.txt"), nil, 0o644))
	must(t, os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644))
	must(t, os.Mkdir(filepath.Join(dir, "zdir"), 0o755))

	d, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	files, _ := d.Files(JustFiles, nil, false)

	ff := FileFilter{Sort: SortSpec{Field: SortByName, Case: CaseInsensitive}, ListDirsFirst: true, Reverse: true}
	ff.SortFiles(files)

	if !files[0].IsDirectory() {
		t.Errorf("expected directories first, got %v", names(files))
	}
	// Within the (non-directory) files, reverse name order puts b before a.
	var nonDirs []string
	for _, f := range files {
		if !f.IsDirectory() {
			nonDirs = append(nonDirs, f.Name)
		}
	}
	if len(nonDirs) != 2 || nonDirs[0] != "b.txt" || nonDirs[1] != "a.txt" {
		t.Errorf("expected reversed [b.txt a.txt], got %v", nonDirs)
	}
}

func TestReverseFiles(t *testing.T) {
	files := []*File{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	reverseFiles(files)
	if names(files)[0] != "c" || names(files)[2] != "a" {
		t.Errorf("expected reversed order, got %v", names(files))
	}

	// Even-length and empty slices shouldn't panic.
	reverseFiles([]*File{{Name: "x"}, {Name: "y"}})
	reverseFiles(nil)
}
