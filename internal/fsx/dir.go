package fsx

import (
	"os"
	"path/filepath"
	"strings"
)

// Dir provides a cached list of the file paths inside a directory that's
// being listed. It's shared with each of its children so that per-file
// checks -- like "is there a matching .c file for this .o file?" -- can see
// their siblings.
type Dir struct {
	// Path is the path that was read.
	Path string

	// contents holds every entry's full path, exactly as returned by the
	// directory read (before any filtering).
	contents []string
}

// ReadDir reads the entries of the directory at path.
func ReadDir(path string) (*Dir, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	contents := make([]string, 0, len(entries))
	for _, e := range entries {
		contents = append(contents, filepath.Join(path, e.Name()))
	}

	return &Dir{Path: path, contents: contents}, nil
}

// Contains reports whether this directory contains an entry with the given
// path, used to detect e.g. a matching source file for a compiled object.
func (d *Dir) Contains(path string) bool {
	for _, c := range d.contents {
		if c == path {
			return true
		}
	}
	return false
}

// Join appends a child path onto this directory's path.
func (d *Dir) Join(child string) string {
	return filepath.Join(d.Path, child)
}

// DotFilter controls which of a directory's dot-prefixed entries --
// ordinary dotfiles, and the special "." and ".." entries -- are included
// in a listing.
type DotFilter int

const (
	// JustFiles hides anything starting with a dot. This is the default.
	JustFiles DotFilter = iota

	// Dotfiles shows dotfiles, but not "." or "..".
	Dotfiles

	// DotfilesAndDots shows dotfiles, plus "." and "..".
	DotfilesAndDots
)

// DotFilterFromCount turns a count of how many times -a/--all was given
// into a DotFilter: -a shows dotfiles, -aa also shows "." and "..".
func DotFilterFromCount(count int) DotFilter {
	switch {
	case count <= 0:
		return JustFiles
	case count == 1:
		return Dotfiles
	default:
		return DotfilesAndDots
	}
}

func (df DotFilter) showsDotfiles() bool { return df != JustFiles }
func (df DotFilter) showsDots() bool     { return df == DotfilesAndDots }

// Files reads this directory's children as File values, applying the given
// dot filter and optionally consulting a GitCache to skip git-ignored
// entries. Entries that fail to stat are reported as errors alongside the
// path that caused them, rather than aborting the whole listing.
func (d *Dir) Files(dots DotFilter, git *GitCache, gitIgnoring bool) ([]*File, []error) {
	var files []*File
	var errs []error

	if dots.showsDots() {
		if f, err := newDotEntry(d.Path, ".", d); err == nil {
			files = append(files, f)
		} else {
			errs = append(errs, err)
		}

		parentPath := filepath.Join(d.Path, "..")
		if f, err := newDotEntry(parentPath, "..", d); err == nil {
			files = append(files, f)
		} else {
			errs = append(errs, err)
		}
	}

	for _, path := range d.contents {
		name := filenameOf(path)
		if !dots.showsDotfiles() && strings.HasPrefix(name, ".") {
			continue
		}

		if gitIgnoring && git != nil {
			status := git.Status(path, false)
			if status.Unstaged == GitIgnored {
				continue
			}
		}

		f, err := NewFile(path, d, name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		files = append(files, f)
	}

	return files, errs
}
