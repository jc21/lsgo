package theme

import (
	"path/filepath"
	"strings"

	"lsgo/internal/fsx"
	"lsgo/internal/style"
)

// FileColourer decides which style, if any, a filename-based rule (as
// opposed to a filesystem-type rule) should paint a file's name with.
type FileColourer interface {
	ColourFile(f *fsx.File) (style.Style, bool)
}

// FileIconer decides which icon glyph, if any, a filename-based rule
// should show for a file. Used as a fallback after the by-name and
// by-extension icon maps.
type FileIconer interface {
	IconFile(f *fsx.File) (rune, bool)
}

// noFileColours is used when colouring is off, or no extension rules
// apply.
type noFileColours struct{}

func (noFileColours) ColourFile(*fsx.File) (style.Style, bool) { return style.Style{}, false }

// extensionMappings holds the glob-based filename colour rules parsed out
// of LS_COLORS (e.g. "*.zip=31").
type extensionMappings struct {
	patterns []string
	styles   []style.Style
}

func (e *extensionMappings) add(pattern string, s style.Style) {
	e.patterns = append(e.patterns, pattern)
	e.styles = append(e.styles, s)
}

func (e *extensionMappings) nonEmpty() bool { return len(e.patterns) > 0 }

// ColourFile matches patterns most-recently-added first, so that later
// entries in the environment variable override earlier ones.
func (e *extensionMappings) ColourFile(f *fsx.File) (style.Style, bool) {
	for i := len(e.patterns) - 1; i >= 0; i-- {
		if ok, err := filepath.Match(e.patterns[i], f.Name); err == nil && ok {
			return e.styles[i], true
		}
	}
	return style.Style{}, false
}

// chainedColours tries a first, falling back to b -- used to let explicit
// user-configured extension colours take precedence over the built-in
// FileExtensions rules, while keeping both active.
type chainedColours struct {
	first, second FileColourer
}

func (c chainedColours) ColourFile(f *fsx.File) (style.Style, bool) {
	if s, ok := c.first.ColourFile(f); ok {
		return s, ok
	}
	return c.second.ColourFile(f)
}

// FileExtensions is the built-in "what kind of content is this, based on
// its name" classifier: images, videos, archives, build-system files, and
// so on. This is what gives files like Makefile, *.zip, or *.jpg a colour
// (and an icon) without any user configuration.
type FileExtensions struct{}

// These classification lists are reference data: several of their entries
// (e.g. "Makefile", "avif", "class") also recur in icons.go's lookup maps
// and icons_test.go's independent cross-check. Constant-izing every shared
// literal would trade scannable, self-evident tables for indirection with
// no correctness upside, so goconst is suppressed for this block.
//
//nolint:goconst
var (
	immediateNames = []string{
		"Makefile", "Cargo.toml", "SConstruct", "CMakeLists.txt",
		"build.gradle", "pom.xml", "Rakefile", "package.json", "Gruntfile.js",
		"Gruntfile.coffee", "BUILD", "BUILD.bazel", "WORKSPACE", "build.xml", "Podfile",
		"webpack.config.js", "meson.build", "composer.json", "RoboFile.php", "PKGBUILD",
		"Justfile", "Procfile", "Dockerfile", "Containerfile", "Vagrantfile", "Brewfile",
		"Gemfile", "Pipfile", "build.sbt", "mix.exs", "bsconfig.json", "tsconfig.json",
	}

	imageExts = []string{
		"png", "jfi", "jfif", "jif", "jpe", "jpeg", "jpg", "gif", "bmp",
		"tiff", "tif", "ppm", "pgm", "pbm", "pnm", "webp", "raw", "arw",
		"svg", "stl", "eps", "dvi", "ps", "cbr", "jpf", "cbz", "xpm",
		"ico", "cr2", "orf", "nef", "heif", "avif", "jxl", "j2k", "jp2",
		"j2c", "jpx",
	}

	videoExts = []string{
		"avi", "flv", "m2v", "m4v", "mkv", "mov", "mp4", "mpeg",
		"mpg", "ogm", "ogv", "vob", "wmv", "webm", "m2ts", "heic",
	}

	musicExts = []string{"aac", "m4a", "mp3", "ogg", "wma", "mka", "opus"}

	losslessExts = []string{"alac", "ape", "flac", "wav"}

	cryptoExts = []string{"asc", "enc", "gpg", "pgp", "sig", "signature", "pfx", "p12"}

	documentExts = []string{
		"djvu", "doc", "docx", "dvi", "eml", "eps", "fotd", "key",
		"keynote", "numbers", "odp", "odt", "pages", "pdf", "ppt",
		"pptx", "rtf", "xls", "xlsx",
	}

	compressedExts = []string{
		"zip", "tar", "Z", "z", "gz", "bz2", "a", "ar", "7z",
		"iso", "dmg", "tc", "rar", "par", "tgz", "xz", "txz",
		"lz", "tlz", "lzma", "deb", "rpm", "zst", "lz4", "cpio",
	}

	tempExts = []string{"tmp", "swp", "swo", "swn", "bak", "bkp", "bk"}

	compiledExts = []string{"class", "elc", "hi", "o", "pyc", "zwc", "ko"}
)

