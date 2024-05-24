package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

// Elastic Common Schema (ECS) field names
// https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html
const (
	ecsMessageField   = "message"
	ecsLevelField     = "log.level"
	ecsTimestampField = "@timestamp"
)

type Style struct {
	BgColor   *string
	FgColor   *string
	Bold      *bool
	Underline *bool
	Italic    *bool
}

type KeyValueStyle struct {
	Key   *Style
	Value *Style
}

type KeywordConfig struct {
	MessageKeywords   []string
	LevelKeywords     []string
	TimestampKeywords []string
	ErrorKeywords     []string
	FieldKeywords     []string
}

type Config struct {
	LevelStyles     map[string]Style
	FieldStyles     map[string]KeyValueStyle
	MessageStyles   map[string]Style
	TimestampStyles map[string]Style
	Keywords        *KeywordConfig
}

func hasConfigFile() bool {
	env, ok := os.LookupEnv("PRETTY_LOGRUS_HOME")
	if !ok {
		return false
	}

	_, err := os.Stat(env + "/config.json")
	return !os.IsNotExist(err)
}

func readConfigFile() *Config {
	config := &Config{
		FieldStyles:     make(map[string]KeyValueStyle),
		LevelStyles:     make(map[string]Style),
		MessageStyles:   make(map[string]Style),
		TimestampStyles: make(map[string]Style),
		Keywords: &KeywordConfig{
			MessageKeywords:   []string{logrus.FieldKeyMsg, ecsMessageField},
			LevelKeywords:     []string{logrus.FieldKeyLevel, ecsLevelField},
			TimestampKeywords: []string{logrus.FieldKeyTime, ecsTimestampField},
			ErrorKeywords:     []string{logrus.ErrorKey},
			FieldKeywords:     []string{"labels"},
		},
	}

	if !hasConfigFile() {
		logDebug("No config file found\n")
		return config
	}

	configFilePath := os.Getenv("PRETTY_LOGRUS_HOME") + "/config.json"

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		logDebug("Failed to read config file: %v\n", err)
		return config
	}

	if err = json.Unmarshal(content, &config); err != nil {
		logDebug("Failed to unmarshal config file: %v\n", err)
		return config
	}

	if isDebug() {
		configJson, _ := json.MarshalIndent(config, "", "  ")
		fmt.Printf("Read config file: %s\n", configJson)
	}

	return config
}
