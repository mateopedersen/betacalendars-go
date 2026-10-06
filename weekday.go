package betacalendars

import "time"

// WeekStart identifies the first weekday shown in a calendar grid.
type WeekStart uint8

const (
	// Monday is the first weekday in the ISO-style week.
	Monday WeekStart = iota
	// Tuesday is the second weekday in the ISO-style week.
	Tuesday
	// Wednesday is the third weekday in the ISO-style week.
	Wednesday
	// Thursday is the fourth weekday in the ISO-style week.
	Thursday
	// Friday is the fifth weekday in the ISO-style week.
	Friday
	// Saturday is the sixth weekday in the ISO-style week.
	Saturday
	// Sunday is the seventh weekday in the ISO-style week.
	Sunday
)

// Valid reports whether w names one of the seven weekdays.
func (w WeekStart) Valid() bool { return w <= Sunday }

// Weekday returns the corresponding [time.Weekday]. Invalid values return time.Sunday.
func (w WeekStart) Weekday() time.Weekday {
	if !w.Valid() {
		return time.Sunday
	}
	return time.Weekday((int(w) + 1) % 7)
}

// String returns the lowercase weekday name, or "invalid" for an unknown value.
func (w WeekStart) String() string {
	if !w.Valid() {
		return "invalid"
	}
	return [...]string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}[w]
}
