# lsgo

A modern, colourful replacement for `ls`, written in Go with no external
dependencies (standard library only).

`lsgo` colours files by type and extension, understands symlinks and Git
status, and can display its listing as a grid, a single column, a detailed
table, or a tree.

## Install / build

```sh
go install github.com/jc21/lsgo@latest

# or

git clone https://github.com/jc21/lsgo.git
cd lsgo
go build -o lsgo .
```

This produces a single `lsgo` binary with no runtime dependencies.


## Install via Homebrew

```bash
export HOMEBREW_NO_INSTALL_FROM_API=1
brew update
brew tap --force homebrew/core
cd $(brew --repo homebrew/core)
git remote add jc21 https://github.com/jc21/homebrew-core.git
git fetch --all
git checkout jc21/lsgo
brew install --build-from-source lsgo
# and if you need to rebuild:
brew reinstall --build-from-source lsgo
# Switch back to homebrew-core master
git checkout master
```


## Usage

```
lsgo [options] [files...]
```

With no arguments, it lists the current directory as a grid.

### Display options

| Flag | Meaning |
|---|---|
| `-1`, `--oneline` | one entry per line |
| `-G`, `--grid` | grid layout (default) |
| `-l`, `--long` | detailed table view |
| `-R`, `--recurse` | recurse into directories |
| `-T`, `--tree` | recurse into directories as a tree |
| `-x`, `--across` | fill the grid row-first instead of column-first |
| `-F`, `--classify` | append `/`, `*`, `@`, `\|`, or `=` to indicate file type |
| `--color=WHEN` | `always`, `auto` (default), or `never` |
| `--color-scale` | shade file sizes by magnitude |
| `--icons` / `--no-icons` | show/hide an icon glyph before each name |

### Filtering and sorting

| Flag | Meaning |
|---|---|
| `-a`, `--all` | show dotfiles (pass twice to also show `.` and `..`) |
| `-d`, `--list-dirs` | list directories as plain entries, don't descend |
| `-L`, `--level=N` | limit recursion depth |
| `-r`, `--reverse` | reverse the sort order |
| `-s`, `--sort=FIELD` | sort field: `name`, `size`, `extension`, `modified`, `accessed`, `created`, `inode`, `type`, `none`, and more (see `--help`) |
| `--group-directories-first` | list directories before files |
| `-D`, `--only-dirs` | list only directories |
| `--git-ignore` | hide files ignored by Git |
| `-I`, `--ignore-glob=GLOBS` | pipe-separated glob patterns to hide |

### Long view (`-l`) options

| Flag | Meaning |
|---|---|
| `-b`, `--binary` | sizes with binary (1024) prefixes |
| `-B`, `--bytes` | exact byte counts |
| `-g`, `--group` | show the group column |
| `-h`, `--header` | show a header row |
| `-H`, `--links` | show hard link counts |
| `-i`, `--inode` | show inode numbers |
| `-m`/`-u`/`-U`/`--changed` | show modified/accessed/created/changed time |
| `-t`, `--time=FIELD` | pick which timestamp to show |
| `--time-style=STYLE` | `default`, `iso`, `long-iso`, `full-iso` |
| `-S`, `--blocks` | show filesystem block counts |
| `-@`, `--extended` | show extended attributes |
| `--git` | show a two-character Git status column |
| `--octal-permissions` | show permissions in octal |
| `-n`, `--numeric` | show user/group as numeric IDs |
| `--no-permissions`, `--no-filesize`, `--no-user`, `--no-time` | hide the respective column |

Run `lsgo --help` for the full list.

## Environment variables

- `LS_COLORS` — customise the colour scheme.
- `NO_COLOR` — disable colour entirely when set (any value).
- `COLUMNS` — override the detected terminal width.
- `TIME_STYLE` — default for `--time-style`.
- `LS_ICON_SPACING` — number of spaces between an icon and its filename.

## Development

```sh
go build ./...      # build everything
go test ./...        # run the test suite
go vet ./...          # static checks
gofmt -l .            # formatting check
```

See [CLAUDE.md](CLAUDE.md) for a tour of the codebase's architecture and
design decisions.
