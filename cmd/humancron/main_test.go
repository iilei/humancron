package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iilei/humancron"
)

func TestParseAndDescribe(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "weekday range",
			expr: "0 22 ? * MON-FRI *",
			want: "Every Monday through Friday at 22:00",
		},
		{
			name: "weekday range",
			expr: "0 6 ? * TUE-SAT *",
			want: "Every Tuesday through Saturday at 06:00",
		},
		{
			name: "single weekday",
			expr: "30 9 ? * MON *",
			want: "Every Monday at 09:30",
		},
		{
			name: "sunday",
			expr: "15 18 ? * SUN *",
			want: "Every Sunday at 18:15",
		},
		{
			name: "saturday",
			expr: "45 12 ? * SAT *",
			want: "Every Saturday at 12:45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cron, err := humancron.Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			got, err := humancron.Describe(cron)
			if err != nil {
				t.Fatalf("Describe() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("Describe() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseRejectsInvalidExpressions(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{
			name: "wrong field count",
			expr: "0 22 MON-FRI *",
		},
		{
			name: "invalid minute",
			expr: "60 22 ? * MON-FRI *",
		},
		{
			name: "invalid hour",
			expr: "0 24 ? * MON-FRI *",
		},
		{
			name: "invalid weekday",
			expr: "0 22 ? * FOO *",
		},
		{
			name: "invalid weekday range",
			expr: "0 22 ? * MON-FOO *",
		},
		{
			name: "reversed weekday range",
			expr: "0 22 ? * FRI-MON *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := humancron.Parse(tt.expr); err == nil {
				t.Errorf("Parse() accepted invalid expression %q", tt.expr)
			}
		})
	}
}

func TestParseRejectsUnsupportedExpressions(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{
			name: "minute step",
			expr: "*/15 22 ? * MON *",
		},
		{
			name: "hour step",
			expr: "0 8-18/2 ? * MON *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := humancron.Parse(tt.expr); err == nil {
				t.Errorf(
					"Parse() accepted unsupported expression %q",
					tt.expr,
				)
			}
		})
	}
}

func TestParseAcceptsListsRangesAndNumericValues(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{
			name: "minute and hour boundaries",
			expr: "0 23 ? * SUN *",
		},
		{
			name: "day-of-month list and range",
			expr: "0 0 1,15-31 * MON *",
		},
		{
			name: "named and numeric months",
			expr: "0 0 ? JAN,3-12 MON *",
		},
		{
			name: "weekday list and range",
			expr: "0 0 ? * MON,3-5 *",
		},
		{
			name: "year list and range boundaries",
			expr: "0 0 ? * MON 1970,2026-2199",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := humancron.Parse(tt.expr); err != nil {
				t.Errorf("Parse() error = %v", err)
			}
		})
	}
}

func TestParseRejectsInvalidFieldValues(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{name: "empty minute", expr: " 22 ? * MON *"},
		{name: "non-numeric minute", expr: "x 22 ? * MON *"},
		{name: "negative minute", expr: "-1 22 ? * MON *"},
		{name: "non-numeric hour", expr: "0 x ? * MON *"},
		{name: "day-of-month outside range", expr: "0 22 32 * MON *"},
		{name: "day-of-month malformed range", expr: "0 22 1-2-3 * MON *"},
		{name: "day-of-month non-numeric range", expr: "0 22 1-x * MON *"},
		{name: "day-of-month invalid range end", expr: "0 22 1-x * ? *"},
		{name: "day-of-month range exceeds maximum", expr: "0 22 1-32 * ? *"},
		{name: "empty list item", expr: "0 22 1, * MON *"},
		{name: "month outside range", expr: "0 22 ? 13 MON *"},
		{name: "invalid month name", expr: "0 22 ? FOO MON *"},
		{name: "weekday zero", expr: "0 22 ? * 0 *"},
		{name: "weekday outside range", expr: "0 22 ? * 8 *"},
		{name: "weekday malformed range", expr: "0 22 ? * MON-TUE-WED *"},
		{name: "weekday step", expr: "0 22 ? * MON/2 *"},
		{name: "year outside range", expr: "0 22 ? * MON 1969"},
		{name: "year reversed range", expr: "0 22 ? * MON 2027-2026"},
		{name: "year step", expr: "0 22 ? * MON 2026/2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := humancron.Parse(tt.expr); err == nil {
				t.Errorf("Parse() accepted invalid expression %q", tt.expr)
			}
		})
	}
}

