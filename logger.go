package main

import "fmt"

func logDebug(format string, args ...interface{}) {
	if isDebug() {
		fmt.Printf(format, args...)
	}
}
