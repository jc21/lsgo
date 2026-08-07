package cli

import (
	"testing"

	"github.com/jc21/lsgo/internal/fsx"
	"github.com/jc21/lsgo/internal/output"
)

// The four view-mode flags, named since TestModeSelectionPrecedence's table
// (and a couple of parser tests) combine them repeatedly.
const (
	flagOneline = "--oneline"
	flagLong    = "--long"
	flagGrid    = "--grid"
	flagTree    = "--tree"
)

func noEnv(string) (string, bool)    { return "", false }
func noProbe() (int, bool)           { return 0, false }
func fixedProbe(w int) TerminalProbe { return func() (int, bool) { return w, true } }
func envMap(m map[string]string) Env {
	return func(name string) (string, bool) {
		v, ok := m[name]
		return v, ok
	}
}

func parse(t *testing.T, args ...string) *Options {
	t.Helper()
	res, err := ParseArgs(args, noEnv, noProbe)
	if err != nil {
		t.Fatalf("ParseArgs(%v) returned error: %v", args, err)
	}
	if res.ShowHelp || res.ShowVersion {
		t.Fatalf("ParseArgs(%v) unexpectedly requested help/version", args)
	}
	return res.Options
}

// Table mirrors the view-mode precedence cases from the reference
// implementation's own test suite: whichever of --long/--oneline/--grid/
// --tree appears last on the command line decides the base view, with
// --long additionally pulling grid/tree combinations into a table view.
func TestModeSelectionPrecedence(t *testing.T) {
	cases := []struct {
		args []string
		want ViewMode
	}{
		{[]string{}, ModeGrid},
		{[]string{"-G"}, ModeGrid},
		{[]string{flagOneline}, ModeLines},
		{[]string{"-1"}, ModeLines},
		{[]string{flagLong}, ModeDetails},
		{[]string{"-l"}, ModeDetails},
		{[]string{flagLong, flagGrid}, ModeDetails},
		{[]string{"-lG"}, ModeDetails},
		{[]string{flagLong, "--across"}, ModeDetails},
		{[]string{flagLong, flagGrid, flagOneline}, ModeLines},
		{[]string{flagLong, flagGrid, flagTree}, ModeDetails},
		{[]string{flagTree, flagGrid, flagLong}, ModeDetails},
		{[]string{flagTree, flagLong, flagGrid}, ModeDetails},
		{[]string{flagOneline, flagTree}, ModeDetails},
		{[]string{flagOneline, flagGrid}, ModeGrid},
		{[]string{flagTree, flagGrid}, ModeGrid},
		{[]string{"--header"}, ModeGrid}, // no effect without --long
	}

	for _, c := range cases {
		got := parse(t, c.args...).Mode
		if got != c.want {
			t.Errorf("args %v: mode = %v, want %v", c.args, got, c.want)
		}
	}
}

func TestTreeWithoutLongHasNoTable(t *testing.T) {
	opts := parse(t, flagTree)
	if opts.Mode != ModeDetails {
		t.Fatalf("expected details mode, got %v", opts.Mode)
	}
	if opts.TableOptions != nil {
		t.Error("expected no table for --tree without --long")
	}
	if opts.DirAction.Kind != DirRecurse || !opts.DirAction.Recurse.Tree {
		t.Error("expected tree recursion to be active")
	}
}

func TestTreeGivesWayToGridModeDropsRecursion(t *testing.T) {
	// "--tree --grid" resolves to plain Grid mode (grid wins), so tree
	// recursion shouldn't be requested at all: it wouldn't be honoured
	// by the grid renderer anyway.
	opts := parse(t, flagTree, flagGrid)
	if opts.DirAction.Kind == DirRecurse && opts.DirAction.Recurse.Tree {
		t.Error("expected tree recursion not to be requested when grid mode wins")
	}
}

func TestDotFilterFromAllCount(t *testing.T) {
	if got := parse(t).Filter.DotFilter; got != 0 {
		t.Errorf("expected JustFiles by default, got %v", got)
	}
	if got := parse(t, "-a").Filter.DotFilter; got != 1 {
		t.Errorf("expected Dotfiles for -a, got %v", got)
	}
	if got := parse(t, "-aa").Filter.DotFilter; got != 2 {
		t.Errorf("expected DotfilesAndDots for -aa, got %v", got)
	}
}

