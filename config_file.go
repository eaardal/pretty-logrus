package main

import (
	"encoding/json"
	"fmt"
	"os"
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

type Config struct {
	LevelStyles     map[string]Style
	FieldStyles     map[string]KeyValueStyle
	MessageStyles   map[string]Style
	TimestampStyles map[string]Style
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
	defaultConfig := &Config{
		FieldStyles: make(map[string]KeyValueStyle),
	}

	if !hasConfigFile() {
		logDebug("No config file found\n")
		return defaultConfig
	}

	configFilePath := os.Getenv("PRETTY_LOGRUS_HOME") + "/config.json"

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		logDebug("Failed to read config file: %v\n", err)
		return defaultConfig
	}

	var config Config
	if err = json.Unmarshal(content, &config); err != nil {
		logDebug("Failed to unmarshal config file: %v\n", err)
		return defaultConfig
	}

	if isDebug() {
		configJson, _ := json.MarshalIndent(config, "", "  ")
		fmt.Printf("Read config file: %s\n", configJson)
	}

	return &config
}
