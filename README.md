# Pretty Logrus

Takes JSON-formatted [logrus](https://github.com/sirupsen/logrus) log messages as input and prints them back out in a more human readable format.

Build it:

```shell
go build -o plr main.go
```

Put the `plr` executable somewhere on your PATH.

Usage:

```shell
kubectl logs <pod> | plr
```

Options:

- `--multi-line`: Print output on multiple lines with log message and level first and then each field/data-entry on separate lines
- `--no-data`: Don't show logged data fields (additional key-value pairs of arbitrary data)
- `--level <level>`: Only show log messages with matching level. Values (logrus levels): `trace` | `debug` | `info` | `warning` | `error` | `fatal` | `panic`
- `--field <field>`: Only show this specific data field
- `--fields <field>,<field>`: Only show specific data fields separated by comma
- `--except <field>,<field>`: Don't show this particular field or fields separated by comma
- `--trunc <field>=<num chars or substr>`: Truncate the content of this field by an index or substring. Several usage examples:
  - `--trunc message=50`: Print the first 50 characters in the message field
  - `--trunc message="\n"`: Print everything up until the first line break in the message field
  - `--trunc message="\t"`: Print everything up until the first tab character in the message field
  - `--trunc message=mytext`: Print everything up until the first occurrence of the phrase 'mytext' in the message field.
  - `--trunc message="stop it"`: Print everything up until the first occurrence of the phrase 'stop it' in the message field.
  - `--trunc message=" "`: Print everything up until the first empty space in the message field.
- `--where <field>=<value>`: Only show log messages where the value occurs. Several usage examples:
  - `--where <field>=<value>`: Only show log messages where the specific field has the given value
  - `--where <field>=<value>,<field>=<value>`: Specify multiple conditions separated by comma
  - `--where <value>`: Only show log messages where the value occurs in any data field or the message field. Value can be a partial phrase or text.

# Changelog

> :hammer_and_wrench: - Enhancements, improvements  
> :sparkles: - New features, additions  
> :bug: - Bug fixes  
> :boom: - Breaking changes  
> :scissors: - Remove features, deletions

## v1.1.5

:calendar: 2024-04-08

:sparkles: Added `--where` flag to filter log messages based on field values. Example: `--where trace.id=1234`. See usage examples above for more variations.

## v1.1.4

:calendar: 2023-03-13

- :hammer_and_wrench: Support tailing like `kubectl logs -f | plr`

## v1.1.3

:calendar: 2022-10-24

- :hammer_and_wrench: A field named `labels` will be treated as a map of other sub-fields and each sub-field under labels will be printed as `labels.<subfield>=[<value>]`. Example:

Before
```
labels=[map[string]struct{ foo: "bar", abc: "def" }]
```

After
```
labels.foo=[bar] labels.abc=[def]
```

## v1.1.2

:calendar: 2022-10-06

- :bug: If a log line cannot be parsed (i.e. if it's not a line with JSON on the expected logrus format), it will be printed as-is instead of being ignored by the error handler.

## v1.1.1

:calendar: 2022-10-04

- :hammer_and_wrench: Truncate flag now supports substrings, newline and tab characters in addition to a character index.

## v1.1.0

:calendar: 2022-10-04

- :sparkles: Added `--trunc` flag to limit the output length of a certain field. The value of the given field name will be cut off at the given character index. Only one field can be truncated at a time. Example: `--trunc service.name=10`.
- :bug: Fixed a bug where long lines would be skipped entirely and the line would be lost without any warning or information.

## v1.0.0

No changelog at this time
