package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"lsgo/internal/fsx"
	"lsgo/internal/output"
	"lsgo/internal/theme"
)

// ViewMode selects which of lsgo's overall layouts is used.
type ViewMode int

const (
	ModeGrid ViewMode = iota
	ModeLines
	ModeDetails
)

// DirActionKind selects what happens when a directory is encountered.
type DirActionKind int

const (
	DirList DirActionKind = iota
	DirAsFile
	DirRecurse
)

// DirAction bundles the directory-handling decision with its recursion
// settings (only meaningful when Kind is DirRecurse).
type DirAction struct {
	Kind    DirActionKind
	Recurse output.RecurseOptions
}

// Options is the fully-resolved configuration for one run, derived from
// command-line flags and the environment.
type Options struct {
	Mode       ViewMode
	GridAcross bool

	DirAction DirAction

	Filter          fsx.FileFilter
	FileNameOptions output.FileNameOptions

	// TableOptions is nil when the details view has no table columns
	// (a plain --tree listing without --long).
	TableOptions *output.TableOptions
	Header       bool
	ShowXattrs   bool

	ColorMode  theme.UseColours
	ColorScale theme.ColourScale

	// TerminalWidth is the console width to lay the grid out for, or 0
	// if it couldn't be determined (stdout isn't a terminal and
	// COLUMNS isn't set) -- in which case the grid view falls back to
	// one file per line.
	TerminalWidth int
	HasWidth      bool

	Paths []string
}

// Result is what Parse produces: either a fully resolved Options, or a
// request to print help/version text instead of listing anything.
type Result struct {
	Options     *Options
	ShowHelp    bool
	ShowVersion bool
}

// Env abstracts environment variable lookups, so option resolution can be
// tested without depending on the real process environment.
type Env func(name string) (string, bool)

// OSEnv looks variables up via os.LookupEnv.
func OSEnv(name string) (string, bool) { return os.LookupEnv(name) }

// TerminalProbe reports the terminal width for the given file descriptor,
// abstracted the same way as Env so tests can supply a fixed answer
// instead of depending on the real stdout.
type TerminalProbe func() (int, bool)

// ParseArgs parses argv (excluding the program name) and the environment
// into a Result.
func ParseArgs(argv []string, env Env, probe TerminalProbe) (*Result, error) {
	flags, err := Parse(argv)
	if err != nil {
		return nil, err
	}

	if flags.Has("help") {
		return &Result{ShowHelp: true}, nil
	}
	if flags.Has("version") {
		return &Result{ShowVersion: true}, nil
	}

	opts := &Options{Paths: flags.Free}

	mode, hasTable, canTree := deduceMode(flags)
	opts.Mode = mode
	opts.GridAcross = flags.Has("across")

	dirAction, err := deduceDirAction(flags, canTree)
	if err != nil {
		return nil, err
	}
	opts.DirAction = dirAction

	filter, err := deduceFilter(flags)
	if err != nil {
		return nil, err
	}
	opts.Filter = filter

	opts.FileNameOptions = output.FileNameOptions{
		Classify:  output.Classify(flags.Has("classify")),
		ShowIcons: deduceShowIcons(flags, env),
	}

	if hasTable {
		table, err := deduceTableOptions(flags, env)
		if err != nil {
			return nil, err
		}
		opts.TableOptions = &table
		opts.Header = flags.Has("header")
	}
	opts.ShowXattrs = flags.Has("extended")

	colorMode, err := deduceColorMode(flags, env)
	if err != nil {
		return nil, err
	}
	opts.ColorMode = colorMode
	opts.ColorScale = theme.ScaleFixed
	if flags.Has("color-scale") {
		opts.ColorScale = theme.ScaleGradient
	}

	if cols, ok := env("COLUMNS"); ok && cols != "" {
		n, err := strconv.Atoi(cols)
		if err != nil {
			return nil, &ParseError{Msg: fmt.Sprintf("invalid COLUMNS value %q", cols)}
		}
		opts.TerminalWidth, opts.HasWidth = n, true
	} else if probe != nil {
		opts.TerminalWidth, opts.HasWidth = probe()
	}

	return &Result{Options: opts}, nil
}

