package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

// traceGroup is a set of log entries that share the same grouping value (e.g.
// the same trace id), in display order.
type traceGroup struct {
	Key     string
	Entries []*LogEntry
}

// groupHeaderRule is the bracketing rule drawn on either side of a group header.
const groupHeaderRule = "══"

// timeLayouts are the timestamp formats tried, in order, when ordering entries
// within and across groups. Unparseable timestamps fall back to arrival order.
var timeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
}

// podHashSuffix matches the ReplicaSet-hash + pod-suffix that Kubernetes appends
// to a Deployment's pods, e.g. "-7d8f9b6c5-x2k4p". podOrdinalSuffix matches the
// "-0" style ordinal appended to a StatefulSet's pods.
var (
	podHashSuffix    = regexp.MustCompile(`-[a-z0-9]{6,10}-[a-z0-9]{5}$`)
	podOrdinalSuffix = regexp.MustCompile(`-[0-9]+$`)
)

// groupKeyFor returns the grouping value for an entry: the value of the first
// configured field that is present and non-empty. Grouping is by value, so the
// same id carried under different field names across apps (e.g. trace.id vs
// labels.trace.id) collapses into a single group.
func groupKeyFor(entry *LogEntry, fields []string) (string, bool) {
	for _, field := range fields {
		if value, ok := entry.Fields[field]; ok && value != "" {
			return value, true
		}
	}
	return "", false
}

// groupEntries buckets entries by their grouping value. Entries within a group
// are ordered by timestamp, and groups are ordered by their earliest entry.
// Entries carrying none of the configured fields are returned separately so the
// caller can render them in a trailing "ungrouped" section.
func groupEntries(entries []*LogEntry, fields []string) (groups []*traceGroup, ungrouped []*LogEntry) {
	index := make(map[string]*traceGroup)

	for _, entry := range entries {
		key, ok := groupKeyFor(entry, fields)
		if !ok {
			ungrouped = append(ungrouped, entry)
			continue
		}

		group, exists := index[key]
		if !exists {
			group = &traceGroup{Key: key}
			index[key] = group
			groups = append(groups, group)
		}
		group.Entries = append(group.Entries, entry)
	}

	for _, group := range groups {
		sortEntriesByTime(group.Entries)
	}
	sortEntriesByTime(ungrouped)

	sort.SliceStable(groups, func(i, j int) bool {
		return lessByTime(groups[i].Entries[0], groups[j].Entries[0])
	})

	return groups, ungrouped
}

// sortEntriesByTime stably orders entries by timestamp, keeping arrival order
// for entries whose timestamps are equal or unparseable.
func sortEntriesByTime(entries []*LogEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return lessByTime(entries[i], entries[j])
	})
}

// lessByTime orders two entries by parsed timestamp, with the original line
// number as a stable tiebreaker. Entries with a parseable timestamp sort before
// those without so unparseable lines drift to the end of their group rather than
// reordering the timed ones.
func lessByTime(a, b *LogEntry) bool {
	ta, okA := parseEntryTime(a)
	tb, okB := parseEntryTime(b)

	if okA && okB {
		if !ta.Equal(tb) {
			return ta.Before(tb)
		}
		return a.LineNumber < b.LineNumber
	}

	if okA != okB {
		return okA
	}

	return a.LineNumber < b.LineNumber
}

// parseEntryTime parses an entry's timestamp against the known layouts.
func parseEntryTime(entry *LogEntry) (time.Time, bool) {
	if entry.Time == "" {
		return time.Time{}, false
	}

	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, entry.Time); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

// hopPath returns the distinct app names the entries pass through, in
// first-appearance order, so a header can show how a call travelled across
// services. Pods with no identity (unprefixed input) are skipped.
func hopPath(entries []*LogEntry) []string {
	var path []string
	seen := make(map[string]struct{})

	for _, entry := range entries {
		if entry.PodID == "" {
			continue
		}

		app := trimPodHash(entry.PodID)
		if _, dup := seen[app]; dup {
			continue
		}

		seen[app] = struct{}{}
		path = append(path, app)
	}

	return path
}

// trimPodHash reduces a pod name to its owning workload name by stripping the
// Kubernetes-generated suffix (ReplicaSet hash + pod suffix for Deployments, or
// the ordinal for StatefulSets). Names that match neither shape are returned
// unchanged so an unusual naming scheme is shown verbatim rather than mangled.
func trimPodHash(podID string) string {
	if loc := podHashSuffix.FindStringIndex(podID); loc != nil {
		return podID[:loc[0]]
	}
	if loc := podOrdinalSuffix.FindStringIndex(podID); loc != nil {
		return podID[:loc[0]]
	}
	return podID
}

// renderGroups prints each trace group, separated by a blank line, followed by
// any ungrouped entries. Individual entries reuse the normal single/multi-line
// rendering so all styling, filtering and flags keep working inside a group.
func renderGroups(args Args, config Config, groups []*traceGroup, ungrouped []*LogEntry, colorizer *PodColorizer) {
	label := groupLabel(args.GroupBy)
	printedAny := false

	for _, group := range groups {
		if printedAny {
			fmt.Println()
		}
		printedAny = true

		fmt.Println(formatGroupHeader(label, group.Key, group.Entries))
		for _, entry := range group.Entries {
			printEntry(args, config, entry, colorizer)
		}
	}

	if len(ungrouped) > 0 {
		if printedAny {
			fmt.Println()
		}

		fmt.Println(formatUngroupedHeader(len(ungrouped)))
		for _, entry := range ungrouped {
			printEntry(args, config, entry, colorizer)
		}
	}
}

// formatGroupHeader builds the header line for a trace group, e.g.
// "══ trace.id=abc123 · 5 lines · gateway → billing → auth ══". The app
// hop-path is omitted when no entry carries a pod identity.
func formatGroupHeader(label, key string, entries []*LogEntry) string {
	parts := []string{
		fmt.Sprintf("%s=%s", label, key),
		lineCount(len(entries)),
	}

	if path := hopPath(entries); len(path) > 0 {
		parts = append(parts, strings.Join(path, " → "))
	}

	return groupHeaderStyle().Sprintf("%s %s %s", groupHeaderRule, strings.Join(parts, " · "), groupHeaderRule)
}

// formatUngroupedHeader builds the header for the trailing section of entries
// that carried none of the configured grouping fields.
func formatUngroupedHeader(count int) string {
	return groupHeaderStyle().Sprintf("%s ungrouped · %s %s", groupHeaderRule, lineCount(count), groupHeaderRule)
}

func groupHeaderStyle() *color.Color {
	return color.New(color.FgHiWhite, color.Bold)
}

// groupLabel is the field name shown in group headers. The first configured
// field is used as the canonical label so the header stays consistent even when
// entries matched via a different fallback field.
func groupLabel(fields []string) string {
	if len(fields) > 0 {
		return fields[0]
	}
	return "group"
}

func lineCount(count int) string {
	if count == 1 {
		return "1 line"
	}
	return fmt.Sprintf("%d lines", count)
}
