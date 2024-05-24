package main

import (
	"fmt"
	"strings"
)

type LogEntry struct {
	LineNumber      int
	OriginalLogLine []byte
	Time            string
	Level           string
	Message         string
	Fields          map[string]string
	IsParsed        bool
}

func (l *LogEntry) setFromJsonMap(logMap map[string]interface{}, keywords KeywordConfig) {
	for key, value := range logMap {
		match := false

		for _, levelKeyword := range keywords.LevelKeywords {
			if strings.ToLower(key) == levelKeyword {
				l.Level = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, messageKeyword := range keywords.MessageKeywords {
			if strings.ToLower(key) == messageKeyword {
				l.Message = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, timeKeyword := range keywords.TimestampKeywords {
			if strings.ToLower(key) == timeKeyword {
				l.Time = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, errorKeyword := range keywords.ErrorKeywords {
			if strings.ToLower(key) == errorKeyword {
				switch val := value.(type) {
				case string:
					l.Fields[key] = val
				case map[string]interface{}:
					for errKey, errValue := range val {
						l.Fields[key+"."+errKey] = errValue.(string)
					}
				}
				match = true
				break
			}
		}

		for _, dataFieldKeyword := range keywords.FieldKeywords {
			if strings.ToLower(key) == dataFieldKeyword {
				switch val := value.(type) {
				case string:
					l.Fields[key] = val
				case map[string]interface{}:
					for dataFieldKey, dataFieldValue := range val {
						l.Fields[key+"."+dataFieldKey] = fmt.Sprintf("%v", dataFieldValue)
					}
				}
				match = true
				break
			}
		}

		if !match {
			l.Fields[key] = fmt.Sprintf("%v", value)
		}
	}

	l.IsParsed = true
}

func (l *LogEntry) setOriginalLogLine(line []byte) {
	copy(l.OriginalLogLine, line)
	l.IsParsed = false
}
