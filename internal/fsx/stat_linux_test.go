package fsx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStatLinuxFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	must(t, os.WriteFile(path, []byte("hello world"), 0o644))

	f, err := NewFile(path, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	if f.Inode() == 0 {
		t.Error("expected a non-zero inode for a real file")
	}
	if f.LinkCount() == 0 {
		t.Error("expected at least one hard link")
	}
	if f.UID() == 0 && os.Geteuid() != 0 {
		t.Error("expected non-root UID to be reported for a file we created")
	}

	if blocks, ok := f.Blocks(); !ok || blocks < 0 {
		t.Errorf("Blocks() = %d, %v; want a valid non-negative count", blocks, ok)
	}

	if f.AccessedTime().IsZero() {
		t.Error("expected a non-zero accessed time")
	}
	if f.ChangedTime().IsZero() {
		t.Error("expected a non-zero changed time")
	}
	if _, ok := f.CreatedTime(); ok {
		t.Error("Linux stat(2) has no birth time; CreatedTime should report false")
	}
}

func TestBlocksFalseForDirectory(t *testing.T) {
	dir := t.TempDir()
	f, err := NewFile(dir, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := f.Blocks(); ok {
		t.Error("expected Blocks() to report false for a directory")
	}
}

func TestDeviceIDsForNonDevice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	must(t, os.WriteFile(path, nil, 0o644))

	f, err := NewFile(path, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	major, minor := f.DeviceIDs()
	if major != 0 || minor != 0 {
		t.Errorf("DeviceIDs() for a regular file = %d,%d, want 0,0", major, minor)
	}
}
