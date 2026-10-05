// Package humancron parses and describes a limited subset of AWS/EventBridge cron expressions.
package humancron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	fieldCount       = 6
	rangeBoundsCount = 2
	maxMinute        = 59
	maxHour          = 23
	maxDayOfMonth    = 31
	minYear          = 1970
	maxYear          = 2199
	firstWeekday     = int(time.Sunday) + 1
	lastWeekday      = int(time.Saturday) + 1
)

var (
	weekdays = map[string]int{
		"SUN": int(time.Sunday) + 1,
		"MON": int(time.Monday) + 1,
		"TUE": int(time.Tuesday) + 1,
		"WED": int(time.Wednesday) + 1,
		"THU": int(time.Thursday) + 1,
		"FRI": int(time.Friday) + 1,
		"SAT": int(time.Saturday) + 1,
	}
	weekdayNames = [...]string{
		"Sunday",
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
	}
	months = map[string]int{
		"JAN": int(time.January),
		"FEB": int(time.February),
		"MAR": int(time.March),
		"APR": int(time.April),
		"MAY": int(time.May),
		"JUN": int(time.June),
		"JUL": int(time.July),
		"AUG": int(time.August),
		"SEP": int(time.September),
		"OCT": int(time.October),
		"NOV": int(time.November),
		"DEC": int(time.December),
	}
)

type Cron struct {
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
	Year       string
}

func Parse(expr string) (Cron, error) {
	fields := strings.Fields(expr)
	if len(fields) != fieldCount {
		return Cron{}, fmt.Errorf("expected 6 fields, got %d", len(fields))
	}

	cron := Cron{
		Minute:     fields[0],
		Hour:       fields[1],
		DayOfMonth: fields[2],
		Month:      fields[3],
		DayOfWeek:  fields[4],
		Year:       fields[5],
	}

	if err := validateMinute(cron.Minute); err != nil {
		return Cron{}, err
	}
	if err := validateHour(cron.Hour); err != nil {
		return Cron{}, err
	}
	if err := validateDayOfMonth(cron.DayOfMonth); err != nil {
		return Cron{}, err
	}
	if err := validateMonth(cron.Month); err != nil {
		return Cron{}, err
	}
	if err := validateDayOfWeek(cron.DayOfWeek); err != nil {
		return Cron{}, err
	}
	if err := validateYear(cron.Year); err != nil {
		return Cron{}, err
	}

	return cron, nil
}

func validateMinute(s string) error {
	return validateNumericField(s, 0, maxMinute, "minute")
}

func validateHour(s string) error {
	return validateNumericField(s, 0, maxHour, "hour")
}

func validateDayOfMonth(s string) error {
	if s == "?" || s == "*" {
		return nil
	}
	return validateListOrRange(s, 1, maxDayOfMonth, "day-of-month")
}

func validateMonth(s string) error {
	if s == "*" {
		return nil
	}
	return validateNamedOrNumericField(s, months, int(time.January), int(time.December), "month")
}

func validateDayOfWeek(s string) error {
	if s == "*" || s == "?" {
		return nil
	}
	return validateNamedOrNumericField(s, weekdays, firstWeekday, lastWeekday, "day-of-week")
}

func validateYear(s string) error {
	if s == "*" {
		return nil
	}
	return validateListOrRange(s, minYear, maxYear, "year")
}

func validateNumericField(s string, lower, upper int, name string) error {
	if s == "" {
		return fmt.Errorf("%s cannot be empty", name)
	}
	_, err := parseNamedOrNumericValue(s, nil, lower, upper, name)
	return err
}

func validateNamedOrNumericField(s string, names map[string]int, lower, upper int, name string) error {
	for part := range strings.SplitSeq(s, ",") {
		if err := validateNamedOrNumericPart(part, names, lower, upper, name); err != nil {
			return err
		}
	}
	return nil
}

func validateNamedOrNumericPart(s string, names map[string]int, lower, upper int, name string) error {
	if strings.Contains(s, "-") {
		return validateNamedRange(s, names, lower, upper, name)
	}
	if strings.Contains(s, "/") {
		return fmt.Errorf("unsupported %s step expression: %q", name, s)
	}
	_, err := parseNamedOrNumericValue(s, names, lower, upper, name)
	return err
}

func validateNamedRange(s string, names map[string]int, lower, upper int, name string) error {
	parts := strings.Split(s, "-")
	if len(parts) != rangeBoundsCount {
		return fmt.Errorf("invalid %s: %q", name, s)
	}
	from, err := parseNamedOrNumericValue(parts[0], names, lower, upper, name)
	if err != nil {
		return err
	}
	to, err := parseNamedOrNumericValue(parts[1], names, lower, upper, name)
	if err != nil {
		return err
	}
	if from > to {
		return fmt.Errorf("invalid %s range: %q", name, s)
	}
	return nil
}

func parseNamedOrNumericValue(s string, names map[string]int, lower, upper int, name string) (int, error) {
	if value, ok := names[strings.ToUpper(s)]; ok {
		return value, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %q", name, s)
	}
	if n < lower || n > upper {
		return 0, fmt.Errorf("invalid %s: %q: must be between %d and %d", name, s, lower, upper)
	}
	return n, nil
}

func validateListOrRange(s string, lower, upper int, name string) error {
	for part := range strings.SplitSeq(s, ",") {
		if err := validateNumericPart(part, s, lower, upper, name); err != nil {
			return err
		}
	}
	return nil
}

func validateNumericPart(part, field string, lower, upper int, name string) error {
	if strings.Contains(part, "-") {
		return validateNumericRange(part, field, lower, upper, name)
	}
	if strings.Contains(part, "/") {
		return fmt.Errorf("unsupported %s step expression: %q", name, part)
	}
	n, err := strconv.Atoi(part)
	if err != nil || n < lower || n > upper {
		return fmt.Errorf("invalid %s: %q", name, field)
	}
	return nil
}

func validateNumericRange(part, field string, lower, upper int, name string) error {
	parts := strings.Split(part, "-")
	if len(parts) != rangeBoundsCount {
		return fmt.Errorf("invalid %s: %q", name, field)
	}
	from, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid %s: %q", name, field)
	}
	to, err := strconv.Atoi(parts[1])
	if err != nil || from < lower || from > upper || to < lower || to > upper {
		return fmt.Errorf("invalid %s: %q", name, field)
	}
	if from > to {
		return fmt.Errorf("invalid %s range: %q", name, field)
	}
	return nil
}
