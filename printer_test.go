package main

import (
	"strings"
	"testing"
)

func TestPodPrefix(t *testing.T) {
	t.Run("returns empty string when colorizer is nil (feature disabled)", func(t *testing.T) {
		if got := podPrefix(nil, "svc-1"); got != "" {
			t.Errorf("podPrefix(nil, ...) = %q, want empty", got)
		}
	})

	t.Run("returns empty string when pod id is empty", func(t *testing.T) {
		if got := podPrefix(newTestPodColorizer(), ""); got != "" {
			t.Errorf("podPrefix(_, \"\") = %q, want empty", got)
		}
	})

	t.Run("returns bracketed pod id followed by a separating space", func(t *testing.T) {
		got := podPrefix(newTestPodColorizer(), "svc-1")

		if !strings.Contains(got, "[svc-1]") {
			t.Errorf("podPrefix = %q, want it to contain %q", got, "[svc-1]")
		}
		if !strings.HasSuffix(got, " ") {
			t.Errorf("podPrefix = %q, want it to end with a space", got)
		}
	})
}
