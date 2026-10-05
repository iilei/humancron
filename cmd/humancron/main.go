package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/iilei/humancron"
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
		return 2
	}

	if *jsonInput {
		if err := runJSON(stdin, stdout); err != nil {
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

	writeDescriptions(expressions, stdout)
	return 0
}

func runJSON(input io.Reader, output io.Writer) error {
	var expressions []string
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&expressions); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	writeDescriptions(expressions, output)
	return nil
}

func writeDescriptions(expressions []string, output io.Writer) {
	for _, expr := range expressions {
		cron, err := humancron.Parse(expr)
		if err != nil {
			fmt.Fprintf(output, "%s -> INVALID: %v\n", expr, err)
			continue
		}

		description, err := humancron.Describe(cron)
		if err != nil {
			fmt.Fprintf(output, "%s -> UNSUPPORTED: %v\n", expr, err)
			continue
		}

		fmt.Fprintf(output, "%s -> %s\n", expr, description)
	}
}
