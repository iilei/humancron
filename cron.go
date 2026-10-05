package humancron

import (
	"fmt"
	"strconv"
	"strings"
)

type Cron struct {
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
	Year       string
}

var weekdays = map[string]int{
	"SUN": 1,
	"MON": 2,
	"TUE": 3,
	"WED": 4,
	"THU": 5,
	"FRI": 6,
	"SAT": 7,
}

var weekdayNames = [...]string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

var months = map[string]int{
	"JAN": 1,
	"FEB": 2,
	"MAR": 3,
	"APR": 4,
	"MAY": 5,
	"JUN": 6,
	"JUL": 7,
	"AUG": 8,
	"SEP": 9,
	"OCT": 10,
	"NOV": 11,
	"DEC": 12,
}

func Parse(expr string) (Cron, error) {
	fields := strings.Fields(expr)

	if len(fields) != 6 {
		return Cron{}, fmt.Errorf(
			"expected 6 fields, got %d",
			len(fields),
		)
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
	return validateNumericField(s, 0, 59, "minute")
}

func validateHour(s string) error {
	return validateNumericField(s, 0, 23, "hour")
}

func validateDayOfMonth(s string) error {
	if s == "?" || s == "*" {
		return nil
	}

	return validateListOrRange(s, 1, 31, "day-of-month")
}

func validateMonth(s string) error {
	if s == "*" {
		return nil
	}

	return validateNamedOrNumericField(s, months, 1, 12, "month")
}

func validateDayOfWeek(s string) error {
	if s == "*" || s == "?" {
		return nil
	}

	return validateNamedOrNumericField(s, weekdays, 1, 7, "day-of-week")
}

func validateYear(s string) error {
	if s == "*" {
		return nil
	}

	return validateListOrRange(s, 1970, 2199, "year")
}

func validateNumericField(
	s string,
	min int,
	max int,
	name string,
) error {
	if s == "" {
		return fmt.Errorf("%s cannot be empty", name)
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf(
			"invalid %s: %q",
			name,
			s,
		)
	}

	if n < min || n > max {
		return fmt.Errorf(
			"invalid %s: %q: must be between %d and %d",
			name,
			s,
			min,
			max,
		)
	}

	return nil
}

func validateNamedOrNumericField(
	s string,
	names map[string]int,
	min int,
	max int,
	name string,
) error {
	for _, part := range strings.Split(s, ",") {
		if err := validateNamedOrNumericPart(
			part,
			names,
			min,
			max,
			name,
		); err != nil {
			return err
		}
	}

	return nil
}

func validateNamedOrNumericPart(
	s string,
	names map[string]int,
	min int,
	max int,
	name string,
) error {
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")

		if len(parts) != 2 {
			return fmt.Errorf(
				"invalid %s: %q",
				name,
				s,
			)
		}

		from, err := parseNamedOrNumericValue(
			parts[0],
			names,
			min,
			max,
			name,
		)
		if err != nil {
			return err
		}

		to, err := parseNamedOrNumericValue(
			parts[1],
			names,
			min,
			max,
			name,
		)
		if err != nil {
			return err
		}

		if from > to {
			return fmt.Errorf(
				"invalid %s range: %q",
				name,
				s,
			)
		}

		return nil
	}

	if strings.Contains(s, "/") {
		return fmt.Errorf(
			"unsupported %s step expression: %q",
			name,
			s,
		)
	}

	_, err := parseNamedOrNumericValue(
		s,
		names,
		min,
		max,
		name,
	)

	return err
}

func parseNamedOrNumericValue(
	s string,
	names map[string]int,
	min int,
	max int,
	name string,
) (int, error) {
	if value, ok := names[strings.ToUpper(s)]; ok {
		return value, nil
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %q",
			name,
			s,
		)
	}

	if n < min || n > max {
		return 0, fmt.Errorf(
			"invalid %s: %q: must be between %d and %d",
			name,
			s,
			min,
			max,
		)
	}

	return n, nil
}

func validateListOrRange(
	s string,
	min int,
	max int,
	name string,
) error {
	for _, part := range strings.Split(s, ",") {
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")

			if len(rangeParts) != 2 {
				return fmt.Errorf(
					"invalid %s: %q",
					name,
					s,
				)
			}

			from, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return fmt.Errorf(
					"invalid %s: %q",
					name,
					s,
				)
			}

			to, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return fmt.Errorf(
					"invalid %s: %q",
					name,
					s,
				)
			}

			if from < min || from > max || to < min || to > max {
				return fmt.Errorf(
					"invalid %s: %q",
					name,
					s,
				)
			}

			if from > to {
				return fmt.Errorf(
					"invalid %s range: %q",
					name,
					s,
				)
			}

			continue
		}

		if strings.Contains(part, "/") {
			return fmt.Errorf(
				"unsupported %s step expression: %q",
				name,
				part,
			)
		}

		n, err := strconv.Atoi(part)
		if err != nil {
			return fmt.Errorf(
				"invalid %s: %q",
				name,
				s,
			)
		}

		if n < min || n > max {
			return fmt.Errorf(
				"invalid %s: %q",
				name,
				s,
			)
		}
	}

	return nil
}
