package output

import (
	"os/user"
	"strconv"
	"time"

	"github.com/jc21/lsgo/internal/fsx"
	"github.com/jc21/lsgo/internal/style"
	"github.com/jc21/lsgo/internal/theme"
)

// TimeType selects which of a file's timestamps a Timestamp column shows.
type TimeType int

const (
	TimeModified TimeType = iota
	TimeChanged
	TimeAccessed
	TimeCreated
)

func (t TimeType) header() string {
	switch t {
	case TimeChanged:
		return "Date Changed"
	case TimeAccessed:
		return "Date Accessed"
	case TimeCreated:
		return "Date Created"
	default:
		return "Date Modified"
	}
}

// TimeTypes records which timestamp columns are enabled; more than one
// may be shown at once.
type TimeTypes struct {
	Modified, Changed, Accessed, Created bool
}

// DefaultTimeTypes shows just the modified time, the most common case.
func DefaultTimeTypes() TimeTypes { return TimeTypes{Modified: true} }

// UserFormat selects how the user/group columns display an owner: by
// name (the default) or by raw numeric ID.
type UserFormat int

const (
	UserName UserFormat = iota
	UserNumeric
)

// ColumnKind identifies which piece of file metadata a Column shows.
type ColumnKind int

const (
	ColInode ColumnKind = iota
	ColOctal
	ColPermissions
	ColHardLinks
	ColFileSize
	ColBlocks
	ColUser
	ColGroup
	ColTimestamp
	ColGitStatus
)

// Column is one column of the long-view table.
type Column struct {
	Kind ColumnKind
	Time TimeType // only meaningful when Kind == ColTimestamp
}

// Header returns the column's title, used when --header is given.
func (c Column) Header() string {
	switch c.Kind {
	case ColInode:
		return "inode"
	case ColOctal:
		return "Octal"
	case ColPermissions:
		return "Permissions"
	case ColHardLinks:
		return "Links"
	case ColFileSize:
		return "Size"
	case ColBlocks:
		return "Blocks"
	case ColUser:
		return "User"
	case ColGroup:
		return "Group"
	case ColTimestamp:
		return c.Time.header()
	case ColGitStatus:
		return "Git"
	default:
		return ""
	}
}

// alignLeft reports whether this column's text should be left-aligned
// (padding on the right) rather than right-aligned (padding on the left,
// as for numeric columns).
func (c Column) alignLeft() bool {
	switch c.Kind {
	case ColFileSize, ColHardLinks, ColInode, ColBlocks, ColGitStatus:
		return false
	default:
		return true
	}
}

// TableColumns decides which columns appear in the long view, and in
// which order -- everything defaults to on except the columns explicitly
// suppressed with --no-permissions/--no-filesize/--no-user, or the
// optional ones (inode, links, blocks, group, git, octal) which default
// to off.
type TableColumns struct {
	TimeTypes TimeTypes

	Inode, Links, Blocks, Group, Git, Octal bool

	Permissions, Filesize, User bool
}

// DefaultTableColumns is the column set shown by plain `-l`.
func DefaultTableColumns() TableColumns {
	return TableColumns{
		TimeTypes:   DefaultTimeTypes(),
		Permissions: true,
		Filesize:    true,
		User:        true,
	}
}

// Collect builds the ordered column list for these settings. gitEnabled
// additionally gates the Git column on whether a repository was actually
// found for what's being listed -- there's no point reserving a column of
// dashes when --git was passed outside of any repository.
func (tc TableColumns) Collect(gitEnabled bool) []Column {
	var cols []Column

	if tc.Inode {
		cols = append(cols, Column{Kind: ColInode})
	}
	if tc.Octal {
		cols = append(cols, Column{Kind: ColOctal})
	}
	if tc.Permissions {
		cols = append(cols, Column{Kind: ColPermissions})
	}
	if tc.Links {
		cols = append(cols, Column{Kind: ColHardLinks})
	}
	if tc.Filesize {
		cols = append(cols, Column{Kind: ColFileSize})
	}
	if tc.Blocks {
		cols = append(cols, Column{Kind: ColBlocks})
	}
	if tc.User {
		cols = append(cols, Column{Kind: ColUser})
	}
	if tc.Group {
		cols = append(cols, Column{Kind: ColGroup})
	}
	if tc.TimeTypes.Modified {
		cols = append(cols, Column{Kind: ColTimestamp, Time: TimeModified})
	}
	if tc.TimeTypes.Changed {
		cols = append(cols, Column{Kind: ColTimestamp, Time: TimeChanged})
	}
	if tc.TimeTypes.Created {
		cols = append(cols, Column{Kind: ColTimestamp, Time: TimeCreated})
	}
	if tc.TimeTypes.Accessed {
		cols = append(cols, Column{Kind: ColTimestamp, Time: TimeAccessed})
	}
	if tc.Git && gitEnabled {
		cols = append(cols, Column{Kind: ColGitStatus})
	}

	return cols
}

