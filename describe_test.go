package humancron_test

import (
	"testing"

	"github.com/iilei/humancron"
)

func TestDescribeRejectsNil(t *testing.T) {
	description, err := humancron.Describe(nil)
	if err == nil || err.Error() != "cron cannot be nil" || description != "" {
		t.Fatalf("Describe(nil) = (%q, %v), want empty description and nil-input error", description, err)
	}
}

func TestDescribeDoesNotMutateCron(t *testing.T) {
	cron, err := humancron.Parse("30 9 ? * MON *")
	if err != nil {
		t.Fatal(err)
	}
	before := cron

	description, err := humancron.Describe(&cron)
	if err != nil {
		t.Fatal(err)
	}
	if description != "Every Monday at 09:30" {
		t.Fatalf("Describe() = %q, want Monday at 09:30", description)
	}
	if cron != before {
		t.Fatalf("Describe() mutated Cron: got %+v, want %+v", cron, before)
	}
}