func TestSortFieldParsing(t *testing.T) {
	opts := parse(t, "--sort=size")
	if opts.Filter.Sort.Field != fsx.SortBySize {
		t.Errorf("expected SortBySize, got %v", opts.Filter.Sort.Field)
	}
}

func TestInvalidSortFieldErrors(t *testing.T) {
	_, err := ParseArgs([]string{"--sort=nonsense"}, noEnv, noProbe)
	if err == nil {
		t.Fatal("expected an error for an invalid --sort value")
	}
}

func TestIgnoreGlobSplitsOnPipe(t *testing.T) {
	opts := parse(t, "-I", "*.ogg|*.mp3")
	if !opts.Filter.IgnorePatterns.IsIgnored("song.ogg") || !opts.Filter.IgnorePatterns.IsIgnored("song.mp3") {
		t.Error("expected both pipe-separated patterns to be active")
	}
	if opts.Filter.IgnorePatterns.IsIgnored("song.wav") {
		t.Error("did not expect song.wav to be ignored")
	}
}

func TestColorModeDefaultsToAutomatic(t *testing.T) {
	opts := parse(t)
	if opts.ColorMode != 0 { // theme.ColourAutomatic == 0
		t.Errorf("expected automatic colour mode by default, got %v", opts.ColorMode)
	}
}

func TestColorModeNoColorEnv(t *testing.T) {
	res, err := ParseArgs(nil, envMap(map[string]string{"NO_COLOR": "1"}), noProbe)
	if err != nil {
		t.Fatal(err)
	}
	if res.Options.ColorMode != 2 { // theme.ColourNever == 2
		t.Errorf("expected NO_COLOR to force ColourNever, got %v", res.Options.ColorMode)
	}
}

func TestSizeFormatFlags(t *testing.T) {
	opts := parse(t, flagLong, "--binary")
	if opts.TableOptions.SizeFormat != output.BinarySize {
		t.Errorf("expected binary size format, got %v", opts.TableOptions.SizeFormat)
	}
}

func TestNoPermissionsSuppressesColumn(t *testing.T) {
	opts := parse(t, flagLong, "--no-permissions")
	for _, c := range opts.TableOptions.Columns.Collect(false) {
		if c.Kind == output.ColPermissions {
			t.Fatal("expected permissions column to be suppressed")
		}
	}
}

func TestGitColumnRequiresGitFlagAndRepo(t *testing.T) {
	opts := parse(t, flagLong, "--git")
	cols := opts.TableOptions.Columns.Collect(false)
	for _, c := range cols {
		if c.Kind == output.ColGitStatus {
			t.Fatal("expected git column to be suppressed when no repo is present")
		}
	}
	cols = opts.TableOptions.Columns.Collect(true)
	found := false
	for _, c := range cols {
		if c.Kind == output.ColGitStatus {
			found = true
		}
	}
	if !found {
		t.Fatal("expected git column to appear once a repo is present")
	}
}

func TestHelpAndVersionShortCircuit(t *testing.T) {
	res, err := ParseArgs([]string{"--help"}, noEnv, noProbe)
	if err != nil || !res.ShowHelp {
		t.Errorf("expected --help to short-circuit, got %+v err=%v", res, err)
	}

	res, err = ParseArgs([]string{"-v"}, noEnv, noProbe)
	if err != nil || !res.ShowVersion {
		t.Errorf("expected -v to short-circuit, got %+v err=%v", res, err)
	}
}

func TestColumnsEnvOverridesWidth(t *testing.T) {
	res, err := ParseArgs(nil, envMap(map[string]string{"COLUMNS": "42"}), fixedProbe(999))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Options.HasWidth || res.Options.TerminalWidth != 42 {
		t.Errorf("expected COLUMNS env to win over probe, got width=%d hasWidth=%v",
			res.Options.TerminalWidth, res.Options.HasWidth)
	}
}
