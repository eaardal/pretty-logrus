package main

import "testing"

func TestParseLogLine_PodPrefix(t *testing.T) {
	config := *newDefaultConfig()

	t.Run("extracts pod id and parses the json body of a prefixed line", func(t *testing.T) {
		line := []byte(`[pod/my-service-abc/main] {"level":"info","msg":"hello"}`)

		entry := parseLogLine(line, 1, config)

		if entry.PodID != "my-service-abc" {
			t.Errorf("PodID = %q, want %q", entry.PodID, "my-service-abc")
		}
		if !entry.IsParsed {
			t.Errorf("IsParsed = false, want true for a valid json body")
		}
		if entry.Message != "hello" {
			t.Errorf("Message = %q, want %q", entry.Message, "hello")
		}
	})

	t.Run("leaves pod id empty for an unprefixed line", func(t *testing.T) {
		line := []byte(`{"level":"info","msg":"hello"}`)

		entry := parseLogLine(line, 1, config)

		if entry.PodID != "" {
			t.Errorf("PodID = %q, want empty", entry.PodID)
		}
		if !entry.IsParsed {
			t.Errorf("IsParsed = false, want true")
		}
	})

	t.Run("strips prefix from raw line when body is not json", func(t *testing.T) {
		line := []byte("[pod/svc-1/main] this is not json")

		entry := parseLogLine(line, 1, config)

		if entry.PodID != "svc-1" {
			t.Errorf("PodID = %q, want %q", entry.PodID, "svc-1")
		}
		if entry.IsParsed {
			t.Errorf("IsParsed = true, want false for a non-json body")
		}
		if string(entry.OriginalLogLine) != "this is not json" {
			t.Errorf("OriginalLogLine = %q, want %q", string(entry.OriginalLogLine), "this is not json")
		}
	})
}
