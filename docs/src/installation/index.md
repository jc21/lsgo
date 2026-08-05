---
outline: deep
---

# Installation

`lsgo` has zero external dependencies, which keeps installation to a single
step: build it with the Go toolchain. There's no package manager listing or
prebuilt binary release yet — building from source is currently the only
supported path, and it's a fast one since there's nothing to fetch.

## Prerequisites

- [Go](https://go.dev/dl/) 1.22 or newer

Check your version with:

```sh
go version
```

## Build from source

Clone the repository and build the binary:

```sh
git clone https://github.com/jc21/lsgo.git
cd lsgo
go build -o lsgo .
```

This produces a single `lsgo` executable in the current directory with no
runtime dependencies. Move it anywhere on your `PATH`, for example:

```sh
sudo mv lsgo /usr/local/bin/
```

## Install straight to your GOPATH

If you'd rather skip the manual move, `go install` places the binary in
`$(go env GOBIN)` (or `$(go env GOPATH)/bin` if `GOBIN` isn't set) — make
sure that directory is on your `PATH`:

```sh
git clone https://github.com/jc21/lsgo.git
cd lsgo
go install .
```

## Verifying the install

```sh
lsgo --version
```

## Cross-compiling

Since `lsgo` has no cgo dependencies on Linux (Darwin's extended-attribute
support shells out to the `xattr` CLI tool rather than using cgo — see the
project's `CLAUDE.md` for why), standard Go cross-compilation works as
expected:

```sh
GOOS=linux GOARCH=arm64 go build -o lsgo-linux-arm64 .
```

## Using `lsgo` as your default `ls`

Once it's on your `PATH`, an alias is the easiest way to reach for it out of
habit. Add this to your shell's rc file (`.bashrc`, `.zshrc`, etc.):

```sh
alias ls='lsgo'
```
