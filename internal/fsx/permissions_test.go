package fsx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTypeOfAndTypeChar(t *testing.T) {
	dir := t.TempDir()

	regular := filepath.Join(dir, "file.txt")
	must(t, os.WriteFile(regular, nil, 0o644))
	sub := filepath.Join(dir, "sub")
	must(t, os.Mkdir(sub, 0o755))
	link := filepath.Join(dir, "link")
	must(t, os.Symlink(regular, link))

	cases := []struct {
		path     string
		wantType Type
		wantChar byte
	}{
		{regular, TypeFile, '-'},
		{sub, TypeDirectory, 'd'},
		{link, TypeLink, 'l'},
	}

	for _, c := range cases {
		f, err := NewFile(c.path, nil, "")
		if err != nil {
			t.Fatal(err)
		}
		if got := TypeOf(f); got != c.wantType {
			t.Errorf("TypeOf(%s) = %v, want %v", c.path, got, c.wantType)
		}
		if got := c.wantType.TypeChar(); got != c.wantChar {
			t.Errorf("TypeChar() for %s = %q, want %q", c.path, got, c.wantChar)
		}
	}

	// Every other Type value falls back to '-' except the ones with their
	// own indicator.
	for typ, want := range map[Type]byte{
		TypePipe:        'p',
		TypeSocket:      's',
		TypeCharDevice:  'c',
		TypeBlockDevice: 'b',
		TypeSpecial:     '-',
	} {
		if got := typ.TypeChar(); got != want {
			t.Errorf("TypeChar() for %v = %q, want %q", typ, got, want)
		}
	}
}

func TestPermissionsOfAndOctal(t *testing.T) {
	dir := t.TempDir()

	// os.FileMode's setuid/setgid/sticky bits (os.ModeSetuid etc) live at
	// different bit positions than the raw 04000/02000/01000 Unix mode
	// bits, so they have to be named explicitly here rather than written
	// as a plain octal literal.
	cases := []struct {
		name string
		mode os.FileMode
		want string
	}{
		{"plain.txt", 0o644, "0644"},
		{"exe.sh", 0o755, "0755"},
		{"setuid", os.ModeSetuid | 0o644, "4644"},
		{"setgid", os.ModeSetgid | 0o644, "2644"},
		{"sticky", os.ModeSticky | 0o644, "1644"},
		{"allbits", os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o777, "7777"},
	}

	for _, c := range cases {
		path := filepath.Join(dir, c.name)
		must(t, os.WriteFile(path, nil, 0o644))
		// The special bits are set via a separate chmod (rather than in
		// the file-creation mode) since some kernels silently drop them
		// at creation time regardless of the mode passed to open(2).
		must(t, os.Chmod(path, c.mode))

		f, err := NewFile(path, nil, "")
		if err != nil {
			t.Fatal(err)
		}
		p := PermissionsOf(f)
		if got := p.Octal(); got != c.want {
			t.Errorf("Octal() for %s (mode %o) = %q, want %q", c.name, c.mode, got, c.want)
		}
	}

	// Spot-check individual bit decoding on the "allbits" file (already
	// chmod'd to 07777 above).
	path := filepath.Join(dir, "allbits")
	f, err := NewFile(path, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	p := PermissionsOf(f)
	for name, got := range map[string]bool{
		"UserRead": p.UserRead, "UserWrite": p.UserWrite, "UserExecute": p.UserExecute,
		"GroupRead": p.GroupRead, "GroupWrite": p.GroupWrite, "GroupExecute": p.GroupExecute,
		"OtherRead": p.OtherRead, "OtherWrite": p.OtherWrite, "OtherExecute": p.OtherExecute,
		"Sticky": p.Sticky, "Setgid": p.Setgid, "Setuid": p.Setuid,
	} {
		if !got {
			t.Errorf("expected %s to be set on mode 07777", name)
		}
	}
}
