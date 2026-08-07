package output

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jc21/lsgo/internal/fsx"
	"github.com/jc21/lsgo/internal/xattr"
)

func TestIsTooDeep(t *testing.T) {
	one := 1
	r := RecurseOptions{MaxDepth: &one}
	if r.IsTooDeep(1) {
		t.Error("depth 1 should not be too deep for MaxDepth 1")
	}
	if !r.IsTooDeep(2) {
		t.Error("depth 2 should be too deep for MaxDepth 1")
	}

	unlimited := RecurseOptions{}
	if unlimited.IsTooDeep(1000) {
		t.Error("nil MaxDepth should never be too deep")
	}
}

func TestHeaderNameCell(t *testing.T) {
	th := colourfulTheme()
	cell := headerNameCell(th)
	if !strings.Contains(cell.String(), "Name") || cell.Width != len("Name") {
		t.Errorf("headerNameCell() = %q, want to contain Name (width 4)", cell.String())
	}
}

func filesInDir(t *testing.T, dir string) []*fsx.File {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	files := make([]*fsx.File, len(entries))
	for i, e := range entries {
		f, err := fsx.NewFile(filepath.Join(dir, e.Name()), nil, "")
		if err != nil {
			t.Fatal(err)
		}
		files[i] = f
	}
	return files
}

func TestDetailsRendererPlainNoTable(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), nil, 0o644)

	r := &DetailsRenderer{Theme: th}
	var buf bytes.Buffer
	if err := r.Render(&buf, filesInDir(t, dir)); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "b.txt") {
		t.Errorf("expected both filenames in tree-only output, got %q", out)
	}
}

func TestDetailsRendererWithTableAndHeader(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi"), 0o644)

	opts := fullTableOptions()
	r := &DetailsRenderer{
		Theme:  th,
		Filter: fsx.FileFilter{},
		Options: DetailsOptions{
			Table:  &opts,
			Header: true,
		},
	}
	var buf bytes.Buffer
	if err := r.Render(&buf, filesInDir(t, dir)); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Name") {
		t.Errorf("expected header row containing 'Name', got %q", out)
	}
	if !strings.Contains(out, "a.txt") {
		t.Errorf("expected a.txt in output, got %q", out)
	}
}

func TestDetailsRendererTreeRecursion(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "nested.txt"), nil, 0o644)
	os.WriteFile(filepath.Join(dir, "top.txt"), nil, 0o644)

	r := &DetailsRenderer{
		Theme:   th,
		Filter:  fsx.FileFilter{},
		Recurse: &RecurseOptions{Tree: true},
	}
	var buf bytes.Buffer
	if err := r.Render(&buf, filesInDir(t, dir)); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "top.txt") || !strings.Contains(out, "nested.txt") {
		t.Errorf("expected both top-level and nested files, got %q", out)
	}
	if !strings.Contains(out, "──") {
		t.Errorf("expected tree branch characters, got %q", out)
	}
}

func TestDetailsRendererTreeMaxDepthStopsRecursion(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "nested.txt"), nil, 0o644)

	zero := 0
	r := &DetailsRenderer{
		Theme:   th,
		Filter:  fsx.FileFilter{},
		Recurse: &RecurseOptions{Tree: true, MaxDepth: &zero},
	}
	var buf bytes.Buffer
	if err := r.Render(&buf, filesInDir(t, dir)); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "nested.txt") {
		t.Errorf("expected recursion to stop at depth 0, got %q", buf.String())
	}
}

func TestDetailsRendererShowXattrs(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644)

	r := &DetailsRenderer{
		Theme:   th,
		Filter:  fsx.FileFilter{},
		Options: DetailsOptions{ShowXattrs: true},
	}
	var buf bytes.Buffer
	// Mostly exercised for coverage of the ShowXattrs branch in addFiles;
	// files with no extended attributes just produce a normal row.
	if err := r.Render(&buf, filesInDir(t, dir)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "a.txt") {
		t.Errorf("expected a.txt in output, got %q", buf.String())
	}
}

func TestXattrCell(t *testing.T) {
	th := colourfulTheme()
	cell := xattrCell(th, xattr.Attribute{Name: "user.test", Size: 5})
	if !strings.Contains(cell.String(), "user.test") || !strings.Contains(cell.String(), "5") {
		t.Errorf("expected attribute name and size in cell, got %q", cell.String())
	}
}