func TestDescribeWeekdayListsAndNumericRanges(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "weekday list",
			expr: "0 0 ? * MON,WED,FRI *",
			want: "Every Monday, Wednesday, Friday at 00:00",
		},
		{
			name: "numeric weekday range",
			expr: "59 23 ? * 2-6 *",
			want: "Every Monday through Friday at 23:59",
		},
		{
			name: "numeric Sunday",
			expr: "0 0 ? * 1 *",
			want: "Every Sunday at 00:00",
		},
		{
			name: "numeric Saturday",
			expr: "0 0 ? * 7 *",
			want: "Every Saturday at 00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cron, err := humancron.Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			got, err := humancron.Describe(cron)
			if err != nil {
				t.Fatalf("Describe() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Describe() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDescribeMonthlyDayOfMonth(t *testing.T) {
	tests := []struct {
		day  string
		want string
	}{
		{day: "1", want: "1st"},
		{day: "2", want: "2nd"},
		{day: "3", want: "3rd"},
		{day: "4", want: "4th"},
		{day: "11", want: "11th"},
		{day: "12", want: "12th"},
		{day: "13", want: "13th"},
		{day: "21", want: "21st"},
		{day: "22", want: "22nd"},
		{day: "23", want: "23rd"},
		{day: "31", want: "31st"},
	}

	for _, tt := range tests {
		t.Run(tt.day, func(t *testing.T) {
			cron, err := humancron.Parse("15 14 " + tt.day + " * ? *")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			got, err := humancron.Describe(cron)
			if err != nil {
				t.Fatalf("Describe() error = %v", err)
			}
			want := "Every month on the " + tt.want + " at 14:15"
			if got != want {
				t.Errorf("Describe() = %q, want %q", got, want)
			}
		})
	}
}

func TestDescribeRejectsInvalidWeekdayAndTime(t *testing.T) {
	tests := []struct {
		name string
		cron humancron.Cron
	}{
		{
			name: "malformed weekday range",
			cron: humancron.Cron{
				Minute:     "0",
				Hour:       "0",
				DayOfMonth: "?",
				Month:      "*",
				DayOfWeek:  "MON-TUE-WED",
				Year:       "*",
			},
		},
		{
			name: "unknown weekday",
			cron: humancron.Cron{Minute: "0", Hour: "0", DayOfMonth: "?", Month: "*", DayOfWeek: "FUNDAY", Year: "*"},
		},
		{
			name: "weekday outside supported numeric range",
			cron: humancron.Cron{Minute: "0", Hour: "0", DayOfMonth: "?", Month: "*", DayOfWeek: "8", Year: "*"},
		},
		{
			name: "invalid minute",
			cron: humancron.Cron{Minute: "x", Hour: "0", DayOfMonth: "?", Month: "*", DayOfWeek: "MON", Year: "*"},
		},
		{
			name: "invalid hour",
			cron: humancron.Cron{Minute: "0", Hour: "x", DayOfMonth: "?", Month: "*", DayOfWeek: "MON", Year: "*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := humancron.Describe(tt.cron); err == nil {
				t.Errorf("Describe() accepted invalid cron %+v", tt.cron)
			}
		})
	}
}

func TestDescribeRejectsUnsupportedExpressions(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{
			name: "day-of-month list",
			expr: "0 22 1,15 * ? *",
		},
		{
			name: "day-of-month with weekday",
			expr: "0 22 15 * MON *",
		},
		{
			name: "specific month",
			expr: "0 22 ? JAN MON *",
		},
		{
			name: "specific year",
			expr: "0 22 ? * MON 2026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cron, err := humancron.Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if _, err := humancron.Describe(cron); err == nil {
				t.Errorf(
					"Describe() accepted unsupported expression %q",
					tt.expr,
				)
			}
		})
	}
}

func TestRunCLIDirectExpressions(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{
		"0 22 ? * MON-FRI *",
		"60 22 ? * MON *",
		"15 14 1 * ? *",
		"0 22 1,15 * ? *",
	}, strings.NewReader(""), &stdout, &stderr)

	want := "0 22 ? * MON-FRI * -> Every Monday through Friday at 22:00\n" +
		"60 22 ? * MON * -> INVALID: invalid minute: \"60\": must be between 0 and 59\n" +
		"15 14 1 * ? * -> Every month on the 1st at 14:15\n" +
		"0 22 1,15 * ? * -> UNSUPPORTED: invalid day-of-month: \"1,15\"\n"
	if code != 0 || stdout.String() != want || stderr.Len() != 0 {
		t.Errorf("runCLI() = (%d, %q, %q), want (0, %q, empty stderr)", code, stdout.String(), stderr.String(), want)
	}
}

func TestRunCLIJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	input := `["0 22 ? * MON-FRI *","60 22 ? * MON *","15 14 1 * ? *"]`
	code := runCLI([]string{"--json"}, strings.NewReader(input), &stdout, &stderr)
	want := "0 22 ? * MON-FRI * -> Every Monday through Friday at 22:00\n" +
		"60 22 ? * MON * -> INVALID: invalid minute: \"60\": must be between 0 and 59\n" +
		"15 14 1 * ? * -> Every month on the 1st at 14:15\n"
	if code != 0 || stdout.String() != want || stderr.Len() != 0 {
		t.Errorf(
			"runCLI(--json) = (%d, %q, %q), want (0, %q, empty stderr)",
			code,
			stdout.String(),
			stderr.String(),
			want,
		)
	}
}

func TestRunCLIRejectsInvalidJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"--json"}, strings.NewReader("not json"), &stdout, &stderr)
	if code != 1 || !strings.HasPrefix(stderr.String(), "invalid JSON:") || stdout.Len() != 0 {
		t.Errorf(
			"runCLI(--json) = (%d, %q, %q), want status 1 and invalid JSON error",
			code,
			stdout.String(),
			stderr.String(),
		)
	}
}

func TestRunCLIWithoutExpressions(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCLI(nil, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "usage:") {
		t.Errorf("runCLI() = (%d, %q, %q), want status 1 and usage error", code, stdout.String(), stderr.String())
	}
}

func TestRunCLIRejectsUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"--unknown"}, strings.NewReader(""), &stdout, &stderr)
	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "unknown") {
		t.Errorf("runCLI() = (%d, %q, %q), want status 2 and flag error", code, stdout.String(), stderr.String())
	}
}
