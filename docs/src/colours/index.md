---
outline: deep
---

# Colours & Theming

`lsgo` colours everything — filenames, permission bits, sizes, dates, Git
status — out of the box, then layers any `LS_COLORS` overrides on top.

## When colour is used

By default (`--color=auto`), colour is only emitted when standard output is
a terminal — piping `lsgo` into another program or a file produces plain
text automatically. Override this with:

```sh
lsgo --color=always   # force colour, even when piped
lsgo --color=never    # never colour, even in a terminal
```

Setting the `NO_COLOR` environment variable to any value disables colour
entirely, regardless of `--color`, following the [no-color.org](https://no-color.org/)
convention.

## `LS_COLORS`

`lsgo` reads the standard `LS_COLORS` environment variable — the same one
GNU `ls`, `exa`, `eza`, and friends use — and understands both the
recognised UI keys and arbitrary filename glob rules.

Format: a `:`-separated list of `key=value` pairs, where each value is a
`;`-separated list of SGR parameter codes (the same numbers used in raw
ANSI escapes, without the `\x1b[` / `m`):

```sh
export LS_COLORS="di=1;34:ln=36:*.md=1;33:*.log=2"
```

### Recognised UI keys

| Key | Applies to |
|---|---|
| `di` | Directories |
| `fi` | Regular files |
| `ln` | Symbolic links |
| `pi` | Named pipes (FIFOs) |
| `so` | Sockets |
| `bd` | Block devices |
| `cd` | Character devices |
| `ex` | Executable files |
| `or` | Broken/orphaned symlinks |

Any other key (`*.txt=`, `*.tar.gz=`, etc.) is treated as a filename glob
rule rather than a UI element, and applies on top of `lsgo`'s built-in
per-extension colouring. Keys `LS_COLORS` defines that have no equivalent
here (`MULTIHARDLINK`, `DOOR`, and similar) are accepted without error but
don't change anything, so a full `dircolors`-style `LS_COLORS` string
doesn't produce spurious warnings.

### Supported SGR codes

| Code(s) | Meaning |
|---|---|
| `1` `2` `3` `4` `5` `7` `8` `9` | Bold, dim, italic, underline, blink, reverse, hidden, strikethrough |
| `30`–`37` | Standard foreground colours |
| `40`–`47` | Standard background colours |
| `38;5;N` / `48;5;N` | 256-colour palette foreground/background (`N` 0–255) |
| `38;2;R;G;B` / `48;2;R;G;B` | 24-bit truecolour foreground/background |

Unrecognised or malformed codes are ignored rather than raising an error,
matching the tolerant parsing of the reference `LS_COLORS` implementations.

## `--color-scale`

Shades file sizes by magnitude instead of a single fixed colour, so it's
easier to spot the largest files in a long listing at a glance:

```sh
lsgo -l --color-scale
```

## Icons

Pass `--icons` to show a small glyph before each filename, resolved by
extension and by well-known filenames (`Makefile`, `Dockerfile`, and so
on). These are Nerd Font private-use-area code points — they render as
tofu/placeholder boxes without a [patched Nerd Font](https://www.nerdfonts.com/)
installed in your terminal.

`--no-icons` always wins if both are given.

Control the gap between an icon and its filename with:

```sh
export LS_ICON_SPACING=2   # default is 1
```

## Other relevant environment variables

| Variable | Effect |
|---|---|
| `LS_COLORS` | Customise the colour scheme, as above |
| `NO_COLOR` | Disable colour entirely when set, to any value |
| `COLUMNS` | Override the detected terminal width |
| `TIME_STYLE` | Default for `--time-style` (`default`, `iso`, `long-iso`, `full-iso`) |
| `LS_ICON_SPACING` | Number of spaces between an icon and its filename |
