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

func printLogEntries(ctx context.Context, args Args, config Config, logEntries <-chan *LogEntry) {
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
				printMultiLine(args, config, logEntry)
			} else {
				printSingleLine(args, config, logEntry)
			}
		}
	}
}

func printSingleLine(args Args, config Config, logEntry *LogEntry) {
	var fields []string

	addField := func(fieldName, fieldValue string) {
		value := fmtValue(args.Truncate, fieldName, fieldValue)
		styledFieldName := applyFieldNameStyle(fieldName, config.FieldStyles, args.HighlightKey)
		styledFieldValue := applyFieldValueStyle(fieldName, value, config.FieldStyles, args.HighlightValue)
		field := fmt.Sprintf("%s=[%s]", styledFieldName, styledFieldValue)
		fields = append(fields, field)
	}

	if noData == nil || *noData == false {
		for fieldName, fieldValue := range logEntry.Fields {
			if len(args.IncludedFields) > 0 {
				if isFieldInList(args.IncludedFields, fieldName) {
					addField(fieldName, fieldValue)
				}
			} else if len(args.ExcludedFields) > 0 {
				if !isFieldInList(args.ExcludedFields, fieldName) {
					addField(fieldName, fieldValue)
				}
			} else {
				addField(fieldName, fieldValue)
			}
		}
	}

	level := applyLevelStyle(logEntry.Level, config.LevelStyles)
	timestamp := applyTimestampStyle(logEntry.Time, config.TimestampStyles)
	message := applyMessageStyle(fmtMessage(args.Truncate, logEntry.Message), config.MessageStyles)

	if len(fields) > 0 {
		sortedFields := sortFieldsAlphabetically(fields)
		fieldsString := strings.Join(sortedFields, ", ")
		fmt.Printf("[%s] %s - %s - %s\n", level, timestamp, message, fieldsString)
	} else {
		fmt.Printf("[%s] %s - %s\n", level, timestamp, message)
	}
}

func printMultiLine(args Args, config Config, logEntry *LogEntry) {
	var fields []string

	addField := func(fieldName, fieldValue string) {
		value := fmtValue(args.Truncate, fieldName, fieldValue)
		styledFieldName := applyFieldNameStyle(fieldName, config.FieldStyles, args.HighlightKey)
		styledFieldValue := applyFieldValueStyle(fieldName, value, config.FieldStyles, args.HighlightValue)
		field := fmt.Sprintf("  %s: %s", styledFieldName, styledFieldValue)
		fields = append(fields, field)
	}

	if noData == nil || *noData == false {
		for fieldName, fieldValue := range logEntry.Fields {
			if len(args.IncludedFields) > 0 {
				if isFieldInList(args.IncludedFields, fieldName) {
					addField(fieldName, fieldValue)
				}
			} else if len(args.ExcludedFields) > 0 {
				if !isFieldInList(args.ExcludedFields, fieldName) {
					addField(fieldName, fieldValue)
				}
			} else {
				addField(fieldName, fieldValue)
			}
		}
	}

	level := applyLevelStyle(logEntry.Level, config.LevelStyles)
	timestamp := applyTimestampStyle(logEntry.Time, config.TimestampStyles)
	message := applyMessageStyle(fmtMessage(args.Truncate, logEntry.Message), config.MessageStyles)

	fmt.Printf("[%s] %s - %s\n", level, timestamp, message)

	if len(fields) > 0 {
		sortedFields := sortFieldsAlphabetically(fields)
		fieldsString := strings.Join(sortedFields, "\n")
		fmt.Println(fieldsString)
	}
}

func isFieldInList(list map[string]struct{}, fieldName string) bool {
	if _, found := list[fieldName]; found {
		logDebug("Field '%s' is explicitly in list\n", fieldName)
		return true
	}

	for key := range list {
		if strings.HasSuffix(key, "*") {
			cleanKey := strings.TrimSuffix(key, "*")

			if strings.HasPrefix(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of trailing wildcard '%s'\n", fieldName, key)
				return true
			}
		}

		if strings.HasPrefix(key, "*") {
			cleanKey := strings.TrimPrefix(key, "*")

			if strings.HasSuffix(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of leading wildcard '%s'\n", fieldName, key)
				return true
			}
		}

		if strings.HasPrefix(key, "*") && strings.HasSuffix(key, "*") {
			cleanKey := strings.Trim(key, "*")

			if strings.Contains(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of leading and trailing wildcard '%s'\n", fieldName, key)
				return true
			}
		}
	}

	if isDebug() {
		fmt.Printf("Field '%s' is not in list found\n", fieldName)
	}
	return false
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
	return shouldShowLogLineForLevelFilter(logEntry, args) && shouldShowLogLineForWhereFilter(args.WhereFields, logEntry)
}

func shouldShowLogLineForLevelFilter(logEntry *LogEntry, args Args) bool {
	if args.MinLogLevel == "" && args.MaxLogLevel == "" && args.LogLevel == "" {
		return true
	}

	logEntrySeverity := logLevelToSeverity[logEntry.Level]
	logLevelSeverity := logLevelToSeverity[args.LogLevel]
	minLogLevelSeverity := logLevelToSeverity[args.MinLogLevel]
	maxLogLevelSeverity := logLevelToSeverity[args.MaxLogLevel]

	if logLevelSeverity > 0 {
		return logEntrySeverity == logLevelSeverity
	}

	if minLogLevelSeverity > 0 {
		return logEntrySeverity >= minLogLevelSeverity
	}

	if maxLogLevelSeverity > 0 {
		return logEntrySeverity <= maxLogLevelSeverity
	}

	return false
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

func sortFieldsAlphabetically(fields []string) []string {
	sort.Strings(fields)
	return fields
}
