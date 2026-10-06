package betacalendars

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ErrInvalidDate indicates that a date is outside the supported Gregorian range
// or has an invalid year, month, or day.
var ErrInvalidDate = errors.New("betacalendars: invalid date")

// ErrDateOutOfRange indicates that date arithmetic would leave years 1 through 9999.
var ErrDateOutOfRange = errors.New("betacalendars: date out of range")

// Date is a civil Gregorian date with no time or time zone.
//
// Its zero value is invalid. Use [NewDate] to construct a valid date.
type Date struct {
	year  int
	month time.Month
	day   int
}

// NewDate constructs a validated civil date in the proleptic Gregorian calendar.
// Years 1 through 9999 are supported.
func NewDate(year int, month time.Month, day int) (Date, error) {
	if year < 1 || year > 9999 || month < time.January || month > time.December {
		return Date{}, ErrInvalidDate
	}
	max, _ := DaysInMonth(year, month)
	if day < 1 || day > max {
		return Date{}, ErrInvalidDate
	}
	return Date{year: year, month: month, day: day}, nil
}

// Year returns the Gregorian year. It returns zero for the invalid zero value.
func (d Date) Year() int { return d.year }

// Month returns the Gregorian month. It returns zero for the invalid zero value.
func (d Date) Month() time.Month { return d.month }

// Day returns the day of the month. It returns zero for the invalid zero value.
func (d Date) Day() int { return d.day }

// IsValid reports whether d is a valid date constructed within the supported range.
func (d Date) IsValid() bool {
	if d.year < 1 || d.year > 9999 || d.month < time.January || d.month > time.December || d.day < 1 {
		return false
	}
	max, _ := DaysInMonth(d.year, d.month)
	return d.day <= max
}

// String returns d in zero-padded YYYY-MM-DD form. The invalid zero value returns "".
func (d Date) String() string {
	if !d.IsValid() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.year, int(d.month), d.day)
}

// GoString returns a Go-syntax representation of d.
func (d Date) GoString() string {
	if !d.IsValid() {
		return "betacalendars.Date{}"
	}
	return fmt.Sprintf("betacalendars.Date(%q)", d.String())
}

// Compare compares d and other lexicographically by year, month, and day.
// Invalid zero values sort before valid dates and compare equal to each other.
func (d Date) Compare(other Date) int {
	if !d.IsValid() || !other.IsValid() {
		switch {
		case !d.IsValid() && !other.IsValid():
			return 0
		case !d.IsValid():
			return -1
		default:
			return 1
		}
	}
	if d.year != other.year {
		return compareInt(d.year, other.year)
	}
	if d.month != other.month {
		return compareInt(int(d.month), int(other.month))
	}
	return compareInt(d.day, other.day)
}

// Before reports whether d precedes other. Invalid values follow [Date.Compare] ordering.
func (d Date) Before(other Date) bool { return d.Compare(other) < 0 }

// After reports whether d follows other. Invalid values follow [Date.Compare] ordering.
func (d Date) After(other Date) bool { return d.Compare(other) > 0 }

// Equal reports whether d and other represent the same date.
func (d Date) Equal(other Date) bool { return d.Compare(other) == 0 }

// Weekday returns the weekday of d. The invalid zero value returns time.Sunday.
func (d Date) Weekday() time.Weekday {
	if !d.IsValid() {
		return time.Sunday
	}
	weekday := (daysFromCivil(d.year, int(d.month), d.day) + 4) % 7 // 1970-01-01 was Thursday.
	if weekday < 0 {
		weekday += 7
	}
	return time.Weekday(weekday)
}

// DayOfYear returns the one-based ordinal day of d, or zero for an invalid value.
func (d Date) DayOfYear() int {
	if !d.IsValid() {
		return 0
	}
	return int(daysFromCivil(d.year, int(d.month), d.day)-daysFromCivil(d.year, 1, 1)) + 1
}

// ISOWeek returns the ISO 8601 week number and ISO week-year of d.
// Invalid values return (0, 0).
func (d Date) ISOWeek() (week, weekYear int) {
	if !d.IsValid() {
		return 0, 0
	}
	weekYear, week = time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC).ISOWeek()
	return week, weekYear
}

