# Human-friendly Cron expressions

`humancron` is a small Go command-line tool that turns a limited subset of
AWS/EventBridge cron expressions into human-readable descriptions. It is not a
general-purpose cron parser: an expression can be syntactically valid but still
outside the subset that this tool can describe.

## Supported expressions

The input uses the six-field AWS/EventBridge layout:

```text
minute hour day-of-month month day-of-week year
```

The describer currently supports:

- A numeric minute from `0` to `59` and hour from `0` to `23`.
- A weekday expression with `?` for day-of-month and `*` for month and year.
  Weekdays can be names (`SUN` through `SAT`), EventBridge numbers (`1` for
  Sunday through `7` for Saturday), comma-separated weekdays, or increasing
  ranges.
- A single numeric day of the month from `1` to `31` with `?` for day-of-week,
  plus `*` for month and year.

Examples:

| Expression | Description |
| --- | --- |
| `0 22 ? * MON-FRI *` | Every Monday through Friday at 22:00 |
| `30 9 ? * MON *` | Every Monday at 09:30 |
| `0 0 ? * MON,WED,FRI *` | Every Monday, Wednesday, Friday at 00:00 |
| `59 23 ? * 2-6 *` | Every Monday through Friday at 23:59 |
| `15 14 1 * ? *` | Every month on the 1st at 14:15 |

Weekday numbers use EventBridge's `1-7` numbering, where `1` is Sunday and `7`
is Saturday. Weekday names are case-insensitive.

## Project layout

- `cron.go` and `describe.go` provide the `humancron` Go package for parsing and
  describing expressions.
- `cmd/humancron` contains the CLI entrypoint and its tests.

## Run from the command line

Pass one or more expressions as quoted arguments:

```bash
go run ./cmd/humancron \
  '0 22 ? * MON-FRI *' \
  '0 6 ? * TUE-SAT *' \
  '15 14 1 * ? *'
```

Output:

```text
0 22 ? * MON-FRI * -> Every Monday through Friday at 22:00
0 6 ? * TUE-SAT * -> Every Tuesday through Saturday at 06:00
15 14 1 * ? * -> Every month on the 1st at 14:15
```

Expressions can also be read as a JSON array of strings from standard input:

```bash
printf '%s\n' '["0 22 ? * MON-FRI *","15 14 1 * ? *"]' |
  go run ./cmd/humancron --json
```

The JSON mode prints one result per expression in the same format as the
argument mode.

## What is not supported

The parser validates more field values than the describer can currently
express. For example, lists or ranges in day-of-month, specific months or
years, and wildcard weekdays are not describable. These inputs are reported
as `UNSUPPORTED`, rather than being turned into an inaccurate description.
Invalid field values are reported as `INVALID`.

For example:

```bash
go run ./cmd/humancron \
  '0 22 1,15 * ? *' \
  '0 22 ? JAN MON *' \
  '0 22 ? * * *'
```

Output:

```text
0 22 1,15 * ? * -> UNSUPPORTED: invalid day-of-month: "1,15"
0 22 ? JAN MON * -> UNSUPPORTED: only '*' is currently supported for month
0 22 ? * * * -> UNSUPPORTED: invalid weekday: "*"
```

Step expressions such as `*/15` are rejected by the parser. In command-line
argument mode, each expression is reported independently and processing
continues, and invalid or unsupported expressions do not by themselves produce
a non-zero exit status. Malformed JSON or command-line usage/flag errors do
produce a non-zero exit status.
