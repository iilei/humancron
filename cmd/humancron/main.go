// Command humancron describes AWS/EventBridge cron expressions.
package main

import (
	"encoding/json"
	"errors"
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

type jsonResult struct {
	Description string `json:"description"`
	Expression  string `json:"expression"`
	Status      string `json:"status"`
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func runCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("humancron", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonMode := flags.Bool(
		"json",
		false,
		"emit a JSON result for one expression argument, or read an expression object from stdin",
	)
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}

	if *jsonMode {
		if len(flags.Args()) > 1 {
			fmt.Fprintln(stderr, "--json accepts at most one expression argument")
			return exitUsage
		}
		if err := runJSON(flags.Args(), stdin, stdout, stderr); err != nil {
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

func runJSON(args []string, input io.Reader, output, stderr io.Writer) error {
	var expr string
	if len(args) == 1 {
		expr = args[0]
	} else {
		query, err := readJSONQuery(input)
		if err != nil {
			return err
		}
		expr = query["expression"]
	}

	warnNumericWeekday(expr, stderr)
	cron, err := humancron.Parse(expr)
	if err != nil {
		return fmt.Errorf("INVALID: %w", err)
	}
	description, err := humancron.Describe(&cron)
	if err != nil {
		return fmt.Errorf("UNSUPPORTED: %w", err)
	}
	if err := json.NewEncoder(output).Encode(jsonResult{
		Description: description,
		Expression:  expr,
		Status:      "ok",
	}); err != nil {
		return fmt.Errorf("writing JSON result: %w", err)
	}
	return nil
}

func readJSONQuery(input io.Reader) (map[string]string, error) {
	var query map[string]string
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&query); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return nil, errors.New("invalid JSON: expected exactly one query object")
	}
	if strings.TrimSpace(query["expression"]) == "" {
		return nil, errors.New("invalid JSON: expression must be a non-empty string")
	}
	return query, nil
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
