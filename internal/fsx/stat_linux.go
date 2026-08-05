package fsx

import (
	"syscall"
	"time"
)

// statT extracts the raw stat_t embedded in a Go FileInfo. Returns false if
// the platform-specific data isn't available (shouldn't happen on linux).
func (f *File) statT() (*syscall.Stat_t, bool) {
	st, ok := f.Info.Sys().(*syscall.Stat_t)
	return st, ok
}

// Inode returns the file's inode number.
func (f *File) Inode() uint64 {
	if st, ok := f.statT(); ok {
		return st.Ino
	}
	return 0
}

// LinkCount returns the number of hard links to this file.
func (f *File) LinkCount() uint64 {
	if st, ok := f.statT(); ok {
		return uint64(st.Nlink)
	}
	return 1
}

// UID returns the numeric ID of the file's owning user.
func (f *File) UID() uint32 {
	if st, ok := f.statT(); ok {
		return st.Uid
	}
	return 0
}

// GID returns the numeric ID of the file's owning group.
func (f *File) GID() uint32 {
	if st, ok := f.statT(); ok {
		return st.Gid
	}
	return 0
}

// Blocks returns the number of filesystem blocks this file occupies, and
// whether that value is meaningful for this file type.
func (f *File) Blocks() (int64, bool) {
	if !f.IsRegularFile() && !f.IsLink() {
		return 0, false
	}
	if st, ok := f.statT(); ok {
		return st.Blocks, true
	}
	return 0, false
}

// DeviceIDs returns the major/minor device numbers for a character or
// block device file.
func (f *File) DeviceIDs() (major, minor uint32) {
	st, ok := f.statT()
	if !ok {
		return 0, 0
	}
	rdev := st.Rdev
	return uint32((rdev >> 24) & 0xff), uint32(rdev & 0xff)
}

// AccessedTime returns the file's last-accessed time.
func (f *File) AccessedTime() time.Time {
	if st, ok := f.statT(); ok {
		return time.Unix(st.Atim.Sec, st.Atim.Nsec)
	}
	return time.Time{}
}

// ChangedTime returns the file's last inode-changed time (ctime).
func (f *File) ChangedTime() time.Time {
	if st, ok := f.statT(); ok {
		return time.Unix(st.Ctim.Sec, st.Ctim.Nsec)
	}
	return time.Time{}
}

// CreatedTime returns the file's creation ("birth") time. Linux's classic
// stat(2) doesn't expose one, so this always reports false; lsgo falls
// back to the modified time for --created / --time=created on this
// platform, same as the modified time is used for --changed on Windows in
// the original implementation.
func (f *File) CreatedTime() (time.Time, bool) {
	return time.Time{}, false
}
