package fixtures

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/mateopedersen/betacalendars-go"
	"github.com/mateopedersen/betacalendars-go/boundary"
	"github.com/mateopedersen/betacalendars-go/grid"
)

// ErrInvalidDataset indicates a dataset with an unsupported kind or invalid years.
var ErrInvalidDataset = errors.New("fixtures: invalid dataset")

// LeapYear is one year and its Gregorian leap-year classification.
type LeapYear struct {
	// Year is the Gregorian year.
	Year int `json:"year"`
	// IsLeap reports whether February contains 29 days.
	IsLeap bool `json:"is_leap"`
	// FebruaryDays is 28 or 29.
	FebruaryDays int `json:"february_days"`
}

// Dataset is a stable JSON or CSV representation of calendar fixture data.
type Dataset struct {
	// Schema is the fixture schema version, currently 1.
	Schema int `json:"schema"`
	// Kind identifies month, year, leap_years, boundaries, or iso_week_boundaries.
	Kind string `json:"kind"`
	// Year and Month identify single-month or single-year fixtures when applicable.
	Year  int `json:"year,omitempty"`
	Month int `json:"month,omitempty"`
	// MonthGrid contains one month fixture when Kind is month.
	MonthGrid *grid.MonthGrid `json:"month_grid,omitempty"`
	// YearGrid contains twelve month fixtures when Kind is year.
	YearGrid *grid.YearGrid `json:"year_grid,omitempty"`
	// LeapYears contains sorted leap-year classification records.
	LeapYears []LeapYear `json:"leap_years,omitempty"`
	// Boundaries contains sorted boundary reports.
	Boundaries []boundary.Report `json:"boundaries,omitempty"`
}

// Month creates a deterministic grid fixture for a single month.
func Month(year int, month time.Month, options grid.Options) (Dataset, error) {
	result, err := grid.Month(year, month, options)
	if err != nil {
		return Dataset{}, err
	}
	return Dataset{Schema: 1, Kind: "month", Year: year, Month: int(month), MonthGrid: &result}, nil
}

// Year creates deterministic January-through-December grid fixtures.
func Year(year int, options grid.Options) (Dataset, error) {
	result, err := grid.Year(year, options)
	if err != nil {
		return Dataset{}, err
	}
	return Dataset{Schema: 1, Kind: "year", Year: year, YearGrid: &result}, nil
}

// LeapYears classifies the supplied years in ascending order.
func LeapYears(years ...int) (Dataset, error) {
	normalized, err := normalizeYears(years)
	if err != nil {
		return Dataset{}, err
	}
	records := make([]LeapYear, 0, len(normalized))
	for _, year := range normalized {
		leap := betacalendars.IsLeapYear(year)
		days := 28
		if leap {
			days = 29
		}
		records = append(records, LeapYear{Year: year, IsLeap: leap, FebruaryDays: days})
	}
	return Dataset{Schema: 1, Kind: "leap_years", LeapYears: records}, nil
}

// Boundaries generates structured transition fixtures for the supplied years.
func Boundaries(years ...int) (Dataset, error) {
	normalized, err := normalizeYears(years)
	if err != nil {
		return Dataset{}, err
	}
	reports := make([]boundary.Report, 0, len(normalized))
	for _, year := range normalized {
		report, err := boundary.ForYear(year)
		if err != nil {
			return Dataset{}, err
		}
		reports = append(reports, report)
	}
	return Dataset{Schema: 1, Kind: "boundaries", Boundaries: reports}, nil
}

// ISOWeekBoundaries generates reports containing ISO week-year transitions
// for the supplied years.
func ISOWeekBoundaries(years ...int) (Dataset, error) {
	normalized, err := normalizeYears(years)
	if err != nil {
		return Dataset{}, err
	}
	reports := make([]boundary.Report, 0, len(normalized))
	for _, year := range normalized {
		report, err := boundary.ForYear(year)
		if err != nil {
			return Dataset{}, err
		}
		reports = append(reports, boundary.Report{Year: year, ISOWeekYearTransitions: report.ISOWeekYearTransitions})
	}
	return Dataset{Schema: 1, Kind: "iso_week_boundaries", Boundaries: reports}, nil
}

