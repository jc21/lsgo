package xattr

import (
	"os/exec"
	"strings"
)

// Enabled reports whether extended attribute lookups are supported on
// this platform.
const Enabled = true

// List returns the extended attributes set on path.
//
// macOS's syscall package doesn't expose listxattr/getxattr, so this
// shells out to the system "xattr" tool rather than hand-rolling the raw
// syscall numbers -- a little slower for files with many attributes, but
// exercises exactly the same code path a user would from a shell.
func List(path string) ([]Attribute, error) {
	out, err := exec.Command("xattr", path).Output()
	if err != nil {
		return nil, err
	}

	var attrs []Attribute
	for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if name == "" {
			continue
		}

		size := 0
		if val, err := exec.Command("xattr", "-p", name, path).Output(); err == nil {
			size = len(val)
		}
		attrs = append(attrs, Attribute{Name: name, Size: size})
	}
	return attrs, nil
}
