// Package theme turns the user's colour preferences (the default palette,
// plus any LS_COLORS overrides) into a set of style.Style values that the
// output renderers paint text with.
package theme

import "lsgo/internal/style"

// FileKinds holds the styles used to colour a filename based on what kind
// of filesystem entry it is.
type FileKinds struct {
	Normal      style.Style
	Directory   style.Style
	Symlink     style.Style
	Pipe        style.Style
	BlockDevice style.Style
	CharDevice  style.Style
	Socket      style.Style
	Special     style.Style
	Executable  style.Style
}

// Permissions holds the per-bit styles used in the "rwxrwxrwx" column.
type Permissions struct {
	UserRead, UserWrite                 style.Style
	UserExecuteFile, UserExecuteOther   style.Style
	GroupRead, GroupWrite, GroupExecute style.Style
	OtherRead, OtherWrite, OtherExecute style.Style
	SpecialUserFile, SpecialOther       style.Style
	Attribute                           style.Style
}

// Size holds the styles used for the file-size column, split by magnitude
// so --color-scale can shade larger sizes differently.
type Size struct {
	Major, Minor style.Style

	NumberByte, NumberKilo, NumberMega, NumberGiga, NumberHuge style.Style
	UnitByte, UnitKilo, UnitMega, UnitGiga, UnitHuge           style.Style
}

// Users holds the styles for the user/group columns, distinguishing "you"
// from everyone else.
type Users struct {
	UserYou, UserSomeoneElse  style.Style
	GroupYours, GroupNotYours style.Style
}

// Links holds the styles for the hard-link-count column.
type Links struct {
	Normal        style.Style
	MultiLinkFile style.Style
}

// Git holds the styles for the two-character Git status column.
type Git struct {
	New, Modified, Deleted, Renamed, TypeChange, Ignored, Conflicted style.Style
}

// UIStyles is every configurable style in the interface, aside from the
// per-extension filename colouring handled separately by a FileColourer.
type UIStyles struct {
	FileKinds   FileKinds
	Permissions Permissions
	Size        Size
	Users       Users
	Links       Links
	Git         Git

	Punctuation style.Style
	Date        style.Style
	Inode       style.Style
	Blocks      style.Style
	Octal       style.Style
	Header      style.Style

	SymlinkPath       style.Style
	ControlChar       style.Style
	BrokenSymlink     style.Style
	BrokenPathOverlay style.Style
}

// ColourScale selects between a single fixed colour for all file sizes, or
// a gradient of colours by magnitude (--color-scale).
type ColourScale int

const (
	ScaleFixed ColourScale = iota
	ScaleGradient
)

// Plain is every style unset, used when colour output is disabled.
func Plain() UIStyles { return UIStyles{} }

// DefaultUIStyles is lsgo's built-in colour scheme.
func DefaultUIStyles(scale ColourScale) UIStyles {
	return UIStyles{
		FileKinds: FileKinds{
			Normal:      style.Style{},
			Directory:   style.Blue.Bold(),
			Symlink:     style.Cyan.Normal(),
			Pipe:        style.Yellow.Normal(),
			BlockDevice: style.Yellow.Bold(),
			CharDevice:  style.Yellow.Bold(),
			Socket:      style.Red.Bold(),
			Special:     style.Yellow.Normal(),
			Executable:  style.Green.Bold(),
		},

		Permissions: Permissions{
			UserRead:         style.Yellow.Bold(),
			UserWrite:        style.Red.Bold(),
			UserExecuteFile:  style.Green.Bold().SetUnderline(),
			UserExecuteOther: style.Green.Bold(),

			GroupRead:    style.Yellow.Normal(),
			GroupWrite:   style.Red.Normal(),
			GroupExecute: style.Green.Normal(),

			OtherRead:    style.Yellow.Normal(),
			OtherWrite:   style.Red.Normal(),
			OtherExecute: style.Green.Normal(),

			SpecialUserFile: style.Purple.Normal(),
			SpecialOther:    style.Purple.Normal(),

			Attribute: style.Style{},
		},

		Size: sizeStyles(scale),

		Users: Users{
			UserYou:         style.Yellow.Bold(),
			UserSomeoneElse: style.Style{},
			GroupYours:      style.Yellow.Bold(),
			GroupNotYours:   style.Style{},
		},

		Links: Links{
			Normal:        style.Red.Bold(),
			MultiLinkFile: style.Red.On(style.Yellow),
		},

		Git: Git{
			New:        style.Green.Normal(),
			Modified:   style.Blue.Normal(),
			Deleted:    style.Red.Normal(),
			Renamed:    style.Yellow.Normal(),
			TypeChange: style.Purple.Normal(),
			Ignored:    style.Style{Dim: true},
			Conflicted: style.Red.Normal(),
		},

		Punctuation: style.Fixed(244).Normal(),
		Date:        style.Blue.Normal(),
		Inode:       style.Purple.Normal(),
		Blocks:      style.Cyan.Normal(),
		Octal:       style.Purple.Normal(),
		Header:      style.Style{Underline: true},

		SymlinkPath:       style.Cyan.Normal(),
		ControlChar:       style.Red.Normal(),
		BrokenSymlink:     style.Red.Normal(),
		BrokenPathOverlay: style.Style{Underline: true},
	}
}

func sizeStyles(scale ColourScale) Size {
	if scale == ScaleGradient {
		return Size{
			Major: style.Green.Bold(),
			Minor: style.Green.Normal(),

			NumberByte: style.Fixed(118).Normal(),
			NumberKilo: style.Fixed(190).Normal(),
			NumberMega: style.Fixed(226).Normal(),
			NumberGiga: style.Fixed(220).Normal(),
			NumberHuge: style.Fixed(214).Normal(),

			UnitByte: style.Green.Normal(),
			UnitKilo: style.Green.Normal(),
			UnitMega: style.Green.Normal(),
			UnitGiga: style.Green.Normal(),
			UnitHuge: style.Green.Normal(),
		}
	}

	return Size{
		Major: style.Green.Bold(),
		Minor: style.Green.Normal(),

		NumberByte: style.Green.Bold(),
		NumberKilo: style.Green.Bold(),
		NumberMega: style.Green.Bold(),
		NumberGiga: style.Green.Bold(),
		NumberHuge: style.Green.Bold(),

		UnitByte: style.Green.Normal(),
		UnitKilo: style.Green.Normal(),
		UnitMega: style.Green.Normal(),
		UnitGiga: style.Green.Normal(),
		UnitHuge: style.Green.Normal(),
	}
}

// SetLS applies one LS_COLORS-style key/value pair to this UIStyles,
// reporting whether the key was recognised. Some keys LS_COLORS defines
// -- MULTIHARDLINK, DOOR, and so on -- have no equivalent here and are
// simply accepted (so a full LS_COLORS string doesn't produce spurious
// warnings) without changing anything.
func (u *UIStyles) SetLS(key string, s style.Style) bool {
	switch key {
	case "di":
		u.FileKinds.Directory = s
	case "ex":
		u.FileKinds.Executable = s
	case "fi":
		u.FileKinds.Normal = s
	case "pi":
		u.FileKinds.Pipe = s
	case "so":
		u.FileKinds.Socket = s
	case "bd":
		u.FileKinds.BlockDevice = s
	case "cd":
		u.FileKinds.CharDevice = s
	case "ln":
		u.FileKinds.Symlink = s
	case "or":
		u.BrokenSymlink = s
	default:
		return false
	}
	return true
}
