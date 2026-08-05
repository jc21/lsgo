//go:build darwin

package termwidth

// TIOCGWINSZ on Darwin/BSD.
const tiocgwinszValue = 0x40087468
