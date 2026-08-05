package output

import (
	"fmt"
	"time"
)

// TimeFormat selects how timestamp columns are rendered.
type TimeFormat int

const (
	// DefaultTimeFormat shows "Mon D HH:MM" for timestamps within the
	// current year, and "Mon D  YYYY" for anything older -- the same
	// shape as `ls -l`'s default.
	DefaultTimeFormat TimeFormat = iota
	// ISOTimeFormat shows "MM-DD HH:MM" for recent timestamps, and
	// "YYYY-MM-DD" for older ones.
	ISOTimeFormat
	// LongISOTimeFormat always shows "YYYY-MM-DD HH:MM".
	LongISOTimeFormat
	// FullISOTimeFormat always shows "YYYY-MM-DD HH:MM:SS.NNNNNNNNN +ZZZZ".
	FullISOTimeFormat
)

// FormatTime renders t according to format, relative to now (used to
// decide whether DefaultTimeFormat/ISOTimeFormat should show a time-of-day
// or a year).
func FormatTime(format TimeFormat, t, now time.Time) string {
	if t.IsZero() {
		return "-"
	}

	recent := t.Year() == now.Year()

	switch format {
	case ISOTimeFormat:
		if recent {
			return fmt.Sprintf("%02d-%02d %02d:%02d", t.Month(), t.Day(), t.Hour(), t.Minute())
		}
		return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())

	case LongISOTimeFormat:
		return fmt.Sprintf("%04d-%02d-%02d %02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())

	case FullISOTimeFormat:
		_, offset := t.Zone()
		sign := byte('+')
		if offset < 0 {
			sign = '-'
			offset = -offset
		}
		return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%09d %c%02d%02d",
			t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
			sign, offset/3600, (offset%3600)/60)

	default: // DefaultTimeFormat
		if recent {
			return fmt.Sprintf("%s %2d %02d:%02d", t.Month().String()[:3], t.Day(), t.Hour(), t.Minute())
		}
		return fmt.Sprintf("%s %2d  %04d", t.Month().String()[:3], t.Day(), t.Year())
	}
}
