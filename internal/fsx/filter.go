package fsx

import (
	"path/filepath"
	"slices"
	"time"
)

// SortCase distinguishes case-sensitive from case-insensitive natural sort
// comparisons.
type SortCase int

const (
	// CaseSensitive sorts uppercase before lowercase ('A' before 'a').
	CaseSensitive SortCase = iota
	// CaseInsensitive treats 'A' and 'a' as equal.
	CaseInsensitive
)

// SortField is the metadata field files are ordered by.
type SortField int

const (
	SortByName SortField = iota
	SortByNameMixHidden
	SortByExtension
	SortBySize
	SortByInode
	SortByModified
	SortByAccessed
	SortByChanged
	SortByCreated
	SortByType
	SortByModifiedAge // newest first is reversed relative to SortByModified
	SortUnsorted
)

// SortSpec pairs a field with the case-sensitivity to use for name/
// extension comparisons.
type SortSpec struct {
	Field SortField
	Case  SortCase
}

// DefaultSortSpec is used when the user doesn't pass --sort: names,
// compared case-insensitively.
func DefaultSortSpec() SortSpec {
	return SortSpec{Field: SortByName, Case: CaseInsensitive}
}

// Compare orders two files according to this sort spec. It does not handle
// reversal or directories-first grouping; see FileFilter.SortFiles for
// that.
func (s SortSpec) Compare(a, b *File) int {
	switch s.Field {
	case SortByName:
		if s.Case == CaseSensitive {
			return naturalCompare(a.Name, b.Name)
		}
		return naturalCompareFold(a.Name, b.Name)

	case SortByNameMixHidden:
		na, nb := stripDot(a.Name), stripDot(b.Name)
		if s.Case == CaseSensitive {
			return naturalCompare(na, nb)
		}
		return naturalCompareFold(na, nb)

	case SortByExtension:
		if a.Ext != b.Ext {
			if a.Ext < b.Ext {
				return -1
			}
			return 1
		}
		if s.Case == CaseSensitive {
			return naturalCompare(a.Name, b.Name)
		}
		return naturalCompareFold(a.Name, b.Name)

	case SortBySize:
		return compareInt64(a.Info.Size(), b.Info.Size())

	case SortByInode:
		return compareUint64(a.Inode(), b.Inode())

	case SortByModified:
		return compareTime(a.ModTime(), b.ModTime())

	case SortByModifiedAge:
		// Flip a and b: the "age" ordering is the reverse of modified time.
		return compareTime(b.ModTime(), a.ModTime())

	case SortByAccessed:
		return compareTime(a.AccessedTime(), b.AccessedTime())

	case SortByChanged:
		return compareTime(a.ChangedTime(), b.ChangedTime())

	case SortByCreated:
		ta, _ := a.CreatedTime()
		tb, _ := b.CreatedTime()
		return compareTime(ta, tb)

	case SortByType:
		ta, tb := TypeOf(a), TypeOf(b)
		if ta != tb {
			if ta < tb {
				return -1
			}
			return 1
		}
		return naturalCompare(a.Name, b.Name)
	}

	// SortUnsorted (and any other unhandled field) leaves files in
	// whatever order they were already in.
	return 0
}

func stripDot(name string) string {
	if len(name) > 0 && name[0] == '.' {
		return name[1:]
	}
	return name
}

func compareInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareUint64(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareTime(a, b time.Time) int {
	switch {
	case a.Before(b):
		return -1
	case a.After(b):
		return 1
	default:
		return 0
	}
}

// GitIgnoreMode controls whether git-ignored files are hidden.
type GitIgnoreMode int

const (
	GitIgnoreOff GitIgnoreMode = iota
	GitIgnoreCheckAndIgnore
)

// IgnorePatterns is a set of shell glob patterns; any filename matching one
// of them is excluded from a listing.
type IgnorePatterns struct {
	patterns []string
}

// NewIgnorePatterns parses a list of glob strings. Invalid patterns are
// skipped (filepath.Match's only error is ErrBadPattern, which we treat as
// "never matches" rather than aborting the whole listing).
func NewIgnorePatterns(patterns []string) IgnorePatterns {
	return IgnorePatterns{patterns: patterns}
}

// IsIgnored reports whether the given filename matches any of the patterns.
func (ip IgnorePatterns) IsIgnored(name string) bool {
	for _, p := range ip.patterns {
		if ok, err := filepath.Match(p, name); err == nil && ok {
			return true
		}
	}
	return false
}

// FileFilter bundles every "which files to show, and in what order"
// setting into one value that gets applied uniformly to whatever list of
// files is currently being displayed.
type FileFilter struct {
	ListDirsFirst  bool
	Sort           SortSpec
	Reverse        bool
	OnlyDirs       bool
	DotFilter      DotFilter
	IgnorePatterns IgnorePatterns
	GitIgnore      GitIgnoreMode
}

// FilterChildFiles removes ignored files (and, if OnlyDirs is set,
// non-directories) from a directory's children.
func (ff FileFilter) FilterChildFiles(files []*File) []*File {
	out := files[:0]
	for _, f := range files {
		if ff.IgnorePatterns.IsIgnored(f.Name) {
			continue
		}
		if ff.OnlyDirs && !f.IsDirectory() {
			continue
		}
		out = append(out, f)
	}
	return out
}

// FilterArgumentFiles removes ignored files from the list of names given
// directly on the command line. Unlike FilterChildFiles, this never
// applies OnlyDirs -- explicitly-named files are always shown.
func (ff FileFilter) FilterArgumentFiles(files []*File) []*File {
	out := files[:0]
	for _, f := range files {
		if ff.IgnorePatterns.IsIgnored(f.Name) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// SortFiles orders files in place according to the sort field, reversal,
// and directories-first settings.
func (ff FileFilter) SortFiles(files []*File) {
	slices.SortStableFunc(files, ff.Sort.Compare)

	if ff.Reverse {
		reverseFiles(files)
	}

	if ff.ListDirsFirst {
		slices.SortStableFunc(files, func(a, b *File) int {
			switch {
			case a.PointsToDirectory() && !b.PointsToDirectory():
				return -1
			case b.PointsToDirectory() && !a.PointsToDirectory():
				return 1
			default:
				return 0
			}
		})
	}
}

func reverseFiles(files []*File) {
	for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
		files[i], files[j] = files[j], files[i]
	}
}
