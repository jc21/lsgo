---
outline: deep
---

# Guide

`lsgo` is a colourful `ls` replacement, written in Go with **zero external
dependencies** — standard library only, by design. It colours files by type
and extension, understands symlinks and Git status, and can display its
listing as a grid, a single column, a detailed table, or a tree.

- [Installation](/installation/)
- [Usage & Flags](/usage/)
- [View Modes & Sorting](/view-modes/)
- [Colours & Theming](/colours/)

## Project Goal

`ls` replacements are common, but most of them lean on a handful of
third-party crates or packages for terminal width detection, colour parsing,
or extended attributes. `lsgo` was built in an environment with no module
proxy access, so every one of those pieces is hand-rolled against the Go
standard library instead. The result is a single binary with no `go.sum` to
audit, and that constraint turned out to be worth keeping even where it
wasn't strictly necessary.

## Features

- Four view modes: grid (default), one-per-line, long table, and tree —
  freely combinable, with [predictable last-flag-wins precedence](/view-modes/)
  when more than one is given
- Colours files by type and extension out of the box, fully overridable via
  the standard `LS_COLORS` environment variable
- Optional Nerd Font icon glyphs next to each filename
- Git-aware: an optional per-file status column (`--git`), and the option to
  hide anything your repository already ignores (`--git-ignore`)
- Natural-order, case-insensitive name sorting by default (`file9` before
  `file10`, `apps` before `Documents`), with size, extension, timestamp,
  type, and no-op sort fields available via `--sort`
- A detailed long view with per-column opt-outs, extended attribute listing,
  binary/decimal size prefixes, configurable timestamp fields and styles,
  and numeric or named user/group columns

## Quick Start

Build from source with the Go toolchain:

```sh
go build -o lsgo .
```

This produces a single `lsgo` binary with no runtime dependencies. See
[Installation](/installation/) for other ways to get it onto your `PATH`.

Once installed, run it with no arguments to list the current directory as a
grid:

```sh
lsgo
```

Add `-l` for a detailed table, `-T` for a tree, or `-1` for one entry per
line:

```sh
lsgo -l
lsgo -T
lsgo -1
```

Run `lsgo --help` at any time for the full, authoritative flag reference.

## Contributing

Pull requests are welcome. `go test ./...`, `go vet ./...`, and
`gofmt -l .` should all be clean before you open one — see the project's
`CLAUDE.md` for a tour of the codebase's architecture and the reasoning
behind its more deliberate design decisions.

## Getting Support

[Found a bug, or have a feature request?](https://github.com/jc21/lsgo/issues)
Open an issue on GitHub.
