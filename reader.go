package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// https://stackoverflow.com/a/49704981
// https://flaviocopes.com/go-shell-pipes/
func readStdin(ctx context.Context, logEntryCh chan<- *LogEntry) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Check that stdin is not a terminal, implying that we are reading from a pipe
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		readAndParseStdin(ctx, logEntryCh)
	} else {
		log.Fatalf("Expected to find content from stdin piped from another command. Example usage: kubectl logs <pod> | plr")
	}
}

func readAndParseStdin(ctx context.Context, logEntryCh chan<- *LogEntry) {
	reader := bufio.NewReader(os.Stdin)

	line, readErr := reader.ReadBytes('\n')
	if readErr != nil {
		log.Fatalf("failed to read from stdin: %v", readErr)
		return
	}

	lineCount := 0

	for readErr == nil {
		lineCount++

		logEntry := parseLogLine(line, lineCount)
		sendToPrinter(ctx, logEntryCh, logEntry)

		line, readErr = reader.ReadBytes('\n')
	}

	close(logEntryCh)
}

func parseLogLine(line []byte, lineCount int) *LogEntry {
	logEntry := &LogEntry{
		LineNumber:      lineCount,
		OriginalLogLine: line,
		Fields:          make(map[string]string),
	}

	parsedLogLine := make(map[string]interface{}, 0)

	if err := json.Unmarshal(line, &parsedLogLine); err != nil {
		logEntry.setOriginalLogLine(line)
	} else {
		logEntry.setFromJsonMap(parsedLogLine)
	}

	if isDebug() {
		fmt.Printf("==== BEGIN DEBUG LINE %d ====\n", lineCount)
		fmt.Printf("[RAW INPUT]: %q\n", string(line))
		j, _ := json.MarshalIndent(logEntry, "", "  ")
		fmt.Printf("[PARSED LOG ENTRY]: %s\n", string(j))
		fmt.Printf("==== END DEBUG LINE %d ====\n", lineCount)
	}

	return logEntry
}

func sendToPrinter(ctx context.Context, logEntryCh chan<- *LogEntry, logEntry *LogEntry) {
	select {
	case <-ctx.Done():
		return
	case logEntryCh <- logEntry:
	}
}
