---
outline: deep
---

# Usage & Flags

```
lsgo [options] [files...]
```

With no arguments, `lsgo` lists the current directory as a grid. Pass one
or more paths to list specific files or directories instead.

This page is a reference for every flag. `lsgo --help` always reflects the
exact wording shipped in your build, so treat that as the final word if
anything here ever drifts.

## Display options

| Flag | Meaning |
|---|---|
| `-1`, `--oneline` | display one entry per line |
| `-G`, `--grid` | display entries as a grid (default) |
| `-l`, `--long` | display extended details and attributes |
| `-R`, `--recurse` | recurse into directories |
| `-T`, `--tree` | recurse into directories as a tree |
| `-x`, `--across` | fill the grid across, rather than downwards |
| `-F`, `--classify` | display a type indicator (`/`, `*`, `@`, `\|`, `=`) next to file names |
| `--color=WHEN` | when to colour output: `always`, `auto` (default), `never` |
| `--color-scale` | highlight levels of file sizes distinctly |
| `--icons` | display icons next to file names |
| `--no-icons` | don't display icons (overrides `--icons`) |

`-1`, `-G`, `-l`, and `-T` are all "view mode" flags — see
[View Modes & Sorting](/view-modes/) for exactly how they interact when more
than one is given on the same command line.

## Filtering options

| Flag | Meaning |
|---|---|
| `-a`, `--all` | show hidden and 'dot' files (pass twice to also show `.` and `..`) |
| `-d`, `--list-dirs` | list directories like regular files, don't descend |
| `-L`, `--level=DEPTH` | limit the depth of recursion |
| `-r`, `--reverse` | reverse the sort order |
| `-s`, `--sort=FIELD` | which field to sort by — see [View Modes & Sorting](/view-modes/) |
| `--group-directories-first` | list directories before other files |
| `-D`, `--only-dirs` | list only directories |
| `--git-ignore` | ignore files mentioned in `.gitignore` |
| `-I`, `--ignore-glob=GLOBS` | glob patterns (pipe-separated) to ignore |

## Long view options

These only take effect alongside `-l`/`--long`.

| Flag | Meaning |
|---|---|
| `-b`, `--binary` | list sizes with binary (1024-based) prefixes |
| `-B`, `--bytes` | list sizes in plain bytes |
| `-g`, `--group` | list each file's group |
| `-h`, `--header` | add a header row |
| `-H`, `--links` | list each file's hard link count |
| `-i`, `--inode` | list each file's inode number |
| `-m`, `--modified` | use the modified timestamp |
| `-S`, `--blocks` | list each file's number of filesystem blocks |
| `-t`, `--time=FIELD` | which timestamp to use: `modified`, `changed`, `accessed`, `created` |
| `-u`, `--accessed` | use the accessed timestamp |
| `-U`, `--created` | use the created timestamp |
| `-@`, `--extended` | list each file's extended attributes |
| `--changed` | use the changed timestamp |
| `--git` | list each file's Git status |
| `--time-style=STYLE` | how to format timestamps: `default`, `iso`, `long-iso`, `full-iso` |
| `--no-permissions` | suppress the permissions column |
| `--octal-permissions` | list permissions in octal |
| `--no-filesize` | suppress the file size column |
| `--no-user` | suppress the user column |
| `--no-time` | suppress the time column |
| `-n`, `--numeric` | list user/group as numeric IDs |

## Other

| Flag | Meaning |
|---|---|
| `-v`, `--version` | show version information and exit |
| `-?`, `--help` | show the full help text and exit |

## Examples

Detailed table, newest files first, with a header row:

```sh
lsgo -lh --sort=newest
```

A tree, three levels deep, skipping anything Git already ignores:

```sh
lsgo -T -L3 --git-ignore
```

Grid view with icons and Git status baked into a long table:

```sh
lsgo -l --git --icons
```

One entry per line, hidden files included, directories grouped first:

```sh
lsgo -1a --group-directories-first
```