// deduceMode picks the overall view and reports whether it should include
// a metadata table (long view), and whether tree recursion is meaningful
// for it. See the package doc comment on view-mode precedence.
func deduceMode(flags *ParsedFlags) (mode ViewMode, hasTable bool, canTree bool) {
	winner, ok := flags.LastOf("long", "oneline", "grid", "tree")
	if !ok {
		return ModeGrid, false, false
	}

	switch winner {
	case "oneline":
		return ModeLines, false, false
	case "long":
		return ModeDetails, true, true
	case "tree":
		if flags.Has("long") {
			return ModeDetails, true, true
		}
		return ModeDetails, false, true
	default: // "grid"
		if flags.Has("long") {
			return ModeDetails, true, true
		}
		return ModeGrid, false, false
	}
}

func deduceDirAction(flags *ParsedFlags, canTree bool) (DirAction, error) {
	var level *int
	if v, ok := flags.Value("level"); ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			return DirAction{}, &ParseError{Msg: fmt.Sprintf("invalid value for --level: %q", v)}
		}
		level = &n
	}

	treeFlag := flags.Has("tree")
	recurseFlag := flags.Has("recurse")
	asFile := flags.Has("list-dirs")

	switch {
	case treeFlag && canTree:
		return DirAction{Kind: DirRecurse, Recurse: output.RecurseOptions{Tree: true, MaxDepth: level}}, nil
	case recurseFlag:
		return DirAction{Kind: DirRecurse, Recurse: output.RecurseOptions{Tree: false, MaxDepth: level}}, nil
	case asFile:
		return DirAction{Kind: DirAsFile}, nil
	default:
		return DirAction{Kind: DirList}, nil
	}
}

func deduceFilter(flags *ParsedFlags) (fsx.FileFilter, error) {
	sortSpec, err := deduceSortSpec(flags)
	if err != nil {
		return fsx.FileFilter{}, err
	}

	var patterns []string
	if v, ok := flags.Value("ignore-glob"); ok {
		patterns = strings.Split(v, "|")
	}

	gitIgnore := fsx.GitIgnoreOff
	if flags.Has("git-ignore") {
		gitIgnore = fsx.GitIgnoreCheckAndIgnore
	}

	return fsx.FileFilter{
		ListDirsFirst:  flags.Has("group-directories-first"),
		Sort:           sortSpec,
		Reverse:        flags.Has("reverse"),
		OnlyDirs:       flags.Has("only-dirs"),
		DotFilter:      fsx.DotFilterFromCount(flags.Count("all")),
		IgnorePatterns: fsx.NewIgnorePatterns(patterns),
		GitIgnore:      gitIgnore,
	}, nil
}

// sortValueSize is the --sort value for ordering by file size. Named since
// parser_test.go also uses it as its example --sort value.
const sortValueSize = "size"

func deduceSortSpec(flags *ParsedFlags) (fsx.SortSpec, error) {
	word, ok := flags.Value("sort")
	if !ok {
		return fsx.DefaultSortSpec(), nil
	}

	switch word {
	case "name", "filename":
		return fsx.SortSpec{Field: fsx.SortByName, Case: fsx.CaseInsensitive}, nil
	case "Name", "Filename":
		return fsx.SortSpec{Field: fsx.SortByName, Case: fsx.CaseSensitive}, nil
	case ".name", ".filename":
		return fsx.SortSpec{Field: fsx.SortByNameMixHidden, Case: fsx.CaseInsensitive}, nil
	case ".Name", ".Filename":
		return fsx.SortSpec{Field: fsx.SortByNameMixHidden, Case: fsx.CaseSensitive}, nil
	case sortValueSize, "filesize":
		return fsx.SortSpec{Field: fsx.SortBySize}, nil
	case "ext", "extension":
		return fsx.SortSpec{Field: fsx.SortByExtension, Case: fsx.CaseInsensitive}, nil
	case "Ext", "Extension":
		return fsx.SortSpec{Field: fsx.SortByExtension, Case: fsx.CaseSensitive}, nil
	case "date", "time", "mod", longModified, "new", "newest":
		return fsx.SortSpec{Field: fsx.SortByModified}, nil
	case "age", "old", "oldest":
		return fsx.SortSpec{Field: fsx.SortByModifiedAge}, nil
	case "ch", longChanged:
		return fsx.SortSpec{Field: fsx.SortByChanged}, nil
	case "acc", longAccessed:
		return fsx.SortSpec{Field: fsx.SortByAccessed}, nil
	case "cr", longCreated:
		return fsx.SortSpec{Field: fsx.SortByCreated}, nil
	case "inode":
		return fsx.SortSpec{Field: fsx.SortByInode}, nil
	case "type":
		return fsx.SortSpec{Field: fsx.SortByType}, nil
	case "none":
		return fsx.SortSpec{Field: fsx.SortUnsorted}, nil
	default:
		return fsx.SortSpec{}, &ParseError{Msg: fmt.Sprintf("invalid value for --sort: %q", word)}
	}
}

