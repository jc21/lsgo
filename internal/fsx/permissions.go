package fsx

import "os"

// Type is a file's base filesystem type, as shown in the leftmost column of
// the permissions field. Its declaration order doubles as the sort order
// used by "--sort=type".
type Type int

const (
	TypeDirectory Type = iota
	TypeFile
	TypeLink
	TypePipe
	TypeSocket
	TypeCharDevice
	TypeBlockDevice
	TypeSpecial
)

// TypeOf classifies a file's base type.
func TypeOf(f *File) Type {
	switch {
	case f.IsRegularFile():
		return TypeFile
	case f.IsDirectory():
		return TypeDirectory
	case f.IsPipe():
		return TypePipe
	case f.IsLink():
		return TypeLink
	case f.IsCharDevice():
		return TypeCharDevice
	case f.IsBlockDevice():
		return TypeBlockDevice
	case f.IsSocket():
		return TypeSocket
	default:
		return TypeSpecial
	}
}

// TypeChar returns the traditional single-character type indicator used as
// the first character of the permissions column (like ls -l).
func (t Type) TypeChar() byte {
	switch t {
	case TypeDirectory:
		return 'd'
	case TypeLink:
		return 'l'
	case TypePipe:
		return 'p'
	case TypeSocket:
		return 's'
	case TypeCharDevice:
		return 'c'
	case TypeBlockDevice:
		return 'b'
	default:
		return '-'
	}
}

// Permissions is the Unix permission bitfield, with one flag per bit,
// mirroring what the rwxrwxrwx permissions column displays.
type Permissions struct {
	UserRead, UserWrite, UserExecute    bool
	GroupRead, GroupWrite, GroupExecute bool
	OtherRead, OtherWrite, OtherExecute bool
	Sticky, Setgid, Setuid              bool
}

// PermissionsOf extracts the permission bitfield from a file's mode.
func PermissionsOf(f *File) Permissions {
	mode := f.Info.Mode()
	perm := mode.Perm()

	return Permissions{
		UserRead:    perm&0o400 != 0,
		UserWrite:   perm&0o200 != 0,
		UserExecute: perm&0o100 != 0,

		GroupRead:    perm&0o040 != 0,
		GroupWrite:   perm&0o020 != 0,
		GroupExecute: perm&0o010 != 0,

		OtherRead:    perm&0o004 != 0,
		OtherWrite:   perm&0o002 != 0,
		OtherExecute: perm&0o001 != 0,

		Sticky: mode&os.ModeSticky != 0,
		Setgid: mode&os.ModeSetgid != 0,
		Setuid: mode&os.ModeSetuid != 0,
	}
}

// Octal renders the permissions as a 4-digit octal string, e.g. "0755".
func (p Permissions) Octal() string {
	bits := func(r, w, x bool) int {
		n := 0
		if r {
			n += 4
		}
		if w {
			n += 2
		}
		if x {
			n += 1
		}
		return n
	}

	special := bits(p.Setuid, p.Setgid, p.Sticky)
	owner := bits(p.UserRead, p.UserWrite, p.UserExecute)
	group := bits(p.GroupRead, p.GroupWrite, p.GroupExecute)
	other := bits(p.OtherRead, p.OtherWrite, p.OtherExecute)

	digits := "01234567"
	return string([]byte{
		digits[special], digits[owner], digits[group], digits[other],
	})
}
