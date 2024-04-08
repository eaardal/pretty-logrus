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
}

func parseArgs() *Args {
	args := &Args{}

	args.IncludedFields = parseFieldArg()
	args.IncludedFields = parseFieldsArg()
	args.ExcludedFields = parseExceptArg()
	args.Truncate = parseTruncArg()
	args.WhereFields = parseWhereArg()

	if isDebug() {
		fmt.Printf("Raw args/flags: %+v\n", os.Args)
		fmt.Println("Parsed args/flags:")
		fmt.Printf("Included fields: %+v\n", args.IncludedFields)
		fmt.Printf("Excluded fields: %+v\n", args.ExcludedFields)
		fmt.Printf("Truncate: %+v\n", args.Truncate)
		fmt.Printf("Where: %+v\n", args.WhereFields)
	}

	return args
}

func parseFieldArg() map[string]struct{} {
	if fieldFilter != nil && *fieldFilter != "" {
		includedFields := make(map[string]struct{})
		includedFields[*fieldFilter] = struct{}{}
		return includedFields
	}
	return nil
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
				parts := strings.Split(whereClause, "=")
				if len(parts) != 2 {
					log.Fatalf("Invalid format for --where flag: %s, expected [fieldname]=[value]. Example: --where trace.id=abc. Specify multiple: --where trace.id=abc,something=else,more=stuff", *whereFlag)
				}
				whereFields[parts[0]] = parts[1]
			}
		} else {
			parts := strings.Split(*whereFlag, "=")
			if len(parts) != 2 {
				log.Fatalf("Invalid format for --where flag: %s, expected [fieldname]=[value]. Example: --where trace.id=abc. Specify multiple: --where trace.id=abc,something=else,more=stuff", *whereFlag)
			}
			whereFields[parts[0]] = parts[1]
		}

		return whereFields
	}

	return nil
}
