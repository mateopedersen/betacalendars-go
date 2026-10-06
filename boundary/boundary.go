package boundary

import (
	"fmt"
	"time"

	"github.com/mateopedersen/betacalendars-go"
)

// MonthTransition describes consecutive month-end and month-start dates.
type MonthTransition struct {
	// From is the final date of the earlier month.
	From betacalendars.Date
	// To is the first date of the following month.
	To betacalendars.Date
	// FromMonthDays and ToMonthDays are the lengths of the two months.
	FromMonthDays int
	ToMonthDays   int
	// MonthLengthChanged reports whether the adjacent month lengths differ.
	MonthLengthChanged bool
	// YearChanged reports whether the transition crosses December to January.
	YearChanged bool
}

// LeapDayTransition describes February 28, February 29, and March 1.
type LeapDayTransition struct {
	// Before is February 28 in a leap year.
	Before betacalendars.Date
	// LeapDay is February 29.
	LeapDay betacalendars.Date
	// After is March 1.
	After betacalendars.Date
}

// ISOWeekYearTransition describes consecutive dates whose ISO week-year changes.
type ISOWeekYearTransition struct {
	// Before and After are consecutive civil dates.
	Before betacalendars.Date
	After  betacalendars.Date
	// BeforeISOWeek and AfterISOWeek are the ISO week numbers.
	BeforeISOWeek int
	AfterISOWeek  int
	// BeforeISOYear and AfterISOYear are the ISO week-years.
	BeforeISOYear int
	AfterISOYear  int
}

// Report contains calendar boundary cases for one requested Gregorian year.
type Report struct {
	// Year is the requested year.
	Year int
	// MonthTransitions contains the eleven within-year transitions plus the
	// December-to-January transition when the following year is representable.
	MonthTransitions []MonthTransition
	// MonthLengthChanges contains the subset of MonthTransitions where lengths differ.
	MonthLengthChanges []MonthTransition
	// YearTransition is December 31 to January 1 of the following year.
	YearTransition MonthTransition
	// HasYearTransition is false only for year 9999, whose next year is unsupported.
	HasYearTransition bool
	// LeapDayTransitions contains one entry in leap years and none otherwise.
	LeapDayTransitions []LeapDayTransition
	// ISOWeekYearTransitions contains ISO week-year changes touching the requested year.
	ISOWeekYearTransitions []ISOWeekYearTransition
}

// ForYear reports month-length changes, the year turn, leap-day context, and
// ISO week-year changes for year. Civil-date support is limited to 1 through 9999.
func ForYear(year int) (Report, error) {
	if year < 1 || year > 9999 {
		return Report{}, betacalendars.ErrInvalidDate
	}
	report := Report{Year: year, MonthTransitions: make([]MonthTransition, 0, 12), MonthLengthChanges: make([]MonthTransition, 0, 6)}
	for month := time.January; month < time.December; month++ {
		transition, err := monthTransition(year, month, year, month+1)
		if err != nil {
			return Report{}, err
		}
		addMonthTransition(&report, transition)
	}
	if year < 9999 {
		transition, err := monthTransition(year, time.December, year+1, time.January)
		if err != nil {
			return Report{}, err
		}
		report.YearTransition = transition
		report.HasYearTransition = true
		addMonthTransition(&report, transition)
	}
	if betacalendars.IsLeapYear(year) {
		before, _ := betacalendars.NewDate(year, time.February, 28)
		leap, _ := betacalendars.NewDate(year, time.February, 29)
		after, _ := betacalendars.NewDate(year, time.March, 1)
		report.LeapDayTransitions = []LeapDayTransition{{Before: before, LeapDay: leap, After: after}}
	} else {
		report.LeapDayTransitions = []LeapDayTransition{}
	}
	transitions, err := isoTransitionsTouching(year)
	if err != nil {
		return Report{}, err
	}
	report.ISOWeekYearTransitions = transitions
	return report, nil
}

func monthTransition(fromYear int, fromMonth time.Month, toYear int, toMonth time.Month) (MonthTransition, error) {
	fromDays, err := betacalendars.DaysInMonth(fromYear, fromMonth)
	if err != nil {
		return MonthTransition{}, err
	}
	toDays, err := betacalendars.DaysInMonth(toYear, toMonth)
	if err != nil {
		return MonthTransition{}, err
	}
	from, _ := betacalendars.NewDate(fromYear, fromMonth, fromDays)
	to, _ := betacalendars.NewDate(toYear, toMonth, 1)
	return MonthTransition{From: from, To: to, FromMonthDays: fromDays, ToMonthDays: toDays, MonthLengthChanged: fromDays != toDays, YearChanged: fromYear != toYear}, nil
}

func addMonthTransition(report *Report, transition MonthTransition) {
	report.MonthTransitions = append(report.MonthTransitions, transition)
	if transition.MonthLengthChanged {
		report.MonthLengthChanges = append(report.MonthLengthChanges, transition)
	}
}

func isoTransitionsTouching(year int) ([]ISOWeekYearTransition, error) {
	startYear := year
	if year > 1 {
		startYear--
	}
	endYear := year
	if year < 9999 {
		endYear++
	}
	start, _ := betacalendars.NewDate(startYear, time.January, 1)
	end, _ := betacalendars.NewDate(endYear, time.December, 31)
	result := make([]ISOWeekYearTransition, 0, 2)
	previous := start
	previousWeek, previousISOYear := previous.ISOWeek()
	for current, err := previous.AddDays(1); err == nil && !current.After(end); current, err = current.AddDays(1) {
		week, isoYear := current.ISOWeek()
		if week != previousWeek && isoYear != previousISOYear && (previous.Year() == year || current.Year() == year) {
			result = append(result, ISOWeekYearTransition{Before: previous, After: current, BeforeISOWeek: previousWeek, AfterISOWeek: week, BeforeISOYear: previousISOYear, AfterISOYear: isoYear})
		}
		previous, previousWeek, previousISOYear = current, week, isoYear
	}
	if len(result) == 0 {
		result = []ISOWeekYearTransition{}
	}
	return result, nil
}

// String returns a concise description useful in test failure output.
func (t MonthTransition) String() string {
	return fmt.Sprintf("%s (%d days) -> %s (%d days)", t.From, t.FromMonthDays, t.To, t.ToMonthDays)
}
