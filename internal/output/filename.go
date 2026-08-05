package output

import (
	"path/filepath"

	"lsgo/internal/fsx"
	"lsgo/internal/style"
	"lsgo/internal/theme"
)

// Classify controls whether a one-character type indicator (*, /, |, @, =)
// is appended after a filename.
type Classify bool

const (
	JustFilenames     Classify = false
	AddFileIndicators Classify = true
)

// ShowIcons controls whether (and with how much trailing spacing) an icon
// glyph is shown before a filename.
type ShowIcons struct {
	On     bool
	Spaces int
}

// FileNameOptions bundles the per-invocation filename-rendering settings
// that are the same for every file being listed.
type FileNameOptions struct {
	Classify  Classify
	ShowIcons ShowIcons
}

// FileNameStyle returns the style a filename should be painted with, based
// on its filesystem type and (for regular files) the theme's extension
// rules. Broken symlinks are handled by the caller, since only it knows
// whether it's about to show the link's target (in which case the arrow,
// not the name, carries the "broken" indication).
func FileNameStyle(th *theme.Theme, f *fsx.File) style.Style {
	switch {
	case f.IsDirectory():
		return th.UI.FileKinds.Directory
	case f.IsExecutableFile():
		return th.UI.FileKinds.Executable
	case f.IsLink():
		return th.UI.FileKinds.Symlink
	case f.IsPipe():
		return th.UI.FileKinds.Pipe
	case f.IsBlockDevice():
		return th.UI.FileKinds.BlockDevice
	case f.IsCharDevice():
		return th.UI.FileKinds.CharDevice
	case f.IsSocket():
		return th.UI.FileKinds.Socket
	case !f.IsRegularFile():
		return th.UI.FileKinds.Special
	default:
		if th.Colourer != nil {
			if s, ok := th.Colourer.ColourFile(f); ok {
				return s
			}
		}
		return th.UI.FileKinds.Normal
	}
}

// classifyChar returns the type-indicator character appended after a
// filename when Classify is on, if this file's type has one.
func classifyChar(f *fsx.File) string {
	switch {
	case f.IsExecutableFile():
		return "*"
	case f.IsDirectory():
		return "/"
	case f.IsPipe():
		return "|"
	case f.IsLink():
		return "@"
	case f.IsSocket():
		return "="
	default:
		return ""
	}
}

// RenderFileName paints a file's name (and, if requested, an icon and/or
// classify indicator) into a cell. withLinkPath additionally appends
// " -> target" for symlinks, as used by the lines and details views (but
// not the grid view, which has no room for it).
func RenderFileName(th *theme.Theme, opts FileNameOptions, f *fsx.File, withLinkPath bool) Cell {
	var cell Cell

	// A broken symlink's *filename* only takes on the "broken" colour
	// when we're not also about to show its target: in that case (the
	// grid view), the filename is the only place left to signal
	// brokenness. When the target is shown (lines/details view via
	// withLinkPath), the filename keeps its normal symlink colour and
	// the arrow carries the broken indication instead.
	var fileStyle style.Style
	if withLinkPath {
		fileStyle = FileNameStyle(th, f)
	} else {
		fileStyle = brokenAwareStyle(th, f)
	}

	if opts.ShowIcons.On {
		iconStyle := iconifyStyle(fileStyle)
		icon := theme.IconForFile(f, th.Iconer)
		cell.Text(iconStyle, string(icon))
		if opts.ShowIcons.Spaces > 0 {
			cell.Spaces(opts.ShowIcons.Spaces)
		} else {
			cell.Spaces(1)
		}
	}

	if f.Name != "" {
		escapeInto(&cell, f.Name, fileStyle, th.UI.ControlChar)
	}

	if withLinkPath && f.IsLink() {
		renderLinkTarget(th, &cell, f)

		if opts.Classify == AddFileIndicators {
			if target, ok := f.LinkTarget(); ok {
				if class := classifyChar(target); class != "" {
					cell.Plain(class)
				}
			}
		}
	} else if opts.Classify == AddFileIndicators {
		if class := classifyChar(f); class != "" {
			cell.Plain(class)
		}
	}

	return cell
}

// brokenAwareStyle is FileNameStyle, except that a broken (or unreadable)
// symlink is coloured as broken rather than as a plain symlink -- but only
// when there's nowhere else (i.e. no link target arrow) to show that fact.
func brokenAwareStyle(th *theme.Theme, f *fsx.File) style.Style {
	if f.IsLink() {
		if _, err, broken := f.LinkTargetDetailed(); err == nil && broken {
			return th.UI.BrokenSymlink
		}
	}
	return FileNameStyle(th, f)
}

func renderLinkTarget(th *theme.Theme, cell *Cell, f *fsx.File) {
	target, err, broken := f.LinkTargetDetailed()

	switch {
	case err != nil:
		// Couldn't even read the link; say nothing further, matching
		// the reference behaviour of just leaving the name as-is.
		return

	case broken:
		cell.Plain(" ")
		cell.Text(th.UI.BrokenSymlink, "->")
		cell.Plain(" ")
		if raw, err := f.LinkTargetRaw(); err == nil {
			escapeInto(cell, raw, style.ApplyOverlay(th.UI.BrokenSymlink, th.UI.BrokenPathOverlay),
				style.ApplyOverlay(th.UI.ControlChar, th.UI.BrokenPathOverlay))
		}

	default:
		cell.Plain(" ")
		cell.Text(th.UI.Punctuation, "->")
		cell.Plain(" ")

		if dir := filepath.Dir(target.Path); dir != "." && dir != "" {
			escapeInto(cell, dir, th.UI.SymlinkPath, th.UI.ControlChar)
			cell.Text(th.UI.SymlinkPath, string(filepath.Separator))
		}

		targetStyle := FileNameStyle(th, target)
		escapeInto(cell, target.Name, targetStyle, th.UI.ControlChar)
	}
}

// iconifyStyle picks the colour used to paint a file's icon: the
// filename's background colour if it has one, else its foreground colour,
// else no colour at all. Bold/underline/etc are deliberately dropped, since
// they tend to make icon glyphs look wrong.
func iconifyStyle(s style.Style) style.Style {
	if bg := s.Background; bg != (style.Colour{}) {
		return style.Style{Foreground: bg}
	}
	if fg := s.Foreground; fg != (style.Colour{}) {
		return style.Style{Foreground: fg}
	}
	return style.Style{}
}
