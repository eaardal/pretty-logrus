package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"strconv"
	"strings"
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
var debugFlag = flag.Bool("debug", false, "Print verbose debug information")

var includedFields map[string]struct{}
var excludedFields map[string]struct{}

type Truncate struct {
	FieldName string
	NumChars  int
	Substr    string
}

func (t *Truncate) Truncate(value string) string {
	if t.NumChars > -1 {
		return t.truncAtNumChars(value)
	}

	if t.Substr != "" {
		return t.truncAtSubstr(value)
	}

	return value
}

func (t *Truncate) truncAtNumChars(value string) string {
	if truncate.NumChars == -1 {
		return value
	}

	if len(value) > truncate.NumChars {
		return value[:truncate.NumChars]
	}
	return value
}

func (t *Truncate) truncAtSubstr(value string) string {
	if truncate.Substr == "" {
		return value
	}

	// If flag is --trunc message="\n" (truncate at newline character) then it'll be read as "\\n" at this point. To match it against newline char \n in a text, we must correct it.
	if strings.Contains(t.Substr, "\\n") {
		t.Substr = "\n"
	}

	// If flag is --trunc message="\t" (truncate at tab character) then it'll be read as "\\t" at this point. To match it against a tab char \t in a text, we must correct it.
	if strings.Contains(t.Substr, "\\t") {
		t.Substr = "\t"
	}

	if indexOfSubstr := strings.Index(value, truncate.Substr); indexOfSubstr > -1 {
		return value[:indexOfSubstr]
	}
	return value
}

var truncate *Truncate

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
var dataFieldKeywords = []string{"labels"}

type LogEntry struct {
	OriginalLogLine []byte
	Time            string
	Level           string
	Message         string
	Fields          map[string]string
	IsParsed        bool
}

func (l *LogEntry) UseParsedLogLine(logMap map[string]interface{}) {
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

		for _, dataFieldKeyword := range dataFieldKeywords {
			if strings.ToLower(key) == dataFieldKeyword {
				switch val := value.(type) {
				case string:
					l.Fields[key] = val
				case map[string]interface{}:
					for dataFieldKey, dataFieldValue := range val {
						l.Fields[key+"."+dataFieldKey] = fmt.Sprintf("%v", dataFieldValue)
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

	l.IsParsed = true
}

func (l *LogEntry) UseOriginalLogLine(line []byte) {
	copy(l.OriginalLogLine, line)
	l.IsParsed = false
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

	if truncateFlag != nil && *truncateFlag != "" {
		parts := strings.Split(*truncateFlag, "=")
		if len(parts) != 2 {
			log.Fatalf("Invalid format for --trunc flag: %s, expected [fieldname]=[number of chars to include]. Example: --trunc message=50", *truncateFlag)
		}

		truncate = &Truncate{
			FieldName: parts[0],
			NumChars:  -1,
		}

		if numChars, err := strconv.Atoi(parts[1]); err != nil {
			truncate.Substr = parts[1]
		} else {
			truncate.NumChars = numChars
		}
	}
}

// https://stackoverflow.com/a/49704981
// https://flaviocopes.com/go-shell-pipes/
func readStdin() {
	logEntryCh := make(chan *LogEntry, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx := context.Background()
	go func() {
		defer wg.Done()
		printLogEntries(ctx, logEntryCh)
	}()

	reader := bufio.NewReader(os.Stdin)

	line, readErr := reader.ReadBytes('\n')
	if readErr != nil {
		log.Fatalf("failed to read from stdin: %v", readErr)
		return
	}

	for readErr == nil {
		logEntry := &LogEntry{
			OriginalLogLine: line,
			Fields:          make(map[string]string),
		}

		parsedLogLine := make(map[string]interface{}, 0)

		if err := json.Unmarshal(line, &parsedLogLine); err != nil {
			logEntry.UseOriginalLogLine(line)
		} else {
			logEntry.UseParsedLogLine(parsedLogLine)
		}

		select {
		case <-ctx.Done():
			return
		case logEntryCh <- logEntry:
		}
		line, readErr = reader.ReadBytes('\n')
	}

	close(logEntryCh)
	// wait until last line is printed
	wg.Wait()
}

func printLogEntries(ctx context.Context, logEntries <-chan *LogEntry) {
	for {
		select {
		case <-ctx.Done():
			return
		case logEntry, ok := <-logEntries:
			if !ok {
				return
			}

			if !shouldShowLogLine(logEntry) {
				continue
			}

			if debugFlag != nil && *debugFlag {
				//fmt.Printf("[ORIGINAL LINE %d]: %s\n", i+1, string(logEntry.OriginalLogLine))
				fmt.Printf("[ORIGINAL LINE]: %s\n", string(logEntry.OriginalLogLine))
			}

			if !logEntry.IsParsed {
				println(string(logEntry.OriginalLogLine))
				continue
			}

			if multiLine != nil && *multiLine {
				printMultiLine(logEntry)
			} else {
				printSingleLine(logEntry)
			}
		}
	}
}

func printSingleLine(logEntry *LogEntry) {
	fields := ""

	addField := func(fieldName, fieldValue string) {
		value := fmtValue(fieldName, fieldValue)
		field := fmt.Sprintf("%s=[%s]", yellow(fieldName), green(value))
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
		fmt.Printf("[%s] %s - %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(fmtMessage(logEntry.Message)), fields)
	} else {
		fmt.Printf("[%s] %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(fmtMessage(logEntry.Message)))
	}
}

func printMultiLine(logEntry *LogEntry) {
	fields := ""

	addField := func(fieldName, fieldValue string) {
		value := fmtValue(fieldName, fieldValue)
		field := fmt.Sprintf("  %s: %s", yellow(fieldName), green(fmt.Sprintf("%v", value)))
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

	fmt.Printf("[%s] %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(fmtMessage(logEntry.Message)))

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

func fmtValue(key, value string) string {
	if truncate != nil && truncate.FieldName == key {
		return truncate.Truncate(value)
	}
	return value
}

func fmtMessage(message string) string {
	return fmtValue("message", message)
}

func shouldShowLogLine(logEntry *LogEntry) bool {
	hasLevelFilter := levelFilter != nil && *levelFilter != ""
	return !hasLevelFilter || (hasLevelFilter && *levelFilter == logEntry.Level)
}
