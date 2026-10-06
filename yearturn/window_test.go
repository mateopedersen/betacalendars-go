package yearturn_test

import (
	"fmt"
	"testing"

	"github.com/mateopedersen/betacalendars-go/yearturn"
)

func TestCrossYearMonthLengths(t *testing.T) {
	for _, test := range []struct {
		year int
		want [4]int
	}{{2026, [4]int{30, 31, 31, 28}}, {2027, [4]int{30, 31, 31, 29}}} {
		window, err := yearturn.New(test.year)
		if err != nil {
			t.Fatal(err)
		}
		if got := window.DayCounts(); got != test.want {
			t.Errorf("year turn %d day counts = %v, want %v", test.year, got, test.want)
		}
		if window.Months[0].Month.String() != "November" || window.Months[3].Month.String() != "February" {
			t.Errorf("unexpected month order: %+v", window.Months)
		}
	}
}

func TestYearTurnRange(t *testing.T) {
	if _, err := yearturn.New(0); err == nil {
		t.Fatal("expected invalid year")
	}
	if _, err := yearturn.New(9999); err == nil {
		t.Fatal("expected out-of-range next year")
	}
}

func FuzzYearTurn(f *testing.F) {
	f.Add(uint16(2026))
	f.Add(uint16(2027))
	f.Fuzz(func(t *testing.T, year uint16) {
		y := int(year%9998) + 1
		window, err := yearturn.New(y)
		if err != nil {
			t.Fatal(err)
		}
		counts := window.DayCounts()
		if counts[0] != 30 || counts[1] != 31 || counts[2] != 31 || (counts[3] != 28 && counts[3] != 29) {
			t.Fatalf("year turn %d has %v", y, counts)
		}
	})
}

func ExampleNew() {
	window, _ := yearturn.New(2026)
	fmt.Println(window.DayCounts())
	// Output: [30 31 31 28]
}
