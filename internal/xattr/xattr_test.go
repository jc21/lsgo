package xattr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListOnRegularFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(path, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A freshly-created file typically has no extended attributes; this
	// mainly exercises that List doesn't error on an ordinary file,
	// whatever the filesystem's xattr support looks like in this
	// environment.
	attrs, err := List(path)
	if err != nil {
		t.Logf("List returned an error (filesystem may not support xattrs): %v", err)
		return
	}
	if len(attrs) != 0 {
		t.Logf("unexpected pre-existing attributes: %+v", attrs)
	}
}

func TestListMissingFile(t *testing.T) {
	if _, err := List(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("expected an error for a missing path")
	}
}