// AddDays returns the date n civil days from d. It never consults a time zone.
func (d Date) AddDays(n int) (Date, error) {
	if !d.IsValid() {
		return Date{}, ErrInvalidDate
	}
	ordinal := daysFromCivil(d.year, int(d.month), d.day) + int64(n)
	year, month, day := civilFromDays(ordinal)
	if year < 1 || year > 9999 {
		return Date{}, ErrDateOutOfRange
	}
	return Date{year: year, month: time.Month(month), day: day}, nil
}

// DaysBetween returns the signed number of civil days from start to end.
func DaysBetween(start, end Date) (int, error) {
	if !start.IsValid() || !end.IsValid() {
		return 0, ErrInvalidDate
	}
	return int(daysFromCivil(end.year, int(end.month), end.day) - daysFromCivil(start.year, int(start.month), start.day)), nil
}

// IsLeapYear reports whether year is a leap year in the proleptic Gregorian calendar.
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in month for a supported year.
func DaysInMonth(year int, month time.Month) (int, error) {
	if year < 1 || year > 9999 || month < time.January || month > time.December {
		return 0, ErrInvalidDate
	}
	days := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if month == time.February && IsLeapYear(year) {
		return 29, nil
	}
	return days[month-1], nil
}

// MarshalText implements [encoding.TextMarshaler] using YYYY-MM-DD.
func (d Date) MarshalText() ([]byte, error) {
	if !d.IsValid() {
		return nil, ErrInvalidDate
	}
	return []byte(d.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler]. It leaves d unchanged on error.
func (d *Date) UnmarshalText(text []byte) error {
	if d == nil {
		return ErrInvalidDate
	}
	parsed, err := ParseDate(string(text))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalJSON encodes d as a YYYY-MM-DD JSON string.
func (d Date) MarshalJSON() ([]byte, error) {
	if !d.IsValid() {
		return nil, ErrInvalidDate
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON decodes a YYYY-MM-DD JSON string. It leaves d unchanged on error.
func (d *Date) UnmarshalJSON(data []byte) error {
	if d == nil {
		return ErrInvalidDate
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return fmt.Errorf("%w: expected JSON string", ErrInvalidDate)
	}
	return d.UnmarshalText([]byte(text))
}

// ParseDate parses an exact, zero-padded YYYY-MM-DD civil date.
func ParseDate(text string) (Date, error) {
	if len(text) != 10 || text[4] != '-' || text[7] != '-' {
		return Date{}, ErrInvalidDate
	}
	for i, c := range text {
		if i == 4 || i == 7 {
			continue
		}
		if c < '0' || c > '9' {
			return Date{}, ErrInvalidDate
		}
	}
	year, err := strconv.Atoi(text[:4])
	if err != nil {
		return Date{}, ErrInvalidDate
	}
	month, err := strconv.Atoi(text[5:7])
	if err != nil {
		return Date{}, ErrInvalidDate
	}
	day, err := strconv.Atoi(text[8:10])
	if err != nil {
		return Date{}, ErrInvalidDate
	}
	return NewDate(year, time.Month(month), day)
}

func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// daysFromCivil returns days relative to 1970-01-01 using Gregorian arithmetic.
func daysFromCivil(year, month, day int) int64 {
	y := int64(year)
	m := int64(month)
	if m <= 2 {
		y--
	}
	era := y / 400
	if y < 0 && y%400 != 0 {
		era--
	}
	yoe := y - era*400
	mp := m
	if m > 2 {
		mp = m - 3
	} else {
		mp = m + 9
	}
	doy := (153*mp+2)/5 + int64(day) - 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
}

func civilFromDays(days int64) (year, month, day int) {
	z := days + 719468
	era := z / 146097
	if z < 0 && z%146097 != 0 {
		era--
	}
	doe := z - era*146097
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365
	y := yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100)
	mp := (5*doy + 2) / 153
	d := doy - (153*mp+2)/5 + 1
	m := mp + 3
	if mp >= 10 {
		m = mp - 9
	}
	if m <= 2 {
		y++
	}
	return int(y), int(m), int(d)
}
