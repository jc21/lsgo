package theme

import "testing"

// TestIconCodePoints cross-checks every icon in the by-name and
// by-extension tables against its expected Unicode code point.
//
// Every expected value below is a plain hex integer literal (0xXXXX),
// deliberately never a character literal. Private-use-area glyphs render
// identically -- as a font's arbitrary picture for that slot, or a tofu
// box -- regardless of their actual code point, so a reference table
// built by "typing what it looks like" can't catch a transposed hex
// digit. A bare integer can't be misread the same way.
func TestIconCodePoints(t *testing.T) {
	wantByName := map[string]uint32{
		".Trash": 0xf1f8, ".atom": 0xe764, ".bashprofile": 0xe615,
		".bashrc": 0xf489, ".git": 0xf1d3, ".gitattributes": 0xf1d3,
		".gitconfig": 0xf1d3, ".github": 0xf408, ".gitignore": 0xf1d3,
		".gitmodules": 0xf1d3, ".rvm": 0xe21e, ".vimrc": 0xe62b,
		".vscode": 0xe70c, ".zshrc": 0xf489, "Cargo.lock": 0xe7a8,
		"bin": 0xe5fc, "config": 0xe5fc, "docker-compose.yml": 0xf308,
		"Dockerfile": 0xf308, "ds_store": 0xf179, "gitignore_global": 0xf1d3,
		"go.mod": 0xe626, "go.sum": 0xe626, "gradle": 0xe256,
		"gruntfile.coffee": 0xe611, "gruntfile.js": 0xe611, "gruntfile.ls": 0xe611,
		"gulpfile.coffee": 0xe610, "gulpfile.js": 0xe610, "gulpfile.ls": 0xe610,
		"hidden": 0xf023, "include": 0xe5fc, "lib": 0xf121,
		"localized": 0xf179, "Makefile": 0xf489, "node_modules": 0xe718,
		"npmignore": 0xe71e, "PKGBUILD": 0xf303, "rubydoc": 0xe73b,
		"yarn.lock": 0xe718,
	}
	checkMap(t, "iconsByName", runify(wantByName), iconsByName)

	wantDirs := map[string]uint32{
		"bin": 0xe5fc, ".git": 0xf1d3, ".idea": 0xe7b5,
	}
	checkMap(t, "directoryIconsByName", runify(wantDirs), directoryIconsByName)

	wantSpecial := map[string]uint32{
		"IconAudio": 0xf001, "IconImage": 0xf1c5, "IconVideo": 0xf03d,
		"IconFolder": 0xf115, "IconFile": 0xf016, "IconFallback": 0xf15b,
	}
	gotSpecial := map[string]rune{
		"IconAudio": IconAudio, "IconImage": IconImage, "IconVideo": IconVideo,
		"IconFolder": IconFolder, "IconFile": IconFile, "IconFallback": IconFallback,
	}
	checkMap(t, "special icons", runify(wantSpecial), gotSpecial)

	// The five entries using 0xf0XXX code points live in Unicode's
	// Supplementary Private Use Area-A (above 0xFFFF).
	wantByExt := map[string]uint32{
		"ai": 0xe7b4, "android": 0xe70e, "apk": 0xe70e, "apple": 0xf179,
		"avi": 0xf03d, "avif": 0xf1c5, "avro": 0xe60b, "awk": 0xf489,
		"bash": 0xf489, "bash_history": 0xf489, "bash_profile": 0xf489,
		"bashrc": 0xf489, "bat": 0xf17a, "bats": 0xf489, "bmp": 0xf1c5,
		"bz": 0xf410, "bz2": 0xf410, "c": 0xe61e, "c++": 0xe61d,
		"cab": 0xe70f, "cc": 0xe61d, "cfg": 0xe615, "class": 0xe256,
		"clj": 0xe768, "cljs": 0xe76a, "cls": 0xf034, "cmd": 0xe70f,
		"coffee": 0xf0f4, "conf": 0xe615, "cp": 0xe61d, "cpio": 0xf410,
		"cpp": 0xe61d, "cs": 0xf031b, "csh": 0xf489, "cshtml": 0xf1fa,
		"csproj": 0xf031b, "css": 0xe749, "csv": 0xf1c3, "csx": 0xf031b,
		"cxx": 0xe61d, "d": 0xe7af, "dart": 0xe798, "db": 0xf1c0,
		"deb": 0xe77d, "diff": 0xf440, "djvu": 0xf02d, "dll": 0xe70f,
		"doc": 0xf1c2, "docx": 0xf1c2, "ds_store": 0xf179, "DS_store": 0xf179,
		"dump": 0xf1c0, "ebook": 0xe28b, "ebuild": 0xf30d, "editorconfig": 0xe615,
		"ejs": 0xe618, "elm": 0xe62c, "env": 0xf462, "eot": 0xf031,
		"epub": 0xe28a, "erb": 0xe73b, "erl": 0xe7b1, "ex": 0xe62d,
		"exe": 0xf17a, "exs": 0xe62d, "fish": 0xf489, "flac": 0xf001,
		"flv": 0xf03d, "font": 0xf031, "fs": 0xe7a7, "fsi": 0xe7a7,
		"fsx": 0xe7a7, "gdoc": 0xf1c2, "gem": 0xe21e, "gemfile": 0xe21e,
		"gemspec": 0xe21e, "gform": 0xf298, "gif": 0xf1c5, "git": 0xf1d3,
		"gitattributes": 0xf1d3, "gitignore": 0xf1d3, "gitmodules": 0xf1d3,
		"go": 0xe626, "gradle": 0xe256, "groovy": 0xe775, "gsheet": 0xf1c3,
		"gslides": 0xf1c4, "guardfile": 0xe21e, "gz": 0xf410, "h": 0xf0fd,
		"hbs": 0xe60f, "hpp": 0xf0fd, "hs": 0xe777, "htm": 0xf13b,
		"html": 0xf13b, "hxx": 0xf0fd, "ico": 0xf1c5, "image": 0xf1c5,
		"img": 0xe271, "iml": 0xe7b5, "ini": 0xf17a, "ipynb": 0xe678,
		"iso": 0xe271, "j2c": 0xf1c5, "j2k": 0xf1c5, "jad": 0xe256,
		"jar": 0xe256, "java": 0xe256, "jfi": 0xf1c5, "jfif": 0xf1c5,
		"jif": 0xf1c5, "jl": 0xe624, "jmd": 0xf48a, "jp2": 0xf1c5,
		"jpe": 0xf1c5, "jpeg": 0xf1c5, "jpg": 0xf1c5, "jpx": 0xf1c5,
		"js": 0xe74e, "json": 0xe60b, "jsx": 0xe7ba, "jxl": 0xf1c5,
		"ksh": 0xf489, "latex": 0xf034, "less": 0xe758, "lhs": 0xe777,
		"license": 0xf0219, "localized": 0xf179, "lock": 0xf023, "log": 0xf18d,
		"lua": 0xe620, "lz": 0xf410, "lz4": 0xf410, "lzh": 0xf410,
		"lzma": 0xf410, "lzo": 0xf410, "m": 0xe61e, "mm": 0xe61d,
		"m4a": 0xf001, "markdown": 0xf48a, "md": 0xf48a, "mjs": 0xe74e,
		"mk": 0xf489, "mkd": 0xf48a, "mkv": 0xf03d, "mobi": 0xe28b,
		"mov": 0xf03d, "mp3": 0xf001, "mp4": 0xf03d, "msi": 0xe70f,
		"mustache": 0xe60f, "nix": 0xf313, "node": 0xf0399, "npmignore": 0xe71e,
		"odp": 0xf1c4, "ods": 0xf1c3, "odt": 0xf1c2, "ogg": 0xf001,
		"ogv": 0xf03d, "otf": 0xf031, "part": 0xf43a, "patch": 0xf440,
		"pdf": 0xf1c1, "php": 0xe73d, "pl": 0xe769, "plx": 0xe769,
		"pm": 0xe769, "png": 0xf1c5, "pod": 0xe769, "ppt": 0xf1c4,
		"pptx": 0xf1c4, "procfile": 0xe21e, "properties": 0xe60b, "ps1": 0xf489,
		"psd": 0xe7b8, "pxm": 0xf1c5, "py": 0xe606, "pyc": 0xe606,
		"r": 0xf25d, "rakefile": 0xe21e, "rar": 0xf410, "razor": 0xf1fa,
		"rb": 0xe21e, "rdata": 0xf25d, "rdb": 0xe76d, "rdoc": 0xf48a,
		"rds": 0xf25d, "readme": 0xf48a, "rlib": 0xe7a8, "rmd": 0xf48a,
		"rpm": 0xe7bb, "rs": 0xe7a8, "rspec": 0xe21e, "rspec_parallel": 0xe21e,
		"rspec_status": 0xe21e, "rss": 0xf09e, "rtf": 0xf0219, "ru": 0xe21e,
		"rubydoc": 0xe73b, "sass": 0xe603, "scala": 0xe737, "scss": 0xe749,
		"sh": 0xf489, "shell": 0xf489, "slim": 0xe73b, "sln": 0xe70c,
		"so": 0xf17c, "sql": 0xf1c0, "sqlite3": 0xe7c4, "sty": 0xf034,
		"styl": 0xe600, "stylus": 0xe600, "svg": 0xf1c5, "swift": 0xe755,
		"t": 0xe769, "tar": 0xf410, "taz": 0xf410, "tbz": 0xf410,
		"tbz2": 0xf410, "tex": 0xf034, "tgz": 0xf410, "tiff": 0xf1c5,
		"tlz": 0xf410, "toml": 0xe615, "torrent": 0xe275, "ts": 0xe628,
		"tsv": 0xf1c3, "tsx": 0xe7ba, "ttf": 0xf031, "twig": 0xe61c,
		"txt": 0xf15c, "txz": 0xf410, "tz": 0xf410, "tzo": 0xf410,
		"video": 0xf03d, "vim": 0xe62b, "vue": 0xf0844, "war": 0xe256,
		"wav": 0xf001, "webm": 0xf03d, "webp": 0xf1c5, "windows": 0xf17a,
		"woff": 0xf031, "woff2": 0xf031, "xhtml": 0xf13b, "xls": 0xf1c3,
		"xlsx": 0xf1c3, "xml": 0xf05c0, "xul": 0xf05c0, "xz": 0xf410,
		"yaml": 0xf481, "yml": 0xf481, "zip": 0xf410, "zsh": 0xf489,
		"zsh-theme": 0xf489, "zshrc": 0xf489, "zst": 0xf410,
	}
	checkMap(t, "iconsByExtension", runify(wantByExt), iconsByExtension)
}

func runify(m map[string]uint32) map[string]rune {
	out := make(map[string]rune, len(m))
	for k, v := range m {
		out[k] = rune(v)
	}
	return out
}

func checkMap(t *testing.T, label string, want, got map[string]rune) {
	t.Helper()

	for key, wantRune := range want {
		gotRune, ok := got[key]
		if !ok {
			t.Errorf("%s[%q]: missing entry, want %U", label, key, wantRune)
			continue
		}
		if gotRune != wantRune {
			t.Errorf("%s[%q] = %U, want %U", label, key, gotRune, wantRune)
		}
	}

	for key := range got {
		if _, ok := want[key]; !ok {
			t.Errorf("%s[%q]: unexpected entry %U not in expected set", label, key, got[key])
		}
	}
}