func deduceTableOptions(flags *ParsedFlags, env Env) (output.TableOptions, error) {
	sizeFormat := output.DecimalSize
	if winner, ok := flags.LastOf("binary", "bytes"); ok {
		if winner == "binary" {
			sizeFormat = output.BinarySize
		} else {
			sizeFormat = output.JustBytes
		}
	}

	timeFormat, err := deduceTimeFormat(flags, env)
	if err != nil {
		return output.TableOptions{}, err
	}

	userFormat := output.UserName
	if flags.Has("numeric") {
		userFormat = output.UserNumeric
	}

	timeTypes, err := deduceTimeTypes(flags)
	if err != nil {
		return output.TableOptions{}, err
	}

	columns := output.TableColumns{
		TimeTypes:   timeTypes,
		Inode:       flags.Has("inode"),
		Links:       flags.Has("links"),
		Blocks:      flags.Has("blocks"),
		Group:       flags.Has("group"),
		Git:         flags.Has("git"),
		Octal:       flags.Has("octal-permissions"),
		Permissions: !flags.Has("no-permissions"),
		Filesize:    !flags.Has("no-filesize"),
		User:        !flags.Has("no-user"),
	}

	return output.TableOptions{
		SizeFormat: sizeFormat,
		TimeFormat: timeFormat,
		UserFormat: userFormat,
		Columns:    columns,
	}, nil
}

func deduceTimeTypes(flags *ParsedFlags) (output.TimeTypes, error) {
	if flags.Has("no-time") {
		return output.TimeTypes{}, nil
	}

	if word, ok := flags.Value("time"); ok {
		switch word {
		case "mod", longModified:
			return output.TimeTypes{Modified: true}, nil
		case "ch", longChanged:
			return output.TimeTypes{Changed: true}, nil
		case "acc", longAccessed:
			return output.TimeTypes{Accessed: true}, nil
		case "cr", longCreated:
			return output.TimeTypes{Created: true}, nil
		default:
			return output.TimeTypes{}, &ParseError{Msg: fmt.Sprintf("invalid value for --time: %q", word)}
		}
	}

	tt := output.TimeTypes{
		Modified: flags.Has(longModified),
		Changed:  flags.Has(longChanged),
		Accessed: flags.Has(longAccessed),
		Created:  flags.Has(longCreated),
	}
	if tt.Modified || tt.Changed || tt.Accessed || tt.Created {
		return tt, nil
	}
	return output.DefaultTimeTypes(), nil
}

func deduceTimeFormat(flags *ParsedFlags, env Env) (output.TimeFormat, error) {
	word, ok := flags.Value("time-style")
	if !ok {
		if v, present := env("TIME_STYLE"); present && v != "" {
			word, ok = v, true
		}
	}
	if !ok {
		return output.DefaultTimeFormat, nil
	}

	switch word {
	case "default":
		return output.DefaultTimeFormat, nil
	case "iso":
		return output.ISOTimeFormat, nil
	case "long-iso":
		return output.LongISOTimeFormat, nil
	case "full-iso":
		return output.FullISOTimeFormat, nil
	default:
		return 0, &ParseError{Msg: fmt.Sprintf("invalid value for --time-style: %q", word)}
	}
}

func deduceColorMode(flags *ParsedFlags, env Env) (theme.UseColours, error) {
	word, ok := flags.Value("color")
	if !ok {
		if _, present := env("NO_COLOR"); present {
			return theme.ColourNever, nil
		}
		return theme.ColourAutomatic, nil
	}

	switch word {
	case "always":
		return theme.ColourAlways, nil
	case "auto", "automatic":
		return theme.ColourAutomatic, nil
	case "never":
		return theme.ColourNever, nil
	default:
		return 0, &ParseError{Msg: fmt.Sprintf("invalid value for --color: %q", word)}
	}
}

func deduceShowIcons(flags *ParsedFlags, env Env) output.ShowIcons {
	if flags.Has("no-icons") || !flags.Has("icons") {
		return output.ShowIcons{On: false}
	}

	spacing := 1
	if v, ok := env("LS_ICON_SPACING"); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			spacing = n
		}
	}
	return output.ShowIcons{On: true, Spaces: spacing}
}
