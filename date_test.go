package betacalendars_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/mateopedersen/betacalendars-go"
)

func TestDateValidationAndLeapMatrix(t *testing.T) {
	for _, test := range []struct {
		year int
		leap bool
	}{{1900, false}, {2000, true}, {2024, true}, {2027, false}, {2028, true}, {2100, false}, {2400, true}} {
		if got := betacalendars.IsLeapYear(test.year); got != test.leap {
			t.Errorf("IsLeapYear(%d) = %v, want %v", test.year, got, test.leap)
		}
	}
	for _, text := range []string{"2027-02-29", "2027-04-31", "2027-13-01", "2027-2-01", "0000-01-01"} {
		if _, err := betacalendars.ParseDate(text); !errors.Is(err, betacalendars.ErrInvalidDate) {
			t.Errorf("ParseDate(%q) error = %v", text, err)
		}
	}
}

func TestDateArithmeticAndISOWeek(t *testing.T) {
	start, _ := betacalendars.ParseDate("2026-12-31")
	next, err := start.AddDays(1)
	if err != nil || next.String() != "2027-01-01" {
		t.Fatalf("AddDays = %v, %v", next, err)
	}
	if days, _ := betacalendars.DaysBetween(start, next); days != 1 {
		t.Fatalf("DaysBetween = %d", days)
	}
	week, year := next.ISOWeek()
	if week != 53 || year != 2026 {
		t.Fatalf("ISOWeek = %d-W%d, want 2026-W53", week, year)
	}
	inverse, err := next.AddDays(-1)
	if err != nil || !inverse.Equal(start) {
		t.Fatalf("inverse AddDays = %v, %v", inverse, err)
	}
	if start.Weekday() != time.Thursday {
		t.Fatalf("weekday = %s", start.Weekday())
	}
	if start.DayOfYear() != 365 {
		t.Fatalf("DayOfYear = %d", start.DayOfYear())
	}
}

func TestDateTextAndJSONRoundTrip(t *testing.T) {
	want, _ := betacalendars.ParseDate("2000-02-29")
	text, err := want.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var fromText betacalendars.Date
	if err := fromText.UnmarshalText(text); err != nil || !fromText.Equal(want) {
		t.Fatalf("text round trip = %v, %v", fromText, err)
	}
	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var fromJSON betacalendars.Date
	if err := json.Unmarshal(encoded, &fromJSON); err != nil || !fromJSON.Equal(want) {
		t.Fatalf("JSON round trip = %v, %v", fromJSON, err)
	}
	if string(encoded) != `"2000-02-29"` {
		t.Fatalf("JSON = %s", encoded)
	}
}

func TestDateRangeEdges(t *testing.T) {
	first, _ := betacalendars.NewDate(1, time.January, 1)
	if _, err := first.AddDays(-1); !errors.Is(err, betacalendars.ErrDateOutOfRange) {
		t.Fatalf("AddDays before range = %v", err)
	}
	last, _ := betacalendars.NewDate(9999, time.December, 31)
	if _, err := last.AddDays(1); !errors.Is(err, betacalendars.ErrDateOutOfRange) {
		t.Fatalf("AddDays after range = %v", err)
	}
}

func FuzzDateRoundTrip(f *testing.F) {
	for _, seed := range []string{"2024-02-29", "1900-02-28", "0001-01-01", "9999-12-31", "not-a-date"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		date, err := betacalendars.ParseDate(text)
		if err != nil {
			return
		}
		roundTrip, err := betacalendars.ParseDate(date.String())
		if err != nil || !date.Equal(roundTrip) {
			t.Fatalf("round trip %q => %q, %v", text, date, err)
		}
	})
}

func FuzzAddDaysInverse(f *testing.F) {
	f.Add(uint16(2024), uint8(2), uint8(29), int16(17))
	f.Add(uint16(1900), uint8(3), uint8(1), int16(-1))
	f.Fuzz(func(t *testing.T, year uint16, month, day uint8, delta int16) {
		y := int(year%9999) + 1
		m := time.Month(month%12 + 1)
		max, _ := betacalendars.DaysInMonth(y, m)
		d, err := betacalendars.NewDate(y, m, int(day%uint8(max))+1)
		if err != nil {
			t.Fatal(err)
		}
		n := int(delta) % 3650
		shifted, err := d.AddDays(n)
		if err != nil {
			return
		}
		back, err := shifted.AddDays(-n)
		if err != nil || !d.Equal(back) {
			t.Fatalf("%s + %d then back = %s, %v", d, n, back, err)
		}
	})
}
