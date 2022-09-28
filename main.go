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
var levelFilter = flag.String("level", "", "Only show log messages with matching level. Values (logrus levels): trace|debug|info|warning|error|fatal|panic")
var fieldFilter = flag.String("field", "", "Only show this specific data field")
var fieldsFilter = flag.String("fields", "", "Only show specific data fields separated by comma")
var exceptFieldsFilter = flag.String("except", "", "Don't show this particular field or fields separated by comma")

var includedFields map[string]struct{}
var excludedFields map[string]struct{}

// Elastic Common Schema (ECS) field names
// https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html
const (
	ecsMessageField   = "message"
	ecsLevelField     = "log.level"
	ecsTimestampField = "@timestamp"
)

var messageKeywords = []string{logrus.FieldKeyMsg, ecsMessageField}
var levelKeywords = []string{logrus.FieldKeyLevel, ecsLevelField}
var timeKeywords = []string{logrus.FieldKeyTime, ecsTimestampField}
var errorKeywords = []string{logrus.ErrorKey}

type LogEntry struct {
	OriginalJson []byte
	Time         string
	Level        string
	Message      string
	Fields       map[string]string
}

func (l *LogEntry) FromMap(logMap map[string]interface{}) {
	for key, value := range logMap {
		match := false

		for _, levelKeyword := range levelKeywords {
			if strings.ToLower(key) == levelKeyword {
				l.Level = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, messageKeyword := range messageKeywords {
			if strings.ToLower(key) == messageKeyword {
				l.Message = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, timeKeyword := range timeKeywords {
			if strings.ToLower(key) == timeKeyword {
				l.Time = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, errorKeyword := range errorKeywords {
			if strings.ToLower(key) == errorKeyword {
				switch val := value.(type) {
				case string:
					l.Fields[key] = val
				case map[string]interface{}:
					for errKey, errValue := range val {
						l.Fields[key+"."+errKey] = errValue.(string)
					}
				}
				match = true
				break
			}
		}

		if !match {
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
		includedFields = make(map[string]struct{})
		includedFields[*fieldFilter] = struct{}{}
	} else if fieldsFilter != nil && *fieldsFilter != "" {
		includedFields = make(map[string]struct{})
		for _, f := range strings.Split(*fieldsFilter, ",") {
			includedFields[f] = struct{}{}
		}
	} else if exceptFieldsFilter != nil && *exceptFieldsFilter != "" {
		excludedFields = make(map[string]struct{})
		for _, f := range strings.Split(*exceptFieldsFilter, ",") {
			excludedFields[f] = struct{}{}
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
		//fmt.Printf("[LINE %d]: json: %s\n\n", lineCount, line)

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

func printSingleLine(logEntry *LogEntry) {
	fields := ""

	addField := func(fieldName, fieldValue string) {
		field := fmt.Sprintf("%s=[%s]", yellow(fieldName), green(fieldValue))
		if fields == "" {
			fields = field
		} else {
			fields = fmt.Sprintf("%s, %s", fields, field)
		}
	}

	if noData == nil || *noData == false {
		for fieldName, fieldValue := range logEntry.Fields {
			if len(includedFields) > 0 {
				if _, included := includedFields[fieldName]; included {
					addField(fieldName, fieldValue)
				}
			} else if len(excludedFields) > 0 {
				if _, excluded := excludedFields[fieldName]; !excluded {
					addField(fieldName, fieldValue)
				}
			} else {
				addField(fieldName, fieldValue)
			}
		}
	}

	if len(fields) > 0 {
		fmt.Printf("[%s] %s - %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(logEntry.Message), fields)
	} else {
		fmt.Printf("[%s] %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(logEntry.Message))
	}
}

func printMultiLine(logEntry *LogEntry) {
	fields := ""

	addField := func(fieldName, fieldValue string) {
		field := fmt.Sprintf("  %s: %s", yellow(fieldName), green(fmt.Sprintf("%v", fieldValue)))
		if fields == "" {
			fields = field
		} else {
			fields = fmt.Sprintf("%s\n%s", fields, field)
		}
	}

	if noData == nil || *noData == false {
		for fieldName, fieldValue := range logEntry.Fields {
			if len(includedFields) > 0 {
				if _, included := includedFields[fieldName]; included {
					addField(fieldName, fieldValue)
				}
			} else if len(excludedFields) > 0 {
				if _, excluded := excludedFields[fieldName]; !excluded {
					addField(fieldName, fieldValue)
				}
			} else {
				addField(fieldName, fieldValue)
			}
		}
	}

	fmt.Printf("[%s] %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(logEntry.Message))

	if len(fields) > 0 {
		fmt.Println(fields)
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
