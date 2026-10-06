package betacalendars_test

import (
	"fmt"
	"time"

	"github.com/mateopedersen/betacalendars-go"
)

func ExampleNewDate() {
	date, _ := betacalendars.NewDate(2027, time.January, 1)
	fmt.Println(date, date.Weekday())
	// Output: 2027-01-01 Friday
}
