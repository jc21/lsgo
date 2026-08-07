package theme

import "github.com/jc21/lsgo/internal/fsx"

// Icon code points are written as explicit \u/\U escapes (rather than
// literal glyphs) throughout this file, since they're drawn from Nerd
// Font's private-use-area ranges and don't render meaningfully as plain
// text -- an escape is the only unambiguous way to write them down.
const (
	IconAudio    = ''
	IconImage    = ''
	IconVideo    = ''
	IconFolder   = ''
	IconFile     = '' // shown for files with no extension at all
	IconFallback = '' // shown for an extension with no icon mapping
)

// iconsByName covers files and directories that are recognised by their
// exact name rather than their extension -- dotfiles, and directories like
// "node_modules" that have a well-known meaning.
//
// Names recur across this map, the extension map below, filetype.go's
// classification lists, and icons_test.go's independent cross-check -- see
// the comment on iconsByExtension below for why goconst is suppressed here.
//
//nolint:goconst
var iconsByName = map[string]rune{
	".Trash":             '',
	".atom":              '',
	".bashprofile":       '',
	".bashrc":            '',
	".git":               '',
	".gitattributes":     '',
	".gitconfig":         '',
	".github":            '',
	".gitignore":         '',
	".gitmodules":        '',
	".rvm":               '',
	".vimrc":             '',
	".vscode":            '',
	".zshrc":             '',
	"Cargo.lock":         '',
	"bin":                '',
	"config":             '',
	"docker-compose.yml": '',
	"Dockerfile":         '',
	"ds_store":           '',
	"gitignore_global":   '',
	"go.mod":             '',
	"go.sum":             '',
	"gradle":             '',
	"gruntfile.coffee":   '',
	"gruntfile.js":       '',
	"gruntfile.ls":       '',
	"gulpfile.coffee":    '',
	"gulpfile.js":        '',
	"gulpfile.ls":        '',
	"hidden":             '',
	"include":            '',
	"lib":                '',
	"localized":          '',
	"Makefile":           '',
	"node_modules":       '',
	"npmignore":          '',
	"PKGBUILD":           '',
	"rubydoc":            '',
	"yarn.lock":          '',
}

// directoryIconsByName covers directories with a special icon, checked
// only once a path is known to point to a directory.
var directoryIconsByName = map[string]rune{
	"bin":   '',
	".git":  '',
	".idea": '',
}

