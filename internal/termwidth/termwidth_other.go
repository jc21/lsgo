//go:build !darwin && !linux

package termwidth

// Width always reports failure on platforms without an ioctl-based
// implementation; callers fall back to the COLUMNS environment variable
// or a fixed default.
func Width(fd uintptr) (int, bool) { return 0, false }

func IsTerminal(fd uintptr) bool { return false }

func Stdout() (int, bool) { return 0, false }
