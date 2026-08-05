# CLAUDE.md

Guidance for future Claude Code sessions working in this repository.

## What this is

`lsgo` is a colourful `ls` replacement, written in Go with **zero
external dependencies** — standard library only, by design. The single
binary at the repo root (`main.go`) is a thin wrapper; essentially all
logic lives under `internal/`, organised as a pipeline:

```
cli    → parses argv + environment into a fully-resolved Options value
fsx    → filesystem access: File/Dir wrappers, sorting, filtering, Git status
theme  → colour resolution: default palette + LS_COLORS overrides
output → rendering: grid / one-per-line / long table / tree
app    → orchestration: walks the given paths, calls into output per directory
style  → tiny ANSI styling primitive (Colour + Style, chainable, Paint())
```

`internal/*` is intentionally not importable from outside this module —
that's a deliberate boundary, not an oversight; there is no public API
surface here, only a CLI.

## Why zero dependencies

This was built in an environment with no module proxy access, so every
piece — ANSI styling, terminal-width detection, extended-attribute
lookups — is hand-rolled against the standard library rather than reused
from an existing package. That constraint turned out fine long-term too:
the binary is a single static-ish executable with no `go.sum` to audit.
If dependencies ever become available, the most natural candidates to
swap in would be `golang.org/x/term` (for `internal/termwidth`) and
`golang.org/x/sys/unix` (for `internal/xattr` on Darwin, which currently
shells out to the `xattr` CLI tool since Go's `syscall` package doesn't
expose `listxattr`/`getxattr` on that platform the way it does on Linux).

## The view-mode precedence rule (the trickiest part of the CLI)

`-l`/`--long`, `-1`/`--oneline`, `-G`/`--grid`, and `-T`/`--tree` can all
be combined, and **whichever one appears last on the command line wins**
the base view — `lsgo --oneline --long` shows a long table,
`lsgo --long --oneline` shows one-per-line. This is implemented in
`internal/cli/parser.go`'s `ParsedFlags.order` map (every flag occurrence
gets an increasing sequence number) and consumed by
`ParsedFlags.LastOf(...)` in `internal/cli/options.go`'s `deduceMode`.

The full resolution table (see `TestModeSelectionPrecedence` in
`internal/cli/options_test.go` for the exhaustive cases) is:

- No view flag at all → grid.
- Last-of-the-four is `oneline` → one-per-line, full stop (even if
  `--long` appeared earlier).
- Last-of-the-four is `long` → long table.
- Last-of-the-four is `tree` → tree view; it gets a metadata table too
  *only if* `--long` was given anywhere on the line (order doesn't
  matter for that check, just presence).
- Last-of-the-four is `grid` → plain grid, *unless* `--long` was given
  anywhere, in which case it becomes a long table (the "grid" part is a
  known simplification — see below).

Tree recursion is only ever honoured when the final mode is the details
view (`internal/cli/options.go`'s `deduceDirAction` takes a `canTree`
bool for exactly this reason) — `--tree --grid` resolves to plain grid
mode and silently does *not* recurse, since a flat grid has nowhere to
draw branches.

## Known, deliberate simplifications

- **No combined grid-of-tables layout.** When both `--long` and `--grid`
  are effectively active, this implementation falls back to a plain long
  table rather than packing multiple side-by-side detail tables into the
  terminal width. Still fully functional, just not the space-optimised
  layout.
- **No "strict" duplicate-flag checking.** Repeated or conflicting flags
  always resolve last-one-wins; there's no opt-in mode that errors on
  `--sort=name --sort=size` the way some prior art does. This matches
  the default (non-strict) behaviour either way.
- **The `@` extended-attribute indicator only appears with `-@`.** Some
  reference implementations show the `@` marker in permissions whenever
  a file *has* attributes, independent of whether `-@` was passed, and
  only gate the full attribute *listing* on the flag. Doing that here
  would mean an xattr lookup (a subprocess call on Darwin, see below) for
  every file in every `-l` listing, which is a real performance cost for
  a rarely-used feature — so the whole thing, indicator included, is
  gated on `-@`.
