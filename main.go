package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
)

var multiLine = flag.Bool("multi-line", false, "Print output on multiple lines with log message and level first and then each field/data-entry on separate lines")
var noData = flag.Bool("no-data", false, "Don't show data fields (additional key-value pairs of arbitrary data)")
var levelFilter = flag.String("level", "", "Only show log messages with matching level. Values (logrus levels): trace|debug|info|warning|error|fatal|panic")
var fieldsFilter = flag.String("fields", "", "Only show specific data fields separated by comma")
var exceptFieldsFilter = flag.String("except", "", "Don't show this particular field or fields separated by comma")
var truncateFlag = flag.String("trunc", "", "Truncate the content of this field by x number of characters. Example: --trunc message=50")
var whereFlag = flag.String("where", "", "Filter log entries based on a condition. Example: --where trace.id=abc")
var debugFlag = flag.Bool("debug", false, "Print verbose debug information")
var highlightKey = flag.String("highlight-key", "", "Highlight the specified key in the output")
var highlightValue = flag.String("highlight-value", "", "Highlight the specified value in the output")
var minLevelFilter = flag.String("min-level", "", "Only show log messages with this level or higher")
var maxLevelFilter = flag.String("max-level", "", "Only show log messages with this level or lower")

var flagAliases = map[string]string{
	"multi-line":      "M",
	"level":           "L",
	"fields":          "F",
	"except":          "E",
	"where":           "W",
	"highlight-key":   "K",
	"highlight-value": "V",
}

func applyFlagAliases() {
	for long, short := range flagAliases {
		flagSet := flag.Lookup(long)
		logDebug("Checking alias %s -> %s: %+v", short, long, flagSet)
		flag.Var(flagSet.Value, short, fmt.Sprintf("Alias for --%s", long))
	}
}

func main() {
	applyFlagAliases()
	flag.Parse()

	args, err := parseArgs()
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		return
	}

	config := getConfig()

	ctx := context.Background()
	logEntryCh := make(chan *LogEntry, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		printLogEntries(ctx, *args, *config, logEntryCh)
	}()

	readStdin(ctx, *config, logEntryCh)

	wg.Wait()
}
