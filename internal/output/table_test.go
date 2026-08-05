package output

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"lsgo/internal/fsx"
)

func fullTableOptions() TableOptions {
	return TableOptions{
		SizeFormat: DecimalSize,
		TimeFormat: DefaultTimeFormat,
		UserFormat: UserName,
		Columns: TableColumns{
			TimeTypes:   TimeTypes{Modified: true, Changed: true, Accessed: true, Created: true},
			Inode:       true,
			Links:       true,
			Blocks:      true,
			Group:       true,
			Git:         true,
			Octal:       true,
			Permissions: true,
			Filesize:    true,
			User:        true,
		},
	}
}

func TestDefaultTableColumnsAndCollect(t *testing.T) {
	cols := DefaultTableColumns().Collect(false)
	if len(cols) == 0 {
		t.Fatal("expected at least one default column")
	}

	full := fullTableOptions().Columns.Collect(true)
	if len(full) != 13 {
		t.Errorf("expected 13 columns with everything enabled, got %d", len(full))
	}

	// Git column is gated on gitEnabled even when TableColumns.Git is set.
	withoutGit := fullTableOptions().Columns.Collect(false)
	if len(withoutGit) != len(full)-1 {
		t.Error("expected git column to be dropped when gitEnabled=false")
	}
}

func TestColumnHeaderAndAlign(t *testing.T) {
	cases := []struct {
		col       Column
		header    string
		alignLeft bool
	}{
		{Column{Kind: ColInode}, "inode", false},
		{Column{Kind: ColOctal}, "Octal", true},
		{Column{Kind: ColPermissions}, "Permissions", true},
		{Column{Kind: ColHardLinks}, "Links", false},
		{Column{Kind: ColFileSize}, "Size", false},
		{Column{Kind: ColBlocks}, "Blocks", false},
		{Column{Kind: ColUser}, "User", true},
		{Column{Kind: ColGroup}, "Group", true},
		{Column{Kind: ColTimestamp, Time: TimeModified}, "Date Modified", true},
		{Column{Kind: ColTimestamp, Time: TimeChanged}, "Date Changed", true},
		{Column{Kind: ColTimestamp, Time: TimeAccessed}, "Date Accessed", true},
		{Column{Kind: ColTimestamp, Time: TimeCreated}, "Date Created", true},
		{Column{Kind: ColGitStatus}, "Git", false},
		{Column{Kind: ColumnKind(999)}, "", true},
	}
	for _, c := range cases {
		if got := c.col.Header(); got != c.header {
			t.Errorf("Header() for %+v = %q, want %q", c.col, got, c.header)
		}
		if got := c.col.alignLeft(); got != c.alignLeft {
			t.Errorf("alignLeft() for %+v = %v, want %v", c.col, got, c.alignLeft)
		}
	}
}

// setupTableFixture creates a small directory tree exercising most file
// kinds and permission-bit combinations that the long-view table renders.
func setupTableFixture(t *testing.T) (files []*fsx.File) {
	t.Helper()
	dir := t.TempDir()

	// The special bits (setuid/setgid/sticky) are applied via a separate
	// chmod, rather than the file-creation mode, since some kernels
	// silently drop them at creation time regardless of what's passed to
	// open(2).
	write := func(name string, mode os.FileMode) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("hello world"), mode&os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
		return path
	}

	paths := make([]string, 0, 8)
	paths = append(paths,
		write("plain.txt", 0o644),
		write("exe.sh", 0o755),
		write("setuid", os.ModeSetuid|0o644),
		write("setgid", os.ModeSetgid|0o644),
		write("sticky", os.ModeSticky|0o644),
		write("allbits", os.ModeSetuid|os.ModeSetgid|os.ModeSticky|0o777),
	)

	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	paths = append(paths, subdir)

	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(paths[0], link); err != nil {
		t.Fatal(err)
	}
	paths = append(paths, link)

	files = make([]*fsx.File, len(paths))
	for i, p := range paths {
		f, err := fsx.NewFile(p, nil, "")
		if err != nil {
			t.Fatal(err)
		}
		files[i] = f
	}
	return files
}

