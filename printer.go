package main

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"sort"
	"strings"
)

var cyan = color.CyanString
var blue = color.BlueString
var yellow = color.YellowString
var green = color.GreenString
var white = color.WhiteString
var red = color.RedString

func printLogEntries(ctx context.Context, args Args, logEntries <-chan *LogEntry) {
	for {
		select {
		case <-ctx.Done():
			return
		case logEntry, ok := <-logEntries:
			if !ok {
				return
			}

			if !shouldShowLogLine(args, logEntry) {
				if isDebug() {
					fmt.Printf("Not showing log entry %d\n", logEntry.LineNumber)
				}
				continue
			}

			if !logEntry.IsParsed {
				println(string(logEntry.OriginalLogLine))
				continue
			}

			if multiLine != nil && *multiLine {
				printMultiLine(args, logEntry)
			} else {
				printSingleLine(args, logEntry)
			}
		}
	}
}

func printSingleLine(args Args, logEntry *LogEntry) {
	var fields []string

	addField := func(fieldName, fieldValue string) {
		value := fmtValue(args.Truncate, fieldName, fieldValue)
		field := fmt.Sprintf("%s=[%s]", yellow(fieldName), green(value))
		fields = append(fields, field)
		//if len(fields) == 0 {
		//	fields = append(fields, field)
		//} else {
		//	fields = append(fields, fmt.Sprintf("%s, %s", fields, field))
		//}
	}

	if noData == nil || *noData == false {
		for fieldName, fieldValue := range logEntry.Fields {
			if len(args.IncludedFields) > 0 {
				if _, included := args.IncludedFields[fieldName]; included {
					addField(fieldName, fieldValue)
				}
			} else if len(args.ExcludedFields) > 0 {
				if _, excluded := args.ExcludedFields[fieldName]; !excluded {
					addField(fieldName, fieldValue)
				}
			} else {
				addField(fieldName, fieldValue)
			}
		}
	}

	sort.Strings(fields)
	fieldsString := strings.Join(fields, ", ")
	//log.Printf("fieldsString: %s", fieldsString)

	if len(fields) > 0 {
		fmt.Printf("[%s] %s - %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(fmtMessage(args.Truncate, logEntry.Message)), fieldsString)
	} else {
		fmt.Printf("[%s] %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(fmtMessage(args.Truncate, logEntry.Message)))
	}
}

func printMultiLine(args Args, logEntry *LogEntry) {
	var fields []string

	addField := func(fieldName, fieldValue string) {
		value := fmtValue(args.Truncate, fieldName, fieldValue)
		field := fmt.Sprintf("  %s: %s", yellow(fieldName), green(fmt.Sprintf("%v", value)))
		fields = append(fields, field)
		//if len(fields) == 0 {
		//	fields = append(fields, field)
		//} else {
		//	fields = append(fields, fmt.Sprintf("%s\n%s", fields, field))
		//}
	}

	if noData == nil || *noData == false {
		for fieldName, fieldValue := range logEntry.Fields {
			if len(args.IncludedFields) > 0 {
				if _, included := args.IncludedFields[fieldName]; included {
					addField(fieldName, fieldValue)
				}
			} else if len(args.ExcludedFields) > 0 {
				if _, excluded := args.ExcludedFields[fieldName]; !excluded {
					addField(fieldName, fieldValue)
				}
			} else {
				addField(fieldName, fieldValue)
			}
		}
	}

	sort.Strings(fields)
	fieldsString := strings.Join(fields, "\n")

	fmt.Printf("[%s] %s - %s\n", formatLevel(logEntry), blue(logEntry.Time), white(fmtMessage(args.Truncate, logEntry.Message)))

	if len(fields) > 0 {
		fmt.Println(fieldsString)
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

func fmtValue(truncate *Truncate, key, value string) string {
	if truncate != nil && truncate.FieldName == key {
		return truncate.Truncate(value)
	}
	return value
}

func fmtMessage(truncate *Truncate, message string) string {
	return fmtValue(truncate, "message", message)
}

func shouldShowLogLine(args Args, logEntry *LogEntry) bool {
	return shouldShowLogLineForLevelFilter(logEntry) && shouldShowLogLineForWhereFilter(args.WhereFields, logEntry)
}

func shouldShowLogLineForLevelFilter(logEntry *LogEntry) bool {
	if levelFilter == nil || *levelFilter == "" {
		return true
	}

	return logEntry.Level == *levelFilter
}

func shouldShowLogLineForWhereFilter(whereFields map[string]string, logEntry *LogEntry) bool {
	if whereFields == nil {
		return true
	}

	for field, value := range whereFields {
		if field == AnyField {
			// Check if the value is in any of the data fields
			for _, fieldValue := range logEntry.Fields {
				if strings.Contains(fieldValue, value) {
					return true
				}
			}

			// Check if the value is in the log message
			if strings.Contains(logEntry.Message, value) {
				return true
			}
		} else {
			// Check if the value is in the specific data field
			if fieldValue, ok := logEntry.Fields[field]; ok {
				if fieldValue == value {
					return true
				}
			}
		}
	}

	return false
}
