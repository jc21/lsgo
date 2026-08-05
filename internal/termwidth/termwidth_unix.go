//go:build darwin || linux

// Package termwidth detects the terminal's width, and whether a file
// descriptor is attached to a terminal at all, without depending on
// anything outside the standard library.
package termwidth

import (
	"os"
	"syscall"
	"unsafe"
)

// tiocgwinsz is the ioctl request number for "get window size". Its value
// is fixed by each platform's tty driver and has been stable for decades;
// darwin and linux happen to disagree on the exact number, hence the
// per-OS files.
const tiocgwinsz = tiocgwinszValue

type winsize struct {
	Row, Col, Xpixel, Ypixel uint16
}

// Width returns the terminal width in columns for the given file
// descriptor, and whether it succeeded in reading one at all (which also
// serves as an "is this a terminal?" check: non-terminals fail the
// ioctl).
func Width(fd uintptr) (int, bool) {
	var ws winsize
	// This unsafe.Pointer is the standard, audited pattern for a raw
	// TIOCGWINSZ ioctl: syscall.Syscall needs a uintptr to &ws's memory to
	// fill in, and there's no other way to get terminal width without an
	// external dependency (see CLAUDE.md "Why zero dependencies").
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, tiocgwinsz, uintptr(unsafe.Pointer(&ws))) //nolint:gosec
	if errno != 0 || ws.Col == 0 {
		return 0, false
	}
	return int(ws.Col), true
}

// IsTerminal reports whether fd is attached to a terminal.
func IsTerminal(fd uintptr) bool {
	_, ok := Width(fd)
	return ok
}

// Stdout is a convenience wrapper around Width(os.Stdout.Fd()).
func Stdout() (int, bool) {
	return Width(os.Stdout.Fd())
}
