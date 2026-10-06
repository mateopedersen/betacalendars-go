package recurrence

import (
	"errors"
	"fmt"
	"time"

	"github.com/mateopedersen/betacalendars-go"
)

// ErrInvalidRule reports an invalid recurrence rule or expansion option.
var ErrInvalidRule = errors.New("recurrence: invalid rule")

// ErrOccurrenceLimit reports that an expansion would exceed its configured limit.
var ErrOccurrenceLimit = errors.New("recurrence: occurrence limit exceeded")

// InvalidDayPolicy controls monthly or annual rules whose requested day does not exist.
type InvalidDayPolicy uint8

const (
	// Skip omits periods without the requested day.
	Skip InvalidDayPolicy = iota
	// Clamp moves the occurrence to the final day of a short month.
	Clamp
	// Error stops expansion when an eligible period lacks the requested day.
	Error
)

type ruleKind uint8

const (
	kindDaily ruleKind = iota + 1
	kindWeekly
	kindMonthlyDay
	kindNthWeekday
	kindLastWeekday
	kindAnnual
)

// Rule is a validated, immutable recurrence description created by a constructor.
// Its zero value is invalid.
type Rule struct {
	kind     ruleKind
	interval int
	weekdays []time.Weekday
	day      int
	month    time.Month
	ordinal  int
	weekday  time.Weekday
	policy   InvalidDayPolicy
}

// Valid reports whether r was created with a valid recurrence constructor.
func (r Rule) Valid() bool {
	if r.kind < kindDaily || r.kind > kindAnnual || r.interval < 1 {
		return false
	}
	return true
}

// Daily creates a rule that occurs every interval days, anchored at the expansion start.
func Daily(interval int) (Rule, error) {
	if interval < 1 {
		return Rule{}, ErrInvalidRule
	}
	return Rule{kind: kindDaily, interval: interval}, nil
}

// Weekly creates a rule for weekdays in every interval-th week, anchored to the
// week containing the expansion start. At least one valid weekday is required.
func Weekly(interval int, weekdays ...time.Weekday) (Rule, error) {
	if interval < 1 || len(weekdays) == 0 {
		return Rule{}, ErrInvalidRule
	}
	copyDays := append([]time.Weekday(nil), weekdays...)
	seen := map[time.Weekday]bool{}
	for _, day := range copyDays {
		if day < time.Sunday || day > time.Saturday || seen[day] {
			return Rule{}, ErrInvalidRule
		}
		seen[day] = true
	}
	return Rule{kind: kindWeekly, interval: interval, weekdays: copyDays}, nil
}

// MonthlyDay creates a rule for day in every interval-th month, anchored to the
// expansion start month. policy handles months shorter than day.
func MonthlyDay(day, interval int, policy InvalidDayPolicy) (Rule, error) {
	if day < 1 || day > 31 || interval < 1 || !validPolicy(policy) {
		return Rule{}, ErrInvalidRule
	}
	return Rule{kind: kindMonthlyDay, day: day, interval: interval, policy: policy}, nil
}

// NthWeekday creates a rule for the nth weekday in every interval-th month.
// ordinal must be between 1 and 5; months without that occurrence are skipped.
func NthWeekday(ordinal int, weekday time.Weekday, interval int) (Rule, error) {
	if ordinal < 1 || ordinal > 5 || !validWeekday(weekday) || interval < 1 {
		return Rule{}, ErrInvalidRule
	}
	return Rule{kind: kindNthWeekday, ordinal: ordinal, weekday: weekday, interval: interval}, nil
}

// LastWeekday creates a rule for the final weekday in every interval-th month.
func LastWeekday(weekday time.Weekday, interval int) (Rule, error) {
	if !validWeekday(weekday) || interval < 1 {
		return Rule{}, ErrInvalidRule
	}
	return Rule{kind: kindLastWeekday, weekday: weekday, interval: interval}, nil
}

// Annual creates a rule for month and day in every year. policy handles
// February 29 in non-leap years and any other invalid month day.
func Annual(month time.Month, day int, policy InvalidDayPolicy) (Rule, error) {
	if month < time.January || month > time.December || day < 1 || day > 31 || !validPolicy(policy) {
		return Rule{}, ErrInvalidRule
	}
	return Rule{kind: kindAnnual, interval: 1, month: month, day: day, policy: policy}, nil
}

