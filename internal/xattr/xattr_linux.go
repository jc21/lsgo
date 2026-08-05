package xattr

import (
	"strings"
	"syscall"
)

// Enabled reports whether extended attribute lookups are supported on
// this platform.
const Enabled = true

// List returns the extended attributes set on path.
func List(path string) ([]Attribute, error) {
	size, err := syscall.Listxattr(path, nil)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, nil
	}

	buf := make([]byte, size)
	n, err := syscall.Listxattr(path, buf)
	if err != nil {
		return nil, err
	}

	var attrs []Attribute
	for _, name := range strings.Split(string(buf[:n]), "\x00") {
		if name == "" {
			continue
		}
		valSize, err := syscall.Getxattr(path, name, nil)
		if err != nil {
			valSize = 0
		}
		attrs = append(attrs, Attribute{Name: name, Size: valSize})
	}
	return attrs, nil
}