// iconsByExtension is lsgo's big "what kind of file is this, by extension"
// icon table.
//
// icons_test.go deliberately duplicates every entry here as an independent
// hex-literal cross-check (see CLAUDE.md "Icon glyphs" -- these private-use
// glyphs look identical to each other regardless of code point, so a
// transposed hex digit needs an independently-typed reference to catch it).
// Extension strings also recur in filetype.go's classification lists.
// Constant-izing them would hide these plain, scannable literals behind
// indirection for no correctness benefit, so dupl and goconst are
// suppressed for this table and its icons_test.go counterpart.
//
//nolint:dupl,goconst
var iconsByExtension = map[string]rune{
	"ai": '', "android": '', "apk": '', "apple": '',
	"avi": '', "avif": '', "avro": '', "awk": '',
	"bash": '', "bash_history": '', "bash_profile": '',
	"bashrc": '', "bat": '', "bats": '', "bmp": '',
	"bz": '', "bz2": '', "c": '', "c++": '',
	"cab": '', "cc": '', "cfg": '', "class": '',
	"clj": '', "cljs": '', "cls": '', "cmd": '',
	"coffee": '', "conf": '', "cp": '', "cpio": '',
	"cpp": '', "cs": '\U000f031b', "csh": '', "cshtml": '',
	"csproj": '\U000f031b', "css": '', "csv": '', "csx": '\U000f031b',
	"cxx": '', "d": '', "dart": '', "db": '',
	"deb": '', "diff": '', "djvu": '', "dll": '',
	"doc": '', "docx": '', "ds_store": '', "DS_store": '',
	"dump": '', "ebook": '', "ebuild": '', "editorconfig": '',
	"ejs": '', "elm": '', "env": '', "eot": '',
	"epub": '', "erb": '', "erl": '', "ex": '',
	"exe": '', "exs": '', "fish": '', "flac": '',
	"flv": '', "font": '', "fs": '', "fsi": '',
	"fsx": '', "gdoc": '', "gem": '', "gemfile": '',
	"gemspec": '', "gform": '', "gif": '', "git": '',
	"gitattributes": '', "gitignore": '', "gitmodules": '',
	"go": '', "gradle": '', "groovy": '', "gsheet": '',
	"gslides": '', "guardfile": '', "gz": '', "h": '',
	"hbs": '', "hpp": '', "hs": '', "htm": '',
	"html": '', "hxx": '', "ico": '', "image": '',
	"img": '', "iml": '', "ini": '', "ipynb": '',
	"iso": '', "j2c": '', "j2k": '', "jad": '',
	"jar": '', "java": '', "jfi": '', "jfif": '',
	"jif": '', "jl": '', "jmd": '', "jp2": '',
	"jpe": '', "jpeg": '', "jpg": '', "jpx": '',
	"js": '', "json": '', "jsx": '', "jxl": '',
	"ksh": '', "latex": '', "less": '', "lhs": '',
	"license": '\U000f0219', "localized": '', "lock": '', "log": '',
	"lua": '', "lz": '', "lz4": '', "lzh": '',
	"lzma": '', "lzo": '', "m": '', "mm": '',
	"m4a": '', "markdown": '', "md": '', "mjs": '',
	"mk": '', "mkd": '', "mkv": '', "mobi": '',
	"mov": '', "mp3": '', "mp4": '', "msi": '',
	"mustache": '', "nix": '', "node": '\U000f0399', "npmignore": '',
	"odp": '', "ods": '', "odt": '', "ogg": '',
	"ogv": '', "otf": '', "part": '', "patch": '',
	"pdf": '', "php": '', "pl": '', "plx": '',
	"pm": '', "png": '', "pod": '', "ppt": '',
	"pptx": '', "procfile": '', "properties": '', "ps1": '',
	"psd": '', "pxm": '', "py": '', "pyc": '',
	"r": '', "rakefile": '', "rar": '', "razor": '',
	"rb": '', "rdata": '', "rdb": '', "rdoc": '',
	"rds": '', "readme": '', "rlib": '', "rmd": '',
	"rpm": '', "rs": '', "rspec": '', "rspec_parallel": '',
	"rspec_status": '', "rss": '', "rtf": '\U000f0219', "ru": '',
	"rubydoc": '', "sass": '', "scala": '', "scss": '',
	"sh": '', "shell": '', "slim": '', "sln": '',
	"so": '', "sql": '', "sqlite3": '', "sty": '',
	"styl": '', "stylus": '', "svg": '', "swift": '',
	"t": '', "tar": '', "taz": '', "tbz": '',
	"tbz2": '', "tex": '', "tgz": '', "tiff": '',
	"tlz": '', "toml": '', "torrent": '', "ts": '',
	"tsv": '', "tsx": '', "ttf": '', "twig": '',
	"txt": '', "txz": '', "tz": '', "tzo": '',
	"video": '', "vim": '', "vue": '\U000f0844', "war": '',
	"wav": '', "webm": '', "webp": '', "windows": '',
	"woff": '', "woff2": '', "xhtml": '', "xls": '',
	"xlsx": '', "xml": '\U000f05c0', "xul": '\U000f05c0', "xz": '',
	"yaml": '', "yml": '', "zip": '', "zsh": '',
	"zsh-theme": '', "zshrc": '', "zst": '',
}

// IconForFile picks the icon glyph to display next to a file's name,
// consulting (in order) the exact-name map, directory special-cases, the
// FileIconer (extension-class icons like audio/image/video), the
// extension map, and finally a generic file/folder fallback.
func IconForFile(f *fsx.File, iconer FileIconer) rune {
	if icon, ok := iconsByName[f.Name]; ok {
		return icon
	}

	if f.PointsToDirectory() {
		if icon, ok := directoryIconsByName[f.Name]; ok {
			return icon
		}
		return IconFolder
	}

	if iconer != nil {
		if icon, ok := iconer.IconFile(f); ok {
			return icon
		}
	}

	if f.Ext != "" {
		if icon, ok := iconsByExtension[f.Ext]; ok {
			return icon
		}
		return IconFallback
	}

	return IconFile
}