// TableOptions bundles every setting that affects how the long-view
// columns are formatted.
type TableOptions struct {
	SizeFormat SizeFormat
	TimeFormat TimeFormat
	UserFormat UserFormat
	Columns    TableColumns
}

// Table accumulates rows for the long view and tracks each column's
// widest cell so far, so that every row can be padded consistently once
// rendering begins.
type Table struct {
	theme   *theme.Theme
	git     *fsx.GitCache
	options TableOptions
	columns []Column
	widths  []int
	now     time.Time

	users  map[uint32]string
	groups map[uint32]string
	me     currentIdentity
}

type currentIdentity struct {
	uid, gid uint32
	groupIDs map[uint32]bool
}

// NewTable creates a table for the given options. git may be nil if no
// Git column is needed.
func NewTable(th *theme.Theme, git *fsx.GitCache, options TableOptions) *Table {
	cols := options.Columns.Collect(git != nil)
	return &Table{
		theme:   th,
		git:     git,
		options: options,
		columns: cols,
		widths:  make([]int, len(cols)),
		now:     time.Now(),
		users:   make(map[uint32]string),
		groups:  make(map[uint32]string),
		me:      loadCurrentIdentity(),
	}
}

func loadCurrentIdentity() currentIdentity {
	id := currentIdentity{groupIDs: make(map[uint32]bool)}

	u, err := user.Current()
	if err != nil {
		return id
	}
	if uid, err := strconv.ParseUint(u.Uid, 10, 32); err == nil {
		id.uid = uint32(uid)
	}
	if gid, err := strconv.ParseUint(u.Gid, 10, 32); err == nil {
		id.gid = uint32(gid)
		id.groupIDs[id.gid] = true
	}
	if ids, err := u.GroupIds(); err == nil {
		for _, g := range ids {
			if gid, err := strconv.ParseUint(g, 10, 32); err == nil {
				id.groupIDs[uint32(gid)] = true
			}
		}
	}
	return id
}

// Columns returns the resolved column list, e.g. for building a header
// row.
func (t *Table) Columns() []Column { return t.columns }

// HeaderRow renders the --header row.
func (t *Table) HeaderRow() []Cell {
	row := make([]Cell, len(t.columns))
	for i, c := range t.columns {
		var cell Cell
		cell.Text(t.theme.UI.Header, c.Header())
		row[i] = cell
	}
	return row
}

// RowForFile builds the row of cells describing one file, one per
// configured column. hasXattrs indicates whether the file has any
// extended attributes (only meaningful when the caller looked them up,
// i.e. --extended was given).
func (t *Table) RowForFile(f *fsx.File, hasXattrs bool) []Cell {
	row := make([]Cell, len(t.columns))
	for i, c := range t.columns {
		row[i] = t.renderColumn(f, c, hasXattrs)
	}
	return row
}

// AddWidths updates the running column-width tracking with a row that's
// about to be displayed.
func (t *Table) AddWidths(row []Cell) {
	for i, cell := range row {
		if cell.Width > t.widths[i] {
			t.widths[i] = cell.Width
		}
	}
}

// Render pads and concatenates a row's cells into the final text that
// precedes the filename column, using each column's tracked width and
// alignment.
func (t *Table) Render(row []Cell) Cell {
	var out Cell
	for i, cell := range row {
		pad := t.widths[i] - cell.Width
		if t.columns[i].alignLeft() {
			out.Append(cell)
			out.Spaces(pad)
		} else {
			out.Spaces(pad)
			out.Append(cell)
		}
		out.Spaces(1)
	}
	return out
}

