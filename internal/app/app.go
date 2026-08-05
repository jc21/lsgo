// Package app wires together the cli, fsx, theme, and output packages into
// lsgo's actual run loop: work out what was named on the command line,
// print any plain files first, then print each directory (recursively, if
// asked) in whichever view was selected.
package app

import (
	"fmt"
	"io"

	"lsgo/internal/cli"
	"lsgo/internal/fsx"
	"lsgo/internal/output"
	"lsgo/internal/theme"
)

// Exit codes, matching the reference implementation's conventions.
const (
	ExitSuccess    = 0
	ExitRuntimeErr = 1
	ExitOptionsErr = 3
)

// App holds everything a run needs: the resolved options, the rendering
// theme, an optional Git status cache, and where to write output.
type App struct {
	Opts   *cli.Options
	Theme  theme.Theme
	Git    *fsx.GitCache
	Stdout io.Writer
	Stderr io.Writer
}

// New builds an App from resolved options. isTerminal should reflect
// whether Stdout is attached to a terminal (used to resolve
// --color=auto), and lsColors is the raw LS_COLORS environment value.
func New(opts *cli.Options, isTerminal bool, lsColors string, stdout, stderr io.Writer) *App {
	th := theme.Build(opts.ColorMode, opts.ColorScale, isTerminal, lsColors)

	var git *fsx.GitCache
	if shouldScanForGit(opts) {
		git = fsx.NewGitCache()
		git.Prime(opts.Paths)
	}

	return &App{Opts: opts, Theme: th, Git: git, Stdout: stdout, Stderr: stderr}
}

func shouldScanForGit(opts *cli.Options) bool {
	if opts.Filter.GitIgnore == fsx.GitIgnoreCheckAndIgnore {
		return true
	}
	return opts.TableOptions != nil && opts.TableOptions.Columns.Git
}

// Run lists everything named in Opts.Paths (defaulting to the current
// directory) and returns the process exit code.
func (a *App) Run() int {
	paths := a.Opts.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var files []*fsx.File
	var dirs []*fsx.Dir
	exitStatus := ExitSuccess

	for _, p := range paths {
		f, err := fsx.NewFile(p, nil, "")
		if err != nil {
			exitStatus = ExitRuntimeErr
			fmt.Fprintf(a.Stderr, "%s: %s\n", p, unwrapStatErr(err))
			continue
		}

		if f.PointsToDirectory() && a.Opts.DirAction.Kind != cli.DirAsFile {
			d, err := fsx.ReadDir(f.Path)
			if err != nil {
				exitStatus = ExitRuntimeErr
				fmt.Fprintf(a.Stderr, "%s: %s\n", p, err)
				continue
			}
			dirs = append(dirs, d)
		} else {
			files = append(files, f)
		}
	}

	noFiles := len(files) == 0
	isOnlyDir := len(dirs) == 1 && noFiles

	files = a.Opts.Filter.FilterArgumentFiles(files)
	a.printFiles(nil, files)

	a.printDirs(dirs, true, isOnlyDir, 1)

	return exitStatus
}

func unwrapStatErr(err error) error {
	if se, ok := err.(*fsx.StatError); ok {
		return se.Err
	}
	return err
}

// printDirs prints each directory's header (unless it's the sole thing
// being listed) and contents, recursing into subdirectories itself when
// --recurse was given without --tree (tree recursion is instead handled
// inline by the details renderer, since it needs to interleave rows).
func (a *App) printDirs(dirs []*fsx.Dir, first, isOnlyDir bool, depth int) {
	for i, dir := range dirs {
		// no leading blank line before the very first thing printed
		if !first || i != 0 {
			_, _ = fmt.Fprintln(a.Stdout)
		}

		if !isOnlyDir {
			fmt.Fprintf(a.Stdout, "%s:\n", dir.Path)
		}

		gitIgnoring := a.Opts.Filter.GitIgnore == fsx.GitIgnoreCheckAndIgnore
		children, errs := dir.Files(a.Opts.Filter.DotFilter, a.Git, gitIgnoring)
		for _, e := range errs {
			fmt.Fprintf(a.Stderr, "[%s]\n", e)
		}

		children = a.Opts.Filter.FilterChildFiles(children)
		a.Opts.Filter.SortFiles(children)

		if a.Opts.DirAction.Kind == cli.DirRecurse && !a.Opts.DirAction.Recurse.Tree &&
			!a.Opts.DirAction.Recurse.IsTooDeep(depth) {
			var childDirs []*fsx.Dir
			for _, c := range children {
				if c.IsDirectory() && !c.IsDotEntry {
					if cd, err := fsx.ReadDir(c.Path); err == nil {
						childDirs = append(childDirs, cd)
					} else {
						fmt.Fprintf(a.Stderr, "%s: %s\n", c.Path, err)
					}
				}
			}

			a.printFiles(dir, children)
			a.printDirs(childDirs, false, false, depth+1)
			continue
		}

		a.printFiles(dir, children)
	}
}

// printFiles renders one set of files (either the loose files named
// directly on the command line, or one directory's children) using
// whichever view was selected.
func (a *App) printFiles(_ *fsx.Dir, files []*fsx.File) {
	if len(files) == 0 {
		return
	}

	switch a.Opts.Mode {
	case cli.ModeGrid:
		if a.Opts.HasWidth {
			a.Opts.Filter.SortFiles(files)
			cells := make([]output.Cell, len(files))
			for i, f := range files {
				cells[i] = output.RenderFileName(&a.Theme, a.Opts.FileNameOptions, f, false)
			}
			fmt.Fprint(a.Stdout, output.RenderGrid(cells, a.Opts.TerminalWidth, a.Opts.GridAcross))
			return
		}
		fallthrough

	case cli.ModeLines:
		a.Opts.Filter.SortFiles(files)
		fmt.Fprint(a.Stdout, output.RenderOneLine(&a.Theme, a.Opts.FileNameOptions, files))

	case cli.ModeDetails:
		var recurse *output.RecurseOptions
		if a.Opts.DirAction.Kind == cli.DirRecurse {
			r := a.Opts.DirAction.Recurse
			recurse = &r
		}

		renderer := &output.DetailsRenderer{
			Theme:           &a.Theme,
			FileNameOptions: a.Opts.FileNameOptions,
			Options: output.DetailsOptions{
				Table:      a.Opts.TableOptions,
				Header:     a.Opts.Header,
				ShowXattrs: a.Opts.ShowXattrs,
			},
			Recurse:     recurse,
			Filter:      a.Opts.Filter,
			Git:         a.Git,
			GitIgnoring: a.Opts.Filter.GitIgnore == fsx.GitIgnoreCheckAndIgnore,
		}

		if err := renderer.Render(a.Stdout, files); err != nil {
			_, _ = fmt.Fprintln(a.Stderr, err)
		}
	}
}
