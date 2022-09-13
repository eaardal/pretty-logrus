package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"strings"
)

var cyan = color.CyanString
var blue = color.BlueString
var yellow = color.YellowString
var green = color.GreenString
var white = color.WhiteString
var red = color.RedString

var multiLine = flag.Bool("multi-line", false, "Print output on multiple lines with log message and level first and then each field/data-entry on separate lines")
var noData = flag.Bool("no-data", false, "Don't show data fields (additional key-value pairs of arbitrary data)")
var levelFilter = flag.String("level", "", "Only show log messages with matching level. Values: trace|debug|info|warning|error|fatal|panic")
var fieldFilter = flag.String("field", "", "Only show this specific data field")
var fieldsFilter = flag.String("fields", "", "Only show specific data fields separated by comma")

var fieldsByName map[string]struct{}

type LogEntry struct {
	OriginalJson []byte
	Time         string
	Level        string
	Message      string
	Fields       map[string]string
}

func (l *LogEntry) FromMap(logMap map[string]interface{}) {
	for key, value := range logMap {
		if strings.ToLower(key) == logrus.FieldKeyLevel {
			l.Level = value.(string)
		} else if strings.ToLower(key) == logrus.FieldKeyMsg || key == "message" {
			l.Message = value.(string)
		} else if strings.ToLower(key) == logrus.FieldKeyTime {
			l.Time = value.(string)
		} else {
			l.Fields[key] = fmt.Sprintf("%v", value)
		}
	}
}

func main() {
	flag.Parse()

	initFields()

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
		return
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		readStdin()
	} else {
		log.Fatalf("Expected to find content from stdin. Example usage: kubectl logs <pod> | plr")
	}
}

func initFields() {
	if fieldFilter != nil && *fieldFilter != "" {
		fieldsByName = make(map[string]struct{})
		fieldsByName[*fieldFilter] = struct{}{}
	} else if fieldsFilter != nil && *fieldsFilter != "" {
		fieldsByName = make(map[string]struct{})
		for _, f := range strings.Split(*fieldsFilter, ",") {
			fieldsByName[f] = struct{}{}
		}
	}
}

// https://stackoverflow.com/a/49704981
// https://flaviocopes.com/go-shell-pipes/
func readStdin() {
	reader := bufio.NewReader(os.Stdin)

	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		log.Fatal(err)
		return
	}

	var logEntries []LogEntry
	lineCount := 1
	for err == nil && !isPrefix {
		// fmt.Printf("[LINE %d]: json: %s\n\n", lineCount, line)

		logEntry := LogEntry{
			OriginalJson: line,
			Fields:       make(map[string]string),
		}

		logMap := make(map[string]interface{}, 0)
		if err := json.Unmarshal(line, &logMap); err != nil {
			log.Fatal(err)
			return
		}

		logEntry.FromMap(logMap)

		if levelFilter != nil && *levelFilter != "" {
			if *levelFilter == logEntry.Level {
				logEntries = append(logEntries, logEntry)
			}
		} else {
			logEntries = append(logEntries, logEntry)
		}

		line, isPrefix, err = reader.ReadLine()
		lineCount++
	}

	printLogEntries(logEntries)
}

func printLogEntries(logEntries []LogEntry) {
	for _, logEntry := range logEntries {
		if multiLine != nil && *multiLine {
			printMultiLine(&logEntry)
		} else {
			printSingleLine(&logEntry)
		}
	}
}

func printMultiLine(logEntry *LogEntry) {
	fmt.Printf("[%s] %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(logEntry.Message))

	if noData == nil || *noData == false {
		for fieldName, fieldValue := range logEntry.Fields {
			if len(fieldsByName) > 0 {
				if _, ok := fieldsByName[fieldName]; ok {
					fmt.Printf("  %s: %s\n", yellow(fieldName), green(fmt.Sprintf("%v", fieldValue)))
				}
			} else {
				fmt.Printf("  %s: %s\n", yellow(fieldName), green(fmt.Sprintf("%v", fieldValue)))
			}
		}
	}
}

func printSingleLine(logEntry *LogEntry) {
	fields := ""

	if noData == nil || *noData == false {
		for fieldName, fieldValue := range logEntry.Fields {
			if len(fieldsByName) > 0 {
				if _, ok := fieldsByName[fieldName]; ok {
					field := fmt.Sprintf("%s=[%s]", yellow(fieldName), green(fieldValue))
					if fields == "" {
						fields = field
					} else {
						fields = fmt.Sprintf("%s, %s", fields, field)
					}
				}
			} else {
				field := fmt.Sprintf("%s=[%s]", yellow(fieldName), green(fieldValue))
				if fields == "" {
					fields = field
				} else {
					fields = fmt.Sprintf("%s, %s", fields, field)
				}
			}
		}

		fmt.Printf("[%s] %s - %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(logEntry.Message), fields)
	} else {
		fmt.Printf("[%s] %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(logEntry.Message))
	}
}

func formatLevel(entry *LogEntry) string {
	level := cyan(entry.Level)
	if entry.Level == "warning" {
		level = yellow(entry.Level)
	}
	if entry.Level == "error" || entry.Level == "fatal" {
		level = red(entry.Level)
	}
	return level
}