func (t *Table) renderColumn(f *fsx.File, c Column, hasXattrs bool) Cell {
	ui := &t.theme.UI

	switch c.Kind {
	case ColPermissions:
		return renderPermissions(ui, f, hasXattrs)

	case ColOctal:
		var cell Cell
		cell.Text(ui.Octal, fsx.PermissionsOf(f).Octal())
		return cell

	case ColHardLinks:
		var cell Cell
		count := f.LinkCount()
		s := ui.Links.Normal
		if f.IsRegularFile() && count > 1 {
			s = ui.Links.MultiLinkFile
		}
		cell.Text(s, formatInt(count))
		return cell

	case ColFileSize:
		return t.renderSize(ui, f)

	case ColBlocks:
		var cell Cell
		if n, ok := f.Blocks(); ok {
			cell.Text(ui.Blocks, strconv.FormatInt(n, 10))
		} else {
			cell.Text(ui.Punctuation, "-")
		}
		return cell

	case ColUser:
		return t.renderUser(ui, f.UID())

	case ColGroup:
		return t.renderGroup(ui, f.GID())

	case ColInode:
		var cell Cell
		cell.Text(ui.Inode, strconv.FormatUint(f.Inode(), 10))
		return cell

	case ColTimestamp:
		return t.renderTimestamp(ui, f, c.Time)

	case ColGitStatus:
		return t.renderGitStatus(ui, f)

	default:
		return Cell{}
	}
}

func renderPermissions(ui *theme.UIStyles, f *fsx.File, hasXattrs bool) Cell {
	var cell Cell

	typ := fsx.TypeOf(f)
	cell.Text(typeStyle(ui, typ), string(typ.TypeChar()))

	p := fsx.PermissionsOf(f)
	isRegular := typ == fsx.TypeFile

	bit := func(on bool, ch string, s style.Style) {
		if on {
			cell.Text(s, ch)
		} else {
			cell.Text(ui.Punctuation, "-")
		}
	}

	bit(p.UserRead, "r", ui.Permissions.UserRead)
	bit(p.UserWrite, "w", ui.Permissions.UserWrite)
	cell.Append(userExecuteBit(ui, p, isRegular))

	bit(p.GroupRead, "r", ui.Permissions.GroupRead)
	bit(p.GroupWrite, "w", ui.Permissions.GroupWrite)
	cell.Append(groupExecuteBit(ui, p))

	bit(p.OtherRead, "r", ui.Permissions.OtherRead)
	bit(p.OtherWrite, "w", ui.Permissions.OtherWrite)
	cell.Append(otherExecuteBit(ui, p))

	if hasXattrs {
		cell.Text(ui.Permissions.Attribute, "@")
	}

	return cell
}

func userExecuteBit(ui *theme.UIStyles, p fsx.Permissions, isRegular bool) Cell {
	var cell Cell
	switch {
	case !p.UserExecute && !p.Setuid:
		cell.Text(ui.Punctuation, "-")
	case p.UserExecute && !p.Setuid && !isRegular:
		cell.Text(ui.Permissions.UserExecuteOther, "x")
	case p.UserExecute && !p.Setuid && isRegular:
		cell.Text(ui.Permissions.UserExecuteFile, "x")
	case !p.UserExecute && p.Setuid:
		cell.Text(ui.Permissions.SpecialOther, "S")
	case p.UserExecute && p.Setuid && !isRegular:
		cell.Text(ui.Permissions.SpecialOther, "s")
	default: // UserExecute && Setuid && isRegular
		cell.Text(ui.Permissions.SpecialUserFile, "s")
	}
	return cell
}

func groupExecuteBit(ui *theme.UIStyles, p fsx.Permissions) Cell {
	var cell Cell
	switch {
	case !p.GroupExecute && !p.Setgid:
		cell.Text(ui.Punctuation, "-")
	case p.GroupExecute && !p.Setgid:
		cell.Text(ui.Permissions.GroupExecute, "x")
	case !p.GroupExecute && p.Setgid:
		cell.Text(ui.Permissions.SpecialOther, "S")
	default:
		cell.Text(ui.Permissions.SpecialOther, "s")
	}
	return cell
}

func otherExecuteBit(ui *theme.UIStyles, p fsx.Permissions) Cell {
	var cell Cell
	switch {
	case !p.OtherExecute && !p.Sticky:
		cell.Text(ui.Punctuation, "-")
	case p.OtherExecute && !p.Sticky:
		cell.Text(ui.Permissions.OtherExecute, "x")
	case !p.OtherExecute && p.Sticky:
		cell.Text(ui.Permissions.SpecialOther, "T")
	default:
		cell.Text(ui.Permissions.SpecialOther, "t")
	}
	return cell
}

func typeStyle(ui *theme.UIStyles, t fsx.Type) style.Style {
	switch t {
	case fsx.TypeDirectory:
		return ui.FileKinds.Directory
	case fsx.TypeLink:
		return ui.FileKinds.Symlink
	case fsx.TypePipe:
		return ui.FileKinds.Pipe
	case fsx.TypeBlockDevice:
		return ui.FileKinds.BlockDevice
	case fsx.TypeCharDevice:
		return ui.FileKinds.CharDevice
	case fsx.TypeSocket:
		return ui.FileKinds.Socket
	case fsx.TypeSpecial:
		return ui.FileKinds.Special
	default:
		return ui.Punctuation
	}
}

