package fsx

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestStatErrorErrorAndUnwrap(t *testing.T) {
	inner := errors.New("boom")
	se := &StatError{Path: "/nope", Err: inner}

	if se.Error() != "/nope: boom" {
		t.Errorf("Error() = %q, want %q", se.Error(), "/nope: boom")
	}
	if !errors.Is(se, inner) {
		t.Error("expected errors.Is to unwrap to the inner error")
	}
}

func TestNewFileMissingPathReturnsStatError(t *testing.T) {
	_, err := NewFile(filepath.Join(t.TempDir(), "does-not-exist"), nil, "")
	if err == nil {
		t.Fatal("expected an error for a missing path")
	}
	var se *StatError
	if !errors.As(err, &se) {
		t.Errorf("expected a *StatError, got %T", err)
	}
}

func TestExtensionIsOneOf(t *testing.T) {
	f := &File{Ext: "txt"}
	if !f.ExtensionIsOneOf("md", "txt", "go") {
		t.Error("expected txt to match")
	}
	if f.ExtensionIsOneOf("md", "go") {
		t.Error("expected no match")
	}

	noExt := &File{Ext: ""}
	if noExt.ExtensionIsOneOf("txt") {
		t.Error("expected a file with no extension to never match")
	}
}

func TestNameIsOneOf(t *testing.T) {
	f := &File{Name: "Makefile"}
	if !f.NameIsOneOf("Dockerfile", "Makefile") {
		t.Error("expected Makefile to match")
	}
	if f.NameIsOneOf("Dockerfile") {
		t.Error("expected no match")
	}
}

func TestModTime(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	must(t, os.WriteFile(path, nil, 0o644))

	f, err := NewFile(path, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if f.ModTime().IsZero() {
		t.Error("expected a non-zero mod time")
	}
}

func TestLinkTargetRaw(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	must(t, os.WriteFile(target, nil, 0o644))
	link := filepath.Join(dir, "link.txt")
	must(t, os.Symlink(target, link))

	f, err := NewFile(link, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := f.LinkTargetRaw()
	if err != nil {
		t.Fatal(err)
	}
	if raw != target {
		t.Errorf("LinkTargetRaw() = %q, want %q", raw, target)
	}
}

func TestReorientRelativeToParentDir(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "target.txt"), nil, 0o644))

	d, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "link.txt")
	must(t, os.Symlink("target.txt", link))

	f, err := NewFile(link, d, "")
	if err != nil {
		t.Fatal(err)
	}
	target, ok := f.LinkTarget()
	if !ok {
		t.Fatal("expected the relative symlink to resolve via its parent dir")
	}
	if target.Name != "target.txt" {
		t.Errorf("resolved name = %q, want target.txt", target.Name)
	}
}

func TestDirContainsAndJoin(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644))

	d, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Contains(filepath.Join(dir, "a.txt")) {
		t.Error("expected Contains to find a.txt")
	}
	if d.Contains(filepath.Join(dir, "nope.txt")) {
		t.Error("expected Contains to report false for a missing entry")
	}
	if got := d.Join("child"); got != filepath.Join(dir, "child") {
		t.Errorf("Join(child) = %q, want %q", got, filepath.Join(dir, "child"))
	}
}
