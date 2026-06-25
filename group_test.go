package main

import (
	"testing"

	"github.com/fatih/color"
)

func groupTestEntry(line int, podID, ts string, fields map[string]string) *LogEntry {
	f := make(map[string]string, len(fields))
	for k, v := range fields {
		f[k] = v
	}
	return &LogEntry{LineNumber: line, PodID: podID, Time: ts, Fields: f, IsParsed: true}
}

func TestGroupEntries(t *testing.T) {
	groupBy := []string{"trace.id", "labels.trace.id"}

	t.Run("groups by value across different field names", func(t *testing.T) {
		e1 := groupTestEntry(1, "gateway-x", "2026-06-25T12:00:01Z", map[string]string{"trace.id": "ABC"})
		e2 := groupTestEntry(2, "billing-y", "2026-06-25T12:00:02Z", map[string]string{"labels.trace.id": "ABC"})

		groups, ungrouped := groupEntries([]*LogEntry{e1, e2}, groupBy)

		if len(groups) != 1 {
			t.Fatalf("got %d groups, want 1", len(groups))
		}
		if groups[0].Key != "ABC" {
			t.Errorf("group key = %q, want %q", groups[0].Key, "ABC")
		}
		if len(groups[0].Entries) != 2 {
			t.Errorf("group has %d entries, want 2", len(groups[0].Entries))
		}
		if len(ungrouped) != 0 {
			t.Errorf("got %d ungrouped, want 0", len(ungrouped))
		}
	})

	t.Run("first configured field wins when several are present", func(t *testing.T) {
		e := groupTestEntry(1, "gateway-x", "2026-06-25T12:00:01Z", map[string]string{
			"trace.id":        "ABC",
			"labels.trace.id": "XYZ",
		})

		groups, _ := groupEntries([]*LogEntry{e}, groupBy)

		if len(groups) != 1 || groups[0].Key != "ABC" {
			t.Fatalf("got groups %+v, want a single group keyed ABC", groups)
		}
	})

	t.Run("entries without any grouping field go to ungrouped", func(t *testing.T) {
		e := groupTestEntry(1, "gateway-x", "2026-06-25T12:00:01Z", map[string]string{"foo": "bar"})

		groups, ungrouped := groupEntries([]*LogEntry{e}, groupBy)

		if len(groups) != 0 {
			t.Errorf("got %d groups, want 0", len(groups))
		}
		if len(ungrouped) != 1 {
			t.Errorf("got %d ungrouped, want 1", len(ungrouped))
		}
	})

	t.Run("orders groups by earliest timestamp and entries within a group by timestamp", func(t *testing.T) {
		a := groupTestEntry(1, "gateway-x", "2026-06-25T12:00:02Z", map[string]string{"trace.id": "t1"})
		b := groupTestEntry(2, "billing-y", "2026-06-25T12:00:03Z", map[string]string{"trace.id": "t2"})
		c := groupTestEntry(3, "auth-z", "2026-06-25T12:00:01Z", map[string]string{"trace.id": "t1"})

		groups, _ := groupEntries([]*LogEntry{a, b, c}, groupBy)

		if len(groups) != 2 {
			t.Fatalf("got %d groups, want 2", len(groups))
		}
		// t1's earliest entry (c @ :01) precedes t2's earliest (b @ :03).
		if groups[0].Key != "t1" || groups[1].Key != "t2" {
			t.Fatalf("group order = [%s, %s], want [t1, t2]", groups[0].Key, groups[1].Key)
		}
		// Within t1, c (:01) precedes a (:02).
		if groups[0].Entries[0].LineNumber != 3 || groups[0].Entries[1].LineNumber != 1 {
			t.Errorf("within-group order = [%d, %d], want [3, 1]",
				groups[0].Entries[0].LineNumber, groups[0].Entries[1].LineNumber)
		}
	})

	t.Run("falls back to arrival order when timestamps are missing", func(t *testing.T) {
		a := groupTestEntry(1, "gateway-x", "", map[string]string{"trace.id": "t1"})
		b := groupTestEntry(2, "billing-y", "", map[string]string{"trace.id": "t1"})

		groups, _ := groupEntries([]*LogEntry{a, b}, groupBy)

		if len(groups) != 1 {
			t.Fatalf("got %d groups, want 1", len(groups))
		}
		if groups[0].Entries[0].LineNumber != 1 || groups[0].Entries[1].LineNumber != 2 {
			t.Errorf("order = [%d, %d], want arrival order [1, 2]",
				groups[0].Entries[0].LineNumber, groups[0].Entries[1].LineNumber)
		}
	})
}

func TestHopPath(t *testing.T) {
	t.Run("lists distinct apps in first-appearance order, deduping recurrences", func(t *testing.T) {
		entries := []*LogEntry{
			{PodID: "api-gateway-7d8f9b6c5-x2k4p"},
			{PodID: "booking-api-5c4b3a2d1-qq9zz"},
			{PodID: "api-gateway-7d8f9b6c5-x2k4p"}, // recurrence of gateway
		}

		got := hopPath(entries)
		want := []string{"api-gateway", "booking-api"}

		if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("hopPath() = %v, want %v", got, want)
		}
	})

	t.Run("skips entries with no pod id", func(t *testing.T) {
		entries := []*LogEntry{{PodID: ""}, {PodID: ""}}
		if got := hopPath(entries); len(got) != 0 {
			t.Errorf("hopPath() = %v, want empty", got)
		}
	})
}

func TestTrimPodHash(t *testing.T) {
	tests := []struct {
		name  string
		podID string
		want  string
	}{
		{"deployment pod", "api-gateway-7d8f9b6c5-x2k4p", "api-gateway"},
		{"statefulset pod", "postgres-0", "postgres"},
		{"no recognizable suffix", "my-service", "my-service"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimPodHash(tt.podID); got != tt.want {
				t.Errorf("trimPodHash(%q) = %q, want %q", tt.podID, got, tt.want)
			}
		})
	}
}

func TestParseEntryTime(t *testing.T) {
	t.Run("parses RFC3339 with nanoseconds", func(t *testing.T) {
		e := &LogEntry{Time: "2026-06-25T12:00:01.123456Z"}
		if _, ok := parseEntryTime(e); !ok {
			t.Errorf("expected RFC3339Nano timestamp to parse")
		}
	})

	t.Run("reports not-ok for empty or unparseable timestamps", func(t *testing.T) {
		if _, ok := parseEntryTime(&LogEntry{Time: ""}); ok {
			t.Errorf("empty timestamp should not parse")
		}
		if _, ok := parseEntryTime(&LogEntry{Time: "not a time"}); ok {
			t.Errorf("garbage timestamp should not parse")
		}
	})
}

func TestFormatGroupHeader(t *testing.T) {
	color.NoColor = true

	entries := []*LogEntry{
		{PodID: "api-gateway-7d8f9b6c5-x2k4p"},
		{PodID: "booking-api-5c4b3a2d1-qq9zz"},
	}

	got := formatGroupHeader("trace.id", "abc123", entries)
	want := "══ trace.id=abc123 · 2 lines · api-gateway → booking-api ══"

	if got != want {
		t.Errorf("formatGroupHeader() =\n  %q\nwant\n  %q", got, want)
	}
}
