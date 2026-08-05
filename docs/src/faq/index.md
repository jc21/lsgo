---
outline: deep
---

# FAQ

## Why zero external dependencies?

`lsgo` was originally built in an environment with no module proxy access,
so every piece — ANSI styling, terminal-width detection, extended-attribute
lookups — is hand-rolled against the Go standard library rather than
reused from an existing package. That constraint turned out to be worth
keeping: the result is a single static-ish binary with no `go.sum` to
audit. See the [Installation](/installation/) page for build instructions.

## Does it work on Windows?

It cross-compiles cleanly (the platform-specific code — `stat` fields,
terminal-width ioctls, extended attributes — has a fallback implementation
for anything that isn't Linux or Darwin), but that fallback path reports
degraded functionality: no extended-attribute listing, for example.
Linux and Darwin are the actively developed and tested platforms.

## Why do I see boxes or `?` instead of icons?

`--icons` renders glyphs from the private-use area of a
[Nerd Font](https://www.nerdfonts.com/). If your terminal isn't using a
patched Nerd Font, those code points have nothing to render against and
show up as tofu boxes. Install a Nerd Font and select it in your terminal,
or just leave `--icons` off — it's opt-in.

## Why is my sort order different from GNU `ls`?

The default sort is name-based, natural-order, and case-insensitive
(`file9` before `file10`, `apps` before `Documents`), which is different
from GNU `ls`'s default byte-order sort. This is deliberate — see
[View Modes & Sorting](/view-modes/) for the reasoning. Pass
`--sort=Name` (capital N) for case-sensitive, uppercase-first ordering
closer to the traditional behaviour.

## Combining `--long`, `--oneline`, `--grid`, and `--tree` gives a result I didn't expect

Whichever of those four flags appears **last** on the command line decides
the base view. The full precedence table, along with the tree-recursion and
grid-of-tables caveats, is on the [View Modes & Sorting](/view-modes/) page.

## Does `--git` slow down large directories?

`lsgo` shells out to the `git` binary once per repository root per
invocation (`git status --porcelain=v1 --ignored=matching -z`), not once
per file, and caches the result. Listing a single subdirectory of a large
repo costs the same one process spawn as listing the repo root.

## Why does `-@`/`--extended` only show attributes when I pass the flag?

Some other `ls` implementations show an `@` indicator in the permissions
column whenever a file *has* extended attributes, independent of whether
you asked to see them. Doing that here would mean an xattr lookup for
every file in every `-l` listing (a subprocess call on Darwin), which is a
real performance cost for a rarely-used feature — so the whole thing,
indicator included, is gated behind `-@`.

## I found a bug, or want to request a feature

[Open an issue on GitHub](https://github.com/jc21/lsgo/issues). Pull
requests are welcome too — `go test ./...`, `go vet ./...`, and
`gofmt -l .` should all be clean before you open one.
