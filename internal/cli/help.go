package cli

// HelpText is printed for -?/--help.
const HelpText = `Usage:
  lsgo [options] [files...]

lsgo lists the contents of directories, colouring and organising the
output more helpfully than the standard ls.

Display options:
  -1, --oneline           display one entry per line
  -G, --grid              display entries as a grid (default)
  -l, --long              display extended details and attributes
  -R, --recurse           recurse into directories
  -T, --tree              recurse into directories as a tree
  -x, --across            sort the grid across, rather than downwards
  -F, --classify          display a type indicator next to file names
      --color=WHEN        when to colour output: always, auto, never
      --color-scale       highlight levels of file sizes distinctly
      --icons             display icons next to file names
      --no-icons          don't display icons (overrides --icons)

Filtering options:
  -a, --all               show hidden and 'dot' files (pass twice to
                           also show '.' and '..')
  -d, --list-dirs         list directories like regular files
  -L, --level=DEPTH       limit the depth of recursion
  -r, --reverse           reverse the sort order
  -s, --sort=FIELD        which field to sort by (see below)
      --group-directories-first
                           list directories before other files
  -D, --only-dirs         list only directories
      --git-ignore        ignore files mentioned in .gitignore
  -I, --ignore-glob=GLOBS glob patterns (pipe-separated) to ignore

Long view options (only apply with --long):
  -b, --binary            list sizes with binary (1024-based) prefixes
  -B, --bytes             list sizes in plain bytes
  -g, --group             list each file's group
  -h, --header            add a header row
  -H, --links             list each file's hard link count
  -i, --inode             list each file's inode number
  -m, --modified          use the modified timestamp
  -S, --blocks            list each file's number of filesystem blocks
  -t, --time=FIELD        which timestamp to use: modified, changed,
                           accessed, created
  -u, --accessed          use the accessed timestamp
  -U, --created           use the created timestamp
  -@, --extended          list each file's extended attributes
      --changed           use the changed timestamp
      --git               list each file's Git status
      --time-style=STYLE  how to format timestamps: default, iso,
                           long-iso, full-iso
      --no-permissions    suppress the permissions column
      --octal-permissions list permissions in octal
      --no-filesize       suppress the file size column
      --no-user           suppress the user column
      --no-time           suppress the time column
  -n, --numeric           list user/group as numeric IDs

Other:
  -v, --version           show version information and exit
  -?, --help              show this help text and exit

Valid sort fields: name, Name, size, extension, Extension, modified,
changed, accessed, created, inode, type, and none. Fields starting with
a capital letter sort uppercase before lowercase. "modified" has the
aliases date, time, and newest; its reverse ("age", "old", "oldest")
sorts newest-first.
`
