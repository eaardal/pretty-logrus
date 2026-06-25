package main

import (
	"fmt"
	"strings"
)

type LogEntry struct {
	LineNumber      int
	OriginalLogLine []byte
	PodID           string
	Time            string
	Level           string
	Message         string
	Fields          map[string]string
	IsParsed        bool
}

func (l *LogEntry) setFromJsonMap(logMap map[string]interface{}, keywords KeywordConfig) {
	// Flatten the whole document first so nested objects become dotted field
	// names (e.g. {"log":{"origin":{"file":{"name":...}}}} -> log.origin.file.name).
	// Keyword matching then runs against the flattened names, so a keyword like
	// the ECS "log.level" is recognised whether it arrives nested or pre-dotted.
	flat := make(map[string]string, len(logMap))
	flattenJSON(flat, "", logMap)

	for key, value := range flat {
		lowerKey := strings.ToLower(key)

		if matchesAnyKeyword(lowerKey, keywords.LevelKeywords) {
			l.Level = value
			continue
		}

		if matchesAnyKeyword(lowerKey, keywords.MessageKeywords) {
			l.Message = value
			continue
		}

		if matchesAnyKeyword(lowerKey, keywords.TimestampKeywords) {
			l.Time = value
			continue
		}

		l.Fields[key] = value
	}

	l.IsParsed = true
}

// flattenJSON recursively flattens a decoded JSON object into dst, joining
// nested object keys with dots. Non-object values (scalars, arrays, null) are
// stored as their default string form, which preserves how arrays and scalars
// were rendered before recursive flattening was introduced.
func flattenJSON(dst map[string]string, prefix string, m map[string]interface{}) {
	for key, value := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if nested, ok := value.(map[string]interface{}); ok {
			flattenJSON(dst, fullKey, nested)
			continue
		}

		dst[fullKey] = fmt.Sprintf("%v", value)
	}
}

// matchesAnyKeyword reports whether the (already lower-cased) field name equals
// one of the configured keywords.
func matchesAnyKeyword(lowerKey string, keywords []string) bool {
	for _, keyword := range keywords {
		if lowerKey == keyword {
			return true
		}
	}
	return false
}

func (l *LogEntry) setOriginalLogLine(line []byte) {
	copy(l.OriginalLogLine, line)
	l.IsParsed = false
}
