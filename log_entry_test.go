package main

import "testing"

func newTestEntry() *LogEntry {
	return &LogEntry{Fields: make(map[string]string)}
}

func TestSetFromJsonMap_Flattening(t *testing.T) {
	keywords := *newDefaultConfig().Keywords

	t.Run("extracts level, message and time from top-level keywords", func(t *testing.T) {
		entry := newTestEntry()

		entry.setFromJsonMap(map[string]interface{}{
			"level": "info",
			"msg":   "hello",
			"time":  "2026-06-25T12:00:00Z",
		}, keywords)

		if entry.Level != "info" {
			t.Errorf("Level = %q, want %q", entry.Level, "info")
		}
		if entry.Message != "hello" {
			t.Errorf("Message = %q, want %q", entry.Message, "hello")
		}
		if entry.Time != "2026-06-25T12:00:00Z" {
			t.Errorf("Time = %q, want %q", entry.Time, "2026-06-25T12:00:00Z")
		}
		if _, ok := entry.Fields["level"]; ok {
			t.Errorf("level should be extracted, not left in Fields")
		}
	})

	t.Run("flattens deeply nested objects into dotted field names", func(t *testing.T) {
		entry := newTestEntry()

		entry.setFromJsonMap(map[string]interface{}{
			"log": map[string]interface{}{
				"origin": map[string]interface{}{
					"file": map[string]interface{}{
						"name": "main.go",
						"line": float64(42),
					},
				},
			},
		}, keywords)

		if got := entry.Fields["log.origin.file.name"]; got != "main.go" {
			t.Errorf("Fields[log.origin.file.name] = %q, want %q", got, "main.go")
		}
		if got := entry.Fields["log.origin.file.line"]; got != "42" {
			t.Errorf("Fields[log.origin.file.line] = %q, want %q", got, "42")
		}
	})

	t.Run("keeps a one-level field whose inner key already contains a dot", func(t *testing.T) {
		entry := newTestEntry()

		entry.setFromJsonMap(map[string]interface{}{
			"labels": map[string]interface{}{
				"trace.id": "ABC",
			},
		}, keywords)

		if got := entry.Fields["labels.trace.id"]; got != "ABC" {
			t.Errorf("Fields[labels.trace.id] = %q, want %q", got, "ABC")
		}
	})

	t.Run("matches a nested keyword key such as ECS log.level", func(t *testing.T) {
		entry := newTestEntry()

		entry.setFromJsonMap(map[string]interface{}{
			"log": map[string]interface{}{
				"level": "warning",
				"origin": map[string]interface{}{
					"file": map[string]interface{}{"name": "main.go"},
				},
			},
		}, keywords)

		if entry.Level != "warning" {
			t.Errorf("Level = %q, want %q", entry.Level, "warning")
		}
		if _, ok := entry.Fields["log.level"]; ok {
			t.Errorf("log.level should be extracted as Level, not left in Fields")
		}
		if got := entry.Fields["log.origin.file.name"]; got != "main.go" {
			t.Errorf("Fields[log.origin.file.name] = %q, want %q", got, "main.go")
		}
	})

	t.Run("flattens non-string scalars using their value", func(t *testing.T) {
		entry := newTestEntry()

		entry.setFromJsonMap(map[string]interface{}{
			"data": map[string]interface{}{
				"count": float64(5),
				"ok":    true,
			},
		}, keywords)

		if got := entry.Fields["data.count"]; got != "5" {
			t.Errorf("Fields[data.count] = %q, want %q", got, "5")
		}
		if got := entry.Fields["data.ok"]; got != "true" {
			t.Errorf("Fields[data.ok] = %q, want %q", got, "true")
		}
	})

	t.Run("flattens an error object with non-string members without panicking", func(t *testing.T) {
		entry := newTestEntry()

		entry.setFromJsonMap(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "boom",
				"code":    float64(42),
			},
		}, keywords)

		if got := entry.Fields["error.message"]; got != "boom" {
			t.Errorf("Fields[error.message] = %q, want %q", got, "boom")
		}
		if got := entry.Fields["error.code"]; got != "42" {
			t.Errorf("Fields[error.code] = %q, want %q", got, "42")
		}
	})

	t.Run("keeps arrays as a single stringified leaf field", func(t *testing.T) {
		entry := newTestEntry()

		entry.setFromJsonMap(map[string]interface{}{
			"tags": []interface{}{"a", "b"},
		}, keywords)

		if got := entry.Fields["tags"]; got != "[a b]" {
			t.Errorf("Fields[tags] = %q, want %q", got, "[a b]")
		}
	})

	t.Run("keeps top-level scalar fields", func(t *testing.T) {
		entry := newTestEntry()

		entry.setFromJsonMap(map[string]interface{}{
			"foo": "bar",
		}, keywords)

		if got := entry.Fields["foo"]; got != "bar" {
			t.Errorf("Fields[foo] = %q, want %q", got, "bar")
		}
		if !entry.IsParsed {
			t.Errorf("IsParsed = false, want true")
		}
	})
}
