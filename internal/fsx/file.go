// Package fsx wraps filesystem entries with the metadata lsgo needs to
// display and sort them: cached lstat results, a parsed extension, symlink
// resolution, and various "what kind of thing is this" predicates.
package fsx

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// File is a single filesystem entry -- a wrapper around an os.FileInfo
// (from Lstat, so symlinks are reported as symlinks) plus the bits of
// context (name, extension, parent) that get queried repeatedly while
// rendering.
type File struct {
	// Name is the filename portion of the path, including its extension.
	Name string

	// Ext is the lowercased extension (the text after the last '.'), or
	// "" if the file has none. Dotfiles count their remainder as an
	// extension, so ".gitignore" has extension "gitignore".
	Ext string

	// Path is the path this file was reached by; may be relative.
	Path string

	// Info is the cached Lstat result.
	Info os.FileInfo

	// Parent is the directory that produced this file, if any. Files
	// named directly on the command line have no parent.
	Parent *Dir

	// IsDotEntry marks the synthetic "." and ".." entries added when the
	// user passes -a twice.
	IsDotEntry bool
}

// StatError pairs a path with the error that occurred while statting it.
type StatError struct {
	Path string
	Err  error
}

func (e *StatError) Error() string { return e.Path + ": " + e.Err.Error() }
func (e *StatError) Unwrap() error { return e.Err }

// NewFile lstats the given path and wraps the result. The name may be
// supplied explicitly (used for directory children, where it's already
// known); otherwise it's taken from the last path component.
func NewFile(path string, parent *Dir, name string) (*File, error) {
	if name == "" {
		name = filenameOf(path)
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil, &StatError{Path: path, Err: err}
	}

	return &File{
		Name:   name,
		Ext:    extensionOf(name),
		Path:   path,
		Info:   info,
		Parent: parent,
	}, nil
}

// newDotEntry builds the synthetic "." or ".." file used when -aa is given.
func newDotEntry(path, name string, parent *Dir) (*File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, &StatError{Path: path, Err: err}
	}
	return &File{
		Name:       name,
		Ext:        extensionOf(path),
		Path:       path,
		Info:       info,
		Parent:     parent,
		IsDotEntry: true,
	}, nil
}

// filenameOf returns the last path component, falling back to the whole
// path when there's no real basename (e.g. "/" or ".").
func filenameOf(path string) string {
	cleaned := filepath.Clean(path)
	base := filepath.Base(cleaned)
	if base == "." || base == string(filepath.Separator) {
		return path
	}
	return base
}

// extensionOf extracts the lowercased extension from a filename, counting
// the text after the *last* dot -- so dotfiles like ".gitignore" have the
// extension "gitignore".
func extensionOf(name string) string {
	base := filepath.Base(name)
	i := strings.LastIndexByte(base, '.')
	if i < 0 {
		return ""
	}
	return strings.ToLower(base[i+1:])
}

// IsDirectory reports whether this entry is a directory on disk (symlinks
// to directories return false here; see PointsToDirectory).
func (f *File) IsDirectory() bool { return f.Info.IsDir() }

// IsRegularFile reports whether this is a plain file (not a directory,
// symlink, or other special type).
func (f *File) IsRegularFile() bool { return f.Info.Mode().IsRegular() }

// IsLink reports whether this entry is a symlink.
func (f *File) IsLink() bool { return f.Info.Mode()&os.ModeSymlink != 0 }

// IsPipe reports whether this entry is a named pipe (FIFO).
func (f *File) IsPipe() bool { return f.Info.Mode()&os.ModeNamedPipe != 0 }

// IsSocket reports whether this entry is a Unix domain socket.
func (f *File) IsSocket() bool { return f.Info.Mode()&os.ModeSocket != 0 }

// IsCharDevice reports whether this entry is a character device.
func (f *File) IsCharDevice() bool {
	return f.Info.Mode()&os.ModeDevice != 0 && f.Info.Mode()&os.ModeCharDevice != 0
}

// IsBlockDevice reports whether this entry is a block device.
func (f *File) IsBlockDevice() bool {
	return f.Info.Mode()&os.ModeDevice != 0 && f.Info.Mode()&os.ModeCharDevice == 0
}

// IsExecutableFile reports whether this is a regular file with any
// executable bit set for the current user.
func (f *File) IsExecutableFile() bool {
	return f.IsRegularFile() && f.Info.Mode()&0o100 != 0
}

// PointsToDirectory reports whether this file is a directory, or a
// (transitively resolved) symlink to one.
func (f *File) PointsToDirectory() bool {
	if f.IsDirectory() {
		return true
	}
	if f.IsLink() {
		if target, ok := f.LinkTarget(); ok {
			return target.PointsToDirectory()
		}
	}
	return false
}

// LinkTarget resolves a symlink to the File it points at. The second
// return value is false if the file isn't a symlink, the link is broken,
// or it couldn't be read for some other reason -- callers that need to
// distinguish those cases should use LinkTargetDetailed.
func (f *File) LinkTarget() (*File, bool) {
	target, err, broken := f.LinkTargetDetailed()
	if err != nil || broken {
		return nil, false
	}
	return target, true
}

// LinkTargetDetailed resolves a symlink, distinguishing three outcomes: a
// successfully resolved target, a broken link (err is nil, broken is
// true), and a read error (err is non-nil).
func (f *File) LinkTargetDetailed() (target *File, err error, broken bool) {
	raw, readErr := os.Readlink(f.Path)
	if readErr != nil {
		return nil, readErr, false
	}

	absolute := f.reorient(raw)

	info, statErr := os.Stat(absolute)
	if statErr != nil {
		return nil, nil, true
	}

	return &File{
		Name: filenameOf(raw),
		Ext:  extensionOf(raw),
		Path: raw,
		Info: info,
	}, nil, false
}

// LinkTargetRaw returns the raw (unresolved) target path of a symlink,
// used for displaying broken links.
func (f *File) LinkTargetRaw() (string, error) {
	return os.Readlink(f.Path)
}

// reorient turns a (possibly relative) symlink target into a path that can
// be looked up from the current working directory.
func (f *File) reorient(target string) string {
	if filepath.IsAbs(target) {
		return target
	}
	if f.Parent != nil {
		return filepath.Join(f.Parent.Path, target)
	}
	dir := filepath.Dir(f.Path)
	return filepath.Join(dir, target)
}

// ExtensionIsOneOf reports whether this file's extension is any of the
// given (already-lowercased) choices.
func (f *File) ExtensionIsOneOf(choices ...string) bool {
	if f.Ext == "" {
		return false
	}
	for _, c := range choices {
		if f.Ext == c {
			return true
		}
	}
	return false
}

// NameIsOneOf reports whether this file's full name is any of the given
// choices.
func (f *File) NameIsOneOf(choices ...string) bool {
	for _, c := range choices {
		if f.Name == c {
			return true
		}
	}
	return false
}

// ModTime returns the file's last-modified time.
func (f *File) ModTime() time.Time { return f.Info.ModTime() }
