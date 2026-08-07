package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jc21/lsgo/internal/cli"
)

func runApp(t *testing.T, dir string, args ...string) (stdout, stderr string) {
	t.Helper()

	noEnv := func(string) (string, bool) { return "", false }
	noProbe := func() (int, bool) { return 0, false }

	res, err := cli.ParseArgs(args, noEnv, noProbe)
	if err != nil {
		t.Fatalf("ParseArgs(%v): %v", args, err)
	}
	if res.Options == nil {
		t.Fatalf("ParseArgs(%v) produced no options (help/version requested?)", args)
	}
	res.Options.ColorMode = 2 // theme.ColourNever, so output is stable/plain

	var outBuf, errBuf bytes.Buffer
	a := New(res.Options, false, "", &outBuf, &errBuf)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
	}

	a.Run()
	return outBuf.String(), errBuf.String()
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAppListsPlainDirectory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "b.txt"), "")
	writeFile(t, filepath.Join(dir, "a.txt"), "")

	out, _ := runApp(t, dir)
	if out != "a.txt\nb.txt\n" {
		t.Errorf("got %q", out)
	}
}

func TestAppHidesDotfilesByDefault(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "visible.txt"), "")
	writeFile(t, filepath.Join(dir, ".hidden"), "")

	out, _ := runApp(t, dir)
	if strings.Contains(out, ".hidden") {
		t.Errorf("expected .hidden to be hidden, got %q", out)
	}
	if !strings.Contains(out, "visible.txt") {
		t.Errorf("expected visible.txt to be shown, got %q", out)
	}
}

func TestAppLongViewShowsSize(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "hello.txt"), "hello world")

	out, _ := runApp(t, dir, "-l")
	if !strings.Contains(out, "hello.txt") {
		t.Errorf("expected file name in output, got %q", out)
	}
	if !strings.Contains(out, "11") {
		t.Errorf("expected byte size 11 in output, got %q", out)
	}
}

func TestAppReportsStatErrorForMissingPath(t *testing.T) {
	dir := t.TempDir()
	_, errOut := runApp(t, dir, "does-not-exist")
	if !strings.Contains(errOut, "does-not-exist") {
		t.Errorf("expected error to mention the missing path, got %q", errOut)
	}
}

func TestAppOnlyDirsFilter(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "file.txt"), "")
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, _ := runApp(t, dir, "-D")
	if strings.Contains(out, "file.txt") {
		t.Errorf("expected file.txt to be excluded by --only-dirs, got %q", out)
	}
	if !strings.Contains(out, "subdir") {
		t.Errorf("expected subdir to be shown, got %q", out)
	}
}

func TestAppRecurseFlatPrintsHeaders(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "sub", "nested.txt"), "")
	writeFile(t, filepath.Join(dir, "top.txt"), "")

	out, _ := runApp(t, dir, "-R")
	if !strings.Contains(out, "top.txt") {
		t.Errorf("expected top-level file, got %q", out)
	}
	if !strings.Contains(out, "nested.txt") {
		t.Errorf("expected nested file via recursion, got %q", out)
	}
	if !strings.Contains(out, "sub:") {
		t.Errorf("expected a 'sub:' section header, got %q", out)
	}
}

func TestAppTreeView(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "sub", "nested.txt"), "")

	out, _ := runApp(t, dir, "-T")
	if !strings.Contains(out, "└── nested.txt") && !strings.Contains(out, "├── nested.txt") {
		t.Errorf("expected a tree branch before nested.txt, got %q", out)
	}
}

func TestAppSortReverse(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "")
	writeFile(t, filepath.Join(dir, "b.txt"), "")

	out, _ := runApp(t, dir, "-r")
	if out != "b.txt\na.txt\n" {
		t.Errorf("got %q", out)
	}
}

func TestAppIgnoreGlob(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "keep.txt"), "")
	writeFile(t, filepath.Join(dir, "skip.log"), "")

	out, _ := runApp(t, dir, "-I", "*.log")
	if strings.Contains(out, "skip.log") {
		t.Errorf("expected skip.log to be ignored, got %q", out)
	}
	if !strings.Contains(out, "keep.txt") {
		t.Errorf("expected keep.txt to be shown, got %q", out)
	}
}

func TestAppListDirsAsFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, _ := runApp(t, filepath.Dir(dir), "-d", filepath.Base(dir))
	if !strings.Contains(out, filepath.Base(dir)) {
		t.Errorf("expected the directory itself to be listed as an entry, got %q", out)
	}
}
