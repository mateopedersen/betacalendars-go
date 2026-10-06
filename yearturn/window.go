package yearturn

import (
	"fmt"
	"time"

	"github.com/mateopedersen/betacalendars-go"
	"github.com/mateopedersen/betacalendars-go/grid"
)

// Window contains month grids from November through February across a year turn.
type Window struct {
	// Year is the year containing November and December.
	Year int
	// Months contains November, December, January, and February in that order.
	Months [4]grid.MonthGrid
}

// New builds the default Monday-first, fixed-six-week, adjacent-date window
// beginning with November of year. Supported starting years are 1 through 9998.
func New(year int) (Window, error) {
	return NewWithOptions(year, grid.Options{})
}

// NewWithOptions builds the November-through-February window using options for
// each month grid. The starting year must be in 1 through 9998.
func NewWithOptions(year int, options grid.Options) (Window, error) {
	if year < 1 || year > 9998 {
		return Window{}, fmt.Errorf("yearturn: %w", betacalendars.ErrInvalidDate)
	}
	months := [4]grid.MonthGrid{}
	requested := [...]struct {
		year  int
		month time.Month
	}{
		{year, time.November}, {year, time.December}, {year + 1, time.January}, {year + 1, time.February},
	}
	for index, item := range requested {
		month, err := grid.Month(item.year, item.month, options)
		if err != nil {
			return Window{}, err
		}
		months[index] = month
	}
	return Window{Year: year, Months: months}, nil
}

// DayCounts returns the number of days in each month, in window order.
func (w Window) DayCounts() [4]int {
	var counts [4]int
	for i, month := range w.Months {
		for _, cell := range month.Cells {
			if cell.InMonth {
				counts[i]++
			}
		}
	}
	return counts
}
