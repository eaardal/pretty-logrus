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
- `--level=<level>`: Only show log messages with matching level. Values (logrus levels): `trace` | `debug` | `info` | `warning` | `error` | `fatal` | `panic`
- `--field=<field>`: Only show this specific data field
- `--fields=<field>,<field>`: Only show specific data fields separated by comma
- `--except=<field>,<field>`: Don't show this particular field or fields separated by comma
- `--ecs`: Expect log entry to be ECS (Elastic Common Schema) formatted (log.level, message, @timestamp). [ECS reference](https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html)
