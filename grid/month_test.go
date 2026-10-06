package grid_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mateopedersen/betacalendars-go"
	"github.com/mateopedersen/betacalendars-go/grid"
)

func TestFixedAndCompactLayouts(t *testing.T) {
	fixed, err := grid.Month(2027, time.January, grid.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(fixed.Cells) != 42 || fixed.Rows != 6 {
		t.Fatalf("fixed cells/rows = %d/%d", len(fixed.Cells), fixed.Rows)
	}
	for _, test := range []struct {
		year     int
		month    time.Month
		wantRows int
	}{{2021, time.February, 4}, {2027, time.February, 4}, {2026, time.August, 6}} {
		compact, err := grid.Month(test.year, test.month, grid.Options{Layout: grid.Compact})
		if err != nil {
			t.Fatal(err)
		}
		if compact.Rows != test.wantRows {
			t.Errorf("%04d-%02d rows = %d, want %d", test.year, test.month, compact.Rows, test.wantRows)
		}
	}
}

func TestWeekStartsAndMonthDateCoverage(t *testing.T) {
	for weekStart := betacalendars.Monday; weekStart <= betacalendars.Sunday; weekStart++ {
		month, err := grid.Month(2028, time.February, grid.Options{WeekStart: weekStart})
		if err != nil {
			t.Fatal(err)
		}
		if month.Cells[0].Weekday != weekStart.Weekday() {
			t.Errorf("week start %s begins on %s", weekStart, month.Cells[0].Weekday)
		}
		seen := map[string]int{}
		for _, cell := range month.Cells {
			if cell.InMonth {
				seen[cell.Date.String()]++
			}
		}
		if len(seen) != 29 {
			t.Errorf("week start %s has %d distinct February dates", weekStart, len(seen))
		}
		for day := 1; day <= 29; day++ {
			date, _ := betacalendars.NewDate(2028, time.February, day)
			if seen[date.String()] != 1 {
				t.Errorf("%s appears %d times", date, seen[date.String()])
			}
		}
	}
}

func TestBlankOverflowAndAdjacentContinuity(t *testing.T) {
	blank, err := grid.Month(2027, time.January, grid.Options{Overflow: grid.Blank})
	if err != nil {
		t.Fatal(err)
	}
	foundBlank := false
	for _, cell := range blank.Cells {
		if !cell.InMonth && !cell.HasDate {
			foundBlank = true
			if cell.Date.IsValid() {
				t.Fatal("blank cell has a date")
			}
		}
	}
	if !foundBlank {
		t.Fatal("expected blank overflow cells")
	}
	adjacent, err := grid.Month(2027, time.January, grid.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(adjacent.Cells); i++ {
		previous, current := adjacent.Cells[i-1], adjacent.Cells[i]
		if previous.HasDate && current.HasDate {
			next, err := previous.Date.AddDays(1)
			if err != nil || !next.Equal(current.Date) {
				t.Fatalf("nonconsecutive cells: %s, %s", previous.Date, current.Date)
			}
		}
	}
}

func TestYearGrid(t *testing.T) {
	year, err := grid.Year(2028, grid.Options{Layout: grid.Compact})
	if err != nil {
		t.Fatal(err)
	}
	if len(year.Months) != 12 || year.Months[0].Month != time.January || year.Months[11].Month != time.December {
		t.Fatalf("wrong year month list")
	}
}

func FuzzMonthGrid(f *testing.F) {
	f.Add(uint16(2027), uint8(1), uint8(0), uint8(0), uint8(0))
	f.Add(uint16(1900), uint8(2), uint8(6), uint8(1), uint8(1))
	f.Fuzz(func(t *testing.T, year uint16, month, week, layout, overflow uint8) {
		y := int(year%9999) + 1
		m := time.Month(month%12 + 1)
		options := grid.Options{WeekStart: betacalendars.WeekStart(week % 7), Layout: grid.Layout(layout % 2), Overflow: grid.Overflow(overflow % 2)}
		got, err := grid.Month(y, m, options)
		if err != nil {
			t.Fatal(err)
		}
		if options.Layout == grid.FixedSixWeeks && len(got.Cells) != 42 {
			t.Fatalf("fixed layout has %d cells", len(got.Cells))
		}
		counts := map[int]int{}
		for _, cell := range got.Cells {
			if cell.InMonth {
				counts[cell.Date.Day()]++
			}
		}
		days, _ := betacalendars.DaysInMonth(y, m)
		if len(counts) != days {
			t.Fatalf("got %d month days, want %d", len(counts), days)
		}
		for day := 1; day <= days; day++ {
			if counts[day] != 1 {
				t.Fatalf("day %d occurs %d times", day, counts[day])
			}
		}
	})
}

func ExampleMonth() {
	month, _ := grid.Month(2027, time.January, grid.Options{Layout: grid.Compact})
	fmt.Println(month.Rows, len(month.Cells))
	// Output: 5 35
}
