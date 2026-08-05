package output

import (
	"testing"
	"time"
)

func TestFormatTimeDefaultRecentVsOld(t *testing.T) {
	now := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	recent := time.Date(2026, time.June, 29, 16, 16, 0, 0, time.UTC)
	old := time.Date(2014, time.November, 23, 10, 0, 0, 0, time.UTC)

	if got := FormatTime(DefaultTimeFormat, recent, now); got != "Jun 29 16:16" {
		t.Errorf("recent = %q", got)
	}
	if got := FormatTime(DefaultTimeFormat, old, now); got != "Nov 23  2014" {
		t.Errorf("old = %q", got)
	}
}

func TestFormatTimeISO(t *testing.T) {
	now := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	recent := time.Date(2026, time.June, 29, 16, 16, 0, 0, time.UTC)
	old := time.Date(2014, time.November, 23, 10, 0, 0, 0, time.UTC)

	if got := FormatTime(ISOTimeFormat, recent, now); got != "06-29 16:16" {
		t.Errorf("recent ISO = %q", got)
	}
	if got := FormatTime(ISOTimeFormat, old, now); got != "2014-11-23" {
		t.Errorf("old ISO = %q", got)
	}
}

func TestFormatTimeLongISO(t *testing.T) {
	now := time.Now()
	tm := time.Date(2026, time.June, 29, 16, 16, 0, 0, time.UTC)
	if got := FormatTime(LongISOTimeFormat, tm, now); got != "2026-06-29 16:16" {
		t.Errorf("got %q", got)
	}
}

func TestFormatTimeFullISO(t *testing.T) {
	now := time.Now()
	tm := time.Date(2026, time.June, 29, 16, 16, 30, 123456789, time.UTC)
	got := FormatTime(FullISOTimeFormat, tm, now)
	want := "2026-06-29 16:16:30.123456789 +0000"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatTimeZeroIsDash(t *testing.T) {
	if got := FormatTime(DefaultTimeFormat, time.Time{}, time.Now()); got != "-" {
		t.Errorf("expected dash for zero time, got %q", got)
	}
}
