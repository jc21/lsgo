//go:build !linux && !darwin

package fsx

import "time"

// This build supplies degraded (but compiling) implementations of the
// Unix-specific metadata accessors for platforms without a syscall.Stat_t,
// such as Windows. lsgo's column output on these platforms omits the
// fields that can't be determined.

func (f *File) Inode() uint64                    { return 0 }
func (f *File) LinkCount() uint64                { return 1 }
func (f *File) UID() uint32                      { return 0 }
func (f *File) GID() uint32                      { return 0 }
func (f *File) Blocks() (int64, bool)            { return 0, false }
func (f *File) DeviceIDs() (major, minor uint32) { return 0, 0 }
func (f *File) AccessedTime() time.Time          { return f.Info.ModTime() }
func (f *File) ChangedTime() time.Time           { return f.Info.ModTime() }
func (f *File) CreatedTime() (time.Time, bool)   { return time.Time{}, false }
