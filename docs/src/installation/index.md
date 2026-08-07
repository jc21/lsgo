---
outline: deep
---

# Installation Options

1. Download a prebuilt binary from the [releases](https://github.com/jc21/lsgo/releases)
2. Go path installation
3. Build from source
4. Homebrew manual installation

## Go path installation

```sh
go install github.com/jc21/lsgo@latest
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


## Homebrew manual installation

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
