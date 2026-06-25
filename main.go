package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
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
var allFields = flag.Bool("all-fields", false, "Show all fields, including excluded ones from config file")
var noPodID = flag.Bool("no-pod-id", false, "Don't prepend the pod ID to each line when reading kubectl logs fetched with --prefix (e.g. kubectl logs -l <selector> --prefix)")
var groupByFlag = flag.String("group-by", "", "Group log lines by the value of a field (e.g. --group-by trace.id), printing each group together under a header. Accepts a comma-separated fallback list treated as one logical key (e.g. --group-by trace.id,labels.trace.id). Batch mode: reads to end of input, so not for use with kubectl logs -f")

var flagAliases = map[string]string{
	"multi-line":      "M",
	"level":           "L",
	"fields":          "F",
	"except":          "E",
	"where":           "W",
	"group-by":        "G",
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

	if handeled, err := execCommands(); err != nil {
		fmt.Printf("Error executing commands: %v\n", err)
		return
	} else if handeled {
		return
	}

	config := getConfig()

	args, err := parseArgs(config.LogLevelToSeverity)
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		return
	}

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

func execCommands() (bool, error) {
	handled := false

	if len(os.Args) < 2 {
		return handled, nil
	}

	command := os.Args[1]
	logDebug("Executing command: %s", command)

	switch command {
	case "default-config":
		config := newDefaultConfig()
		configJson, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(configJson))
		handled = true
		break
	case "init":
		if !hasHomeEnvVar() {
			fmt.Println("Please set PRETTY_LOGRUS_HOME environment variable to initialize config file")
			return true, nil
		}

		if err := ensureConfigFileExistsIfHomeEnvIsSet(newDefaultConfig()); err != nil {
			return true, err
		}

		fmt.Printf("Config file exists at: %s\n", path.Join(homeEnvDir(), configFileName))
		fmt.Printf("Remember that PRETTY_LOGRUS_HOME environment variable must be set to use this config file. Make sure it's part of your shell config\n")
		handled = true
		break
	}

	return handled, nil
}
