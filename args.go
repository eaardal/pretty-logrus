package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const AnyField = "*"

type Args struct {
	IncludedFields          map[string]struct{}
	ExcludedFields          map[string]struct{}
	Truncate                *Truncate
	WhereFields             map[string]string
	HighlightKey            string
	HighlightValue          string
	AddToMsgIgnoreList      string
	RemoveFromMsgIgnoreList string
	ClearMsgIgnoreList      bool
	ShowMsgIgnoreList       bool
}

func parseArgs() *Args {
	args := &Args{}

	args.IncludedFields = parseFieldsArg()
	args.ExcludedFields = parseExceptArg()
	args.Truncate = parseTruncArg()
	args.WhereFields = parseWhereArg()
	args.HighlightKey = parseHighlightKey()
	args.HighlightValue = parseHighlightValue()
	args.AddToMsgIgnoreList = parseAddToMsgIgnoreList()
	args.RemoveFromMsgIgnoreList = parseRemoveFromMsgIgnoreList()
	args.ClearMsgIgnoreList = parseClearMsgIgnoreList()
	args.ShowMsgIgnoreList = parseShowMsgIgnoreList()

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
	}

	return args
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

func parseAddToMsgIgnoreList() string {
	if addToMsgIgnoreList != nil {
		return *addToMsgIgnoreList
	}
	return ""
}

func parseRemoveFromMsgIgnoreList() string {
	if removeFromMsgIgnoreList != nil {
		return *removeFromMsgIgnoreList
	}
	return ""
}

func parseClearMsgIgnoreList() bool {
	if clearMsgIgnoreList != nil {
		return *clearMsgIgnoreList
	}
	return false
}

func parseShowMsgIgnoreList() bool {
	if showMsgIgnoreList != nil {
		return *showMsgIgnoreList
	}
	return false
}

func isDebug() bool {
	return debugFlag != nil && *debugFlag
}
