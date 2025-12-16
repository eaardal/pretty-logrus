package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// Elastic Common Schema (ECS) field names
// https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html
const (
	ecsMessageField   = "message"
	ecsLevelField     = "log.level"
	ecsTimestampField = "@timestamp"
)

const HomeEnvVar = "PRETTY_LOGRUS_HOME"

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
	}
}

func hasHomeEnvVar() bool {
	_, ok := os.LookupEnv(HomeEnvVar)
	return ok
}

func hasConfigFile() bool {
	env, ok := os.LookupEnv(HomeEnvVar)
	if !ok {
		return false
	}

	_, err := os.Stat(path.Join(env, "config.json"))
	return !os.IsNotExist(err)
}

func readConfigFile(defaultConfig Config) (*Config, error) {
	homeDir := os.Getenv(HomeEnvVar)
	configFilePath := path.Join(homeDir, "config.json")

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	config := defaultConfig
	if err = json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	if isDebug() {
		configJson, _ := json.MarshalIndent(config, "", "  ")
		fmt.Printf("Read config file: %s\n", configJson)
	}

	return &config, nil
}

func getConfig() *Config {
	defaultConfig := newDefaultConfig()

	if !hasConfigFile() {
		logDebug("No config file found\n")
		return defaultConfig
	}

	config, err := readConfigFile(*defaultConfig)
	if err != nil {
		logDebug("Failed to read config file: %v\n", err)
		return defaultConfig
	}

	if config == nil {
		return defaultConfig
	}

	return config
}
