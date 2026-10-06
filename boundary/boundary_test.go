package boundary_test

import (
	"testing"

	"github.com/mateopedersen/betacalendars-go/boundary"
)

func TestYearReport(t *testing.T) {
	report, err := boundary.ForYear(2028)
	if err != nil {
		t.Fatal(err)
	}
	if !report.HasYearTransition || report.YearTransition.From.String() != "2028-12-31" || report.YearTransition.To.String() != "2029-01-01" {
		t.Fatalf("year transition = %+v", report.YearTransition)
	}
	if len(report.MonthTransitions) != 12 || len(report.LeapDayTransitions) != 1 {
		t.Fatalf("report transitions = %d months, %d leap days", len(report.MonthTransitions), len(report.LeapDayTransitions))
	}
	leap := report.LeapDayTransitions[0]
	if leap.Before.String() != "2028-02-28" || leap.LeapDay.String() != "2028-02-29" || leap.After.String() != "2028-03-01" {
		t.Fatalf("leap transition = %+v", leap)
	}
}

func TestISOWeekYearTransition(t *testing.T) {
	report, err := boundary.ForYear(2021)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, transition := range report.ISOWeekYearTransitions {
		if transition.Before.String() == "2021-01-03" && transition.After.String() == "2021-01-04" && transition.BeforeISOWeek == 53 && transition.AfterISOWeek == 1 && transition.BeforeISOYear == 2020 && transition.AfterISOYear == 2021 {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing ISO boundary for 2021: %+v", report.ISOWeekYearTransitions)
	}
	if len(report.LeapDayTransitions) != 0 {
		t.Fatalf("non-leap year has leap transitions")
	}
}

func TestMonthLengthChangesAndRange(t *testing.T) {
	report, err := boundary.ForYear(2027)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.MonthLengthChanges) == 0 {
		t.Fatal("expected month length changes")
	}
	if _, err := boundary.ForYear(10000); err == nil {
		t.Fatal("expected year range error")
	}
}
