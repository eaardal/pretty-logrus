package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

// Elastic Common Schema (ECS) field names
// https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html
const (
	ecsMessageField   = "message"
	ecsLevelField     = "log.level"
	ecsTimestampField = "@timestamp"
)

var messageKeywords = []string{logrus.FieldKeyMsg, ecsMessageField}
var levelKeywords = []string{logrus.FieldKeyLevel, ecsLevelField}
var timeKeywords = []string{logrus.FieldKeyTime, ecsTimestampField}
var errorKeywords = []string{logrus.ErrorKey}
var dataFieldKeywords = []string{"labels"}

type LogEntry struct {
	LineNumber      int
	OriginalLogLine []byte
	Time            string
	Level           string
	Message         string
	Fields          map[string]string
	IsParsed        bool
}

func (l *LogEntry) setFromJsonMap(logMap map[string]interface{}) {
	for key, value := range logMap {
		match := false

		for _, levelKeyword := range levelKeywords {
			if strings.ToLower(key) == levelKeyword {
				l.Level = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, messageKeyword := range messageKeywords {
			if strings.ToLower(key) == messageKeyword {
				l.Message = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, timeKeyword := range timeKeywords {
			if strings.ToLower(key) == timeKeyword {
				l.Time = value.(string)
				match = true
				break
			}
		}

		if match {
			continue
		}

		for _, errorKeyword := range errorKeywords {
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

		for _, dataFieldKeyword := range dataFieldKeywords {
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
