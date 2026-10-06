# betacalendars-go

A dependency-free Go library for Gregorian dates, calendar grids, recurrence rules, boundary fixtures, and November-through-February year-turn views.

- Module: `github.com/mateopedersen/betacalendars-go`
- Go: 1.22 or newer
- License: MIT

## Packages

| Package | Purpose |
| --- | --- |
| `betacalendars` | Validated civil dates, Gregorian leap years, date arithmetic, ISO weeks, and weekday types |
| `grid` | Monday-first or configurable month/year grids, fixed six-week or compact layouts, adjacent or blank overflow cells |
| `recurrence` | Bounded daily, weekly, monthly, nth-weekday, last-weekday, and annual rules |
| `boundary` | Month-length, year-turn, leap-day, and ISO week-year transitions |
| `fixtures` | Deterministic JSON and CSV fixture datasets |
| `yearturn` | November, December, following January, and February window |
| `cmd/betacal` | Optional text, JSON, and CSV command-line tool |

## Quick start

```go
package main

import (
    "fmt"
    "time"

    "github.com/mateopedersen/betacalendars-go"
    "github.com/mateopedersen/betacalendars-go/grid"
)

func main() {
    date, err := betacalendars.NewDate(2028, time.February, 29)
    if err != nil {
        panic(err)
    }
    fmt.Println(date, date.Weekday())

    month, err := grid.Month(2028, time.February, grid.Options{})
    if err != nil {
        panic(err)
    }
    fmt.Printf("%d rows, %d cells\n", month.Rows, len(month.Cells))
}
```

The zero `grid.Options` value produces a Monday-first, fixed six-row grid populated with adjacent dates. Choose `grid.Compact` for four-to-six rows, `grid.Blank` for empty overflow positions, or set `WeekStart` to another weekday.

## Dates and week numbers

`Date` is a civil date without a time zone. Construct it with `NewDate` or parse strict `YYYY-MM-DD` strings with `ParseDate`. Supported years are 1–9999. Date arithmetic reports `ErrDateOutOfRange` rather than wrapping. `ISOWeek` returns `(week, weekYear)`, matching Go's `time.Time.ISOWeek` tuple ordering.

## Recurrence

Create a rule with a constructor and expand it over an inclusive, finite date range:

```go
rule, err := recurrence.Annual(time.February, 29, recurrence.Clamp)
if err != nil { panic(err) }
occurrences, err := recurrence.Expand(rule, recurrence.Options{
    Start: mustDate("2024-01-01"),
    End: mustDate("2028-12-31"),
    Limit: 10,
})
```

Monthly and annual rules accept `Skip`, `Clamp`, or `Error` when the requested day does not exist. Expansion requires a positive occurrence limit and returns an error rather than silently truncating.

## Fixtures and CLI

Fixture JSON is indented, stable by struct field order, and ends in LF. CSV uses deterministic column order and LF records. Run the optional CLI with `go run ./cmd/betacal`:

```text
betacal month 2028 2 [text|json|csv]
betacal year 2028 [text|json|csv]
betacal boundaries 2028 [json|csv]
betacal yearturn 2028 [text|json|csv]
```

## Development

```sh
go test ./...
go vet ./...
go test -race ./...
go mod tidy
```

The CI matrix covers Go 1.22, 1.26, and 1.27 on Linux, macOS, and Windows.

## Versioning

The module path is tied to the public repository path. Published versions are immutable; release a new semantic version tag for changes.
