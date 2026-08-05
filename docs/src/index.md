---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "lsgo"
  tagline: A modern, colourful `ls` replacement, written in Go with zero external dependencies
  image:
    src: /logo.svg
    alt: lsgo logo
  actions:
    - theme: brand
      text: Get Started
      link: /guide/
    - theme: alt
      text: GitHub
      link: https://github.com/jc21/lsgo

features:
  - title: Four Views
    details: Grid, one-per-line, long-table, and tree layouts. Combine the flags and whichever appears last on the command line wins.
  - title: Colourful by Default
    details: Files are coloured by type and extension out of the box, with full LS_COLORS support to override or extend the palette.
  - title: Git-Aware
    details: Pass --git to see a per-file status column, or --git-ignore to hide anything your repository already ignores.
  - title: Zero Dependencies
    details: Standard library only. A single static-ish binary with no go.sum to audit and nothing else to install.
  - title: Nerd Font Icons
    details: Optional per-file icon glyphs, sensible natural-order sorting, and a long view with as many or as few columns as you want.
  - title: Cross-Platform
    details: Built and tested on Linux and Darwin, with graceful fallbacks so the module still cross-compiles cleanly to Windows and beyond.
---
