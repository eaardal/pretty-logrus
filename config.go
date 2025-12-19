package main

import (
	"github.com/sirupsen/logrus"
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
	LevelStyles                     map[string]Style
	FieldStyles                     map[string]KeyValueStyle
	MessageStyles                   map[string]Style
	TimestampStyles                 map[string]Style
	Keywords                        *KeywordConfig
	ExcludeFields                   []string
	ExcludedFieldsWarningText       string
	ExcludedFieldsWarningTextStyles map[string]Style
}

func newDefaultConfig() *Config {
	return &Config{
		FieldStyles:     DefaultFieldStyles,
		LevelStyles:     DefaultLevelStyles,
		MessageStyles:   DefaultMessageStyles,
		TimestampStyles: DefaultTimestampStyles,
		Keywords: &KeywordConfig{
			MessageKeywords:   []string{logrus.FieldKeyMsg, ecsMessageField},
			LevelKeywords:     []string{logrus.FieldKeyLevel, ecsLevelField},
			TimestampKeywords: []string{logrus.FieldKeyTime, ecsTimestampField},
			ErrorKeywords:     []string{logrus.ErrorKey},
			FieldKeywords:     []string{"labels"},
		},
		ExcludeFields:                   []string{},
		ExcludedFieldsWarningText:       "[Some fields excluded]",
		ExcludedFieldsWarningTextStyles: DefaultExcludedWarningTextStyles,
	}
}

func getConfig() *Config {
	defaultConfig := newDefaultConfig()

	if err := ensureConfigFileExistsIfHomeEnvIsSet(defaultConfig); err != nil {
		logDebug("Error ensuring config file exists: ", err)
		return defaultConfig
	}

	if !hasConfigFile() {
		logDebug("No config file found\n")
		return defaultConfig
	}

	configFile, err := readConfigFile(*defaultConfig)
	if err != nil {
		logDebug("Failed to read config file: %v\n", err)
		return defaultConfig
	}

	if configFile == nil {
		return defaultConfig
	}

	logDebug("Loaded config file: %+v\n", configFile)

	return configFile
}
