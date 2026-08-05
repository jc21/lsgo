package termwidth

import (
	"os"
	"testing"
)

// These mostly exercise the "not a terminal" path, since test binaries
// don't normally run attached to one -- but that's still real coverage of
// the syscall plumbing, and Width/IsTerminal are exactly the functions
// meant to report that gracefully rather than erroring.

func TestWidthOnNonTerminal(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if _, ok := Width(r.Fd()); ok {
		t.Log("pipe unexpectedly reported a width (not necessarily wrong, just unusual)")
	}
}

func TestIsTerminalOnNonTerminal(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if IsTerminal(r.Fd()) {
		t.Error("expected a pipe to not be reported as a terminal")
	}
}

func TestStdout(_ *testing.T) {
	// Just exercise the convenience wrapper without asserting a specific
	// outcome, since whether test stdout is a terminal varies by
	// environment.
	Stdout()
}
