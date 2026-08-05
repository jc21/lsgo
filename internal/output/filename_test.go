package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lsgo/internal/fsx"
	"lsgo/internal/theme"
)

func colourfulTheme() *theme.Theme {
	th := theme.Build(theme.ColourAlways, theme.ScaleFixed, true, "")
	return &th
}

func mustFile(t *testing.T, path string) *fsx.File {
	t.Helper()
	f, err := fsx.NewFile(path, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFileNameStyleByType(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()

	regular := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(regular, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(exe, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(regular, link); err != nil {
		t.Fatal(err)
	}

	regularFile := mustFile(t, regular)
	if got := FileNameStyle(th, regularFile); got != th.UI.FileKinds.Normal {
		t.Errorf("regular file style = %+v, want Normal", got)
	}

	exeFile := mustFile(t, exe)
	if got := FileNameStyle(th, exeFile); got != th.UI.FileKinds.Executable {
		t.Errorf("executable file style = %+v, want Executable", got)
	}

	dirFile := mustFile(t, sub)
	if got := FileNameStyle(th, dirFile); got != th.UI.FileKinds.Directory {
		t.Errorf("directory style = %+v, want Directory", got)
	}

	linkFile := mustFile(t, link)
	if got := FileNameStyle(th, linkFile); got != th.UI.FileKinds.Symlink {
		t.Errorf("symlink style = %+v, want Symlink", got)
	}
}

func TestClassifyChar(t *testing.T) {
	dir := t.TempDir()

	regular := filepath.Join(dir, "plain.txt")
	os.WriteFile(regular, nil, 0o644)
	exe := filepath.Join(dir, "run.sh")
	os.WriteFile(exe, nil, 0o755)
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0o755)
	link := filepath.Join(dir, "link")
	os.Symlink(regular, link)

	cases := map[string]string{
		regular: "",
		exe:     "*",
		sub:     "/",
		link:    "@",
	}
	for path, want := range cases {
		f := mustFile(t, path)
		if got := classifyChar(f); got != want {
			t.Errorf("classifyChar(%s) = %q, want %q", path, got, want)
		}
	}
}

func TestRenderFileNamePlain(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	cell := RenderFileName(th, FileNameOptions{}, mustFile(t, path), false)
	if !strings.Contains(cell.String(), "hello.txt") {
		t.Errorf("expected rendered name to contain hello.txt, got %q", cell.String())
	}
}

func TestRenderFileNameWithIconAndClassify(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	path := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(path, nil, 0o755); err != nil {
		t.Fatal(err)
	}

	opts := FileNameOptions{
		Classify:  AddFileIndicators,
		ShowIcons: ShowIcons{On: true, Spaces: 2},
	}
	cell := RenderFileName(th, opts, mustFile(t, path), false)
	got := cell.String()
	if !strings.Contains(got, "run.sh") {
		t.Errorf("expected name in output, got %q", got)
	}
	if !strings.HasSuffix(got, "*") {
		t.Errorf("expected trailing '*' classify indicator, got %q", got)
	}
}

func TestRenderFileNameWorkingSymlinkShowsTarget(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	cell := RenderFileName(th, FileNameOptions{Classify: AddFileIndicators}, mustFile(t, link), true)
	got := cell.String()
	if !strings.Contains(got, "link.txt") || !strings.Contains(got, "target.txt") || !strings.Contains(got, "->") {
		t.Errorf("expected link name, arrow, and target, got %q", got)
	}
}

func TestRenderFileNameBrokenSymlinkShowsBrokenArrow(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	link := filepath.Join(dir, "broken.txt")
	if err := os.Symlink(filepath.Join(dir, "does-not-exist"), link); err != nil {
		t.Fatal(err)
	}

	cell := RenderFileName(th, FileNameOptions{}, mustFile(t, link), true)
	got := cell.String()
	if !strings.Contains(got, "broken.txt") || !strings.Contains(got, "does-not-exist") {
		t.Errorf("expected broken link name and raw target, got %q", got)
	}

	// Without withLinkPath, a broken symlink's *name* carries the broken
	// colour instead.
	plain := RenderFileName(th, FileNameOptions{}, mustFile(t, link), false)
	if !strings.Contains(plain.String(), "broken.txt") {
		t.Errorf("expected name in plain rendering, got %q", plain.String())
	}
}

func TestRenderFileNameUnreadableLinkLeavesNameAsIs(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	// A file that isn't actually a symlink: LinkTargetDetailed's
	// os.Readlink call fails, exercising renderLinkTarget's err != nil path
	// via brokenAwareStyle/FileNameStyle instead (not a link at all).
	path := filepath.Join(dir, "notalink.txt")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cell := RenderFileName(th, FileNameOptions{}, mustFile(t, path), true)
	if !strings.Contains(cell.String(), "notalink.txt") {
		t.Errorf("expected plain file name, got %q", cell.String())
	}
}

func TestIconifyStyle(t *testing.T) {
	th := colourfulTheme()

	// Directory style (Blue.Bold()) has a foreground colour but no
	// background; iconify should keep the colour but drop Bold.
	fg := iconifyStyle(th.UI.FileKinds.Directory)
	if fg.Foreground != th.UI.FileKinds.Directory.Foreground {
		t.Errorf("expected iconify to keep the foreground colour, got %+v", fg)
	}
	if fg.Bold {
		t.Errorf("expected iconify to drop Bold, got %+v", fg)
	}

	withBg := th.UI.Links.MultiLinkFile // Red.On(Yellow): has both fg and bg
	got := iconifyStyle(withBg)
	if got.Foreground != withBg.Background {
		t.Errorf("expected iconify to prefer the background colour, got %+v", got)
	}

	// A style with neither foreground nor background produces an
	// entirely empty style.
	if got := iconifyStyle(theme.Plain().FileKinds.Normal); got != (theme.Plain().FileKinds.Normal) {
		t.Errorf("expected empty style for a style with no colours, got %+v", got)
	}
}

func TestRenderOneLine(t *testing.T) {
	th := colourfulTheme()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), nil, 0o644)

	files := []*fsx.File{
		mustFile(t, filepath.Join(dir, "a.txt")),
		mustFile(t, filepath.Join(dir, "b.txt")),
	}
	out := RenderOneLine(th, FileNameOptions{}, files)
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "b.txt") {
		t.Errorf("expected both names in output, got %q", out)
	}
}
