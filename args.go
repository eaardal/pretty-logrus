package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func parseArgs() {
	if fieldFilter != nil && *fieldFilter != "" {
		// Set up --field
		includedFields = make(map[string]struct{})
		includedFields[*fieldFilter] = struct{}{}
	} else if fieldsFilter != nil && *fieldsFilter != "" {
		// Set up --fields
		includedFields = make(map[string]struct{})
		for _, f := range strings.Split(*fieldsFilter, ",") {
			includedFields[f] = struct{}{}
		}
	} else if exceptFieldsFilter != nil && *exceptFieldsFilter != "" {
		// Set up --except
		excludedFields = make(map[string]struct{})
		for _, f := range strings.Split(*exceptFieldsFilter, ",") {
			excludedFields[f] = struct{}{}
		}
	}

	// Set up --trunc
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

	// Set up --where
	if whereFlag != nil && *whereFlag != "" {
		whereFields = make(map[string]string)

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
	}

	if isDebug() {
		fmt.Println("Initialized fields from flags:")
		fmt.Printf("Included fields: %+v\n", includedFields)
		fmt.Printf("Excluded fields: %+v\n", excludedFields)
		fmt.Printf("Truncate: %+v\n", truncate)
		fmt.Printf("Where: %+v\n", whereFields)
	}
}
