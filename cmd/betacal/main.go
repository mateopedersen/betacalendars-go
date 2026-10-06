// Command betacal prints calendar grids and boundary datasets.
package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/mateopedersen/betacalendars-go"
	"github.com/mateopedersen/betacalendars-go/fixtures"
	"github.com/mateopedersen/betacalendars-go/grid"
	"github.com/mateopedersen/betacalendars-go/yearturn"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string, output *os.File) error {
	if len(args) == 0 {
		return errors.New("usage: betacal month YEAR MONTH [text|json|csv] | year YEAR [text|json|csv] | boundaries YEAR [json|csv] | yearturn YEAR [text|json|csv]")
	}
	format := "text"
	if len(args) > 0 && (args[len(args)-1] == "text" || args[len(args)-1] == "json" || args[len(args)-1] == "csv") {
		format = args[len(args)-1]
		args = args[:len(args)-1]
	}
	switch args[0] {
	case "month":
		if len(args) != 3 {
			return errors.New("usage: betacal month YEAR MONTH [text|json|csv]")
		}
		year, err := strconv.Atoi(args[1])
		if err != nil {
			return betacalendars.ErrInvalidDate
		}
		month, err := strconv.Atoi(args[2])
		if err != nil {
			return betacalendars.ErrInvalidDate
		}
		data, err := fixtures.Month(year, time.Month(month), grid.Options{})
		if err != nil {
			return err
		}
		return printDataset(output, format, data, data.MonthGrid)
	case "year":
		if len(args) != 2 {
			return errors.New("usage: betacal year YEAR [text|json|csv]")
		}
		year, err := strconv.Atoi(args[1])
		if err != nil {
			return betacalendars.ErrInvalidDate
		}
		data, err := fixtures.Year(year, grid.Options{})
		if err != nil {
			return err
		}
		return printDataset(output, format, data, data.YearGrid)
	case "boundaries":
		if len(args) != 2 {
			return errors.New("usage: betacal boundaries YEAR [json|csv]")
		}
		year, err := strconv.Atoi(args[1])
		if err != nil {
			return betacalendars.ErrInvalidDate
		}
		data, err := fixtures.Boundaries(year)
		if err != nil {
			return err
		}
		return printDataset(output, format, data, data.Boundaries)
	case "yearturn":
		if len(args) != 2 {
			return errors.New("usage: betacal yearturn YEAR [text|json|csv]")
		}
		year, err := strconv.Atoi(args[1])
		if err != nil {
			return betacalendars.ErrInvalidDate
		}
		window, err := yearturn.New(year)
		if err != nil {
			return err
		}
		if format == "text" {
			for _, month := range window.Months {
				count := 0
				for _, cell := range month.Cells {
					if cell.InMonth {
						count++
					}
				}
				if _, err := fmt.Fprintf(output, "%04d-%02d %d\n", month.Year, month.Month, count); err != nil {
					return err
				}
			}
			return nil
		}
		return printValue(output, format, struct {
			Year   int               `json:"year"`
			Months [4]grid.MonthGrid `json:"months"`
		}{Year: window.Year, Months: window.Months})
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printDataset(output io.Writer, format string, data fixtures.Dataset, textValue any) error {
	switch format {
	case "json":
		encoded, err := data.JSON()
		if err != nil {
			return err
		}
		_, err = output.Write(encoded)
		return err
	case "csv":
		encoded, err := data.CSV()
		if err != nil {
			return err
		}
		_, err = output.Write(encoded)
		return err
	default:
		return printText(output, textValue)
	}
}

func printValue(output io.Writer, format string, value any) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case "csv":
		writer := csv.NewWriter(output)
		if err := writer.Write([]string{"month_year", "month", "rows", "date", "in_month"}); err != nil {
			return err
		}
		for _, month := range value.(struct {
			Year   int               `json:"year"`
			Months [4]grid.MonthGrid `json:"months"`
		}).Months {
			for _, cell := range month.Cells {
				date := ""
				if cell.HasDate {
					date = cell.Date.String()
				}
				if err := writer.Write([]string{strconv.Itoa(month.Year), month.Month.String(), strconv.Itoa(month.Rows), date, strconv.FormatBool(cell.InMonth)}); err != nil {
					return err
				}
			}
		}
		writer.Flush()
		return writer.Error()
	default:
		return printText(output, value)
	}
}

func printText(output io.Writer, value any) error {
	switch typed := value.(type) {
	case *grid.MonthGrid:
		for row := 0; row < typed.Rows; row++ {
			for col := 0; col < 7; col++ {
				cell := typed.Cells[row*7+col]
				value := "  "
				if cell.HasDate {
					value = fmt.Sprintf("%2d", cell.Date.Day())
				}
				if _, err := fmt.Fprintf(output, "%s ", value); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(output); err != nil {
				return err
			}
		}
		return nil
	case *grid.YearGrid:
		for _, month := range typed.Months {
			if _, err := fmt.Fprintf(output, "%04d-%02d rows=%d cells=%d\n", month.Year, month.Month, month.Rows, len(month.Cells)); err != nil {
				return err
			}
		}
		return nil
	case []interface{}:
		_, err := fmt.Fprintln(output, typed)
		return err
	default:
		_, err := fmt.Fprintln(output, value)
		return err
	}
}
