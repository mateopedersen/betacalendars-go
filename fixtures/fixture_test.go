package fixtures_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/mateopedersen/betacalendars-go/fixtures"
	"github.com/mateopedersen/betacalendars-go/grid"
)

func TestFixtureSerializersAreStable(t *testing.T) {
	month, err := fixtures.Month(2027, time.January, grid.Options{})
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := month.JSON()
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := month.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) || len(firstJSON) == 0 || firstJSON[len(firstJSON)-1] != '\n' {
		t.Fatal("JSON output is not stable LF-terminated")
	}
	csv, err := month.CSV()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(csv, []byte("kind,year,month,row,column,date")) {
		t.Fatalf("unexpected CSV header: %s", csv[:40])
	}
}

func TestLeapAndBoundaryDatasets(t *testing.T) {
	leaps, err := fixtures.LeapYears(2028, 1900, 2000, 2100)
	if err != nil {
		t.Fatal(err)
	}
	data, err := leaps.JSON()
	if err != nil {
		t.Fatal(err)
	}
	wantOrder := []byte(`"year": 1900`)
	if !bytes.Contains(data, wantOrder) {
		t.Fatalf("sorted JSON lacks first year: %s", data)
	}
	boundaries, err := fixtures.ISOWeekBoundaries(2021, 2020)
	if err != nil {
		t.Fatal(err)
	}
	if len(boundaries.Boundaries) != 2 || boundaries.Boundaries[0].Year != 2020 {
		t.Fatalf("boundaries not sorted: %+v", boundaries.Boundaries)
	}
	if _, err := boundaries.CSV(); err != nil {
		t.Fatal(err)
	}
	if _, err := fixtures.LeapYears(2027, 2027); err == nil {
		t.Fatal("duplicate year should fail")
	}
}

func ExampleYear() {
	data, _ := fixtures.Year(2027, grid.Options{Layout: grid.Compact})
	fmt.Println(data.Year, len(data.YearGrid.Months), data.YearGrid.Months[0].Rows)
	// Output: 2027 12 5
}
