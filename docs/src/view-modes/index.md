---
outline: deep
---

# View Modes & Sorting

## The view-mode precedence rule

`-l`/`--long`, `-1`/`--oneline`, `-G`/`--grid`, and `-T`/`--tree` can all be
combined on the same command line, and **whichever one appears last wins**
the base view — `lsgo --oneline --long` shows a long table,
`lsgo --long --oneline` shows one-per-line.

The full resolution table:

| Last-of-the-four flag | Result |
|---|---|
| *(none given)* | Grid |
| `--oneline` | One-per-line, full stop — even if `--long` appeared earlier |
| `--long` | Long table |
| `--tree` | Tree view; gains a metadata table too **only if** `--long` was given *anywhere* on the line (order doesn't matter for that check, just presence) |
| `--grid` | Plain grid, **unless** `--long` was given anywhere, in which case it becomes a long table (see below) |

::: tip Recursion only follows the details view
Tree recursion is only ever honoured when the final resolved mode is a
details view. `--tree --grid` resolves to plain grid mode and silently does
*not* recurse, since a flat grid has nowhere to draw branches.
:::

## Known simplification: no combined grid-of-tables layout

When both `--long` and `--grid` are effectively active, `lsgo` falls back to
a plain long table rather than packing multiple side-by-side detail tables
into the terminal width the way some other implementations do. Still fully
functional, just not the space-optimised layout.

There's also no "strict" duplicate-flag checking — repeated or conflicting
flags always resolve last-one-wins, matching the default (non-strict)
behaviour of most other `ls` implementations. There's no opt-in mode that
errors on something like `--sort=name --sort=size`.

## Sorting

The default sort is by name, **natural order** (so `file9` sorts before
`file10`), and **case-insensitive** by default (`apps` before `Documents`).
This is a deliberate choice to keep `lsgo dir` and `lsgo dir/*`
(shell-glob-expanded, thus shell-sorted) producing similarly-ordered output.

Pass a capitalised `--sort=Name` to opt into case-sensitive,
uppercase-before-lowercase ordering instead.

### Sort fields

| `--sort=` value | Behaviour |
|---|---|
| `name` (default) | Natural order, case-insensitive |
| `Name` | Natural order, case-sensitive (uppercase first) |
| `size` | Smallest to largest |
| `extension` / `Extension` | By file extension, case-insensitive / case-sensitive |
| `modified` | By modification time — aliases: `date`, `time`, `newest` |
| `changed` | By inode-change time |
| `accessed` | By last-access time |
| `created` | By creation time |
| `inode` | By inode number |
| `type` | By file-kind (directory, file, link, pipe, socket, char device, block device, special, in that order), falling back to natural name order within a type |
| `none` | No sorting — directory read order |

Reversing a time-based sort with `-r`/`--reverse` flips it to newest-last;
`modified`'s own reverse aliases (`age`, `old`, `oldest`) sort newest-first
without needing `--reverse` at all.

### Combining `--reverse` with `--group-directories-first`

Sorting always runs in this order: **sort → reverse → group directories
first**. Reversing *then* grouping directories, rather than the other way
around, is what makes `--group-directories-first --reverse` still show
directories before files — just with each group's *contents* reversed
internally — which is almost certainly what you want from combining those
two flags.