// Options bounds an expansion. Start and End are inclusive; Limit must be positive.
// The zero WeekStart is Monday and is used to anchor weekly rules.
type Options struct {
	// Start is both the first eligible date and recurrence anchor.
	Start betacalendars.Date
	// End is the final eligible date.
	End betacalendars.Date
	// Limit is the maximum number of occurrences returned.
	Limit int
	// WeekStart anchors interval weeks and defaults to Monday.
	WeekStart betacalendars.WeekStart
}

// Expand returns occurrences in the inclusive range described by options.
// It stops with ErrOccurrenceLimit instead of silently truncating results.
func Expand(rule Rule, options Options) ([]betacalendars.Date, error) {
	if !rule.Valid() || !options.Start.IsValid() || !options.End.IsValid() || options.End.Before(options.Start) || options.Limit < 1 || !options.WeekStart.Valid() {
		return nil, ErrInvalidRule
	}
	startWeekday := (int(options.Start.Weekday()) - int(options.WeekStart.Weekday()) + 7) % 7
	results := make([]betacalendars.Date, 0)
	for current, offset := options.Start, 0; !current.After(options.End); offset++ {
		match, err := matches(rule, current, options, startWeekday)
		if err != nil {
			return nil, err
		}
		if match {
			if len(results) == options.Limit {
				return nil, ErrOccurrenceLimit
			}
			results = append(results, current)
		}
		next, err := current.AddDays(1)
		if err != nil {
			break
		}
		current = next
	}
	return results, nil
}

func matches(rule Rule, date betacalendars.Date, options Options, startWeekday int) (bool, error) {
	daysFromStart, _ := betacalendars.DaysBetween(options.Start, date)
	switch rule.kind {
	case kindDaily:
		return daysFromStart%rule.interval == 0, nil
	case kindWeekly:
		weekIndex := (daysFromStart + startWeekday) / 7
		if weekIndex%rule.interval != 0 {
			return false, nil
		}
		for _, weekday := range rule.weekdays {
			if date.Weekday() == weekday {
				return true, nil
			}
		}
		return false, nil
	case kindMonthlyDay:
		monthIndex := (date.Year()-options.Start.Year())*12 + int(date.Month()-options.Start.Month())
		if monthIndex < 0 || monthIndex%rule.interval != 0 {
			return false, nil
		}
		days, _ := betacalendars.DaysInMonth(date.Year(), date.Month())
		actual, err := policyDay(rule.day, days, rule.policy, date)
		if err != nil || actual == 0 {
			return false, err
		}
		return date.Day() == actual, nil
	case kindNthWeekday, kindLastWeekday:
		monthIndex := (date.Year()-options.Start.Year())*12 + int(date.Month()-options.Start.Month())
		if monthIndex < 0 || monthIndex%rule.interval != 0 {
			return false, nil
		}
		day := weekdayInMonth(date.Year(), date.Month(), rule.weekday, rule.ordinal, rule.kind == kindLastWeekday)
		return day != 0 && date.Day() == day, nil
	case kindAnnual:
		if date.Month() != rule.month || date.Year() < options.Start.Year() {
			return false, nil
		}
		if (date.Year()-options.Start.Year())%rule.interval != 0 {
			return false, nil
		}
		days, _ := betacalendars.DaysInMonth(date.Year(), date.Month())
		actual, err := policyDay(rule.day, days, rule.policy, date)
		if err != nil || actual == 0 {
			return false, err
		}
		return date.Day() == actual, nil
	default:
		return false, ErrInvalidRule
	}
}

func policyDay(day, days int, policy InvalidDayPolicy, current betacalendars.Date) (int, error) {
	if day <= days {
		return day, nil
	}
	switch policy {
	case Skip:
		return 0, nil
	case Clamp:
		return days, nil
	case Error:
		return 0, fmt.Errorf("%w: day %d does not exist in %04d-%02d", ErrInvalidRule, day, current.Year(), current.Month())
	default:
		return 0, ErrInvalidRule
	}
}

func weekdayInMonth(year int, month time.Month, weekday time.Weekday, ordinal int, last bool) int {
	days, _ := betacalendars.DaysInMonth(year, month)
	if last {
		date, _ := betacalendars.NewDate(year, month, days)
		return days - (int(date.Weekday())-int(weekday)+7)%7
	}
	first, _ := betacalendars.NewDate(year, month, 1)
	day := 1 + (int(weekday)-int(first.Weekday())+7)%7 + (ordinal-1)*7
	if day > days {
		return 0
	}
	return day
}

func validPolicy(policy InvalidDayPolicy) bool { return policy <= Error }
func validWeekday(weekday time.Weekday) bool {
	return weekday >= time.Sunday && weekday <= time.Saturday
}
