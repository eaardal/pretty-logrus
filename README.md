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
- `--trunc <field>=<num chars>`: Truncate the content of this field by x number of characters. Example: `--trunc message=50`

# Changelog

## v1.1.0

:calendar: 2022-10-04

- :sparkles: Added `--trunc` flag to limit the output of a certain field. The value of the given field name will be cut off at the given character index. Only one field can be truncated at a time. Example: `--trunc service.name=10`.
- :bug: Fixed a bug where long lines would be skipped entirely and the line would be lost without any warning or information.

## v1.0.0

No changelog at this time
