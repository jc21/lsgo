package fsx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtensionOf(t *testing.T) {
	cases := map[string]string{
		"fester.dat": "dat",
		".vimrc":     "vimrc",
		"jarlsberg":  "",
		"a.TAR.GZ":   "gz",
	}
	for name, want := range cases {
		if got := extensionOf(name); got != want {
			t.Errorf("extensionOf(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestNewFileAndPredicates(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(filePath, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := NewFile(filePath, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	if f.Name != "hello.txt" {
		t.Errorf("Name = %q, want hello.txt", f.Name)
	}
	if f.Ext != "txt" {
		t.Errorf("Ext = %q, want txt", f.Ext)
	}
	if f.IsDirectory() {
		t.Error("regular file reported as directory")
	}
	if !f.IsRegularFile() {
		t.Error("expected regular file")
	}
	if f.IsExecutableFile() {
		t.Error("file written with 0644 should not be executable")
	}
}

func TestSymlinkResolution(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	f, err := NewFile(link, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if !f.IsLink() {
		t.Fatal("expected symlink to be reported as a link")
	}
	if !f.PointsToDirectory() && f.IsDirectory() {
		t.Error("unexpected directory state")
	}

	resolved, ok := f.LinkTarget()
	if !ok {
		t.Fatal("expected symlink to resolve")
	}
	if resolved.Name != "target.txt" {
		t.Errorf("resolved name = %q, want target.txt", resolved.Name)
	}
}

func TestBrokenSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "broken.txt")
	if err := os.Symlink(filepath.Join(dir, "does-not-exist"), link); err != nil {
		t.Fatal(err)
	}

	f, err := NewFile(link, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	_, _, broken := f.LinkTargetDetailed()
	if !broken {
		t.Error("expected broken symlink to be reported as broken")
	}
}

func TestDirFilesHidesDotfilesByDefault(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "visible.txt"), nil, 0o644))
	must(t, os.WriteFile(filepath.Join(dir, ".hidden"), nil, 0o644))

	d, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	files, errs := d.Files(JustFiles, nil, false)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(files) != 1 || files[0].Name != "visible.txt" {
		t.Errorf("expected only visible.txt, got %+v", names(files))
	}
}

func TestDirFilesShowsDotfilesWithAll(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "visible.txt"), nil, 0o644))
	must(t, os.WriteFile(filepath.Join(dir, ".hidden"), nil, 0o644))

	d, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	files, _ := d.Files(Dotfiles, nil, false)
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %v", names(files))
	}
}

func TestDirFilesShowsDotDotDotWithDoubleAll(t *testing.T) {
	dir := t.TempDir()
	d, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	files, errs := d.Files(DotfilesAndDots, nil, false)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(files) != 2 {
		t.Fatalf("expected just '.' and '..', got %v", names(files))
	}
	if files[0].Name != "." || files[1].Name != ".." {
		t.Errorf("expected '.' then '..', got %v", names(files))
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func names(files []*File) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Name
	}
	return out
}
