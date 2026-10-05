// Command humancron describes AWS/EventBridge cron expressions.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/iilei/humancron"
)

const (
	exitUsage         = 2
	cronFieldCount    = 6
	weekdayFieldIndex = 4
)

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func runCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("humancron", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonInput := flags.Bool(
		"json",
		false,
		"read a JSON array of cron expressions from stdin",
	)
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}

	if *jsonInput {
		if err := runJSON(stdin, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}

	expressions := flags.Args()
	if len(expressions) == 0 {
		fmt.Fprintln(
			stderr,
			"usage: humancron '0 22 ? * MON-FRI *'",
		)
		return 1
	}

	writeDescriptions(expressions, stdout, stderr)
	return 0
}

func runJSON(input io.Reader, output, stderr io.Writer) error {
	var expressions []string
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&expressions); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	writeDescriptions(expressions, output, stderr)
	return nil
}

func writeDescriptions(expressions []string, output, stderr io.Writer) {
	for _, expr := range expressions {
		warnNumericWeekday(expr, stderr)
		cron, err := humancron.Parse(expr)
		if err != nil {
			fmt.Fprintf(output, "%s -> INVALID: %v\n", expr, err)
			continue
		}

		description, err := humancron.Describe(&cron)
		if err != nil {
			fmt.Fprintf(output, "%s -> UNSUPPORTED: %v\n", expr, err)
			continue
		}

		fmt.Fprintf(output, "%s -> %s\n", expr, description)
	}
}

func warnNumericWeekday(expr string, stderr io.Writer) {
	fields := strings.Fields(expr)
	if len(fields) != cronFieldCount || !strings.ContainsAny(fields[weekdayFieldIndex], "0123456789") {
		return
	}

	fmt.Fprintf(stderr,
		"WARNING: %q uses numeric weekdays; AWS/EventBridge uses 1=SUN through 7=SAT, not zero-based numbering. "+
			"Use three-letter English abbreviations (SUN, MON, TUE, WED, THU, FRI, SAT) to avoid ambiguity.\n",
		expr,
	)
}
