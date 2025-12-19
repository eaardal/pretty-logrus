package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Args struct {
	IncludedFields map[string]struct{}
	ExcludedFields map[string]struct{}
	Truncate       *Truncate
	WhereFields    map[string]string
	HighlightKey   string
	HighlightValue string
	LogLevel       string
	MinLogLevel    string
	MaxLogLevel    string
	AllFields      bool
}

func parseArgs(logLevelToSeverity map[string]int) (*Args, error) {
	args := &Args{}

	args.IncludedFields = parseFieldsArg()
	args.ExcludedFields = parseExceptArg()
	args.Truncate = parseTruncArg()
	args.WhereFields = parseWhereArg()
	args.HighlightKey = parseHighlightKey()
	args.HighlightValue = parseHighlightValue()
	args.AllFields = parseAllFieldsArg()

	level, err := parseLogLevel(logLevelToSeverity)
	if err != nil {
		return nil, err
	}
	args.LogLevel = level

	minLevel, err := parseMinLogLevel(logLevelToSeverity)
	if err != nil {
		return nil, err
	}
	args.MinLogLevel = minLevel

	maxLevel, err := parseMaxLogLevel(logLevelToSeverity)
	if err != nil {
		return nil, err
	}
	args.MaxLogLevel = maxLevel

	if isDebug() {
		fmt.Printf("CLI Arguments:\n")
		fmt.Printf("  Raw args/flags: %+v\n", os.Args)
		fmt.Printf("  Parsed args/flags:\n")
		fmt.Printf("    Included fields: %+v\n", args.IncludedFields)
		fmt.Printf("    Excluded fields: %+v\n", args.ExcludedFields)
		fmt.Printf("    Truncate: %+v\n", args.Truncate)
		fmt.Printf("    Where: %+v\n", args.WhereFields)
		fmt.Printf("    Highlight key: %s\n", args.HighlightKey)
		fmt.Printf("    Highlight value: %s\n", args.HighlightValue)
		fmt.Printf("    LogLevel: %s\n", args.LogLevel)
		fmt.Printf("    MinLogLevel: %s\n", args.MinLogLevel)
		fmt.Printf("    MaxLogLevel: %s\n", args.MaxLogLevel)
		fmt.Printf("    AllFields: %t\n", args.AllFields)
	}

	return args, nil
}

func parseAllFieldsArg() bool {
	return allFields != nil && *allFields
}

func parseLogLevel(logLevelToSeverity map[string]int) (string, error) {
	if levelFilter != nil && *levelFilter != "" {
		severity := logLevelToSeverity[*levelFilter]
		if severity <= 0 {
			return "", fmt.Errorf("invalid log level %q, must be one of trace|debug|info|warning|error|fatal|panic", *levelFilter)
		}
		return *levelFilter, nil
	}
	return "", nil
}

func parseMinLogLevel(logLevelToSeverity map[string]int) (string, error) {
	if minLevelFilter != nil && *minLevelFilter != "" {
		severity := logLevelToSeverity[*minLevelFilter]
		if severity <= 0 {
			return "", fmt.Errorf("invalid minimum log level %q, must be one of trace|debug|info|warning|error|fatal|panic", *minLevelFilter)
		}
		return *minLevelFilter, nil
	}
	return "", nil
}

func parseMaxLogLevel(logLevelToSeverity map[string]int) (string, error) {
	if maxLevelFilter != nil && *maxLevelFilter != "" {
		severity := logLevelToSeverity[*maxLevelFilter]
		if severity <= 0 {
			return "", fmt.Errorf("invalid maximum log level %q, must be one of trace|debug|info|warning|error|fatal|panic", *maxLevelFilter)
		}
		return *maxLevelFilter, nil
	}
	return "", nil
}

func parseHighlightKey() string {
	if highlightKey != nil {
		return *highlightKey
	}
	return ""
}

func parseHighlightValue() string {
	if highlightValue != nil {
		return *highlightValue
	}
	return ""
}

func parseFieldsArg() map[string]struct{} {
	if fieldsFilter != nil && *fieldsFilter != "" {
		includedFields := make(map[string]struct{})
		for _, f := range strings.Split(*fieldsFilter, ",") {
			includedFields[f] = struct{}{}
		}
		return includedFields
	}
	return nil
}

func parseExceptArg() map[string]struct{} {
	if exceptFieldsFilter != nil && *exceptFieldsFilter != "" {
		excludedFields := make(map[string]struct{})
		for _, f := range strings.Split(*exceptFieldsFilter, ",") {
			excludedFields[f] = struct{}{}
		}
		return excludedFields
	}
	return nil
}

func parseTruncArg() *Truncate {
	if truncateFlag != nil && *truncateFlag != "" {
		parts := strings.Split(*truncateFlag, "=")
		if len(parts) != 2 {
			log.Fatalf("Invalid format for --trunc flag: %s, expected [fieldname]=[number of chars to include]. Example: --trunc message=50", *truncateFlag)
		}

		truncate := &Truncate{
			FieldName: parts[0],
			NumChars:  -1,
		}

		if numChars, err := strconv.Atoi(parts[1]); err != nil {
			truncate.Substr = parts[1]
		} else {
			truncate.NumChars = numChars
		}

		return truncate
	}

	return nil
}

const AnyField = "*"

func parseWhereArg() map[string]string {
	if whereFlag != nil && *whereFlag != "" {
		whereFields := make(map[string]string)

		// Check if multiple where clauses are specified like `--where trace.id=abc,something=else` etc
		if strings.Contains(*whereFlag, ",") {
			whereClauses := strings.Split(*whereFlag, ",")

			for _, whereClause := range whereClauses {
				key, value := parseWhereClause(whereClause)
				whereFields[key] = value
			}
		} else {
			key, value := parseWhereClause(*whereFlag)
			whereFields[key] = value
		}

		return whereFields
	}

	return nil
}

func parseWhereClause(whereClause string) (string, string) {
	// If we can't find a key=value pair, then assume the value is the entire where clause. We'll look for this value in any message or data field.
	if !strings.Contains(whereClause, "=") {
		return AnyField, whereClause
	}

	parts := strings.Split(whereClause, "=")
	if len(parts) != 2 {
		log.Fatalf("Invalid format for --where flag: %s, expected either a) --where [fieldname]=[value], b) --where trace.id=abc,something=else,more=stuff or c) --where [value]", *whereFlag)
	}
	return parts[0], parts[1]
}

func isDebug() bool {
	return debugFlag != nil && *debugFlag
}