// JSON returns indented JSON with a final LF and stable struct-field ordering.
func (d Dataset) JSON() ([]byte, error) {
	encoded, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

// CSV returns a deterministic comma-separated representation with LF line endings.
func (d Dataset) CSV() ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	writer.UseCRLF = false
	switch d.Kind {
	case "month":
		if d.MonthGrid == nil {
			return nil, ErrInvalidDataset
		}
		if err := writeGridCSV(writer, []grid.MonthGrid{*d.MonthGrid}); err != nil {
			return nil, err
		}
	case "year":
		if d.YearGrid == nil {
			return nil, ErrInvalidDataset
		}
		if err := writeGridCSV(writer, d.YearGrid.Months); err != nil {
			return nil, err
		}
	case "leap_years":
		if err := writer.Write([]string{"kind", "year", "is_leap", "february_days"}); err != nil {
			return nil, err
		}
		for _, record := range d.LeapYears {
			if err := writer.Write([]string{"leap_year", strconv.Itoa(record.Year), strconv.FormatBool(record.IsLeap), strconv.Itoa(record.FebruaryDays)}); err != nil {
				return nil, err
			}
		}
	case "boundaries", "iso_week_boundaries":
		if err := writeBoundaryCSV(writer, d); err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidDataset
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeGridCSV(writer *csv.Writer, months []grid.MonthGrid) error {
	if err := writer.Write([]string{"kind", "year", "month", "row", "column", "date", "weekday", "in_month", "weekend", "month_start", "month_end", "year_start", "year_end", "iso_week", "iso_week_year"}); err != nil {
		return err
	}
	for _, month := range months {
		for _, cell := range month.Cells {
			date := ""
			week, weekYear := "", ""
			if cell.HasDate {
				date = cell.Date.String()
				week = strconv.Itoa(cell.ISOWeek)
				weekYear = strconv.Itoa(cell.ISOWeekYear)
			}
			record := []string{"grid", strconv.Itoa(month.Year), strconv.Itoa(int(month.Month)), strconv.Itoa(cell.Row), strconv.Itoa(cell.Column), date, cell.Weekday.String(), strconv.FormatBool(cell.InMonth), strconv.FormatBool(cell.Weekend), strconv.FormatBool(cell.MonthStart), strconv.FormatBool(cell.MonthEnd), strconv.FormatBool(cell.YearStart), strconv.FormatBool(cell.YearEnd), week, weekYear}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeBoundaryCSV(writer *csv.Writer, dataset Dataset) error {
	if err := writer.Write([]string{"kind", "year", "category", "from", "to", "from_month_days", "to_month_days", "iso_week_before", "iso_year_before", "iso_week_after", "iso_year_after"}); err != nil {
		return err
	}
	for _, report := range dataset.Boundaries {
		for _, transition := range report.MonthTransitions {
			if dataset.Kind == "iso_week_boundaries" {
				continue
			}
			if err := writer.Write([]string{"boundary", strconv.Itoa(report.Year), "month", transition.From.String(), transition.To.String(), strconv.Itoa(transition.FromMonthDays), strconv.Itoa(transition.ToMonthDays), "", "", "", ""}); err != nil {
				return err
			}
		}
		for _, transition := range report.ISOWeekYearTransitions {
			if err := writer.Write([]string{"boundary", strconv.Itoa(report.Year), "iso_week_year", transition.Before.String(), transition.After.String(), "", "", strconv.Itoa(transition.BeforeISOWeek), strconv.Itoa(transition.BeforeISOYear), strconv.Itoa(transition.AfterISOWeek), strconv.Itoa(transition.AfterISOYear)}); err != nil {
				return err
			}
		}
		for _, transition := range report.LeapDayTransitions {
			if dataset.Kind == "iso_week_boundaries" {
				continue
			}
			if err := writer.Write([]string{"boundary", strconv.Itoa(report.Year), "leap_day", transition.Before.String(), transition.After.String(), "28", "31", "", "", "", ""}); err != nil {
				return err
			}
		}
	}
	return nil
}

func normalizeYears(years []int) ([]int, error) {
	copyYears := append([]int(nil), years...)
	for _, year := range copyYears {
		if year < 1 || year > 9999 {
			return nil, fmt.Errorf("%w: year %d", ErrInvalidDataset, year)
		}
	}
	sort.Ints(copyYears)
	for index := 1; index < len(copyYears); index++ {
		if copyYears[index] == copyYears[index-1] {
			return nil, fmt.Errorf("%w: duplicate year %d", ErrInvalidDataset, copyYears[index])
		}
	}
	if copyYears == nil {
		copyYears = []int{}
	}
	return copyYears, nil
}
