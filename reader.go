package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

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

	lineCount := 0

	for readErr == nil {
		lineCount++

		if isDebug() {
			fmt.Printf("==== BEGIN LINE %d ====\n[RAW INPUT]: %q\n", lineCount, string(line))
		}

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
			j, _ := json.MarshalIndent(logEntry, "", "  ")
			fmt.Printf("[PARSED LOG ENTRY]: %s\n", string(j))
			fmt.Printf("==== END LINE %d ====\n", lineCount)
		}

		select {
		case <-ctx.Done():
			return
		case logEntryCh <- logEntry:
		}

		line, readErr = reader.ReadBytes('\n')
	}

	close(logEntryCh)
	wg.Wait()
}
