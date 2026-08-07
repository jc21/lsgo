// Command lsgo lists directory contents, in the same spirit as ls --
// colourised by file type, aware of symlinks and Git status, and able to
// show its output as a grid, a single column, a detailed table, or a
// tree.
package main

import (
	"fmt"
	"os"

	"github.com/jc21/lsgo/internal/app"
	"github.com/jc21/lsgo/internal/cli"
	"github.com/jc21/lsgo/internal/termwidth"
)

var Version = "0.0.0" // overridden at build time

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(argv []string, stdout, stderr *os.File) int {
	isTerminal := termwidth.IsTerminal(stdout.Fd())
	probe := func() (int, bool) { return termwidth.Width(stdout.Fd()) }

	result, err := cli.ParseArgs(argv, cli.OSEnv, probe)
	if err != nil {
		fmt.Fprintf(stderr, "lsgo: %s\n", err)
		return app.ExitOptionsErr
	}

	if result.ShowHelp {
		fmt.Fprint(stdout, cli.HelpText)
		return app.ExitSuccess
	}
	if result.ShowVersion {
		fmt.Fprintf(stdout, "lsgo %s\n", Version)
		return app.ExitSuccess
	}

	lsColors, _ := os.LookupEnv("LS_COLORS")

	a := app.New(result.Options, isTerminal, lsColors, stdout, stderr)
	return a.Run()
}
