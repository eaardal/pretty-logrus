package main

import (
	"context"
	"flag"
	"github.com/fatih/color"
	"log"
	"os"
	"sync"
)

var cyan = color.CyanString
var blue = color.BlueString
var yellow = color.YellowString
var green = color.GreenString
var white = color.WhiteString
var red = color.RedString

var multiLine = flag.Bool("multi-line", false, "Print output on multiple lines with log message and level first and then each field/data-entry on separate lines")
var noData = flag.Bool("no-data", false, "Don't show data fields (additional key-value pairs of arbitrary data)")
var levelFilter = flag.String("level", "", "Only show log messages with matching level. Values (logrus levels): trace|debug|info|warning|error|fatal|panic")
var fieldFilter = flag.String("field", "", "Only show this specific data field")
var fieldsFilter = flag.String("fields", "", "Only show specific data fields separated by comma")
var exceptFieldsFilter = flag.String("except", "", "Don't show this particular field or fields separated by comma")
var truncateFlag = flag.String("trunc", "", "Truncate the content of this field by x number of characters. Example: --trunc message=50")
var whereFlag = flag.String("where", "", "Filter log entries based on a condition. Example: --where trace.id=abc")
var debugFlag = flag.Bool("debug", false, "Print verbose debug information")

func main() {
	flag.Parse()

	args := parseArgs()

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
		return
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		readStdin(*args)
	} else {
		log.Fatalf("Expected to find content from stdin. Example usage: kubectl logs <pod> | plr")
	}
}

func run(args Args) {
	ctx := context.Background()
	logEntryCh := make(chan *LogEntry, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		printLogEntries(ctx, args, logEntryCh)
	}()

	readStdin()

	close(logEntryCh)
	wg.Wait()
}

func isDebug() bool {
	return debugFlag != nil && *debugFlag
}