- **Directory read errors during `--tree` recursion are silently
  skipped** rather than rendered as an inline `<path: error>` row.
  Getting the tree "last branch" connector right when an error row can
  appear after a file's own row is fiddly for very little payoff (this
  mostly only bites on permission-denied subdirectories).
- **Git status shells out to the `git` binary** (`git rev-parse
  --show-toplevel` to find repo roots, `git status --porcelain=v1
  --ignored=matching -z` to get statuses), cached per repository root in
  `internal/fsx/git.go`. This avoids a libgit2/cgo dependency entirely
  and produces identical results to what a user would get running `git
  status` themselves, at the cost of one process spawn per repository
  per invocation (not per file).

## The grid layout algorithm

`internal/output/grid.go`'s `fitGrid` tries column counts from *most*
down to *fewest*, computing each candidate's column widths and total row
width (cell widths + 2-space gutters), and takes the first (i.e. widest)
one that fits the terminal. If even a single column is too wide for the
terminal (one filename longer than the whole width), it reports failure
and the caller (`RenderGrid`) falls back to one file per line — same
fallback used when stdout isn't a terminal and no width could be
determined at all (`Options.HasWidth` false).

Column-major ("down then across", the default) vs row-major (`-x`/
`--across`) layout is just a different index→(row,column) mapping
(`columnOf` in the same file); the width-fitting logic is identical
either way.

## Tree rendering

`internal/output/tree.go`'s `TreeTrunk` is a small state machine: it
keeps one `TreePart` per currently-open depth level, and on each new row
retroactively fixes up the *previous* row's part at that depth (turning
an `Edge`/`Corner` guess into `Line`/`Blank` once it's known whether more
siblings follow). This is what lets `├──`/`└──`/`│  `/`   ` compose
correctly across arbitrarily deep, arbitrarily wide trees without a
lookahead pass. `internal/output/tree_test.go` walks through the exact
sequences worth knowing if this ever needs touching (two-deep nesting,
siblings closing at different depths, etc.) — read those before changing
`NewRow`.

Rendering detail worth remembering: each `TreePart.asciiArt()` string is
already 3 columns wide and tree segments are concatenated with **no**
gap between them; exactly one separating space is added after the whole
prefix, only when not at the root (`internal/output/details.go`'s
`writeRows`). Adding a per-segment space here is a tempting-looking bug —
don't; it visibly breaks multi-level trees.

## Sorting

Default sort is by name, natural-order (so `file9` sorts before
`file10` — see `internal/fsx/natural.go`), **case-insensitive** by
default (`apps` before `Documents`) — this is a deliberate choice to keep
`lsgo dir` and `lsgo dir/*` (shell-glob-expanded, thus
shell-sorted) producing similar-looking orders; explicit
`--sort=Name` (capital N) opts into case-sensitive/uppercase-first
ordering instead. `--sort=type` sorts by the `fsx.Type` enum's
declaration order (directory, file, link, pipe, socket, char device,
block device, special) as a tiebreaker key, then falls back to natural
name order within a type.

`FileFilter.SortFiles` always runs sort → reverse → directories-first,
in that order (see `internal/fsx/filter.go`) — reversing *then* grouping
directories first, rather than the other way around, is what makes
`--group-directories-first --reverse` still show directories before
files (just each group internally reversed), which is almost certainly
what a user wants from combining those two flags.

## Colour resolution

`internal/theme/theme.go`'s `Build` layers two things, in order:
1. The built-in default palette (`DefaultUIStyles`).
2. `LS_COLORS` pairs (`di=`, `fi=`, `*.txt=`, etc.) — recognised UI keys
   update the palette directly; anything else is treated as a filename
   glob rule.

**Gotcha already hit once:** the SGR-code parser (`internal/theme/lscolors.go`)
must still *consume* the tokens after a `38`/`48` even when the resulting
colour is out of range (e.g. `48;5;999`) — the reference behaviour reads
those tokens off the stream unconditionally before validating them, so
skipping the consume-on-failure step would let the invalid tokens get
reinterpreted as fresh (wrong) SGR codes. `parseHighColour`'s `consumed`
return value exists specifically for this; don't short-circuit it to 0
on failure.

