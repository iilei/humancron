package humancron_test

import (
	"testing"

	"github.com/iilei/humancron"
)

func TestParseValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "invalid list value reports the whole field",
			expr: "0 22 1,x * ? *",
			want: `invalid day-of-month: "1,x"`,
		},
		{
			name: "malformed range reports the whole field",
			expr: "0 22 1,2-3-4 * ? *",
			want: `invalid day-of-month: "1,2-3-4"`,
		},
		{
			name: "invalid range start",
			expr: "0 22 x-3 * ? *",
			want: `invalid day-of-month: "x-3"`,
		},
		{
			name: "invalid range end",
			expr: "0 22 1,2-x * ? *",
			want: `invalid day-of-month: "1,2-x"`,
		},
		{
			name: "out-of-bounds range",
			expr: "0 22 1,2-32 * ? *",
			want: `invalid day-of-month: "1,2-32"`,
		},
		{
			name: "reversed range reports the whole field",
			expr: "0 22 1,4-2 * ? *",
			want: `invalid day-of-month range: "1,4-2"`,
		},
		{
			name: "step reports only the list item",
			expr: "0 22 1,2/3 * ? *",
			want: `unsupported day-of-month step expression: "2/3"`,
		},
		{
			name: "named range reports invalid endpoint",
			expr: "0 22 ? * MON-FOO *",
			want: `invalid day-of-week: "FOO"`,
		},
		{
			name: "reversed named range",
			expr: "0 22 ? * FRI-MON *",
			want: `invalid day-of-week range: "FRI-MON"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := humancron.Parse(tt.expr)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("Parse() error = %v, want %q", err, tt.want)
			}
		})
	}
}