func (FileExtensions) isImmediate(f *fsx.File) bool {
	return strings.HasPrefix(strings.ToLower(f.Name), "readme") ||
		strings.HasSuffix(f.Name, ".ninja") ||
		f.NameIsOneOf(immediateNames...)
}

func (FileExtensions) isImage(f *fsx.File) bool      { return f.ExtensionIsOneOf(imageExts...) }
func (FileExtensions) isVideo(f *fsx.File) bool      { return f.ExtensionIsOneOf(videoExts...) }
func (FileExtensions) isMusic(f *fsx.File) bool      { return f.ExtensionIsOneOf(musicExts...) }
func (FileExtensions) isLossless(f *fsx.File) bool   { return f.ExtensionIsOneOf(losslessExts...) }
func (FileExtensions) isCrypto(f *fsx.File) bool     { return f.ExtensionIsOneOf(cryptoExts...) }
func (FileExtensions) isDocument(f *fsx.File) bool   { return f.ExtensionIsOneOf(documentExts...) }
func (FileExtensions) isCompressed(f *fsx.File) bool { return f.ExtensionIsOneOf(compressedExts...) }

func (FileExtensions) isTemp(f *fsx.File) bool {
	if strings.HasSuffix(f.Name, "~") {
		return true
	}
	if strings.HasPrefix(f.Name, "#") && strings.HasSuffix(f.Name, "#") {
		return true
	}
	return f.ExtensionIsOneOf(tempExts...)
}

func (FileExtensions) isCompiled(f *fsx.File) bool {
	if f.ExtensionIsOneOf(compiledExts...) {
		return true
	}
	if f.Parent == nil || f.Ext == "" {
		return false
	}
	// A compiled file (e.g. "main.o") is highlighted if a plausible
	// source file for it (e.g. "main.c") exists alongside it.
	base := strings.TrimSuffix(f.Name, "."+f.Ext)
	for _, srcExt := range sourceExtsFor(f.Ext) {
		if f.Parent.Contains(f.Parent.Join(base + "." + srcExt)) {
			return true
		}
	}
	return false
}

// sourceExtsFor returns the plausible source-file extensions for a given
// compiled-output extension, e.g. a ".pyc" might come from a ".py".
func sourceExtsFor(compiledExt string) []string {
	switch compiledExt {
	case "o":
		return []string{"c", "cc", "cpp", "cxx", "m"}
	case "pyc":
		return []string{"py"}
	case "class":
		return []string{"java"} //nolint:goconst // see the classification lists' doc comment above
	case "hi":
		return []string{"hs"}
	case "elc":
		return []string{"el"}
	default:
		return nil
	}
}

// ColourFile implements FileColourer for the built-in extension rules.
func (e FileExtensions) ColourFile(f *fsx.File) (style.Style, bool) {
	switch {
	case e.isTemp(f):
		return style.Fixed(244).Normal(), true
	case e.isImmediate(f):
		return style.Yellow.Bold().SetUnderline(), true
	case e.isImage(f):
		return style.Fixed(133).Normal(), true
	case e.isVideo(f):
		return style.Fixed(135).Normal(), true
	case e.isMusic(f):
		return style.Fixed(92).Normal(), true
	case e.isLossless(f):
		return style.Fixed(93).Normal(), true
	case e.isCrypto(f):
		return style.Fixed(109).Normal(), true
	case e.isDocument(f):
		return style.Fixed(105).Normal(), true
	case e.isCompressed(f):
		return style.Red.Normal(), true
	case e.isCompiled(f):
		return style.Fixed(137).Normal(), true
	default:
		return style.Style{}, false
	}
}

// IconFile implements FileIconer for the handful of extension classes that
// get their own icon (audio/image/video) rather than falling through to
// the general extension map.
func (e FileExtensions) IconFile(f *fsx.File) (rune, bool) {
	switch {
	case e.isMusic(f), e.isLossless(f):
		return IconAudio, true
	case e.isImage(f):
		return IconImage, true
	case e.isVideo(f):
		return IconVideo, true
	default:
		return 0, false
	}
}
