# Pretty Logrus

Takes JSON-formatted logrus log messages as input and prints them back out in a more human readable format.

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
- To print each log message on only one line, use the `-oneline` flag.
