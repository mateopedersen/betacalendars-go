package grid

import (
	"errors"
	"fmt"
	"time"

	"github.com/mateopedersen/betacalendars-go"
)

// Layout selects the number of rows used for a month grid.
type Layout uint8

const (
	// FixedSixWeeks always emits six rows and 42 cells.
	FixedSixWeeks Layout = iota
	// Compact emits only the four, five, or six rows needed for the month.
	Compact
)

// Overflow controls how cells outside the requested month are represented.
type Overflow uint8

const (
	// AdjacentDates fills overflow cells with actual neighboring civil dates.
	AdjacentDates Overflow = iota
	// Blank leaves overflow dates unset while retaining the grid position and weekday.
	Blank
)

// Options configures a month or year grid. Its zero value uses Monday, a fixed
// six-week layout, and adjacent dates.
type Options struct {
	// WeekStart is the first weekday in each row.
	WeekStart betacalendars.WeekStart
	// Layout selects FixedSixWeeks or Compact.
	Layout Layout
	// Overflow selects AdjacentDates or Blank.
	Overflow Overflow
}

// Cell is one position in a seven-column month grid. Date is invalid when
// HasDate is false, which occurs for blank overflow cells.
type Cell struct {
	// Date is the actual civil date when HasDate is true.
	Date betacalendars.Date
	// HasDate distinguishes real dates from blank overflow positions.
	HasDate bool
	// Row and Column are zero-based grid coordinates.
	Row    int
	Column int
	// Weekday is the weekday of this column.
	Weekday time.Weekday
	// InMonth reports whether Date belongs to the requested month.
	InMonth bool
	// Weekend reports whether the column is Saturday or Sunday.
	Weekend bool
	// MonthStart and MonthEnd mark the first and last days of the requested month.
	MonthStart bool
	MonthEnd   bool
	// YearStart and YearEnd mark January 1 and December 31.
	YearStart bool
	YearEnd   bool
	// ISOWeek and ISOWeekYear contain ISO 8601 values for real dates, or zero for blanks.
	ISOWeek     int
	ISOWeekYear int
}

// MonthGrid is a row-major seven-column grid for one Gregorian month.
type MonthGrid struct {
	// Year and Month identify the requested Gregorian month.
	Year  int
	Month time.Month
	// WeekStart, Layout, and Overflow record the options used to build the grid.
	WeekStart betacalendars.WeekStart
	Layout    Layout
	Overflow  Overflow
	// Rows is 6 for FixedSixWeeks and 4, 5, or 6 for Compact.
	Rows int
	// Cells contains Rows*7 cells in row-major order.
	Cells []Cell
}

// YearGrid contains January through December for one Gregorian year.
type YearGrid struct {
	// Year is the requested Gregorian year.
	Year int
	// Months contains exactly twelve month grids in calendar order.
	Months []MonthGrid
}

var errOptions = errors.New("grid: invalid options")

// Month builds a deterministic grid for year and month.
func Month(year int, month time.Month, options Options) (MonthGrid, error) {
	if _, err := betacalendars.NewDate(year, month, 1); err != nil {
		return MonthGrid{}, fmt.Errorf("grid: %w", err)
	}
	if err := validateOptions(options); err != nil {
		return MonthGrid{}, err
	}
	first, _ := betacalendars.NewDate(year, month, 1)
	days, _ := betacalendars.DaysInMonth(year, month)
	lead := (int(first.Weekday()) - int(options.WeekStart.Weekday()) + 7) % 7
	rows := 6
	if options.Layout == Compact {
		rows = (lead + days + 6) / 7
	}
	cells := make([]Cell, 0, rows*7)
	for index := 0; index < rows*7; index++ {
		row, column := index/7, index%7
		day := index - lead + 1
		weekday := options.WeekStart.Weekday() + time.Weekday(column)
		if weekday > time.Saturday {
			weekday -= 7
		}
		cell := Cell{Row: row, Column: column, Weekday: weekday, Weekend: weekday == time.Saturday || weekday == time.Sunday}
		if day >= 1 && day <= days {
			date, _ := betacalendars.NewDate(year, month, day)
			cells = append(cells, makeCell(date, true, row, column, weekday, true, day == 1, day == days))
			continue
		}
		if options.Overflow == AdjacentDates {
			date, err := first.AddDays(day - 1)
			if err != nil {
				if errors.Is(err, betacalendars.ErrDateOutOfRange) {
					cells = append(cells, cell)
					continue
				}
				return MonthGrid{}, fmt.Errorf("grid: adjacent date: %w", err)
			}
			cells = append(cells, makeCell(date, true, row, column, weekday, false, false, false))
			continue
		}
		cells = append(cells, cell)
	}
	return MonthGrid{Year: year, Month: month, WeekStart: options.WeekStart, Layout: options.Layout, Overflow: options.Overflow, Rows: rows, Cells: cells}, nil
}

// Year builds twelve month grids for year using the same options for every month.
func Year(year int, options Options) (YearGrid, error) {
	if year < 1 || year > 9999 {
		return YearGrid{}, betacalendars.ErrInvalidDate
	}
	if err := validateOptions(options); err != nil {
		return YearGrid{}, err
	}
	months := make([]MonthGrid, 0, 12)
	for month := time.January; month <= time.December; month++ {
		grid, err := Month(year, month, options)
		if err != nil {
			return YearGrid{}, err
		}
		months = append(months, grid)
	}
	return YearGrid{Year: year, Months: months}, nil
}

func makeCell(date betacalendars.Date, hasDate bool, row, column int, weekday time.Weekday, inMonth, monthStart, monthEnd bool) Cell {
	cell := Cell{Date: date, HasDate: hasDate, Row: row, Column: column, Weekday: weekday, InMonth: inMonth, Weekend: weekday == time.Saturday || weekday == time.Sunday, MonthStart: monthStart, MonthEnd: monthEnd}
	if hasDate {
		cell.YearStart = date.Month() == time.January && date.Day() == 1
		cell.YearEnd = date.Month() == time.December && date.Day() == 31
		cell.ISOWeek, cell.ISOWeekYear = date.ISOWeek()
	}
	return cell
}

func validateOptions(options Options) error {
	if !options.WeekStart.Valid() || options.Layout > Compact || options.Overflow > Blank {
		return errOptions
	}
	return nil
}