func TestTableRowForFileAllColumns(t *testing.T) {
	th := colourfulTheme()
	files := setupTableFixture(t)

	table := NewTable(th, nil, fullTableOptions())
	if got := len(table.Columns()); got == 0 {
		t.Fatal("expected non-empty column set")
	}

	header := table.HeaderRow()
	if len(header) != len(table.Columns()) {
		t.Fatalf("header row length = %d, want %d", len(header), len(table.Columns()))
	}

	for _, f := range files {
		row := table.RowForFile(f, false)
		if len(row) != len(table.Columns()) {
			t.Fatalf("row length = %d, want %d for %s", len(row), len(table.Columns()), f.Name)
		}
		table.AddWidths(row)
	}

	// Rendering shouldn't panic and should produce non-empty, padded text
	// for a regular file's row.
	regular, err := fsx.NewFile(files[0].Path, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	row := table.RowForFile(regular, true) // hasXattrs=true to hit the '@' branch
	rendered := table.Render(row)
	if rendered.Width == 0 {
		t.Error("expected non-empty rendered row")
	}
}

func TestTableRenderPermissionsSpecialBits(t *testing.T) {
	th := colourfulTheme()
	files := setupTableFixture(t)
	table := NewTable(th, nil, fullTableOptions())

	// Sanity: every fixture file should produce a permissions cell without
	// panicking, covering user/group/other execute-bit branches including
	// setuid/setgid/sticky combinations.
	for _, f := range files {
		cell := table.RowForFile(f, false)
		if len(cell) == 0 {
			t.Fatalf("expected columns for %s", f.Name)
		}
	}
}

func TestTableRenderSizeVariants(t *testing.T) {
	th := colourfulTheme()
	files := setupTableFixture(t)
	table := NewTable(th, nil, fullTableOptions())

	for _, f := range files {
		cell := table.renderSize(&th.UI, f)
		if f.IsDirectory() {
			if cell.String() == "" {
				t.Error("expected placeholder size text for directory")
			}
			continue
		}
		if cell.Width == 0 {
			t.Errorf("expected non-empty size cell for %s", f.Name)
		}
	}
}

func TestTableRenderUserAndGroup(t *testing.T) {
	th := colourfulTheme()
	files := setupTableFixture(t)
	table := NewTable(th, nil, fullTableOptions())

	me, err := user.Current()
	if err != nil {
		t.Skip("no current user available")
	}
	uid, _ := strconv.ParseUint(me.Uid, 10, 32)

	f := files[0]
	userCell := table.renderUser(&th.UI, uint32(uid))
	if !strings.Contains(userCell.String(), me.Username) && !strings.Contains(userCell.String(), me.Uid) {
		t.Errorf("expected current user's name or uid in cell, got %q", userCell.String())
	}

	groupCell := table.renderGroup(&th.UI, f.GID())
	if groupCell.String() == "" {
		t.Error("expected non-empty group cell")
	}

	// Numeric format forces IDs even when a name is resolvable.
	numeric := fullTableOptions()
	numeric.UserFormat = UserNumeric
	table2 := NewTable(th, nil, numeric)
	numericCell := table2.renderUser(&th.UI, uint32(uid))
	if !strings.Contains(numericCell.String(), strconv.FormatUint(uid, 10)) {
		t.Errorf("expected numeric uid, got %q", numericCell.String())
	}

	// An unresolvable uid falls back to the numeric form even in name mode.
	unresolvable := table.renderUser(&th.UI, 0xFFFFFFFE)
	if !strings.Contains(unresolvable.String(), strconv.FormatUint(0xFFFFFFFE, 10)) {
		t.Errorf("expected numeric fallback for unknown uid, got %q", unresolvable.String())
	}
	unresolvableGroup := table.renderGroup(&th.UI, 0xFFFFFFFE)
	if !strings.Contains(unresolvableGroup.String(), strconv.FormatUint(0xFFFFFFFE, 10)) {
		t.Errorf("expected numeric fallback for unknown gid, got %q", unresolvableGroup.String())
	}
}

func TestTableRenderTimestampVariants(t *testing.T) {
	th := colourfulTheme()
	files := setupTableFixture(t)
	table := NewTable(th, nil, fullTableOptions())
	f := files[0]

	for _, tt := range []TimeType{TimeModified, TimeChanged, TimeAccessed, TimeCreated} {
		cell := table.renderTimestamp(&th.UI, f, tt)
		if cell.String() == "" {
			t.Errorf("expected non-empty timestamp cell for TimeType %d", tt)
		}
	}
}

func TestTableRenderGitStatusWithoutRepo(t *testing.T) {
	th := colourfulTheme()
	files := setupTableFixture(t)

	table := NewTable(th, nil, fullTableOptions())
	cell := table.renderGitStatus(&th.UI, files[0])
	if !strings.Contains(cell.String(), "-") {
		t.Errorf("expected placeholder git status without a cache, got %q", cell.String())
	}
}

func TestTableRenderGitStatusWithRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	th := colourfulTheme()
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	untracked := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(untracked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	git := fsx.NewGitCache()
	git.Prime([]string{dir})
	if !git.HasAnythingFor(dir) {
		t.Skip("git repository wasn't detected; skipping")
	}

	table := NewTable(th, git, fullTableOptions())
	f, err := fsx.NewFile(untracked, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	cell := table.renderGitStatus(&th.UI, f)
	if !strings.Contains(cell.String(), "N") {
		t.Errorf("expected 'N' (new) status for untracked file, got %q", cell.String())
	}
}

func TestGitStatusCharAndStyle(t *testing.T) {
	th := colourfulTheme()
	statuses := []fsx.GitFileStatus{
		fsx.GitNotModified, fsx.GitNew, fsx.GitModified, fsx.GitDeleted,
		fsx.GitRenamed, fsx.GitTypeChange, fsx.GitIgnored, fsx.GitConflicted,
	}
	seen := map[string]bool{}
	for _, s := range statuses {
		seen[gitStatusChar(s)] = true
		_ = gitStatusStyle(&th.UI, s) // exercise every branch without panicking
	}
	if len(seen) != len(statuses) {
		t.Errorf("expected a distinct character per status, got %v", seen)
	}
}

func TestNewTableUsesCurrentIdentity(t *testing.T) {
	th := colourfulTheme()
	table := NewTable(th, nil, fullTableOptions())
	if table.me.groupIDs == nil {
		t.Error("expected current identity's group set to be initialised")
	}
}
