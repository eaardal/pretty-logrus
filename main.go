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

var oneLine = flag.Bool("oneline", false, "Print output on one line (default: false)")

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
		logEntries = append(logEntries, logEntry)

		line, isPrefix, err = reader.ReadLine()
		lineCount++
	}

	printLogEntries(logEntries)
}

func printLogEntries(logEntries []LogEntry) {
	for _, logEntry := range logEntries {
		if oneLine != nil && *oneLine {
			printSingleLine(&logEntry)
		} else {
			printMultiLine(&logEntry)
		}
	}
}

func printMultiLine(logEntry *LogEntry) {
	fmt.Printf("[%s] %s - %s\n", cyan(logEntry.Level), blue(logEntry.Time), white(logEntry.Message))
	for fieldName, fieldValue := range logEntry.Fields {
		fmt.Printf("  %s: %s\n", yellow(fieldName), green(fmt.Sprintf("%v", fieldValue)))
	}
}

func printSingleLine(logEntry *LogEntry) {
	fields := ""

	for fieldName, fieldValue := range logEntry.Fields {
		field := fmt.Sprintf("%s=[%s]", yellow(fieldName), green(fieldValue))
		if fields == "" {
			fields = field
		} else {
			fields = fmt.Sprintf("%s, %s", fields, field)
		}
	}

	fmt.Printf("[%s] %s - %s - %s\n", cyan(logEntry.Level), blue(logEntry.Time), white(logEntry.Message), fields)
}
