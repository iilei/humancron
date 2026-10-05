package humancron

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const decimalBase = 10

func Describe(cron *Cron) (string, error) {
	if cron == nil {
		return "", errors.New("cron cannot be nil")
	}

	if cron.Month != "*" {
		return "", errors.New(
			"only '*' is currently supported for month",
		)
	}

	if cron.Year != "*" {
		return "", errors.New(
			"only '*' is currently supported for year",
		)
	}

	var recurrence string
	if cron.DayOfMonth == "?" {
		days, err := describeWeekdays(cron.DayOfWeek)
		if err != nil {
			return "", err
		}
		recurrence = "Every " + days
	} else {
		if cron.DayOfWeek != "?" {
			return "", errors.New(
				"day-of-week must be '?' when day-of-month is specified",
			)
		}

		day, err := describeDayOfMonth(cron.DayOfMonth)
		if err != nil {
			return "", err
		}
		recurrence = "Every month on the " + day
	}

	minute, err := strconv.Atoi(cron.Minute)
	if err != nil {
		return "", err
	}

	hour, err := strconv.Atoi(cron.Hour)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%s at %02d:%02d",
		recurrence,
		hour,
		minute,
	), nil
}

func describeDayOfMonth(s string) (string, error) {
	day, err := strconv.Atoi(s)
	if err != nil || day < 1 || day > maxDayOfMonth {
		return "", fmt.Errorf("invalid day-of-month: %q", s)
	}

	suffixes := [...]string{"th", "st", "nd", "rd"}
	suffix := suffixes[0]
	if day%100 < 11 || day%100 > 13 {
		if digit := day % decimalBase; digit < len(suffixes) {
			suffix = suffixes[digit]
		}
	}

	return fmt.Sprintf("%d%s", day, suffix), nil
}

func describeWeekdays(s string) (string, error) {
	parts := strings.Split(s, ",")
	names := make([]string, 0, len(parts))

	for _, part := range parts {
		name, err := describeWeekday(part)
		if err != nil {
			return "", err
		}
		names = append(names, name)
	}

	return strings.Join(names, ", "), nil
}

func describeWeekday(s string) (string, error) {
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) != rangeBoundsCount {
			return "", fmt.Errorf(
				"unsupported weekday expression: %q",
				s,
			)
		}

		from, err := parseWeekday(parts[0])
		if err != nil {
			return "", err
		}
		to, err := parseWeekday(parts[1])
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(
			"%s through %s",
			weekdayNames[from],
			weekdayNames[to],
		), nil
	}

	n, err := parseWeekday(s)
	if err != nil {
		return "", err
	}

	return weekdayNames[n], nil
}

func parseWeekday(s string) (int, error) {
	if value, ok := weekdays[strings.ToUpper(s)]; ok {
		return value - 1, nil
	}

	n, err := strconv.Atoi(s)
	if err != nil || n < firstWeekday || n > lastWeekday {
		return 0, fmt.Errorf("invalid weekday: %q", s)
	}

	return n - 1, nil
}