**Gotcha already hit once (size formatting):** `internal/output/sizefmt.go`'s
prefix-scaling loop (`splitPrefix`) counts *completed divisions*, and the
loop's starting index must be `-1`, not `0` — otherwise every size gets
bumped up one magnitude tier (2.1 MB reported as 2.1 GB, etc). This was
caught by `TestRenderSizeDecimal`/`TestRenderSizeBinary`; if those ever
start failing after touching this function, check the off-by-one first.

## Icon glyphs

`internal/theme/icons.go` maps filenames/extensions to Nerd-Font
private-use-area code points. These glyphs render as tofu/placeholder
boxes without a patched font installed, and — critically for
maintenance — **look identical to each other** in most editors regardless
of their actual code point, so a transposed hex digit is invisible on
casual inspection. `internal/theme/icons_test.go` cross-checks every
single entry against a plain hex integer literal (`0xe7b4`, never a rune
literal), specifically so a corrupted code point fails loudly instead of
silently persisting. If you add or change icon mappings, add the
corresponding hex-literal assertion rather than trusting the visual diff.

## Platform-specific files

Build-tag-gated per-OS files exist for three things, each with a Linux
implementation, a Darwin implementation, and a fallback for everything
else (so the whole module still cross-compiles cleanly to Windows, etc.,
just with degraded functionality):

- `internal/fsx/stat_{darwin,linux,other}.go` — inode/link
  count/uid/gid/block count/device numbers/access-change-birth times,
  since `syscall.Stat_t`'s field names and types differ by platform
  (`Nlink` is `uint16` on Darwin vs `uint64` on Linux, Darwin has
  `Birthtimespec` for creation time and Linux's classic `stat(2)`
  doesn't expose one at all, etc).
- `internal/termwidth/{const_darwin,const_linux,termwidth_other}.go` —
  terminal width via a raw `TIOCGWINSZ` ioctl (the request-number
  constant differs by platform: `0x40087468` on Darwin/BSD,
  `0x5413` on Linux). Also doubles as the "is this even a terminal"
  check: a non-terminal fd fails the ioctl.
- `internal/xattr/xattr_{darwin,linux,other}.go` — extended attribute
  listing. Linux gets a direct `syscall.Listxattr`/`Getxattr`. Darwin
  shells out to the `xattr` command-line tool instead (see "why zero
  dependencies" above). Everything else reports no attributes rather
  than erroring.

## Testing

`go test ./...` covers every package except the one-line `main.go`
wrapper. Notable test files if you're about to touch something:

- `internal/output/tree_test.go` — exact branch-character sequences for
  nested trees; treat any diff here as a real regression, not a
  formatting nit.
- `internal/output/grid_test.go` — column-fitting and both fill
  directions (down-then-across vs across).
- `internal/theme/icons_test.go` — see "Icon glyphs" above.
- `internal/cli/options_test.go`'s `TestModeSelectionPrecedence` — the
  full view-mode truth table from "The view-mode precedence rule" above.
- `internal/app/app_test.go` — end-to-end runs against real temp
  directories (colour forced off, since ANSI codes would otherwise
  make string-matching output fragile).

`go vet ./...` and `gofmt -l .` are both clean on the current tree; keep
them that way.

## Extending this

If you add a new flag: `internal/cli/flags.go` (the flag table) →
`internal/cli/parser.go` doesn't need changes for ordinary flags →
`internal/cli/options.go` (deduction logic, likely a new field on
`Options` or `output.TableOptions`) → wire the resolved value through
`internal/app/app.go` into whichever `output` renderer needs it. Try to
keep the `output` package ignorant of `cli` (it already is) — renderers
take plain, already-resolved values, not raw flags, which is what keeps
`internal/output` unit-testable without spinning up argument parsing.
