//go:build !linux && !darwin

package xattr

// Enabled reports whether extended attribute lookups are supported on
// this platform.
const Enabled = false

// List always reports no attributes on unsupported platforms.
func List(path string) ([]Attribute, error) {
	return nil, nil
}