func (t *Table) renderSize(ui *theme.UIStyles, f *fsx.File) Cell {
	switch {
	case f.IsDirectory():
		return RenderNoSize(ui)
	case f.IsCharDevice(), f.IsBlockDevice():
		major, minor := f.DeviceIDs()
		return RenderDeviceIDs(ui, major, minor)
	default:
		// os.FileInfo.Size() is declared int64; guard the negative case
		// explicitly (rather than converting straight to uint64) since a
		// negative size would otherwise wrap to a huge value.
		size := max(f.Info.Size(), 0)
		return RenderSize(ui, t.options.SizeFormat, uint64(size))
	}
}

func (t *Table) renderUser(ui *theme.UIStyles, uid uint32) Cell {
	var cell Cell

	name, ok := t.lookupUser(uid)
	text := name
	if t.options.UserFormat == UserNumeric || !ok {
		text = strconv.FormatUint(uint64(uid), 10)
	}

	s := ui.Users.UserSomeoneElse
	if uid == t.me.uid {
		s = ui.Users.UserYou
	}
	cell.Text(s, text)
	return cell
}

func (t *Table) renderGroup(ui *theme.UIStyles, gid uint32) Cell {
	var cell Cell

	name, ok := t.lookupGroup(gid)
	text := name
	if t.options.UserFormat == UserNumeric || !ok {
		text = strconv.FormatUint(uint64(gid), 10)
	}

	s := ui.Users.GroupNotYours
	if t.me.groupIDs[gid] {
		s = ui.Users.GroupYours
	}
	cell.Text(s, text)
	return cell
}

func (t *Table) lookupUser(uid uint32) (string, bool) {
	if name, ok := t.users[uid]; ok {
		return name, name != ""
	}
	u, err := user.LookupId(strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		t.users[uid] = ""
		return "", false
	}
	t.users[uid] = u.Username
	return u.Username, true
}

func (t *Table) lookupGroup(gid uint32) (string, bool) {
	if name, ok := t.groups[gid]; ok {
		return name, name != ""
	}
	g, err := user.LookupGroupId(strconv.FormatUint(uint64(gid), 10))
	if err != nil {
		t.groups[gid] = ""
		return "", false
	}
	t.groups[gid] = g.Name
	return g.Name, true
}

func (t *Table) renderTimestamp(ui *theme.UIStyles, f *fsx.File, tt TimeType) Cell {
	var cell Cell

	var when time.Time
	switch tt {
	case TimeChanged:
		when = f.ChangedTime()
	case TimeAccessed:
		when = f.AccessedTime()
	case TimeCreated:
		if created, ok := f.CreatedTime(); ok {
			when = created
		} else {
			when = f.ModTime()
		}
	default:
		when = f.ModTime()
	}

	cell.Text(ui.Date, FormatTime(t.options.TimeFormat, when, t.now))
	return cell
}

func (t *Table) renderGitStatus(ui *theme.UIStyles, f *fsx.File) Cell {
	var cell Cell
	var status fsx.GitStatus
	if t.git != nil {
		status = t.git.Status(f.Path, f.IsDirectory())
	}

	cell.Text(gitStatusStyle(ui, status.Staged), gitStatusChar(status.Staged))
	cell.Text(gitStatusStyle(ui, status.Unstaged), gitStatusChar(status.Unstaged))
	return cell
}

func gitStatusChar(s fsx.GitFileStatus) string {
	switch s {
	case fsx.GitNew:
		return "N"
	case fsx.GitModified:
		return "M"
	case fsx.GitDeleted:
		return "D"
	case fsx.GitRenamed:
		return "R"
	case fsx.GitTypeChange:
		return "T"
	case fsx.GitIgnored:
		return "I"
	case fsx.GitConflicted:
		return "U"
	default:
		return "-"
	}
}

func gitStatusStyle(ui *theme.UIStyles, s fsx.GitFileStatus) style.Style {
	switch s {
	case fsx.GitNew:
		return ui.Git.New
	case fsx.GitModified:
		return ui.Git.Modified
	case fsx.GitDeleted:
		return ui.Git.Deleted
	case fsx.GitRenamed:
		return ui.Git.Renamed
	case fsx.GitTypeChange:
		return ui.Git.TypeChange
	case fsx.GitIgnored:
		return ui.Git.Ignored
	case fsx.GitConflicted:
		return ui.Git.Conflicted
	default:
		return ui.Punctuation
	}
}
