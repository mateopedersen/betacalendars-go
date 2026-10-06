package recurrence_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mateopedersen/betacalendars-go"
	"github.com/mateopedersen/betacalendars-go/recurrence"
)

func mustDate(t *testing.T, text string) betacalendars.Date {
	t.Helper()
	d, err := betacalendars.ParseDate(text)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func texts(dates []betacalendars.Date) []string {
	out := make([]string, len(dates))
	for i, d := range dates {
		out[i] = d.String()
	}
	return out
}

func TestDailyAndWeekly(t *testing.T) {
	rule, err := recurrence.Daily(2)
	if err != nil {
		t.Fatal(err)
	}
	got, err := recurrence.Expand(rule, recurrence.Options{Start: mustDate(t, "2027-01-01"), End: mustDate(t, "2027-01-06"), Limit: 10})
	if err != nil || !reflect.DeepEqual(texts(got), []string{"2027-01-01", "2027-01-03", "2027-01-05"}) {
		t.Fatalf("daily = %v, %v", texts(got), err)
	}
	weekly, err := recurrence.Weekly(1, time.Tuesday, time.Thursday)
	if err != nil {
		t.Fatal(err)
	}
	got, err = recurrence.Expand(weekly, recurrence.Options{Start: mustDate(t, "2027-01-06"), End: mustDate(t, "2027-01-14"), Limit: 10})
	if err != nil || !reflect.DeepEqual(texts(got), []string{"2027-01-07", "2027-01-12", "2027-01-14"}) {
		t.Fatalf("weekly = %v, %v", texts(got), err)
	}
}

func TestMonthlyDayPolicies(t *testing.T) {
	start, end := mustDate(t, "2027-01-01"), mustDate(t, "2027-03-31")
	for _, test := range []struct {
		policy  recurrence.InvalidDayPolicy
		want    []string
		wantErr bool
	}{{recurrence.Skip, []string{"2027-01-31", "2027-03-31"}, false}, {recurrence.Clamp, []string{"2027-01-31", "2027-02-28", "2027-03-31"}, false}, {recurrence.Error, nil, true}} {
		rule, _ := recurrence.MonthlyDay(31, 1, test.policy)
		got, err := recurrence.Expand(rule, recurrence.Options{Start: start, End: end, Limit: 10})
		if (err != nil) != test.wantErr {
			t.Fatalf("policy %d error = %v", test.policy, err)
		}
		if !test.wantErr && !reflect.DeepEqual(texts(got), test.want) {
			t.Errorf("policy %d = %v", test.policy, texts(got))
		}
	}
}

func TestWeekdayMonthlyAndAnnualPolicies(t *testing.T) {
	nth, _ := recurrence.NthWeekday(2, time.Tuesday, 1)
	got, err := recurrence.Expand(nth, recurrence.Options{Start: mustDate(t, "2027-01-01"), End: mustDate(t, "2027-03-31"), Limit: 5})
	if err != nil || !reflect.DeepEqual(texts(got), []string{"2027-01-12", "2027-02-09", "2027-03-09"}) {
		t.Fatalf("nth weekday = %v, %v", texts(got), err)
	}
	last, _ := recurrence.LastWeekday(time.Friday, 1)
	got, err = recurrence.Expand(last, recurrence.Options{Start: mustDate(t, "2027-01-01"), End: mustDate(t, "2027-02-28"), Limit: 5})
	if err != nil || !reflect.DeepEqual(texts(got), []string{"2027-01-29", "2027-02-26"}) {
		t.Fatalf("last weekday = %v, %v", texts(got), err)
	}
	annual, _ := recurrence.Annual(time.February, 29, recurrence.Clamp)
	got, err = recurrence.Expand(annual, recurrence.Options{Start: mustDate(t, "2024-01-01"), End: mustDate(t, "2028-12-31"), Limit: 5})
	if err != nil || !reflect.DeepEqual(texts(got), []string{"2024-02-29", "2025-02-28", "2026-02-28", "2027-02-28", "2028-02-29"}) {
		t.Fatalf("annual leap-day = %v, %v", texts(got), err)
	}
}

func TestBoundsAndInvalidInputs(t *testing.T) {
	rule, _ := recurrence.Daily(1)
	_, err := recurrence.Expand(rule, recurrence.Options{Start: mustDate(t, "2027-01-01"), End: mustDate(t, "2027-01-03"), Limit: 2})
	if !errors.Is(err, recurrence.ErrOccurrenceLimit) {
		t.Fatalf("limit error = %v", err)
	}
	if _, err := recurrence.Daily(0); !errors.Is(err, recurrence.ErrInvalidRule) {
		t.Fatalf("invalid interval = %v", err)
	}
	if _, err := recurrence.Expand(rule, recurrence.Options{Start: mustDate(t, "2027-01-03"), End: mustDate(t, "2027-01-01"), Limit: 2}); !errors.Is(err, recurrence.ErrInvalidRule) {
		t.Fatalf("invalid range = %v", err)
	}
}

func ExampleMonthlyDay() {
	rule, _ := recurrence.MonthlyDay(31, 1, recurrence.Clamp)
	start, _ := betacalendars.ParseDate("2027-01-01")
	end, _ := betacalendars.ParseDate("2027-03-31")
	dates, _ := recurrence.Expand(rule, recurrence.Options{Start: start, End: end, Limit: 4})
	for _, date := range dates {
		fmt.Println(date)
	}
	// Output:
	// 2027-01-31
	// 2027-02-28
	// 2027-03-31
}
