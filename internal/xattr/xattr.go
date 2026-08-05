// Package xattr looks up a file's extended attributes for the --extended
// (-@) flag. Support is best-effort and platform-specific; where it isn't
// available, List simply reports no attributes rather than erroring, so
// the rest of the listing is unaffected.
package xattr

// Attribute is one extended attribute: its name, and the size of its
// value in bytes.
type Attribute struct {
	Name string
	Size int
}
